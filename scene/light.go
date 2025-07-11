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

	Ambient           color.RGBA
	AmbientIntensity  float64
	Diffuse           color.RGBA
	DiffuseIntensity  float64
	Specular          color.RGBA
	SpecularIntensity float64

	Intensity float64
}

func NewDirectionalLight(c color.RGBA) *DirectionalLight {
	light := &DirectionalLight{
		Direction:         compute.Vector3{X: -0.2, Y: -1, Z: .3},
		Ambient:           c,
		Diffuse:           c,
		Specular:          c,
		AmbientIntensity:  0.02,
		DiffuseIntensity:  0.5,
		SpecularIntensity: 1.0,
	}
	return light
}

func (l *DirectionalLight) Type() LightType {
	return Directional
}

func (l *DirectionalLight) SetColor(c color.RGBA) {
	l.Ambient = c
	l.Diffuse = c
	l.Specular = c
}

func (l *DirectionalLight) AmbientColor() compute.Vector3 {
	return normalizeColor(l.Ambient, l.AmbientIntensity)
}

func (l *DirectionalLight) DiffuseColor() compute.Vector3 {
	return normalizeColor(l.Diffuse, l.DiffuseIntensity)
}

func (l *DirectionalLight) SpecularColor() compute.Vector3 {
	return normalizeColor(l.Specular, l.SpecularIntensity)
}

type PointLight struct {
	Ambient           color.RGBA
	Diffuse           color.RGBA
	Specular          color.RGBA
	AmbientIntensity  float64
	DiffuseIntensity  float64
	SpecularIntensity float64
	Radius            float64
}

func NewPointLight(c color.RGBA) *PointLight {
	light := &PointLight{
		Ambient:           c,
		Diffuse:           c,
		Specular:          c,
		AmbientIntensity:  0.02,
		DiffuseIntensity:  0.5,
		SpecularIntensity: 1.0,
		Radius:            2.0,
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
	return normalizeColor(l.Ambient, l.AmbientIntensity)
}

func (l *PointLight) DiffuseColor() compute.Vector3 {
	return normalizeColor(l.Diffuse, l.DiffuseIntensity)
}

func (l *PointLight) SpecularColor() compute.Vector3 {
	return normalizeColor(l.Specular, l.SpecularIntensity)
}

type SpotLight struct {
	Direction compute.Vector3

	Ambient           color.RGBA
	Diffuse           color.RGBA
	Specular          color.RGBA
	AmbientIntensity  float64
	DiffuseIntensity  float64
	SpecularIntensity float64

	CutOff      float64
	OuterCutOff float64
}

func NewSpotLight(c color.RGBA) *SpotLight {
	light := &SpotLight{
		Direction:         compute.Vector3{X: 0, Y: -0.2, Z: 1},
		Ambient:           c,
		Diffuse:           c,
		Specular:          c,
		AmbientIntensity:  0.02,
		DiffuseIntensity:  0.8,
		SpecularIntensity: 1.0,
		CutOff:            math.Cos(12.5 * (math.Pi / 180)),
		OuterCutOff:       math.Cos(17.5 * (math.Pi / 180)),
	}
	return light
}

func (l *SpotLight) Type() LightType {
	return Spot
}

func (l *SpotLight) AmbientColor() compute.Vector3 {
	return normalizeColor(l.Ambient, l.AmbientIntensity)
}

func (l *SpotLight) DiffuseColor() compute.Vector3 {
	return normalizeColor(l.Diffuse, l.DiffuseIntensity)
}

func (l *SpotLight) SpecularColor() compute.Vector3 {
	return normalizeColor(l.Specular, l.SpecularIntensity)
}

func normalizeColor(c color.Color, f float64) compute.Vector3 {
	r, g, b, a := c.RGBA()
	red := float64(r) / float64(a)
	blue := float64(b) / float64(a)
	green := float64(g) / float64(a)
	color := compute.Vector3{X: red * f, Y: blue * f, Z: green * f}
	return color
}
