package glsl

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type TokenType uint8

type ShaderParseContext struct {
	isStruct       bool
	isUniformBlock bool
	shaderType     ShaderType
	file           string
	line           int
}

type ShaderToken struct {
	Type            TokenType
	Line            int
	Raw             string
	Comment         string
	ConstantName    string
	ConstantValue   string
	VariableType    string
	VariableName    string
	VariableSize    int
	VariableSizeRaw string
	VersionNumber   int
	VersionName     string
	StructName      string
	Include         string
	Precision       string
	UniformBlock    string
}

const (
	TokenVersion TokenType = iota
	TokenConstant
	TokenInclude
	TokenPrecision
	TokenStruct
	TokenStructField
	TokenEndStruct
	TokenUniformBlock
	TokenUniformBlockField
	TokenEndUniformBlock
	TokenUniform
	TokenAttribute
	TokenVarying
	TokenComment
	TokenMain
	TokenSource
)

const typeRegex = `?P<type>[a-zA-Z0-9_]+`
const varNameRegex = `?P<name>[a-zA-Z0-9_]+`
const varSizeRegex = `\[(?P<size>.+)\]`

var structRegex = regexp.MustCompile(`^struct (?P<name>[a-zA-Z0-9_]+)`)
var structFieldRegex = regexp.MustCompile(fmt.Sprintf(`^\s*(%s) (%s)(%s)?;$`, typeRegex, varNameRegex, varSizeRegex))
var uniformBlockRegex = regexp.MustCompile(`^uniform (?P<name>[a-zA-Z0-9_]+) {$`)
var uniformRegex = regexp.MustCompile(fmt.Sprintf(`^uniform (%s) (%s)(%s)?;$`, typeRegex, varNameRegex, varSizeRegex))
var attributeRegex = regexp.MustCompile(fmt.Sprintf(`^(?:flat )?in (%s) (%s)(%s)?;$`, typeRegex, varNameRegex, varSizeRegex))
var varyingRegex = regexp.MustCompile(fmt.Sprintf(`^(?:flat )?out (%s) (%s)(%s)?;$`, typeRegex, varNameRegex, varSizeRegex))
var constantRegex = regexp.MustCompile(`^#define (?P<name>[a-zA-Z0-9_]+) (?P<value>[a-zA-Z0-9_.]+)$`)
var versionRegex = regexp.MustCompile(`^#version (?P<version>[0-9]+) (?P<ext>[a-zA-Z0-9_.]+)?$`)
var includeRegex = regexp.MustCompile(`^#include (?P<file>[A-Za-z0-9/.]+)$`)
var precisionRegex = regexp.MustCompile(fmt.Sprintf(`^precision (?P<precision>[a-zA-Z0-9_]+) (%s);$`, typeRegex))

func parse(r io.Reader, ctx *ShaderParseContext) ([]*ShaderToken, error) {
	scanner := bufio.NewScanner(r)

	tokens := make([]*ShaderToken, 0)

	for scanner.Scan() {
		line := scanner.Text()
		ctx.line++
		if line != "" {
			token := &ShaderToken{Type: TokenSource, Raw: line, Line: ctx.line}
			token.parse(line, ctx)
			tokens = append(tokens, token)
		}
	}

	return tokens, nil
}

func filterTokens(tokens []*ShaderToken, typ []TokenType) []*ShaderToken {
	results := make([]*ShaderToken, 0)

	for _, t := range tokens {
		for _, rt := range typ {
			if t.Type == rt {
				results = append(results, t)
			}
		}
	}

	return results
}

func insertTokensBefore(tokens []*ShaderToken, insertTokens []*ShaderToken, before []TokenType) []*ShaderToken {
	var index = -1
	for _, b := range before {
		index = slices.IndexFunc(tokens, func(t *ShaderToken) bool {
			return t.Type == b
		})
		if index != -1 {
			break
		}
	}
	if index == -1 {
		return tokens
	}
	return slices.Insert(tokens, index, insertTokens...)
}

func (t *ShaderToken) parse(line string, ctx *ShaderParseContext) {
	trimed := strings.Trim(line, " \n\r")

	if strings.HasPrefix(trimed, "//") {
		t.Type = TokenComment
		t.Comment, _ = strings.CutPrefix(trimed, "//")
		t.Comment = strings.Trim(t.Comment, " \n")
		return
	}

	if strings.HasPrefix(trimed, "void main()") {
		t.Type = TokenMain
		return
	}

	if ctx.isStruct {
		structFieldMatch := structFieldRegex.FindStringSubmatch(line)
		if structFieldMatch != nil {
			t.Type = TokenStructField
			t.VariableType = structFieldMatch[1]
			t.VariableName = structFieldMatch[2]
			if sizeStr := structFieldMatch[4]; sizeStr != "" {
				if size, err := strconv.Atoi(sizeStr); err == nil {
					t.VariableSize = size
				} else {
					t.VariableSizeRaw = sizeStr
				}
			}
			return
		} else {
			t.Type = TokenEndStruct
			ctx.isStruct = false
			return
		}
	}

	if ctx.isUniformBlock {
		structFieldMatch := structFieldRegex.FindStringSubmatch(line)
		if structFieldMatch != nil {
			t.Type = TokenUniformBlockField
			t.VariableType = structFieldMatch[1]
			t.VariableName = structFieldMatch[2]
			if sizeStr := structFieldMatch[4]; sizeStr != "" {
				if size, err := strconv.Atoi(sizeStr); err == nil {
					t.VariableSize = size
				} else {
					t.VariableSizeRaw = sizeStr
				}
			}
			return
		} else {
			t.Type = TokenEndUniformBlock
			ctx.isUniformBlock = false
			return
		}
	}

	includeMatch := includeRegex.FindStringSubmatch(line)
	if includeMatch != nil {
		t.Type = TokenInclude
		t.Include = filepath.Clean(fmt.Sprintf("%s/%s", filepath.Dir(ctx.file), includeMatch[1]))
		return
	}

	constantMatch := constantRegex.FindStringSubmatch(line)
	if constantMatch != nil {
		t.Type = TokenConstant
		t.ConstantName = constantMatch[1]
		t.ConstantValue = constantMatch[2]
		return
	}

	versionMatch := versionRegex.FindStringSubmatch(line)
	if versionMatch != nil {
		t.Type = TokenVersion
		if nb, err := strconv.Atoi(versionMatch[1]); err == nil {
			t.VersionNumber = nb
		}
		t.VersionName = versionMatch[2]
		return
	}

	precisionMatch := precisionRegex.FindStringSubmatch(line)
	if precisionMatch != nil {
		t.Type = TokenPrecision
		t.Precision = precisionMatch[1]
		t.VariableType = precisionMatch[2]
		return
	}

	structMatch := structRegex.FindStringSubmatch(line)
	if structMatch != nil {
		t.Type = TokenStruct
		t.StructName = structMatch[1]
		ctx.isStruct = true
		return
	}

	uniformBlockMatch := uniformBlockRegex.FindStringSubmatch(line)
	if uniformBlockMatch != nil {
		t.Type = TokenUniformBlock
		t.UniformBlock = uniformBlockMatch[1]
		ctx.isUniformBlock = true
		return
	}

	uniformMatch := uniformRegex.FindStringSubmatch(line)
	if uniformMatch != nil {
		t.Type = TokenUniform
		t.VariableType = uniformMatch[1]
		t.VariableName = uniformMatch[2]
		if sizeStr := uniformMatch[4]; sizeStr != "" {
			if size, err := strconv.Atoi(sizeStr); err == nil {
				t.VariableSize = size
			} else {
				t.VariableSizeRaw = sizeStr
			}
		}
		return
	}

	attributeMatch := attributeRegex.FindStringSubmatch(line)
	if attributeMatch != nil {
		if ctx.shaderType == ShaderVertex {
			t.Type = TokenAttribute
		} else {
			t.Type = TokenVarying
		}
		t.VariableType = attributeMatch[1]
		t.VariableName = attributeMatch[2]
		if sizeStr := attributeMatch[4]; sizeStr != "" {
			if size, err := strconv.Atoi(sizeStr); err == nil {
				t.VariableSize = size
			} else {
				t.VariableSizeRaw = sizeStr
			}
		}
		return
	}

	varyingMatch := varyingRegex.FindStringSubmatch(line)
	if varyingMatch != nil {
		t.Type = TokenVarying
		t.VariableType = varyingMatch[1]
		t.VariableName = varyingMatch[2]
		if sizeStr := varyingMatch[4]; sizeStr != "" {
			if size, err := strconv.Atoi(sizeStr); err == nil {
				t.VariableSize = size
			} else {
				t.VariableSizeRaw = sizeStr
			}
		}
		return
	}
}

func (t *ShaderToken) String() string {
	if t.Type == TokenConstant {
		return fmt.Sprintf("%d constant %s %s", t.Line, t.ConstantName, t.ConstantValue)
	} else if t.Type == TokenInclude {
		return fmt.Sprintf("%d include %s", t.Line, t.Include)
	} else if t.Type == TokenVersion {
		return fmt.Sprintf("%d version %d %s", t.Line, t.VersionNumber, t.VersionName)
	} else if t.VariableSizeRaw != "" {
		return fmt.Sprintf("%d %d %s %s [%s]", t.Line, t.Type, t.VariableType, t.VariableName, t.VariableSizeRaw)
	} else if t.VariableSize != 0 {
		return fmt.Sprintf("%d %d %s %s [%d]", t.Line, t.Type, t.VariableType, t.VariableName, t.VariableSize)
	} else if t.Type == TokenStruct {
		return fmt.Sprintf("%d struct %s", t.Line, t.StructName)
	} else if t.Type == TokenUniformBlock {
		return fmt.Sprintf("%d uniform block %s", t.Line, t.UniformBlock)
	} else {
		return fmt.Sprintf("%d %d %s %s", t.Line, t.Type, t.VariableType, t.VariableName)
	}
}
