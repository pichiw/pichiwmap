package pmwgl

import (
	"fmt"

	"github.com/gowasm/gopherwasm/js"
)

// Render to framebuffer first, then framebuffer to screen
func compileShader(gl, shaderType js.Value, shaderSrc string) (js.Value, error) {
	var shader = gl.Call("createShader", shaderType)
	gl.Call("shaderSource", shader, shaderSrc)
	gl.Call("compileShader", shader)

	if !gl.Call("getShaderParameter", shader, gl.Get("COMPILE_STATUS")).Bool() {
		return js.Undefined(), fmt.Errorf("could not compile shader: %v", gl.Call("getShaderInfoLog", shader).String())
	}
	return shader, nil
}

func linkProgram(gl, vertexShader, fragmentShader js.Value) (js.Value, error) {
	var program = gl.Call("createProgram")
	gl.Call("attachShader", program, vertexShader)
	gl.Call("attachShader", program, fragmentShader)
	gl.Call("linkProgram", program)
	if !gl.Call("getProgramParameter", program, gl.Get("LINK_STATUS")).Bool() {
		return js.Undefined(), fmt.Errorf("could not link program: %v", gl.Call("getProgramInfoLog", program).String())
	}

	return program, nil
}

func createProgram(gl js.Value, vertex, fragment string) (js.Value, error) {
	vertexShader, err := compileShader(gl, gl.Get("VERTEX_SHADER"), vertex)
	if err != nil {
		return js.Undefined(), err
	}
	fragShader, err := compileShader(gl, gl.Get("FRAGMENT_SHADER"), fragment)
	if err != nil {
		return js.Undefined(), err
	}

	return linkProgram(gl, vertexShader, fragShader)
}
