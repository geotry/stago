package glsl

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type ShaderType uint8

const (
	ShaderVertex ShaderType = iota
	ShaderFragment
	ShaderPartial
)

type ShaderVariable struct {
	Name   string
	Type   string
	Size   int
	Struct *ShaderStruct
}

type ShaderAttribute struct {
	Name   string
	Type   string
	Size   int
	Struct *ShaderStruct
}

type ShaderUniform struct {
	Name   string        // u_point_light
	Type   string        // PointLight
	Size   int           // 3
	Struct *ShaderStruct // PointLight
}

type ShaderUniformBlock struct {
	Name   string // Camera
	Fields []*ShaderUniformBlockField
}

type ShaderUniformBlockField struct {
	Name   string
	Type   string
	Size   int
	Struct *ShaderStruct
}

type ShaderStruct struct {
	Name   string // direction
	Fields []*ShaderStructField
}

type ShaderStructField struct {
	Name   string
	Type   string
	Size   int
	Struct *ShaderStruct
}

type Shader struct {
	Source        string
	SourceFile    string
	Type          ShaderType
	Attributes    []*ShaderAttribute
	Uniforms      []*ShaderUniform
	UniformBlocks map[string]*ShaderUniformBlock
}

type ShaderParsed struct {
	Name   string
	File   string
	Lines  int
	Source string
	Type   ShaderType
	Tokens []*ShaderToken
}

type Program struct {
	Name     string
	Fragment *Shader
	Vertex   *Shader
}

// Scan a directory with glsl files and return a list of programs
func Scan(dir string) ([]*Program, error) {
	programs := make([]*Program, 0)

	// Load every shader in dir
	shadersParsed, err := loadShaders(dir)
	if err != nil {
		return nil, err
	}

	// Index shaders by their filenames
	shadersIndex := make(map[string]*ShaderParsed)

	for _, shader := range shadersParsed {
		shadersIndex[shader.File] = shader
	}

	programsIndex := make(map[string]*Program)

	// Resolve include directive by merging shader source and tokens
	for _, parsed := range shadersParsed {
		if parsed.Type == ShaderVertex || parsed.Type == ShaderFragment {
			var includeIndex int
			for includeIndex >= 0 {
				includeIndex = slices.IndexFunc(parsed.Tokens, func(t *ShaderToken) bool { return t.Type == TokenInclude })
				if includeIndex >= 0 {
					token := parsed.Tokens[includeIndex]
					partial := shadersIndex[token.Include]
					if partial == nil {
						return nil, fmt.Errorf("file %s does not exist (imported in %s)", token.Include, parsed.File)
					}

					// Remove include token from original parsed shader
					parsed.Tokens = slices.Delete(parsed.Tokens, includeIndex, includeIndex+1)

					// Insert source tokens before main definition
					parsed.Tokens = insertTokensBefore(parsed.Tokens, filterTokens(partial.Tokens, []TokenType{TokenStruct, TokenStructField, TokenEndStruct}), []TokenType{TokenStruct, TokenUniform, TokenMain, TokenSource})
					parsed.Tokens = insertTokensBefore(parsed.Tokens, filterTokens(partial.Tokens, []TokenType{TokenUniformBlock, TokenUniformBlockField, TokenEndUniformBlock}), []TokenType{TokenUniformBlock, TokenUniform, TokenMain, TokenSource})
					parsed.Tokens = insertTokensBefore(parsed.Tokens, filterTokens(partial.Tokens, []TokenType{TokenUniform}), []TokenType{TokenUniform, TokenMain, TokenSource})
					parsed.Tokens = insertTokensBefore(parsed.Tokens, filterTokens(partial.Tokens, []TokenType{TokenAttribute}), []TokenType{TokenAttribute, TokenMain, TokenSource})
					parsed.Tokens = insertTokensBefore(parsed.Tokens, filterTokens(partial.Tokens, []TokenType{TokenSource}), []TokenType{TokenMain, TokenSource})
				}
			}

			// Once includes are resolved, we need to re-arrange tokens

			uniforms := make([]*ShaderUniform, 0)
			attributes := make([]*ShaderAttribute, 0)
			structs := make(map[string]*ShaderStruct, 0)
			uniformBlocks := make(map[string]*ShaderUniformBlock, 0)

			// index constants to replace size
			constants := make(map[string]string)
			for _, t := range parsed.Tokens {
				if t.Type == TokenConstant {
					constants[t.ConstantName] = t.ConstantValue
				}
			}

			for _, t := range parsed.Tokens {
				if t.VariableSizeRaw != "" && constants[t.VariableSizeRaw] != "" {
					s, err := strconv.Atoi(constants[t.VariableSizeRaw])
					if err != nil {
						return nil, err
					}
					t.VariableSize = s
				}
			}

			var currentStruct *ShaderStruct
			for _, t := range parsed.Tokens {
				if currentStruct != nil {
					if t.Type != TokenStructField {
						currentStruct = nil
					} else {
						currentStruct.Fields = append(currentStruct.Fields, &ShaderStructField{
							Name: t.VariableName,
							Type: t.VariableType,
							Size: t.VariableSize,
						})
					}
				}
				if t.Type == TokenStruct {
					currentStruct = &ShaderStruct{Name: t.StructName, Fields: make([]*ShaderStructField, 0)}
					structs[t.StructName] = currentStruct
				}
			}

			var currentUniformBlock *ShaderUniformBlock
			for _, t := range parsed.Tokens {
				if currentUniformBlock != nil {
					if t.Type != TokenUniformBlockField {
						currentUniformBlock = nil
					} else {
						currentUniformBlock.Fields = append(currentUniformBlock.Fields, &ShaderUniformBlockField{
							Name: t.VariableName,
							Type: t.VariableType,
							Size: t.VariableSize,
						})
					}
				}
				if t.Type == TokenUniformBlock {
					currentUniformBlock = &ShaderUniformBlock{Name: t.UniformBlock, Fields: make([]*ShaderUniformBlockField, 0)}
					uniformBlocks[t.UniformBlock] = currentUniformBlock
				}
			}

			// todo: do a second pass on struct for nested struct on fields

			// Resolve all uniforms and attributes
			for _, t := range parsed.Tokens {
				if t.Type == TokenUniform || t.Type == TokenAttribute {
					variable := &ShaderVariable{
						Name: t.VariableName,
						Type: t.VariableType,
						Size: t.VariableSize,
					}
					if structs[t.VariableType] != nil {
						variable.Struct = structs[t.VariableType]
					}
					if t.Type == TokenUniform {
						uniforms = append(uniforms, &ShaderUniform{Name: variable.Name, Type: variable.Type, Size: variable.Size, Struct: variable.Struct})
					} else {
						attributes = append(attributes, &ShaderAttribute{Name: variable.Name, Type: variable.Type, Size: variable.Size, Struct: variable.Struct})
					}
				}
			}

			shader := &Shader{
				Type:          parsed.Type,
				SourceFile:    parsed.File,
				Uniforms:      uniforms,
				Attributes:    attributes,
				Source:        writeSource(parsed.Tokens),
				UniformBlocks: uniformBlocks,
			}

			if programsIndex[parsed.Name] == nil {
				programsIndex[parsed.Name] = &Program{Name: parsed.Name}
			}
			if parsed.Type == ShaderVertex {
				programsIndex[parsed.Name].Vertex = shader
			}
			if parsed.Type == ShaderFragment {
				programsIndex[parsed.Name].Fragment = shader
			}
		}
	}

	for _, pg := range programsIndex {
		programs = append(programs, pg)
	}

	return programs, nil
}

func loadShaders(dir string) ([]*ShaderParsed, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	shaders := make([]*ShaderParsed, 0)

	for _, entry := range entries {
		entryFullname := fmt.Sprintf("%s/%s", dir, entry.Name())
		ext := filepath.Ext(entryFullname)
		if entry.IsDir() {
			subShaders, err := loadShaders(entryFullname)
			if err != nil {
				return nil, err
			}
			shaders = append(shaders, subShaders...)
		} else if ext == ".glsl" {
			shader, err := load(entryFullname)
			if err != nil {
				return nil, err
			}
			shaders = append(shaders, shader)
		}
	}

	return shaders, nil
}

// Load and parse a single shader file
func load(file string) (*ShaderParsed, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	filename := filepath.Base(file)
	// Create parsing context based on file name
	ctx := &ShaderParseContext{
		file: filepath.Clean(file),
	}
	switch filename {
	case "vertex.glsl":
		ctx.shaderType = ShaderVertex
	case "fragment.glsl":
		ctx.shaderType = ShaderFragment
	default:
		ctx.shaderType = ShaderPartial
	}

	source, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	tokens, err := parse(f, ctx)
	if err != nil {
		return nil, err
	}

	return &ShaderParsed{
		Name:   filepath.Base(filepath.Dir(file)),
		File:   filepath.Clean(file),
		Source: string(source),
		Type:   ctx.shaderType,
		Tokens: tokens,
		Lines:  ctx.line,
	}, nil
}

func writeSource(tokens []*ShaderToken) string {
	str := ""

	// slices.SortFunc(tokens, func(t1 *ShaderToken, t2 *ShaderToken) int {
	// 	if t1.Line < t2.Line {
	// 		return -1
	// 	}
	// 	if t1.Line > t2.Line {
	// 		return 1
	// 	}
	// 	return 0
	// })

	for _, t := range tokens {
		if t.Type == TokenComment {
			continue
		}
		str = fmt.Sprintf("%s\n%s", str, t.Raw)
	}

	return strings.Trim(str, "\n\r")
}

func (p *Program) String() string {
	str := fmt.Sprintf("program=%s", p.Name)

	if p.Vertex != nil {
		str = fmt.Sprintf("%s\n[vertex=%s]", str, p.Vertex.SourceFile)
		for _, attribute := range p.Vertex.Attributes {
			str = fmt.Sprintf("%s\n\tattribute %s %s %d", str, attribute.Name, attribute.Type, attribute.Size)
		}
		for _, uniform := range p.Vertex.Uniforms {
			str = fmt.Sprintf("%s\n\tuniform %s %s %d", str, uniform.Name, uniform.Type, uniform.Size)
		}
	}
	if p.Fragment != nil {
		str = fmt.Sprintf("%s\n[fragment=%s]", str, p.Fragment.SourceFile)
		for _, uniform := range p.Fragment.Uniforms {
			str = fmt.Sprintf("%s\n\tuniform %s %s %d", str, uniform.Name, uniform.Type, uniform.Size)
		}
	}

	return str
}
