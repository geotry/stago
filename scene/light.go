package scene

import (
	"image/color"
	"math"

	"github.com/geotry/rass/compute"
)

type LightType uint8

const (
	Directional LightType = iota
	Point
	Spot
)

type Light interface {
	Type() LightType
	AmbientColor() compute.Vector3
	DiffuseColor() compute.Vector3
	SpecularColor() compute.Vector3
}

type DirectionalLight struct {
	Direction compute.Vector3
	Ambient   color.RGBA
	Diffuse   color.RGBA
	Specular  color.RGBA

	viewMatrix *compute.Matrix4
}

func NewDirectionalLight(c color.RGBA, a, d, s uint8) *DirectionalLight {
	light := &DirectionalLight{
		Direction: compute.Vector3{X: .5, Y: -.8, Z: 0},
		Ambient:   color.RGBA{R: c.R, G: c.G, B: c.B, A: a},
		Diffuse:   color.RGBA{R: c.R, G: c.G, B: c.B, A: d},
		Specular:  color.RGBA{R: c.R, G: c.G, B: c.B, A: s},

		viewMatrix: compute.NewMatrix4(),
	}
	return light
}

func (l *DirectionalLight) Type() LightType {
	return Directional
}

func (l *DirectionalLight) AmbientColor() compute.Vector3 {
	return normalizeColor(l.Ambient)
}

func (l *DirectionalLight) DiffuseColor() compute.Vector3 {
	return normalizeColor(l.Diffuse)
}

func (l *DirectionalLight) SpecularColor() compute.Vector3 {
	return normalizeColor(l.Specular)
}

func (l *DirectionalLight) ViewMatrix(position compute.Point) compute.Matrix {
	l.viewMatrix.Reset()
	l.viewMatrix.LookAt(position, position.Add(l.Direction))
	return l.viewMatrix.Out
}

type PointLight struct {
	Ambient  color.RGBA
	Diffuse  color.RGBA
	Specular color.RGBA
	Radius   float64
}

func NewPointLight(c color.RGBA, a, d, s uint8) *PointLight {
	light := &PointLight{
		Ambient:  color.RGBA{R: c.R, G: c.G, B: c.B, A: a},
		Diffuse:  color.RGBA{R: c.R, G: c.G, B: c.B, A: d},
		Specular: color.RGBA{R: c.R, G: c.G, B: c.B, A: s},
		Radius:   2.0,
	}
	return light
}

func (l *PointLight) Type() LightType {
	return Point
}

func (l *PointLight) SetColor(c color.RGBA) {
	l.Ambient = c
	l.Diffuse = c
	l.Specular = c
}

func (l *PointLight) AmbientColor() compute.Vector3 {
	return normalizeColor(l.Ambient)
}

func (l *PointLight) DiffuseColor() compute.Vector3 {
	return normalizeColor(l.Diffuse)
}

func (l *PointLight) SpecularColor() compute.Vector3 {
	return normalizeColor(l.Specular)
}

type SpotLight struct {
	Direction   compute.Vector3
	Ambient     color.RGBA
	Diffuse     color.RGBA
	Specular    color.RGBA
	CutOff      float64
	OuterCutOff float64

	viewMatrix *compute.Matrix4
}

func NewSpotLight(c color.RGBA, a, d, s uint8) *SpotLight {
	light := &SpotLight{
		Direction:   compute.Vector3{X: 0, Y: -0.2, Z: 1},
		Ambient:     color.RGBA{R: c.R, G: c.G, B: c.B, A: a},
		Diffuse:     color.RGBA{R: c.R, G: c.G, B: c.B, A: d},
		Specular:    color.RGBA{R: c.R, G: c.G, B: c.B, A: s},
		CutOff:      math.Cos(12.5 * (math.Pi / 180)),
		OuterCutOff: math.Cos(17.5 * (math.Pi / 180)),
		viewMatrix:  compute.NewMatrix4(),
	}
	return light
}

func (l *SpotLight) Type() LightType {
	return Spot
}

func (l *SpotLight) AmbientColor() compute.Vector3 {
	return normalizeColor(l.Ambient)
}

func (l *SpotLight) DiffuseColor() compute.Vector3 {
	return normalizeColor(l.Diffuse)
}

func (l *SpotLight) SpecularColor() compute.Vector3 {
	return normalizeColor(l.Specular)
}

func (l *SpotLight) ViewMatrix(position compute.Point, rotation compute.Rotation) compute.Matrix {
	l.viewMatrix.Reset()
	l.viewMatrix.LookAt(position, position.Add(l.Direction))
	l.viewMatrix.Rotate(rotation.Opposite())
	return l.viewMatrix.Out
}

func normalizeColor(c color.RGBA) compute.Vector3 {
	f := float64(c.A) / 255
	return compute.Vector3{X: (float64(c.R) / 255) * f, Y: (float64(c.G) / 255) * f, Z: (float64(c.B) / 255) * f}
}
