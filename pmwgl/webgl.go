package pmwgl

import (
	"fmt"
	"syscall/js"
)

func NewWebGL(canvasEl js.Value) (*WebGL, error) {
	gl := canvasEl.Call("getContext", "webgl")
	if gl == js.Undefined() {
		gl = canvasEl.Call("getContext", "experimental-webgl")
	}
	if gl == js.Undefined() {
		return nil, ErrNoWebGL
	}

	return &WebGL{
		gl:                 gl,
		m4:                 js.Global().Get("m4"),
		COMPILE_STATUS:     gl.Get("COMPILE_STATUS"),
		LINK_STATUS:        gl.Get("LINK_STATUS"),
		VERTEX_SHADER:      gl.Get("VERTEX_SHADER"),
		FRAGMENT_SHADER:    gl.Get("FRAGMENT_SHADER"),
		ARRAY_BUFFER:       gl.Get("ARRAY_BUFFER"),
		STATIC_DRAW:        gl.Get("STATIC_DRAW"),
		COLOR_BUFFER_BIT:   gl.Get("COLOR_BUFFER_BIT"),
		TEXTURE_2D:         gl.Get("TEXTURE_2D"),
		FLOAT:              gl.Get("FLOAT"),
		TRIANGLES:          gl.Get("TRIANGLES"),
		RGBA:               gl.Get("RGBA"),
		TEXTURE_WRAP_S:     gl.Get("TEXTURE_WRAP_S"),
		TEXTURE_WRAP_T:     gl.Get("TEXTURE_WRAP_T"),
		TEXTURE_MIN_FILTER: gl.Get("TEXTURE_MIN_FILTER"),
		CLAMP_TO_EDGE:      gl.Get("CLAMP_TO_EDGE"),
		LINEAR:             gl.Get("LINEAR"),
		UNSIGNED_BYTE:      gl.Get("UNSIGNED_BYTE"),
	}, nil
}

type WebGL struct {
	gl js.Value
	m4 js.Value

	COMPILE_STATUS     js.Value
	LINK_STATUS        js.Value
	VERTEX_SHADER      js.Value
	FRAGMENT_SHADER    js.Value
	ARRAY_BUFFER       js.Value
	STATIC_DRAW        js.Value
	COLOR_BUFFER_BIT   js.Value
	TEXTURE_2D         js.Value
	FLOAT              js.Value
	TRIANGLES          js.Value
	RGBA               js.Value
	TEXTURE_WRAP_S     js.Value
	TEXTURE_WRAP_T     js.Value
	TEXTURE_MIN_FILTER js.Value
	CLAMP_TO_EDGE      js.Value
	UNSIGNED_BYTE      js.Value
	LINEAR             js.Value
}

func (w *WebGL) CreateShader(shaderType js.Value) js.Value {
	return w.gl.Call("createShader", shaderType)
}

func (w *WebGL) ShaderSource(shader js.Value, src string) {
	w.gl.Call("shaderSource", shader, src)
}

func (w *WebGL) CompileShader(shader js.Value) {
	w.gl.Call("compileShader", shader)
}

func (w *WebGL) GetShaderParameter(shader, param js.Value) js.Value {
	return w.gl.Call("getShaderParameter", shader, param)
}

func (w *WebGL) CreateCompiledShader(shaderType js.Value, shaderSrc string) (js.Value, error) {
	var shader = w.CreateShader(shaderType)
	w.ShaderSource(shader, shaderSrc)
	w.CompileShader(shader)

	if !w.GetShaderParameter(shader, w.COMPILE_STATUS).Bool() {
		return js.Undefined(), fmt.Errorf("could not compile shader: %v", w.GetShaderInfoLog(shader))
	}
	return shader, nil
}

func (w *WebGL) CreateProgram() js.Value {
	return w.gl.Call("createProgram")
}

func (w *WebGL) AttachShader(program, shader js.Value) {
	w.gl.Call("attachShader", program, shader)
}

func (w *WebGL) LinkProgram(program js.Value) {
	w.gl.Call("linkProgram", program)
}

func (w *WebGL) GetProgramParameter(program, param js.Value) js.Value {
	return w.gl.Call("getProgramParameter", program, param)
}

func (w *WebGL) GetShaderInfoLog(shader js.Value) string {
	return w.gl.Call("getShaderInfoLog", shader).String()
}

func (w *WebGL) GetProgramInfoLog(program js.Value) string {
	return w.gl.Call("getProgramInfoLog", program).String()
}

func (w *WebGL) GetAttribLocation(program js.Value, attrib string) js.Value {
	return w.gl.Call("getAttribLocation", program, attrib)
}

func (w *WebGL) GetUniformLocation(program js.Value, uniform string) js.Value {
	return w.gl.Call("getUniformLocation", program, uniform)
}

func (w *WebGL) CreateBuffer() js.Value {
	return w.gl.Call("createBuffer")
}

func (w *WebGL) BindBuffer(t, buffer js.Value) {
	w.gl.Call("bindBuffer", t, buffer)
}

func (w *WebGL) BufferData(t js.Value, data js.TypedArray, drawType js.Value) {
	w.gl.Call("bufferData", t, data, drawType)
}

func (w *WebGL) Viewport(top, left, width, height float64) {
	w.gl.Call("viewport", top, left, width, height)
}

func (w *WebGL) ClearColor(r, g, b, a float64) {
	w.gl.Call("clearColor", r, g, b, a)
}

func (w *WebGL) Clear(bufferBit js.Value) {
	w.gl.Call("clear", bufferBit)
}

func (w *WebGL) Canvas() js.Value {
	return w.gl.Get("canvas")
}

func (w *WebGL) BindTexture(typ, texture js.Value) {
	w.gl.Call("bindTexture", typ, texture)
}
func (w *WebGL) UseProgram(program js.Value) {
	w.gl.Call("useProgram", program)
}
func (w *WebGL) EnableVertexAttribArray(position js.Value) {
	w.gl.Call("enableVertexAttribArray", position)
}
func (w *WebGL) VertexAttribPointer(index js.Value, size int, typ js.Value, normalized bool, stride, offset int) {
	w.gl.Call("vertexAttribPointer", index, size, typ, normalized, stride, offset)
}
func (w *WebGL) Orthographic(left, right, bottom, top, near, far float64) js.Value {
	return w.m4.Call("orthographic", left, right, bottom, top, near, far)
}
func (w *WebGL) Translate(matrix js.Value, dstX, dstY, dstZ float64) js.Value {
	return w.m4.Call("translate", matrix, dstX, dstY, dstZ)
}
func (w *WebGL) Scale(matrix js.Value, width, height, depth float64) js.Value {
	return w.m4.Call("scale", matrix, width, height, depth)
}
func (w *WebGL) UniformMatrix4fv(location js.Value, transpose bool, value js.Value) {
	w.gl.Call("uniformMatrix4fv", location, transpose, value)
}
func (w *WebGL) Uniform1i(location js.Value, v0 float64) {
	w.gl.Call("uniform1i", location, v0)
}
func (w *WebGL) DrawArrays(mode js.Value, first, count int) {
	w.gl.Call("drawArrays", mode, first, count)
}

func (w *WebGL) CreateTexture() js.Value {
	return w.gl.Call("createTexture")
}
func (w *WebGL) TexImage2DColor(target js.Value, level float64, internalformat js.Value, width, height, border float64, format js.Value, typ js.Value, source js.TypedArray) {
	w.gl.Call("texImage2D", target, level, internalformat, width, height, border, format, typ, source)
}

func (w *WebGL) TexImage2DData(target js.Value, level float64, internalformat, format js.Value, typ js.Value, source js.Value) {
	w.gl.Call("texImage2D", target, level, internalformat, format, typ, source)
}

func (w *WebGL) TexParameteri(target, name, param js.Value) {
	w.gl.Call("texParameteri", target, name, param)
}

func (w *WebGL) CreateProgramFromShaders(vertexShader, fragmentShader js.Value) (js.Value, error) {
	var program = w.CreateProgram()
	w.AttachShader(program, vertexShader)
	w.AttachShader(program, fragmentShader)
	w.LinkProgram(program)
	if !w.GetProgramParameter(program, w.LINK_STATUS).Bool() {
		return js.Undefined(), fmt.Errorf("could not link program: %v", w.GetProgramInfoLog(program))
	}

	return program, nil
}

func (w *WebGL) CreateProgramFromSource(vertex, fragment string) (js.Value, error) {
	vertexShader, err := w.CreateCompiledShader(w.VERTEX_SHADER, vertex)
	if err != nil {
		return js.Undefined(), err
	}
	fragShader, err := w.CreateCompiledShader(w.FRAGMENT_SHADER, fragment)
	if err != nil {
		return js.Undefined(), err
	}

	return w.CreateProgramFromShaders(vertexShader, fragShader)
}
