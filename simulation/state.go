package simulation

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/geotry/rass/encoding"
	"github.com/geotry/rass/rendering"
	"github.com/geotry/rass/scene"
)

// Represents the state of the simulation as a block-database.
type State struct {
	buffer *encoding.BlockBuffer

	// Index of blocks by type
	textures             map[int]*encoding.Block
	cameras              map[*scene.Camera]*encoding.Block
	sceneObjects         map[int32]*encoding.Block
	sceneObjectInstances map[uint32]*encoding.Block

	mu sync.RWMutex
}

// List of blocks
type BlockType uint8

const (
	TextureBlock BlockType = iota
	CameraBlock
	SceneObjectBlock
	SceneObjectInstanceBlock
)

const (
	_ = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
)

const BufferSize = 1 * MiB

func NewState() *State {
	return &State{
		buffer:               encoding.NewBlockBuffer(BufferSize),
		textures:             make(map[int]*encoding.Block),
		cameras:              make(map[*scene.Camera]*encoding.Block),
		sceneObjects:         make(map[int32]*encoding.Block),
		sceneObjectInstances: make(map[uint32]*encoding.Block),
	}
}

// func (s *State) LockWrite() {
// 	s.mu.Lock()
// }

// func (s *State) UnlockWrite() {
// 	s.mu.Unlock()
// }

func (s *State) WriteTextureGroup(group *rendering.TextureGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var buf encoding.WritableBlock

	if s.textures[group.Id] != nil {
		buf = s.textures[group.Id]
	} else {
		buf = s.buffer
	}

	depth := len(group.Textures)

	buf.NewBlock(uint8(TextureBlock))
	buf.PutUint8(uint8(group.Id))
	buf.PutUint16(uint16(group.Width))
	buf.PutUint16(uint16(group.Height))
	buf.PutUint8(uint8(depth))
	buf.PutUint8(uint8(group.Model))

	buf.NewArray()
	for _, texture := range group.Textures {
		textureWidth := texture.Width
		textureHeight := len(texture.Pixels) / group.PixelSize / textureWidth
		// Write pixels of the texture in the group canvas
		for i := 0; i < group.Width*group.Height*group.PixelSize; i += group.PixelSize {
			width := group.Width * group.PixelSize
			x := i % width
			y := int(math.Floor(float64(i) / float64(width)))
			if x >= textureWidth || y >= textureHeight {
				switch group.Model {
				case rendering.ALPHA:
					buf.PutUint8(255)
				case rendering.RGB:
					buf.PutUint8(0)
					buf.PutUint8(0)
					buf.PutUint8(0)
				case rendering.RGBA:
					buf.PutUint8(0)
					buf.PutUint8(0)
					buf.PutUint8(0)
					buf.PutUint8(0)
				}
			} else {
				for p := range group.PixelSize {
					buf.PutUint8(texture.Pixels[i+p-(width-textureWidth*group.PixelSize)*y])
				}
			}
		}
	}
	buf.EndArray()

	if s.textures[group.Id] == nil {
		s.textures[group.Id] = buf.EndBlock()
	}
}

func (s *State) WriteTextureGroupOnce(group *rendering.TextureGroup) {
	s.mu.RLock()
	missing := s.textures[group.Id] == nil
	s.mu.RUnlock()
	if missing {
		s.WriteTextureGroup(group)
	}
}

func (s *State) WriteTextureOnce(texture *rendering.Texture) {
	s.mu.RLock()
	missing := s.textures[texture.Group.Id] == nil
	s.mu.RUnlock()
	if missing {
		s.WriteTextureGroup(texture.Group)
	}
}

func (s *State) WriteCamera(camera *scene.Camera) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var buf encoding.WritableBlock
	if s.cameras[camera] != nil {
		buf = s.cameras[camera]
	} else {
		buf = s.buffer
	}

	perspective, ortho := camera.ViewMatrix()

	buf.NewBlock(uint8(CameraBlock))
	buf.PutUint8(2)
	buf.PutMatrix(ortho)
	buf.PutMatrix(perspective)

	if s.cameras[camera] == nil {
		s.cameras[camera] = buf.EndBlock()
	}
}

func (s *State) ReadCamera(camera *scene.Camera) *encoding.Block {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cameras[camera]
}

func (s *State) DeleteCamera(camera *scene.Camera) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cameras[camera] != nil {
		// b := s.cameras[camera]
		// b.Free()
		delete(s.cameras, camera)
	}
}

func (s *State) WriteSceneObjectOnce(obj *scene.SceneObject) {
	if s.sceneObjects[obj.Id] == nil {
		s.WriteSceneObject(obj)
	}
}

func (s *State) WriteSceneObject(obj *scene.SceneObject) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var buf encoding.WritableBlock

	if s.sceneObjects[obj.Id] != nil {
		buf = s.sceneObjects[obj.Id]
	} else {
		buf = s.buffer
	}

	// Encode vertex data (geometry, uv mapping, normals...) in a single block
	// It should be sent upfront like textures, then use ObjectID to reference it in other blocks
	buf.NewBlock(uint8(SceneObjectBlock))
	// Encode Object ID so webgl can index buffer position with ID and update partial buffer
	// without rewriting the whole buffer from zero
	// ObjectID is the SceneObject.Id, not SceneObjectInstance.Id
	buf.PutUint32(uint32(obj.Id))

	buf.PutUint8(uint8(obj.Texture.Id))    // Texture ID
	buf.PutUint8(uint8(obj.Texture.Index)) // Texture index

	buf.PutBool(obj.UIElement)

	// Geometry: 3*4 bytes per vertex + 2 byte for number of vertices
	// For now, use a simple 1x1 quad but can support more advanced shapes
	buf.NewArray()
	for _, p := range obj.Geometry {
		buf.PutVector3Float32(float32(p.X), float32(p.Y), float32(p.Z))
	}
	buf.EndArray()

	// UV Mapping
	// Use texture size / dimension to stretch correctly
	// Todo: translate coordinates when using texture atlases
	rx := float32(float64(obj.Texture.Width) / float64(obj.Texture.Group.Width))
	ry := float32(float64(obj.Texture.Height) / float64(obj.Texture.Group.Height))
	buf.NewArray()
	for _, p := range obj.UV {
		buf.PutVector2Float32(float32(p.X)*rx, float32(p.Y)*ry)
	}
	buf.EndArray()

	// TODO: Normals

	if s.sceneObjects[obj.Id] == nil {
		s.sceneObjects[obj.Id] = buf.EndBlock()
	}
}

func (s *State) ReadSceneObject(obj *scene.SceneObject) *encoding.Block {
	return s.sceneObjects[obj.Id]
}

func (s *State) WriteSceneObjectInstance(obj *scene.SceneObjectInstance) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var buf encoding.WritableBlock

	if s.sceneObjectInstances[obj.Id] != nil {
		buf = s.sceneObjectInstances[obj.Id]
	} else {
		buf = s.buffer
	}

	buf.NewBlock(uint8(SceneObjectInstanceBlock))
	buf.PutUint16(uint16(obj.Id))
	buf.PutUint32(uint32(obj.SceneObject.Id))
	buf.PutMatrix(obj.ModelMatrix())

	if s.sceneObjectInstances[obj.Id] == nil {
		s.sceneObjectInstances[obj.Id] = buf.EndBlock()
	}
}

func (s *State) ReadSceneObjectInstance(obj *scene.SceneObjectInstance) *encoding.Block {
	return s.sceneObjectInstances[obj.Id]
}

func (s *State) ReadAllSceneObjects() []*encoding.Block {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blocks := make([]*encoding.Block, len(s.sceneObjects))
	for i, b := range s.sceneObjects {
		blocks[i] = b
	}
	return blocks
}

func (s *State) CopyTextures(buf []byte) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	offset := 0
	for _, b := range s.textures {
		offset += b.Copy(buf[offset:])
	}
	return offset
}

func (s *State) CopySceneObjects(buf []byte) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	offset := 0
	for _, b := range s.sceneObjects {
		offset += b.Copy(buf[offset:])
	}
	return offset
}

func (s *State) CopySceneObjectInstances(buf []byte) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	offset := 0
	for _, b := range s.sceneObjectInstances {
		offset += b.Copy(buf[offset:])
	}
	return offset
}

func (s *State) CopySceneObjectInstance(buf []byte, obj *scene.SceneObjectInstance) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	offset := 0
	if s.sceneObjectInstances[obj.Id] != nil {
		offset += s.sceneObjectInstances[obj.Id].Copy(buf[offset:])
	}
	return offset
}

func (s *State) CopyCamera(buf []byte, camera *scene.Camera) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	offset := 0
	if s.cameras[camera] != nil {
		offset += s.cameras[camera].Copy(buf[offset:])
	}
	return offset
}

func (s *State) GetTextureRGBA(id int) (*image.RGBA, error) {
	if s.textures[id] == nil {
		return nil, fmt.Errorf("texture with id %v not found", id)
	}

	b := s.textures[id]
	b.Reset()

	_, width, height, depth, model := b.Uint8(), b.Uint16(), b.Uint16(), b.Uint8(), b.Uint8()

	if model != uint8(rendering.RGBA) {
		return nil, fmt.Errorf("texture %v is not in rgba format", id)
	}

	pixels := make([]uint8, b.Uint32())
	for i := range len(pixels) {
		pixels[i] = b.Uint8()
	}

	img := image.NewRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(int(width), int(height)*int(depth))})
	copy(img.Pix, pixels)

	return img, nil
}

func (s *State) GetTexturePaletted(id int, rgba *image.RGBA) (*image.Paletted, error) {
	if s.textures[id] == nil {
		return nil, fmt.Errorf("texture with id %v not found", id)
	}

	b := s.textures[id]
	b.Reset()

	_, width, height, depth, model := b.Uint8(), b.Uint16(), b.Uint16(), b.Uint8(), b.Uint8()

	if model != uint8(rendering.ALPHA) {
		return nil, fmt.Errorf("texture %v is not in alpha format", id)
	}

	pixels := make([]uint8, b.Uint32())
	for i := range len(pixels) {
		pixels[i] = b.Uint8()
	}

	palette := make(color.Palette, len(rgba.Pix)/4)

	for i := range len(rgba.Pix) / 4 {
		r, g, b, a := rgba.Pix[(i*4)], rgba.Pix[(i*4)+1], rgba.Pix[(i*4)+2], rgba.Pix[(i*4)+3]
		palette[i] = color.RGBA{R: r, G: g, B: b, A: a}
	}

	img := image.NewPaletted(
		image.Rectangle{image.Pt(0, 0), image.Pt(int(width), int(height)*int(depth))},
		palette,
	)
	copy(img.Pix, pixels)

	return img, nil
}
