package pmwgl

import (
	"errors"
	"fmt"
	"syscall/js"
)

// ErrNoWebGL is returned from NewTileRenderer when webgl isn't found
var ErrNoWebGL = errors.New("no webgl found")

// NewWebGL creates a new webgl wrapper
func NewWebGL(canvasEl js.Value) (*WebGL, error) {
	gl := canvasEl.Call("getContext", "webgl")
	if gl == js.Undefined() {
		gl = canvasEl.Call("getContext", "experimental-webgl")
	}
	if gl == js.Undefined() {
		return nil, ErrNoWebGL
	}

	return &WebGL{
		gl:               gl,
		CompileStatus:    gl.Get("COMPILE_STATUS"),
		LinkStatus:       gl.Get("LINK_STATUS"),
		VertexShader:     gl.Get("VERTEX_SHADER"),
		FragmentShader:   gl.Get("FRAGMENT_SHADER"),
		ArrayBuffer:      gl.Get("ARRAY_BUFFER"),
		StaticDraw:       gl.Get("STATIC_DRAW"),
		ColorBufferBit:   gl.Get("COLOR_BUFFER_BIT"),
		Texture2D:        gl.Get("TEXTURE_2D"),
		Float:            gl.Get("FLOAT"),
		Triangles:        gl.Get("TRIANGLES"),
		RGBA:             gl.Get("RGBA"),
		TextureWrapS:     gl.Get("TEXTURE_WRAP_S"),
		TextureWrapT:     gl.Get("TEXTURE_WRAP_T"),
		TextureMinFilter: gl.Get("TEXTURE_MIN_FILTER"),
		ClampToEdge:      gl.Get("CLAMP_TO_EDGE"),
		Linear:           gl.Get("LINEAR"),
		UnsignedByte:     gl.Get("UNSIGNED_BYTE"),
	}, nil
}

// WebGL wrapper
type WebGL struct {
	gl js.Value

	CompileStatus    js.Value
	LinkStatus       js.Value
	VertexShader     js.Value
	FragmentShader   js.Value
	ArrayBuffer      js.Value
	StaticDraw       js.Value
	ColorBufferBit   js.Value
	Texture2D        js.Value
	Float            js.Value
	Triangles        js.Value
	RGBA             js.Value
	TextureWrapS     js.Value
	TextureWrapT     js.Value
	TextureMinFilter js.Value
	ClampToEdge      js.Value
	UnsignedByte     js.Value
	Linear           js.Value
}

// CreateShader https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/createShader
// WebGLShader gl.createShader(type);
func (w *WebGL) CreateShader(shaderType js.Value) js.Value {
	return w.gl.Call("createShader", shaderType)
}

// ShaderSource https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/shaderSource
// void gl.shaderSource(shader, source);
func (w *WebGL) ShaderSource(shader js.Value, src string) {
	w.gl.Call("shaderSource", shader, src)
}

// CompileShader https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/compileShader
// void gl.compileShader(shader);
func (w *WebGL) CompileShader(shader js.Value) {
	w.gl.Call("compileShader", shader)
}

// GetShaderParameter https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/getShaderParameter
// any gl.getShaderParameter(shader, pname);
func (w *WebGL) GetShaderParameter(shader, param js.Value) js.Value {
	return w.gl.Call("getShaderParameter", shader, param)
}

// CreateCompiledShader creates and compilesa shader from a source string
func (w *WebGL) CreateCompiledShader(shaderType js.Value, shaderSrc string) (js.Value, error) {
	var shader = w.CreateShader(shaderType)
	w.ShaderSource(shader, shaderSrc)
	w.CompileShader(shader)

	if !w.GetShaderParameter(shader, w.CompileStatus).Bool() {
		return js.Undefined(), fmt.Errorf("could not compile shader: %v", w.GetShaderInfoLog(shader))
	}
	return shader, nil
}

// CreateProgram https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/createProgram
// WebGLProgram gl.createProgram();
func (w *WebGL) CreateProgram() js.Value {
	return w.gl.Call("createProgram")
}

// AttachShader https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/attachShader
// void gl.attachShader(program, shader);
func (w *WebGL) AttachShader(program, shader js.Value) {
	w.gl.Call("attachShader", program, shader)
}

// LinkProgram https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/linkProgram
// void gl.linkProgram(program);
func (w *WebGL) LinkProgram(program js.Value) {
	w.gl.Call("linkProgram", program)
}

// GetProgramParameter https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/getProgramParameter
// any gl.getProgramParameter(program, pname);
func (w *WebGL) GetProgramParameter(program, param js.Value) js.Value {
	return w.gl.Call("getProgramParameter", program, param)
}

// GetShaderInfoLog https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/getShaderInfoLog
// gl.getShaderInfoLog(shader);
func (w *WebGL) GetShaderInfoLog(shader js.Value) string {
	return w.gl.Call("getShaderInfoLog", shader).String()
}

// GetProgramInfoLog https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/getProgramInfoLog
// gl.getProgramInfoLog(program);
func (w *WebGL) GetProgramInfoLog(program js.Value) string {
	return w.gl.Call("getProgramInfoLog", program).String()
}

// GetAttribLocation https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/getAttribLocation
// GLint gl.getAttribLocation(program, name);
func (w *WebGL) GetAttribLocation(program js.Value, attrib string) js.Value {
	return w.gl.Call("getAttribLocation", program, attrib)
}

// GetUniformLocation https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/getUniformLocation
// WebGLUniformLocation = WebGLRenderingContext.getUniformLocation(program, name);
func (w *WebGL) GetUniformLocation(program js.Value, uniform string) js.Value {
	return w.gl.Call("getUniformLocation", program, uniform)
}

// CreateBuffer https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/createBuffer
// WebGLBuffer gl.createBuffer();
func (w *WebGL) CreateBuffer() js.Value {
	return w.gl.Call("createBuffer")
}

// BindBuffer https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/bindBuffer
// void gl.bindBuffer(target, buffer);
func (w *WebGL) BindBuffer(t, buffer js.Value) {
	w.gl.Call("bindBuffer", t, buffer)
}

// BufferData https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/bufferData
// void gl.bufferData(target, ArrayBuffer? srcData, usage);
func (w *WebGL) BufferData(t js.Value, data js.TypedArray, drawType js.Value) {
	w.gl.Call("bufferData", t, data, drawType)
}

// Viewport https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/viewport
// void gl.viewport(x, y, width, height);
func (w *WebGL) Viewport(top, left, width, height float64) {
	w.gl.Call("viewport", top, left, width, height)
}

// ClearColor https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/clearColor
// void gl.clearColor(red, green, blue, alpha);
func (w *WebGL) ClearColor(r, g, b, a float64) {
	w.gl.Call("clearColor", r, g, b, a)
}

// Clear https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/clear
// void gl.clear(mask);
func (w *WebGL) Clear(bufferBit js.Value) {
	w.gl.Call("clear", bufferBit)
}

// Canvas https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/canvas
// gl.canvas;
func (w *WebGL) Canvas() js.Value {
	return w.gl.Get("canvas")
}

// BindTexture https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/bindTexture
// void gl.bindTexture(target, texture);
func (w *WebGL) BindTexture(typ, texture js.Value) {
	w.gl.Call("bindTexture", typ, texture)
}

// UseProgram https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/useProgram
// void gl.useProgram(program);
func (w *WebGL) UseProgram(program js.Value) {
	w.gl.Call("useProgram", program)
}

// EnableVertexAttribArray https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/enableVertexAttribArray
// void gl.enableVertexAttribArray(index);
func (w *WebGL) EnableVertexAttribArray(position js.Value) {
	w.gl.Call("enableVertexAttribArray", position)
}

// VertexAttribPointer https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/vertexAttribPointer
// void gl.vertexAttribPointer(index, size, type, normalized, stride, offset);
func (w *WebGL) VertexAttribPointer(index js.Value, size int, typ js.Value, normalized bool, stride, offset int) {
	w.gl.Call("vertexAttribPointer", index, size, typ, normalized, stride, offset)
}

// UniformMatrix4fv https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/uniformMatrix
// WebGLRenderingContext.uniformMatrix4fv(location, transpose, value);
func (w *WebGL) UniformMatrix4fv(location js.Value, transpose bool, value Matrix4) {
	w.gl.Call("uniformMatrix4fv", location, transpose, js.TypedArrayOf(value[:]))
}

// Uniform1i https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/uniform
// void gl.uniform1i(location, v0);
func (w *WebGL) Uniform1i(location js.Value, v0 float64) {
	w.gl.Call("uniform1i", location, v0)
}

// DrawArrays https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/drawArrays
// void gl.drawArrays(mode, first, count);
func (w *WebGL) DrawArrays(mode js.Value, first, count int) {
	w.gl.Call("drawArrays", mode, first, count)
}

// CreateTexture https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/createTexture
// WebGLTexture gl.createTexture();
func (w *WebGL) CreateTexture() js.Value {
	return w.gl.Call("createTexture")
}

// TexImage2DColor https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/texImage2D
// void gl.texImage2D(target, level, internalformat, width, height, border, format, type, ArrayBufferView? pixels);
func (w *WebGL) TexImage2DColor(target js.Value, level float64, internalformat js.Value, width, height, border float64, format js.Value, typ js.Value, source js.TypedArray) {
	w.gl.Call("texImage2D", target, level, internalformat, width, height, border, format, typ, source)
}

// TexImage2DData https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/texImage2D
// void gl.texImage2D(target, level, internalformat, format, type, ImageData? pixels);
func (w *WebGL) TexImage2DData(target js.Value, level float64, internalformat, format js.Value, typ js.Value, source js.Value) {
	w.gl.Call("texImage2D", target, level, internalformat, format, typ, source)
}

// TexParameteri https://developer.mozilla.org/en-US/docs/Web/API/WebGLRenderingContext/texParameter
// void gl.texParameteri(GLenum target, GLenum pname, GLint param);
func (w *WebGL) TexParameteri(target, name, param js.Value) {
	w.gl.Call("texParameteri", target, name, param)
}

// CreateProgramFromShaders Creates a program from a vertex and fragment shader
func (w *WebGL) CreateProgramFromShaders(vertexShader, fragmentShader js.Value) (js.Value, error) {
	var program = w.CreateProgram()
	w.AttachShader(program, vertexShader)
	w.AttachShader(program, fragmentShader)
	w.LinkProgram(program)
	if !w.GetProgramParameter(program, w.LinkStatus).Bool() {
		return js.Undefined(), fmt.Errorf("could not link program: %v", w.GetProgramInfoLog(program))
	}

	return program, nil
}

// CreateProgramFromSource creates a program from vertex and fragment source
func (w *WebGL) CreateProgramFromSource(vertex, fragment string) (js.Value, error) {
	vertexShader, err := w.CreateCompiledShader(w.VertexShader, vertex)
	if err != nil {
		return js.Undefined(), err
	}
	fragShader, err := w.CreateCompiledShader(w.FragmentShader, fragment)
	if err != nil {
		return js.Undefined(), err
	}

	return w.CreateProgramFromShaders(vertexShader, fragShader)
}
