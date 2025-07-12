package glsljs

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/geotry/rass/glsl"
)

var SourceTpl = CreateTemplate("Source", `
{{- with .Vertex }}
/**
 * GLSL vertex shader (source: {{ .SourceFile }})
 */
const VERTEX_SRC = `+"`"+`
{{ .Source }}
`+"`"+`.trim();
{{- end }}

{{- with .Fragment }}
/**
 * GLSL fragment shader (source: {{ .SourceFile }})
 */
const FRAGMENT_SRC = `+"`"+`
{{ .Source }}
`+"`"+`.trim();
{{- end }}
`, nil)

// Render get, set functions for a uniform field
var UniformFieldGetSetTpl = CreateTemplate("UniformFieldGetSet", `
{{- $args := glArgs .Type }}
    /**
     * Set the value of uniform `+"`"+`{{.Name}}`+"`"+`.
     *
		{{- range $arg := $args }}
     * @param {{"{"}}{{$arg.Type}}{{"}"}} {{ $arg.Name }}
		{{- end }}
		{{- if glUniformMatrix .Type }}
     * @param {boolean} transpose
		{{- end }}
     */
    set({{range $i, $arg := $args }}{{ $arg.Name }}{{ ifNotLast (len $args) $i ", " }}{{end}}{{if glUniformMatrix .Type }}, transpose = false{{end}}) {
      gl.uniform{{ glUniformType .Type }}(locs[`+"`"+`{{.Name}}`+"`"+`],{{if glUniformMatrix .Type }} transpose, {{end}}{{range $i, $arg := $args }}{{ $arg.Name }}{{ ifNotLast (len $args) $i ", " }}{{end}});
    },
    /**
     * Returns the value of uniform `+"`"+`{{.Name}}`+"`"+`.
     *
     * @returns {number}
     */
    get() {
      return gl.getUniform(program, locs[`+"`"+`{{.Name}}`+"`"+`]);
    },
`, []*template.Template{})

var UniformGetSetTpl = CreateTemplate("UniformGetSet", `
{{- $IsArray := index . "IsArray" }}
{{- $Struct := index . "Struct" }}
{{- if $Struct }}
{{- $Parent := . }}
{{- range $Struct.Fields }}
{{- $Name := $Parent.Name }}
{{- if $IsArray }}{{ $Name = print $Name "[${" $Name "i}]" }}{{end}}
{{- $Name := print $Name "." .Name }}
    {{ .Name }}: {
      {{- template "UniformFieldGetSet" (map "Type" .Type "Name" $Name) }}
    },
{{- end }}
{{- else }}
{{- $Name := .Name }}
{{- if $IsArray }}{{ $Name = print $Name "[${" $Name "i}]" }}{{end}}
{{- template "UniformFieldGetSet" (map "Type" .Type "Name" $Name) }}
{{- end }}
`, []*template.Template{UniformFieldGetSetTpl})

var UniformDefTpl = CreateTemplate("UniformDef", `
{{- if gt .Size 0 }}
  const {{ .Name }} = new Array({{ .Size }}).fill(undefined).map((_, {{ .Name }}i) => ({
{{- template "UniformGetSet" (map "Struct" .Struct "Name" .Name "Type" .Type "IsArray" true) }}
  }));
{{- else }}
  const {{ .Name }} = {
{{- template "UniformGetSet" (map "Struct" .Struct "Name" .Name "Type" .Type "IsArray" false) }}
  };
{{end}}
`, []*template.Template{UniformGetSetTpl, UniformFieldGetSetTpl})

// Template to render uniforms as nested objects and arrays with an API to get, set prepare values.
var UniformsTpl = CreateTemplate("Uniforms", `
/**
 * @param {WebGL2RenderingContext} gl
 * @param {WebGLProgram} program
 * @returns
 */
const createUniforms = (gl, program) => {
  const locs = {
{{- range flattenUniforms .Uniforms }}
    [`+"`"+`{{.}}`+"`"+`]: gl.getUniformLocation(program, "{{.}}"),
{{- end }}
  };
{{- range .Uniforms }}
{{- template "UniformDef" . -}}
{{end}}
  return {
{{- range .Uniforms }}
    {{ .Name }},
{{- end }}
  };
};
`, []*template.Template{UniformDefTpl, UniformGetSetTpl, UniformFieldGetSetTpl})

var MainTpl = CreateTemplate("Main", `
/**
 * The {{ .Name }} shader program.
 */
export const {{ .Name | title }}Shader = {
  name: "{{ .Name }}",
  vertex: VERTEX_SRC,
  fragment: FRAGMENT_SRC,
  createUniforms,
};
`, nil)

var UniformBlockTemplate = CreateTemplate("UniformBlock", `
{{- range $name, $block := .UniformBlocks }}
/**
 * The Camera uniform block.
 */
export const {{ $name }} = {
{{- range $field := $block.Fields }}
	 {{- template "UniformDef" $field -}}
{{- end }}
};
{{- end }}
`, []*template.Template{UniformDefTpl, UniformGetSetTpl, UniformFieldGetSetTpl})

func CreateTemplate(name string, layout string, dependencies []*template.Template) *template.Template {
	tpl := template.Must(template.New(name).Funcs(funcs).Parse(layout))

	for _, child := range dependencies {
		template.Must(tpl.AddParseTree(child.Name(), child.Tree))
	}

	return tpl
}

func flattenUniform(uniform *glsl.ShaderUniform) []string {
	names := make([]string, 0)

	if uniform.Size > 0 {
		for i := range uniform.Size {
			names = append(names, flattenUniform(&glsl.ShaderUniform{
				Name:   fmt.Sprintf("%s[%d]", uniform.Name, i),
				Type:   uniform.Type,
				Size:   0,
				Struct: uniform.Struct,
			})...)
		}
	} else if uniform.Struct != nil {
		for _, f := range uniform.Struct.Fields {
			names = append(names, flattenUniform(&glsl.ShaderUniform{
				Name:   fmt.Sprintf("%s.%s", uniform.Name, f.Name),
				Type:   f.Type,
				Size:   f.Size,
				Struct: f.Struct,
			})...)
		}
	} else {
		names = append(names, uniform.Name)
	}

	return names
}

var funcs = template.FuncMap{
	"title": func(s string) string {
		return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	},
	"map": func(kvs ...any) map[any]any {
		m := make(map[any]any)
		for i := 0; i < len(kvs)-1; i += 2 {
			m[kvs[i]] = kvs[i+1]
		}
		return m
	},
	"flattenUniforms": func(uniforms []*glsl.ShaderUniform) []string {
		names := make([]string, 0)
		for _, uniform := range uniforms {
			names = append(names, flattenUniform(uniform)...)
		}
		return names
	},
	"glArgs": func(typ string) []struct {
		Name string
		Type string
	} {
		switch typ {
		case "int", "float", "sampler2D", "sampler2DArray", "sampler2DShadow", "sampler2DArrayShadow":
			return []struct {
				Name string
				Type string
			}{{Name: "value", Type: "number"}}
		case "vec2", "ivec2":
			return []struct {
				Name string
				Type string
			}{{Name: "x", Type: "number"}, {Name: "y", Type: "number"}}
		case "vec3", "ivec3":
			return []struct {
				Name string
				Type string
			}{{Name: "x", Type: "number"}, {Name: "y", Type: "number"}, {Name: "z", Type: "number"}}
		case "vec4", "ivec4":
			return []struct {
				Name string
				Type string
			}{{Name: "x", Type: "number"}, {Name: "y", Type: "number"}, {Name: "z", Type: "number"}, {Name: "w", Type: "number"}}
		case "mat4":
			return []struct {
				Name string
				Type string
			}{{Name: "matrix", Type: "Float32Array"}}
		}
		return []struct {
			Name string
			Type string
		}{}
	},
	"glUniformMatrix": func(typ string) bool {
		return typ == "mat4"
	},
	"glUniformType": func(typ string) string {
		switch typ {
		case "int", "sampler2D", "sampler2DArray", "sampler2DShadow", "sampler2DArrayShadow":
			return "1i"
		case "float":
			return "1f"
		case "ivec2", "vec2":
			return "2f"
		case "ivec3", "vec3":
			return "3f"
		case "ivec4", "vec4":
			return "4f"
		case "mat4":
			return "Matrix4fv"
		}
		return ""
	},
	"ifNotLast": func(l int, i int, s string) string {
		if l == 0 || i == l-1 {
			return ""
		}
		return s
	},
}
