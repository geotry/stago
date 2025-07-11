package rendering

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strconv"
	"sync"
)

type ResourceManager struct {
	Palette  *Texture
	Diffuse  *TextureGroup
	Specular *TextureGroup

	paletteNextColorIndex int
	mu                    sync.Mutex
}

type TextureModel uint8
type TextureRole uint8

const (
	ALPHA TextureModel = iota
	RGB
	RGBA
)

const (
	Diffuse TextureRole = iota
	TPalette
	Specular
	Normal
)

type Texture struct {
	Id            int
	Width, Height int
	Pixels        []uint8
	Index         int // Index in group
	Group         *TextureGroup
}

type Material struct {
	Diffuse   *Texture
	Specular  *Texture
	Shininess float64
}

type TextureGroup struct {
	Id        int
	Width     int
	Height    int
	Model     TextureModel
	PixelSize int
	Role      TextureRole
	Textures  []*Texture
}

func NewResourceManager() *ResourceManager {
	paletteGroup := &TextureGroup{
		Id:        1,
		Width:     256,
		Height:    1,
		Model:     RGBA,
		PixelSize: 4,
		Role:      TPalette,
		Textures:  []*Texture{},
	}
	palette := &Texture{
		Id:     1,
		Width:  256,
		Height: 1,
		Pixels: make([]uint8, 256*4),
		Group:  paletteGroup,
	}
	paletteGroup.Textures = append(paletteGroup.Textures, palette)

	diffuseGroup := &TextureGroup{
		Id:        2,
		Width:     1,
		Height:    1,
		Model:     ALPHA,
		PixelSize: 1,
		Role:      Diffuse,
		Textures:  []*Texture{},
	}

	specularGroup := &TextureGroup{
		Id:        3,
		Width:     1,
		Height:    1,
		Model:     ALPHA,
		PixelSize: 1,
		Role:      Specular,
		Textures:  []*Texture{},
	}

	return &ResourceManager{
		Palette:               palette,
		Diffuse:               diffuseGroup,
		Specular:              specularGroup,
		paletteNextColorIndex: 0,
	}
}

func (rm *ResourceManager) UseRGBPalette(colors []string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(colors) > len(rm.Palette.Pixels)/4 {
		return fmt.Errorf("palette is too large (256 colors max)")
	}

	for i, rgb := range colors {
		r, err := strconv.ParseUint(rgb[:2], 16, 8)
		if err != nil {
			return err
		}
		g, err := strconv.ParseUint(rgb[2:4], 16, 8)
		if err != nil {
			return err
		}
		b, err := strconv.ParseUint(rgb[4:], 16, 8)
		if err != nil {
			return err
		}
		rm.Palette.Pixels[(i * 4)] = uint8(r)
		rm.Palette.Pixels[(i*4)+1] = uint8(g)
		rm.Palette.Pixels[(i*4)+2] = uint8(b)
		rm.Palette.Pixels[(i*4)+3] = 255
		rm.paletteNextColorIndex++
	}

	return nil
}

// Create a new texture from palette colors and returns its identifier
func (rm *ResourceManager) NewMaterialPalette(
	width int,
	diffuse []uint8,
	specular []uint8,
	shininess float64,
) *Material {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Diffuse
	tex := &Texture{
		Id:     rm.Diffuse.Id,
		Index:  len(rm.Diffuse.Textures),
		Pixels: make([]uint8, len(diffuse)),
		Width:  width,
		Height: len(diffuse) / width,
		Group:  rm.Diffuse,
	}
	copy(tex.Pixels, diffuse)
	if tex.Width > rm.Diffuse.Width {
		rm.Diffuse.Width = tex.Width
	}
	if tex.Height > rm.Diffuse.Height {
		rm.Diffuse.Height = tex.Height
	}
	rm.Diffuse.Textures = append(rm.Diffuse.Textures, tex)

	// Specular
	spec := &Texture{
		Id:     rm.Specular.Id,
		Index:  len(rm.Specular.Textures),
		Pixels: make([]uint8, len(specular)),
		Width:  width,
		Height: len(specular) / width,
		Group:  rm.Specular,
	}
	copy(spec.Pixels, specular)
	if spec.Width > rm.Specular.Width {
		rm.Specular.Width = spec.Width
	}
	if spec.Height > rm.Specular.Height {
		rm.Specular.Height = spec.Height
	}
	rm.Specular.Textures = append(rm.Specular.Textures, spec)

	return &Material{
		Diffuse:   tex,
		Specular:  spec,
		Shininess: shininess,
	}
}

// Create a new RGBA texture and returns its identifier
func (rm *ResourceManager) NewMaterialRGBA(
	width int,
	diffuse []uint8,
	shininess float64,
) *Material {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	diff := &Texture{
		Id:     rm.Diffuse.Id,
		Index:  len(rm.Diffuse.Textures),
		Pixels: make([]uint8, len(diffuse)/4),
		Width:  width,
		Height: len(diffuse) / 4 / width,
		Group:  rm.Diffuse,
	}

	// Transform texture in paletted
	for i := 0; i < len(diffuse); i = i + 4 {
		r, g, b, a := diffuse[i], diffuse[i+1], diffuse[i+2], diffuse[i+3]

		// Look for color in palette
		pi := -1
		for j := 0; j < len(rm.Palette.Pixels); j = j + 4 {
			pr, pg, pb, pa := rm.Palette.Pixels[i], rm.Palette.Pixels[i+1], rm.Palette.Pixels[i+2], rm.Palette.Pixels[i+3]
			if pr == r && pg == g && pb == b && pa == a {
				pi = j
				break
			}
		}

		if pi != -1 {
			diff.Pixels[i] = uint8(pi)
		} else if rm.paletteNextColorIndex < len(rm.Palette.Pixels)/4 {
			// Add color to palette
			rm.Palette.Pixels[rm.paletteNextColorIndex] = r
			rm.Palette.Pixels[rm.paletteNextColorIndex+1] = g
			rm.Palette.Pixels[rm.paletteNextColorIndex+2] = b
			rm.Palette.Pixels[rm.paletteNextColorIndex+3] = a
			diff.Pixels[i] = uint8(rm.paletteNextColorIndex)
			rm.paletteNextColorIndex++
		} else {
			// Palette is full
		}
	}

	if diff.Width > rm.Diffuse.Width {
		diff.Width = rm.Diffuse.Width
	}
	if diff.Height > rm.Diffuse.Height {
		diff.Height = rm.Diffuse.Height
	}

	rm.Diffuse.Textures = append(rm.Diffuse.Textures, diff)

	return &Material{
		Diffuse:   diff,
		Specular:  nil,
		Shininess: shininess,
	}
}

func (rm *ResourceManager) NewTextureFromAtlas(
	path string,
	role TextureRole,
	offsetX, offsetY int,
	texWidth, texHeight int,
) *Texture {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil
	}

	img, err := png.Decode(file)
	if err != nil {
		log.Println(err)
		return nil
	}

	var group *TextureGroup

	switch role {
	case Diffuse:
		group = rm.Diffuse
	case Specular:
		group = rm.Specular
	}

	texture := &Texture{
		Id:     group.Id,
		Index:  len(group.Textures),
		Pixels: make([]uint8, texWidth*texHeight),
		Width:  texWidth,
		Height: texHeight,
		Group:  group,
	}

	rm.copyPixels(texture, img, role, offsetX, offsetY)

	if texture.Width > group.Width {
		group.Width = texture.Width
	}
	if texture.Height > group.Height {
		group.Height = texture.Height
	}

	group.Textures = append(group.Textures, texture)

	return texture
}

func (rm *ResourceManager) NewMaterialRGBAFromFile(path string, role TextureRole) *Texture {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil
	}

	img, err := png.Decode(file)
	if err != nil {
		log.Println(err)
		return nil
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	var group *TextureGroup

	switch role {
	case Diffuse:
		group = rm.Diffuse
	case Specular:
		group = rm.Specular
	}

	texture := &Texture{
		Id:     group.Id,
		Index:  len(group.Textures),
		Pixels: make([]uint8, width*height),
		Width:  width,
		Height: height,
		Group:  group,
	}

	rm.copyPixels(texture, img, role, 0, 0)

	if texture.Width > group.Width {
		group.Width = texture.Width
	}
	if texture.Height > group.Height {
		group.Height = texture.Height
	}

	group.Textures = append(group.Textures, texture)

	return texture
}

func (rm *ResourceManager) copyPixels(
	texture *Texture,
	img image.Image,
	role TextureRole,
	offsetX, offsetY int,
) {
	width, height := texture.Width, texture.Height
	for x := range width {
		for y := range height {
			r, g, b, a := img.At(offsetX+x, offsetY+y).RGBA()
			i := y*width + x

			if role == Diffuse {
				if uint8(a>>8) == 0 {
					texture.Pixels[i] = 255
				} else {
					// Look for color in palette
					pi := -1
					for j := 0; j < len(rm.Palette.Pixels); j = j + 4 {
						pr, pg, pb, pa := rm.Palette.Pixels[j], rm.Palette.Pixels[j+1], rm.Palette.Pixels[j+2], rm.Palette.Pixels[j+3]
						if pr == uint8(r>>8) && pg == uint8(g>>8) && pb == uint8(b>>8) && pa == uint8(a>>8) {
							pi = j / 4
							break
						}
					}

					if pi != -1 {
						texture.Pixels[i] = uint8(pi)
					} else if rm.paletteNextColorIndex < len(rm.Palette.Pixels)/4 {
						// Add color to palette
						rm.Palette.Pixels[(rm.paletteNextColorIndex * 4)] = uint8(r >> 8)
						rm.Palette.Pixels[(rm.paletteNextColorIndex*4)+1] = uint8(g >> 8)
						rm.Palette.Pixels[(rm.paletteNextColorIndex*4)+2] = uint8(b >> 8)
						rm.Palette.Pixels[(rm.paletteNextColorIndex*4)+3] = uint8(a >> 8)
						texture.Pixels[i] = uint8(rm.paletteNextColorIndex)
						rm.paletteNextColorIndex++
					} else {
						// Palette is full
					}
				}
			}

			// Take only one component
			if role == Specular {
				texture.Pixels[i] = uint8(r >> 8)
			}
		}
	}
}
