package shaders



type GlType uint8

const (
	mat4 GlType = iota
	vec3
	vec2
	glInt
)

type Uniform struct {
	name string
	Index uint8
	typ GlType
	size uint8
}

type Attribute struct {
	name string
	typ GlType
	Index uint8
	// index of vertex buffer
	buffer uint
	// offset in vertex buffer
	offset uint
	normalized bool
}

type Variable string

type VertexShader struct {
	source string
	uniforms []Uniform
	attributes []Attribute
}

var Model Uniform = Uniform{
	name: "u_model",
	Index: 0,
	typ: mat4,
	size: 32,
}

var View Uniform = Uniform{
	name: "u_view",
	Index: 1,
	typ: mat4,
	size: 32,
}

var Position Attribute = Attribute{
	name: "a_position",
	Index: 0,
	typ: vec3,
	buffer: 0,
	// offset: 0,
}

var UV Attribute = Attribute{
	name: "a_tex_uv",
	Index: 1,
	typ: vec3,
	buffer: 0,
	// offset: 0,
}

func NewVertexShader() *VertexShader {
	return &VertexShader{
		source: ``,
		uniforms: []Uniform{
			{
				name: "u_tex_index",
				typ: glInt,
				size: 32,
			},
			{
				name: "u_model",
				typ: mat4,
				size: 32,
			},
			{
				name: "u_view",
				typ: mat4,
				// size: 2,
			},
			// {
			// 	name: "u_view_index",
			// 	typ: glInt,
			// 	size: 32,
			// },
		},

		attributes: []Attribute{
			{
				name: "a_position",
				typ: vec3,
				buffer: 0,
			},
			{
				name: "a_tex_uv",
				typ: vec2,
				buffer: 0,
			},
		},
	}
}
