package rendering

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"strconv"
	"sync"
)

type ResourceManager struct {
	Palette  *Texture
	Textures []*TextureGroup

	paletteNextColorIndex int
	mu                    sync.Mutex
}

type TextureModel uint8

const (
	ALPHA TextureModel = iota
	RGB
	RGBA
)

type Texture struct {
	Id            int
	Width, Height int
	Pixels        []uint8
	Index         int // Index in group
	Group         *TextureGroup
}

type TextureGroup struct {
	Id        int
	Width     int
	Height    int
	Model     TextureModel
	PixelSize int
	Textures  []*Texture
}

func NewResourceManager() *ResourceManager {
	paletteGroup := &TextureGroup{
		Id:        1,
		Width:     256,
		Height:    1,
		Model:     RGBA,
		PixelSize: 4,
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

	textureGroup := &TextureGroup{
		Id:        2,
		Width:     1,
		Height:    1,
		Model:     ALPHA,
		PixelSize: 1,
		Textures:  []*Texture{},
	}

	return &ResourceManager{
		Palette:               palette,
		Textures:              []*TextureGroup{textureGroup},
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
func (rm *ResourceManager) NewTexturePalette(
	pixels []uint8,
	width int,
) *Texture {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// todo: Find group matching texture specs, or create new group
	group := rm.Textures[0]

	tex := &Texture{
		Id:     group.Id,
		Index:  len(group.Textures),
		Pixels: make([]uint8, len(pixels)),
		Width:  width,
		Height: len(pixels) / width,
		Group:  group,
	}
	copy(tex.Pixels, pixels)

	if tex.Width > group.Width {
		group.Width = tex.Width
	}
	if tex.Height > group.Height {
		group.Height = tex.Height
	}
	group.Textures = append(group.Textures, tex)

	return tex
}

// Create a new RGBA texture and returns its identifier
func (rm *ResourceManager) NewTextureRGBA(
	pixels []uint8,
	width int,
) *Texture {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// todo: Find group matching texture specs, or create new group
	group := rm.Textures[0]

	tex := &Texture{
		Id:     group.Id,
		Index:  len(group.Textures),
		Pixels: make([]uint8, len(pixels)/4),
		Width:  width,
		Height: len(pixels) / 4 / width,
		Group:  group,
	}

	// Transform texture in paletted
	for i := 0; i < len(pixels); i = i + 4 {
		r, g, b, a := pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]

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
			tex.Pixels[i] = uint8(pi)
		} else if rm.paletteNextColorIndex < len(rm.Palette.Pixels)/4 {
			// Add color to palette
			rm.Palette.Pixels[rm.paletteNextColorIndex] = r
			rm.Palette.Pixels[rm.paletteNextColorIndex+1] = g
			rm.Palette.Pixels[rm.paletteNextColorIndex+2] = b
			rm.Palette.Pixels[rm.paletteNextColorIndex+3] = a
			tex.Pixels[i] = uint8(rm.paletteNextColorIndex)
			rm.paletteNextColorIndex++
		} else {
			// Palette is full
		}
	}

	if tex.Width > group.Width {
		group.Width = tex.Width
	}
	if tex.Height > group.Height {
		group.Height = tex.Height
	}

	group.Textures = append(group.Textures, tex)

	return tex
}

func (rm *ResourceManager) NewTextureRGBAFromFile(p string) *Texture {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	file, err := os.Open(p)
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

	// todo: Find group matching texture specs, or create new group
	group := rm.Textures[0]

	tex := &Texture{
		Id:     group.Id,
		Index:  len(group.Textures),
		Pixels: make([]uint8, width*height),
		Width:  width,
		Height: height,
		Group:  group,
	}

	for x := range width {
		for y := range height {
			r, g, b, a := img.At(x, y).RGBA()
			i := y*width + x

			if uint8(a>>8) == 0 {
				tex.Pixels[i] = 255
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
					tex.Pixels[i] = uint8(pi)
				} else if rm.paletteNextColorIndex < len(rm.Palette.Pixels)/4 {
					// Add color to palette
					rm.Palette.Pixels[(rm.paletteNextColorIndex * 4)] = uint8(r >> 8)
					rm.Palette.Pixels[(rm.paletteNextColorIndex*4)+1] = uint8(g >> 8)
					rm.Palette.Pixels[(rm.paletteNextColorIndex*4)+2] = uint8(b >> 8)
					rm.Palette.Pixels[(rm.paletteNextColorIndex*4)+3] = uint8(a >> 8)
					tex.Pixels[i] = uint8(rm.paletteNextColorIndex)
					rm.paletteNextColorIndex++
				} else {
					// Palette is full
				}
			}
		}
	}

	if tex.Width > group.Width {
		group.Width = tex.Width
	}
	if tex.Height > group.Height {
		group.Height = tex.Height
	}

	group.Textures = append(group.Textures, tex)

	return tex
}
