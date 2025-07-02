package simulation

import (
	"github.com/geotry/rass/encoding"
	"github.com/geotry/rass/scene"
	"github.com/geotry/rass/shaders"
)

// Note: It might be better to let the client chose the size of each batch
// Server can still send different passes (to set different uniform, view matrices...)
// But client could split each render pass in smaller batches according to specs.
// The client would need to re-map the buffer indices with object IDs
const MAX_OBJECTS_PER_RENDER_PASS = 32

// Buffer
// Note:
// UpdateBuffer() can be ignored if ReadBuffer() has not been called
// This should avoid having to copy buffer from simulation to session buffer and then session buffer to server
// Server could read directly the session buffer if it's not written at the same time by simulation
// Simulation: keeps a copy of the buffer
// Session: keeps a copy of its own buffer (optionnaly copying simulation buffer?)
// Server: reads Session buffer

// Simulation must not block waiting for session to read buffer
//

// func (s *Session) ReadBuffer() []byte {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	s.readCount++
// 	s.paletteSent = true
// 	s.textureSent = true
// 	// s.cameraSent = true
// 	for i := range s.objectsSent {
// 		s.objectsSent[i] = true
// 	}
// 	// for i := range s.instancesSent {
// 	// 	s.instancesSent[i] = true
// 	// }
// 	buf := make([]byte, s.buffer.Offset())
// 	s.buffer.Copy(buf)
// 	return buf
// }

// func (session *Session) updateBuffer() {
// 	session.mu.Lock()
// 	defer session.mu.Unlock()

// 	session.buffer.Reset()

// 	camera := session.Camera
// 	objects := camera.Scene.Scan(camera)

// 	// Sort objects by UI element to split them, and then by ID to have consistent order
// 	slices.SortFunc(objects, func(a, b *scene.SceneObjectInstance) int {
// 		if a.SceneObject.UIElement && !b.SceneObject.UIElement {
// 			return -1
// 		}
// 		if !a.SceneObject.UIElement && b.SceneObject.UIElement {
// 			return 1
// 		}
// 		if a.Id < b.Id {
// 			return -1
// 		}
// 		return 1
// 	})

// 	// Create render passes
// 	renderPasses := make([][]*scene.SceneObjectInstance, 1)
// 	i := 0
// 	for j, o := range objects {
// 		if len(renderPasses[i]) >= MAX_OBJECTS_PER_RENDER_PASS || j > 0 && objects[j-1].SceneObject.UIElement && !o.SceneObject.UIElement {
// 			renderPasses = append(renderPasses, make([]*scene.SceneObjectInstance, 0))
// 			i++
// 		}
// 		renderPasses[i] = append(renderPasses[i], o)
// 	}

// 	// First render call, send initial data
// 	if session.readCount == 0 {
// 		// sendVertexShader(session.buffer, shaders.NewVertexShader())
// 		// sendTexturePalette(buf, 1, palette)
// 		// sendTexture(buf, 0, ...)
// 		encodeVBO(session.buffer)
// 		for _, o := range objects {
// 			encodeVBOVertices(session.buffer, o.SceneObject)
// 		}
// 	}

// 	// Encode view matrices
// 	// TODO: Skip if matrix did not change since last session render
// 	v0, v1 := session.Camera.ViewMatrix()
// 	encodeViewMatrix(session.buffer, 0, v0)
// 	encodeViewMatrix(session.buffer, 1, v1)
// 	// if session.sim.cameras[camera] != nil {
// 	// 	session.buffer.AppendBlock(session.sim.cameras[camera])
// 	// }

// 	// Encode model matrices
// 	for _, o := range objects {
// 		// TODO: Skip if matrix has not changed
// 		encodeModelMatrix(session.buffer, o)
// 	}

// 	// Render pass sets uniform and attributes using references (ID)
// 	for _, objs := range renderPasses {
// 		encodeRenderPass(session.buffer, objs)
// 	}
// }

// func encodeViewMatrix(buf encoding.WritableBlock, index int, matrix compute.Matrix) {
// 	buf.NewBlock(0)
// 	buf.PutUint8(uint8(index))
// 	buf.PutMatrix(matrix)
// 	buf.EndBlock()
// }

// func encodeModelMatrix(buf encoding.WritableBlock, o *scene.SceneObjectInstance) {
// 	buf.NewBlock(0)
// 	buf.PutUint32(o.Id)
// 	buf.PutMatrix(o.ModelMatrix())
// 	buf.EndBlock()
// }

// func encodeRenderPass(buf encoding.WritableBlock, objects []*scene.SceneObjectInstance) {
// 	buf.NewBlock(0)

// 	buf.PutUint8(uint8(len(objects)))       // Number of objects
// 	buf.PutUint16(uint16(len(objects) * 6)) // Number of vertices

// 	// Model uniform
// 	buf.PutUint8(shaders.Model.Index)
// 	buf.PutUint8(uint8(len(objects)))
// 	for _, o := range objects {
// 		buf.PutUint32(o.Id)
// 	}

// 	// View uniform
// 	buf.PutUint8(shaders.View.Index)
// 	if objects[0].SceneObject.UIElement {
// 		buf.PutUint8(0)
// 	} else {
// 		buf.PutUint8(1)
// 	}

// 	buf.EndBlock()
// }

// func sendTexturePalette(buf encoding.WritableBlock, index uint8, palette []uint8) *encoding.Block {
// 	buf.NewBlock(0)
// 	buf.PutUint8(index) // gl.TEXTURE1
// 	buf.PutUint16(uint16(len(palette)))
// 	buf.PutUint16(1)
// 	buf.PutUint8(1) // length for 2D texture array
// 	buf.NewArray(len(palette))
// 	for _, p := range palette {
// 		buf.PutUint8(p)
// 	}
// 	return buf.EndBlock()
// }

// func sendTexture(buf encoding.WritableBlock, index uint8, texSize compute.Point, textures []*rendering.Texture) *encoding.Block {
// 	buf.NewBlock(0)
// 	texMaxWidth, texMaxHeight := int(texSize.X), int(texSize.Y)
// 	buf.PutUint8(index) // gl.TEXTURE0
// 	buf.PutUint16(uint16(texMaxWidth))
// 	buf.PutUint16(uint16(texMaxHeight))
// 	buf.PutUint8(uint8(len(textures)))
// 	buf.NewArray(texMaxWidth * texMaxHeight * len(textures))
// 	for _, tex := range textures {
// 		texWidth, texHeight := int(tex.Size.X), int(tex.Size.Y)
// 		for i := range texMaxWidth * texMaxHeight {
// 			x := i % texMaxWidth
// 			y := int(math.Floor(float64(i) / float64(texMaxWidth)))
// 			if x >= texWidth || y >= texHeight {
// 				buf.PutUint8(255)
// 			} else {
// 				// texIndex := (y * texWidth) + x
// 				texIndex := i - (texMaxWidth-texWidth)*y
// 				buf.PutUint8(tex.Pixels[texIndex])
// 			}
// 		}
// 	}
// 	return buf.EndBlock()
// }

// Send a vertex buffer object block
// Describe a vertex buffer, its attributes and vertices
// Each group of vertices is associated with an ObjectID and index
// to re-use them accross render passes.
func encodeVBO(buf encoding.WritableBlock) {
	// Describe a vertex buffer
	buf.NewBlock(0)
	buf.PutUint8(0)                      // VBO index
	buf.PutUint8(2)                      // Number of attributes per vertex
	buf.PutUint8(shaders.Position.Index) // First attribute (position)
	buf.PutUint8(3)                      // Size
	buf.PutUint8(shaders.UV.Index)       // Second attribute (uv)
	buf.PutUint8(2)                      // Size
	buf.EndBlock()
}

func encodeVBOVertices(buf encoding.WritableBlock, o *scene.SceneObject) {
	buf.NewBlock(1)
	buf.PutUint8(0)             // VBO index
	buf.PutUint32(uint32(o.Id)) // Object ID
	buf.PutUint8(6)             // Number of vertices for this object
	buf.PutVector3Float32(1.0, 1.0, 0.0)
	buf.PutVector2Float32(1.0, 0.0)
	buf.PutVector3Float32(-1.0, 1.0, 0.0)
	buf.PutVector2Float32(0.0, 0.0)
	buf.PutVector3Float32(-1.0, -1.0, 0.0)
	buf.PutVector2Float32(0.0, 1.0)
	buf.PutVector3Float32(1.0, 1.0, 0.0)
	buf.PutVector2Float32(1.0, 0.0)
	buf.PutVector3Float32(-1.0, -1.0, 0.0)
	buf.PutVector2Float32(0.0, 1.0)
	buf.PutVector3Float32(1.0, -1.0, 0.0)
	buf.PutVector2Float32(1.0, 1.0)
	buf.EndBlock()
}
