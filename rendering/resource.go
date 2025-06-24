package rendering

import (
	"encoding/binary"
	"fmt"
	"image/png"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/geotry/rass/compute"
)

type ResourceManager struct {
	// RGBA colors of the palette
	Palette0     []uint8
	TextureIndex []*PalettedTexture

	// next available palette color
	palette0Next int

	mu sync.Mutex
}

type PalettedTexture struct {
	Id     int
	Size   compute.Size
	Pixels []uint8
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		Palette0:     make([]uint8, 256*4),
		TextureIndex: []*PalettedTexture{},
		palette0Next: 0,
	}
}

func (rm *ResourceManager) UseRGBPalette(colors []string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(colors) > len(rm.Palette0)/4 {
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
		rm.Palette0[(i * 4)] = uint8(r)
		rm.Palette0[(i*4)+1] = uint8(g)
		rm.Palette0[(i*4)+2] = uint8(b)
		rm.Palette0[(i*4)+3] = 255
		rm.palette0Next++
	}

	return nil
}

// Create a new texture from palette colors and returns its identifier
func (rm *ResourceManager) NewTexturePalette(
	pixels []uint8,
	width int,
) *PalettedTexture {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	tex := &PalettedTexture{
		Id:     len(rm.TextureIndex),
		Pixels: make([]uint8, len(pixels)),
		Size:   compute.Point{X: 0, Y: 0, Z: 0},
	}
	if width > 0 {
		tex.Size = compute.Point{X: float64(width), Y: float64(len(pixels) / width)}
	}
	copy(tex.Pixels, pixels)

	rm.TextureIndex = append(rm.TextureIndex, tex)

	return tex
}

// Create a new RGBA texture and returns its identifier
func (rm *ResourceManager) NewTextureRGBA(
	pixels []uint8,
	width int,
) *PalettedTexture {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	tex := &PalettedTexture{
		Id:     len(rm.TextureIndex),
		Size:   compute.Point{X: float64(width), Y: float64(len(pixels) / width)},
		Pixels: make([]uint8, len(pixels)/4),
	}

	// Transform texture in paletted
	for i := 0; i < len(pixels); i = i + 4 {
		r, g, b, a := pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]

		// Look for color in palette
		pi := -1
		for j := 0; j < len(rm.Palette0); j = j + 4 {
			pr, pg, pb, pa := rm.Palette0[i], rm.Palette0[i+1], rm.Palette0[i+2], rm.Palette0[i+3]
			if pr == r && pg == g && pb == b && pa == a {
				pi = j
				break
			}
		}

		if pi != -1 {
			tex.Pixels[i] = uint8(pi)
		} else if rm.palette0Next < len(rm.Palette0)/4 {
			// Add color to palette
			rm.Palette0[rm.palette0Next] = r
			rm.Palette0[rm.palette0Next+1] = g
			rm.Palette0[rm.palette0Next+2] = b
			rm.Palette0[rm.palette0Next+3] = a
			tex.Pixels[i] = uint8(rm.palette0Next)
			rm.palette0Next++
		} else {
			// Palette is full
		}
	}

	rm.TextureIndex = append(rm.TextureIndex, tex)

	return tex
}

func (rm *ResourceManager) NewTextureRGBAFromFile(p string) *PalettedTexture {
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

	tex := &PalettedTexture{
		Id:     len(rm.TextureIndex),
		Size:   compute.Point{X: float64(width), Y: float64(height)},
		Pixels: make([]uint8, width*height),
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
				for j := 0; j < len(rm.Palette0); j = j + 4 {
					pr, pg, pb, pa := rm.Palette0[j], rm.Palette0[j+1], rm.Palette0[j+2], rm.Palette0[j+3]
					if pr == uint8(r>>8) && pg == uint8(g>>8) && pb == uint8(b>>8) && pa == uint8(a>>8) {
						pi = j / 4
						break
					}
				}

				if pi != -1 {
					tex.Pixels[i] = uint8(pi)
				} else if rm.palette0Next < len(rm.Palette0)/4 {
					// Add color to palette
					rm.Palette0[(rm.palette0Next * 4)] = uint8(r >> 8)
					rm.Palette0[(rm.palette0Next*4)+1] = uint8(g >> 8)
					rm.Palette0[(rm.palette0Next*4)+2] = uint8(b >> 8)
					rm.Palette0[(rm.palette0Next*4)+3] = uint8(a >> 8)
					tex.Pixels[i] = uint8(rm.palette0Next)
					rm.palette0Next++
				} else {
					// Palette is full
				}
			}
		}
	}

	rm.TextureIndex = append(rm.TextureIndex, tex)

	return tex
}

func (rm *ResourceManager) EncodePalette() []uint8 {
	return rm.Palette0
}

func (rm *ResourceManager) EncodeTextures() []uint8 {
	// Get max tex width and heigt
	texMaxWidth, texMaxHeight := 0, 0
	for _, t := range rm.TextureIndex {
		if int(t.Size.X) > texMaxWidth {
			texMaxWidth = int(t.Size.X)
		}
		if int(t.Size.Y) > texMaxHeight {
			texMaxHeight = int(t.Size.Y)
		}
	}

	encoded := make([]uint8, 4+4*texMaxWidth*texMaxHeight*len(rm.TextureIndex))
	// Encode max texture width and height in first bytes
	binary.BigEndian.PutUint16(encoded, uint16(texMaxWidth))
	binary.BigEndian.PutUint16(encoded[2:], uint16(texMaxHeight))

	for texi, tex := range rm.TextureIndex {
		texWidth, texHeight := int(tex.Size.X), int(tex.Size.Y)

		// Encode texture width(height) before each texture
		// to know how to use texcoords and scale the texture properly
		indexOffset := 4 + texi*texMaxWidth*texMaxHeight + texi*4
		// log.Printf("index=%v offset=%v width=%v height=%v", texi, indexOffset, uint16(texWidth), uint16(texHeight))
		binary.BigEndian.PutUint16(encoded[indexOffset:], uint16(texWidth))
		binary.BigEndian.PutUint16(encoded[indexOffset+2:], uint16(texHeight))

		for y := range texMaxHeight {

			for x := range texMaxWidth {
				if x < texWidth && y < texHeight {
					texIndex := (y * texWidth) + x
					// add width padding based on max texture width
					i := 4 + indexOffset + texIndex + y*(texMaxWidth-texWidth)
					encoded[i] = tex.Pixels[texIndex]
				}
			}
		}
	}

	return encoded
}
