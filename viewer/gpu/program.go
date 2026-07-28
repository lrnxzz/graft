package gpu

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Program struct {
	id uint32
}

func NewProgram(vertexSource, fragmentSource string) (*Program, error) {
	vertex, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}

	fragment, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	id := gl.CreateProgram()
	gl.AttachShader(id, vertex)
	gl.AttachShader(id, fragment)
	gl.LinkProgram(id)
	gl.DeleteShader(vertex)
	gl.DeleteShader(fragment)

	var status int32
	gl.GetProgramiv(id, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		return nil, fmt.Errorf("gpu: link program: %s", infoLog(id, gl.GetProgramiv, gl.GetProgramInfoLog))
	}

	return &Program{id: id}, nil
}

func (p *Program) Use() {
	gl.UseProgram(p.id)
}

func (p *Program) Mat4(name string, value mgl32.Mat4) {
	gl.UniformMatrix4fv(p.uniform(name), 1, false, &value[0])
}

func (p *Program) Int(name string, value int32) {
	gl.Uniform1i(p.uniform(name), value)
}

func (p *Program) Float(name string, value float32) {
	gl.Uniform1f(p.uniform(name), value)
}

func (p *Program) Vec2(name string, x, y float32) {
	gl.Uniform2f(p.uniform(name), x, y)
}

func (p *Program) uniform(name string) int32 {
	return gl.GetUniformLocation(p.id, gl.Str(name+"\x00"))
}

func compileShader(source string, kind uint32) (uint32, error) {
	shader := gl.CreateShader(kind)

	sources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, sources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		return 0, fmt.Errorf("gpu: compile shader: %s", infoLog(shader, gl.GetShaderiv, gl.GetShaderInfoLog))
	}

	return shader, nil
}

func infoLog(object uint32, lengthOf func(uint32, uint32, *int32), read func(uint32, int32, *int32, *uint8)) string {
	var length int32
	lengthOf(object, gl.INFO_LOG_LENGTH, &length)

	message := strings.Repeat("\x00", int(length+1))
	read(object, length, nil, gl.Str(message))

	return strings.TrimRight(message, "\x00")
}
