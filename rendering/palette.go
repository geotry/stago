package rendering

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

// https://mycours.es/crc/F0D62941

var Palette0 = []string{
	"0B0F00", // #0B0F00
	"281703", // #281703
	"3F0F07", // #3F0F07
	"564907", // #564907
	"3D700E", // #3D700E
	"18872E", // #18872E
	"219E8F", // #219E8F
	"2D69B7", // #2D69B7
	"603BCE", // #603BCE
	"CF4ED8", // #CF4ED8
	"E06494", // #E06494
	"E8A081", // #E8A081
	"EDEB9C", // #EDEB9C
	"CAF4B5", // #CAF4B5
	"D6FFE4", // #D6FFE4
	"EFFEFF", // #EFFEFF
}

type Palette struct {
	Colors color.Palette
	Size   int
}

func NewPalette(list []string) *Palette {
	palette := make(color.Palette, len(list))

	for i, rgb := range list {
		r, _ := strconv.ParseUint(rgb[:2], 16, 8)
		g, _ := strconv.ParseUint(rgb[2:4], 16, 8)
		b, _ := strconv.ParseUint(rgb[4:], 16, 8)
		c := &color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}

		palette[i] = c
	}

	return &Palette{
		Colors: palette,
		Size:   len(list),
	}
}

func NewPalette16(list []string, darker int, lighter int) (*Palette, error) {
	size := len(list) + len(list)*(darker+lighter)
	if size > 256 {
		return nil, fmt.Errorf("palette too large: %d>256", size)
	}

	palette := make(color.Palette, size)

	for i, rgb := range list {
		i0 := i + i*(darker+lighter)

		r, _ := strconv.ParseUint(rgb[:2], 16, 8)
		g, _ := strconv.ParseUint(rgb[2:4], 16, 8)
		b, _ := strconv.ParseUint(rgb[4:], 16, 8)
		c := &color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}

		palette[i0] = c

		for v := range lighter {
			g := Lighten(c, (float64(v+1)/float64(lighter))*0.2)
			palette[i0+1+v] = g
		}
		for v := range darker {
			g := Darken(c, (float64(v+1)/float64(lighter))*0.2)
			palette[i0+1+lighter+v] = g
		}
	}

	return &Palette{
		Colors: palette,
		Size:   len(list),
		// GradientDark: darker,
		// GradientLight: lighter,
	}, nil
}

func (p *Palette) Encode() []uint8 {
	bytes := make([]uint8, len(p.Colors)*4)
	for i, c := range p.Colors {
		r, g, b, a := c.RGBA()
		bytes[i*4] = uint8(r >> 8)
		bytes[i*4+1] = uint8(g >> 8)
		bytes[i*4+2] = uint8(b >> 8)
		bytes[i*4+3] = uint8(a >> 8)
	}
	return bytes
}

func (p *Palette) At(i int) color.Color {
	return p.Colors[i]
}

// Create a new Color palette from a file
func NewPaletteFromFile(filename string) (color.Palette, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(filename)

	var img image.Image

	switch ext {
	case ".png":
		if img, err = png.Decode(file); err != nil {
			return nil, err
		}
	case ".jpg":
	case ".jpeg":
		if img, err = jpeg.Decode(file); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("file extension %v not recognized", ext)
	}

	colors := map[string]color.Color{}

	for x := range img.Bounds().Dx() {
		for y := range img.Bounds().Dy() {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			k := fmt.Sprintf("(%d,%d,%d,%d)", r, g, b, a)
			if colors[k] == nil {
				colors[k] = c
			}
		}
	}

	palette := make([]color.Color, len(colors))
	i := 0
	for _, color := range colors {
		palette[i] = color
		i = i + 1
	}

	return palette, nil
}

// Darken a color by {a} percent
// v = 0 = c
// v = 1 = dark
func Darken(c *color.RGBA, v float64) color.Color {
	h, s, l := ColorToHSL(c)
	dl := max(l-v*(float64(1)-l), 0)
	r, g, b, a := HSLToColor(h, s, dl).RGBA()
	return &color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

// v = 0 = c
// v= 1 = white
func Lighten(c *color.RGBA, v float64) color.Color {
	h, s, l := ColorToHSL(c)
	ll := min(l+v*(1-l), 1)
	r, g, b, a := HSLToColor(h, s, ll).RGBA()
	return &color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func ColorToHSL(c color.Color) (h int, s, l float64) {
	r, g, b, _ := c.RGBA()
	return RGBToHSL(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func HSLToColor(h int, s, l float64) color.Color {
	r, g, b, err := HSLToRGB(h, s, l)
	if err != nil {
		return &color.RGBA{}
	}
	return &color.RGBA{R: r, G: g, B: b, A: 255}
}

func RGBToHSL(r, g, b uint8) (int, float64, float64) {
	var h, s, l float64

	// convert uint32 pre-multiplied value to uint8
	// The r,g,b values are divided by 255 to change the range from 0..255 to 0..1:
	Rnot := float64(r) / 255
	Gnot := float64(g) / 255
	Bnot := float64(b) / 255
	Cmax, Cmin := getMaxMin(Rnot, Gnot, Bnot)
	Δ := Cmax - Cmin
	// Lightness calculation:
	l = (Cmax + Cmin) / 2
	// Hue and Saturation Calculation:
	if Δ == 0 {
		h = 0
		s = 0
	} else {
		switch Cmax {
		case Rnot:
			h = 60 * (math.Mod((Gnot-Bnot)/Δ, 6))
		case Gnot:
			h = 60 * (((Bnot - Rnot) / Δ) + 2)
		case Bnot:
			h = 60 * (((Rnot - Gnot) / Δ) + 4)
		}
		if h < 0 {
			h += 360
		}
		s = Δ / (1 - math.Abs((2*l)-1))
	}

	return int(h), math.Round(s*1000) / 1000, math.Round(l*1000) / 1000
}

func getMaxMin(a, b, c float64) (max, min float64) {
	if a > b {
		max = a
		min = b
	} else {
		max = b
		min = a
	}
	if c > max {
		max = c
	} else if c < min {
		min = c
	}
	return max, min
}

func HSLToRGB(h int, s, l float64) (r, g, b uint8, err error) {
	if h < 0 || h >= 360 ||
		s < 0 || s > 1 ||
		l < 0 || l > 1 {
		return 0, 0, 0, fmt.Errorf("out of range")
	}
	// When 0 ≤ h < 360, 0 ≤ s ≤ 1 and 0 ≤ l ≤ 1:
	C := (1 - math.Abs((2*l)-1)) * s
	X := C * (1 - math.Abs(math.Mod(float64(h)/60, 2)-1))
	m := l - (C / 2)
	var Rnot, Gnot, Bnot float64

	switch {
	case 0 <= h && h < 60:
		Rnot, Gnot, Bnot = C, X, 0
	case 60 <= h && h < 120:
		Rnot, Gnot, Bnot = X, C, 0
	case 120 <= h && h < 180:
		Rnot, Gnot, Bnot = 0, C, X
	case 180 <= h && h < 240:
		Rnot, Gnot, Bnot = 0, X, C
	case 240 <= h && h < 300:
		Rnot, Gnot, Bnot = X, 0, C
	case 300 <= h && h < 360:
		Rnot, Gnot, Bnot = C, 0, X
	}
	r = uint8(math.Round((Rnot + m) * 255))
	g = uint8(math.Round((Gnot + m) * 255))
	b = uint8(math.Round((Bnot + m) * 255))
	return r, g, b, nil
}
