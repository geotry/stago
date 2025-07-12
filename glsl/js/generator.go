package glsljs

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/geotry/rass/glsl"
)

func RenderUniformBlocks(programs []*glsl.Program, outDir string) {
	tpl := "/* Generated file, DO NOT EDIT! */\n"

	// Collect all uniform blocks in shaders
	blocks := make(map[string]*glsl.ShaderUniformBlock)
	for _, pg := range programs {
		for name, b := range pg.Vertex.UniformBlocks {
			if blocks[name] != nil {
				// check if block is similar
			} else {
				blocks[name] = b
			}
		}
		for name, b := range pg.Fragment.UniformBlocks {
			if blocks[name] != nil {
				// check if block is similar
			} else {
				blocks[name] = b
			}
		}
	}

	tpl = fmt.Sprintf("%s%s", tpl, templateCreateUniformBlocks(blocks))

	filename := filepath.Clean(fmt.Sprintf("%s/uniforms.js", outDir))
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.WriteString(tpl)
}

func RenderShader(pg *glsl.Program, outDir string) {
	tpl := "/* Generated file, DO NOT EDIT! */\n"
	tpl = fmt.Sprintf("%s%s", tpl, templateCreateSource(pg))
	tpl = fmt.Sprintf("%s%s", tpl, templateCreateUniform(pg))
	tpl = fmt.Sprintf("%s%s", tpl, templateCreateApi(pg))

	filename := filepath.Clean(fmt.Sprintf("%s/%s.js", outDir, pg.Name))
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.WriteString(tpl)
}

func templateCreateSource(pg *glsl.Program) string {
	var buf bytes.Buffer
	err := SourceTpl.Execute(&buf, pg)
	if err != nil {
		log.Panic(err)
	}
	return buf.String()
}

func templateCreateUniform(pg *glsl.Program) string {
	var buf bytes.Buffer

	uniforms := make([]*glsl.ShaderUniform, 0)
	if pg.Vertex != nil {
		uniforms = append(uniforms, pg.Vertex.Uniforms...)
	}
	if pg.Fragment != nil {
		uniforms = append(uniforms, pg.Fragment.Uniforms...)
	}

	slices.SortFunc(uniforms, func(a, b *glsl.ShaderUniform) int {
		if a.Name > b.Name {
			return 1
		} else {
			return -1
		}
	})

	err := UniformsTpl.Execute(&buf, struct{ Uniforms []*glsl.ShaderUniform }{Uniforms: uniforms})
	if err != nil {
		log.Panic(err)
	}

	return buf.String()
}

func templateCreateApi(pg *glsl.Program) string {
	var buf bytes.Buffer
	err := MainTpl.Execute(&buf, pg)
	if err != nil {
		log.Panic(err)
	}
	return buf.String()
}

func templateCreateUniformBlocks(blocks map[string]*glsl.ShaderUniformBlock) string {
	var buf bytes.Buffer
	err := UniformBlockTemplate.Execute(&buf, struct {
		UniformBlocks map[string]*glsl.ShaderUniformBlock
	}{UniformBlocks: blocks})
	if err != nil {
		log.Panic(err)
	}
	return buf.String()
}
