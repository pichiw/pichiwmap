package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"syscall/js"

	"github.com/pichiw/pichiwmap"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	lat := 49.8951
	lon := -97.1384

	doc := js.Global().Get("document")

	latEl := doc.Call("getElementById", "latitude")
	lonEl := doc.Call("getElementById", "longitude")
	buttonEl := doc.Call("getElementById", "updatePosition")

	latEl.Set("value", strconv.FormatFloat(lat, 'f', 6, 64))
	lonEl.Set("value", strconv.FormatFloat(lon, 'f', 6, 64))

	canvasEl := doc.Call("getElementById", "mycanvas")
	width := doc.Get("body").Get("clientWidth").Int()
	height := doc.Get("body").Get("clientHeight").Int()
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	gl := canvasEl.Call("getContext", "webgl")
	if gl == js.Undefined() {
		gl = canvasEl.Call("getContext", "experimental-webgl")
	}
	// once again
	if gl == js.Undefined() {
		js.Global().Call("alert", "browser might not support webgl")
		return
	}

	program, err := createProgram(gl, vertexShaderSource, fragmentShaderSource)
	if err != nil {
		panic(err)
	}

	positionLocation := gl.Call("getAttribLocation", program, "a_position")
	texcoordLocation := gl.Call("getAttribLocation", program, "a_texcoord")

	matrixLocation := gl.Call("getUniformLocation", program, "u_matrix")
	textureLocation := gl.Call("getUniformLocation", program, "u_texture")

	positionBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), positionBuffer)
	positions := js.TypedArrayOf([]float32{
		0, 0,
		0, 1,
		1, 0,
		1, 0,
		0, 1,
		1, 1,
	})
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), positions, gl.Get("STATIC_DRAW"))

	texCoordBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), texCoordBuffer)
	texcoords := js.TypedArrayOf([]float32{
		0, 0,
		0, 1,
		1, 0,
		1, 0,
		0, 1,
		1, 1,
	})
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), texcoords, gl.Get("STATIC_DRAW"))

	cWidth := gl.Get("canvas").Get("width").Float()
	cHeight := gl.Get("canvas").Get("height").Float()

	gl.Call("viewport", 0, 0, cWidth, cHeight)

	var renderFrame js.Callback

	baseURL, err := url.Parse("https://a.tile.openstreetmap.org")
	if err != nil {
		panic(err)
	}

	var toDraw []*drawInfo
	update := func() {
		toDraw = nil

		tiles := pichiwmap.TilesFromCenter(baseURL, 15, lat, lon, int(cWidth), int(cHeight))

		for _, t := range tiles {
			currentTile := t
			loadImage(gl, currentTile.URL.String(), func(txi *textureInfo) {
				toDraw = append(toDraw, &drawInfo{
					Texture: txi,
					DX:      currentTile.DX,
					DY:      currentTile.DY,
				})
				js.Global().Call("requestAnimationFrame", renderFrame)
			})
		}
	}

	buttonEl.Set("onclick", js.NewCallback(func(v []js.Value) {
		v[0].Call("preventDefault")
		lastLat := lat
		lastLon := lon

		valid := true
		var err error
		lat, err = strconv.ParseFloat(latEl.Get("value").String(), 64)
		if err != nil {
			lat = lastLat
			latEl.Set("value", strconv.FormatFloat(lat, 'f', 6, 64))
			js.Global().Call("alert", "Invalid Lon Value")
			valid = false
		}
		lon, err = strconv.ParseFloat(lonEl.Get("value").String(), 64)
		if err != nil {
			lon = lastLon
			lonEl.Set("value", strconv.FormatFloat(lon, 'f', 6, 64))
			js.Global().Call("alert", "Invalid Lat Value")
			valid = false
		}
		if valid {
			update()
		}
	}))

	renderFrame = js.NewCallback(func(args []js.Value) {
		gl.Call("clearColor", 0, 0, 0, 0)
		gl.Call("clear", gl.Get("COLOR_BUFFER_BIT"))

		centreX := cWidth / 2
		centreY := cHeight / 2
		for _, td := range toDraw {
			drawImage(
				gl,
				js.Global().Get("m4"),
				program,
				positionBuffer,
				positionLocation,
				texCoordBuffer,
				texcoordLocation,
				matrixLocation,
				textureLocation,
				td.Texture,
				centreX-float64(td.DX),
				centreY-float64(td.DY),
			)
		}
	})

	update()
	// Start running
	js.Global().Call("requestAnimationFrame", renderFrame)
	defer renderFrame.Release()

	c := make(chan struct{}, 0)

	<-c
}

type drawInfo struct {
	Texture *textureInfo
	DX      int
	DY      int
}

func drawImage(
	gl,
	m4,
	program,
	positionBuffer,
	positionLocation,
	texcoordBuffer,
	texcoordLocation,
	matrixLocation,
	textureLocation js.Value,
	tex *textureInfo,
	dstX,
	dstY float64,
) {
	gl.Call("bindTexture", gl.Get("TEXTURE_2D"), tex.Texture)

	// Tell WebGL to use our shader program pair
	gl.Call("useProgram", program)

	// Setup the attributes to pull data from our buffers
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), positionBuffer)
	gl.Call("enableVertexAttribArray", positionLocation)
	gl.Call("vertexAttribPointer", positionLocation, 2, gl.Get("FLOAT"), false, 0, 0)
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), texcoordBuffer)
	gl.Call("enableVertexAttribArray", texcoordLocation)
	gl.Call("vertexAttribPointer", texcoordLocation, 2, gl.Get("FLOAT"), false, 0, 0)

	// this matirx will convert from pixels to clip space
	var matrix = m4.Call("orthographic", 0, gl.Get("canvas").Get("width"), gl.Get("canvas").Get("height"), 0, -1, 1)

	// this matrix will translate our quad to dstX, dstY
	matrix = m4.Call("translate", matrix, dstX, dstY, 0)

	// this matrix will scale our 1 unit quad
	// from 1 unit to texWidth, texHeight units
	matrix = m4.Call("scale", matrix, tex.Width, tex.Height, 1)

	// Set the matrix.
	gl.Call("uniformMatrix4fv", matrixLocation, false, matrix)

	// Tell the shader to get the texture from texture unit 0
	gl.Call("uniform1i", textureLocation, 0)

	// draw the quad (2 triangles, 6 vertices)
	gl.Call("drawArrays", gl.Get("TRIANGLES"), 0, 6)
}

type textureInfo struct {
	Width   int // we don't know the size until it loads
	Height  int
	Texture js.Value
}

// https://a.tile.openstreetmap.org/15/5227/11225.png
func loadImage(
	gl js.Value,
	url string,
	onLoad func(txi *textureInfo),
) *textureInfo {
	tex := gl.Call("createTexture")
	gl.Call("bindTexture", gl.Get("TEXTURE_2D"), tex)
	gl.Call("texImage2D", gl.Get("TEXTURE_2D"), 0, gl.Get("RGBA"), 1, 1, 0, gl.Get("RGBA"), gl.Get("UNSIGNED_BYTE"), js.TypedArrayOf([]uint8{0, 0, 255, 255}))
	// let's assume all images are not a power of 2
	gl.Call("texParameteri", gl.Get("TEXTURE_2D"), gl.Get("TEXTURE_WRAP_S"), gl.Get("CLAMP_TO_EDGE"))
	gl.Call("texParameteri", gl.Get("TEXTURE_2D"), gl.Get("TEXTURE_WRAP_T"), gl.Get("CLAMP_TO_EDGE"))
	gl.Call("texParameteri", gl.Get("TEXTURE_2D"), gl.Get("TEXTURE_MIN_FILTER"), gl.Get("LINEAR"))

	txi := &textureInfo{
		Width:   1,
		Height:  1,
		Texture: tex,
	}
	img := js.Global().Get("Image").New()
	img.Call("addEventListener", "load", js.NewCallback(func(args []js.Value) {
		txi.Width = img.Get("width").Int()
		txi.Height = img.Get("height").Int()

		gl.Call("bindTexture", gl.Get("TEXTURE_2D"), txi.Texture)
		gl.Call("texImage2D", gl.Get("TEXTURE_2D"), 0, gl.Get("RGBA"), gl.Get("RGBA"), gl.Get("UNSIGNED_BYTE"), img)
		onLoad(txi)
	}))
	img.Set("crossOrigin", "")
	img.Set("src", url)
	return txi
}

const vertexShaderSource = `
attribute vec4 a_position;
attribute vec2 a_texcoord;
 
uniform mat4 u_matrix;
 
varying vec2 v_texcoord;
 
void main() {
   gl_Position = u_matrix * a_position;
   v_texcoord = a_texcoord;
}
`

const fragmentShaderSource = `
precision mediump float;
 
varying vec2 v_texcoord;
 
uniform sampler2D u_texture;
 
void main() {
   gl_FragColor = texture2D(u_texture, v_texcoord);
}
`

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
	vertexShader, err := compileShader(gl, gl.Get("VERTEX_SHADER"), vertexShaderSource)
	if err != nil {
		return js.Undefined(), err
	}
	fragShader, err := compileShader(gl, gl.Get("FRAGMENT_SHADER"), fragmentShaderSource)
	if err != nil {
		return js.Undefined(), err
	}

	return linkProgram(gl, vertexShader, fragShader)
}
