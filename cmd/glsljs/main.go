package main

import (
	"log"

	"github.com/geotry/rass/glsl"
	glsljs "github.com/geotry/rass/glsl/js"
)

const ShadersDir = "./shaders"
const ShadersOutDir = "./web/src/shaders"

func main() {
	programs, err := glsl.Scan(ShadersDir)
	if err != nil {
		log.Panic(err)
	}
	for _, pg := range programs {
		glsljs.RenderShader(pg, ShadersOutDir)
	}
	glsljs.RenderUniformBlocks(programs, ShadersOutDir)
}
