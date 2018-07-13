package main

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"syscall/js"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pichiw/pichiwmap"
)

func main() {
	lat := 49.8951
	lon := -97.1384
	zoom := 15

	doc := js.Global().Get("document")

	zoomEl := doc.Call("getElementById", "zoom")
	latEl := doc.Call("getElementById", "latitude")
	lonEl := doc.Call("getElementById", "longitude")
	buttonEl := doc.Call("getElementById", "updatePosition")

	updateControls := func() {
		zoomEl.Set("value", strconv.Itoa(zoom))
		latEl.Set("value", strconv.FormatFloat(lat, 'f', 6, 64))
		lonEl.Set("value", strconv.FormatFloat(lon, 'f', 6, 64))
	}

	canvasEl := doc.Call("getElementById", "mycanvas")
	width := doc.Get("body").Get("clientWidth").Int()
	height := doc.Get("body").Get("clientHeight").Int()
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	gl := canvasEl.Call("getContext", "webgl")
	if gl == js.Undefined() {
		gl = canvasEl.Call("getContext", "experimental-webgl")
	}

	cache, err := lru.New(1000)
	if err != nil {
		panic(err)
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

		tiles := pichiwmap.TilesFromCenter(baseURL, zoom, lat, lon, int(cWidth), int(cHeight))

		for _, t := range tiles {
			currentTile := t
			u := currentTile.URL.String()
			v, ok := cache.Get(u)
			if ok {
				txi := v.(*textureInfo)
				toDraw = append(toDraw, &drawInfo{
					Texture: txi,
					DX:      currentTile.DX,
					DY:      currentTile.DY,
				})
			} else {
				txi := loadImage(gl, currentTile.URL.String(), func(txi *textureInfo) {
					toDraw = append(toDraw, &drawInfo{
						Texture: txi,
						DX:      currentTile.DX,
						DY:      currentTile.DY,
					})
					cache.Add(u, txi)
					js.Global().Call("requestAnimationFrame", renderFrame)
				})
				cache.Add(u, txi)
			}
		}
		js.Global().Call("requestAnimationFrame", renderFrame)
		updateControls()
	}

	tlat := lat
	tlon := lon
	step := 0.0005
	var anim func()
	anim = func() {
		if lat == tlat && lon == tlon {
			return
		}

		if tlat > lat {
			lat += step
			if lat > tlat {
				lat = tlat
			}
		} else {
			lat -= step
			if lat < tlat {
				lat = tlat
			}
		}

		if tlon > lon {
			lon += step
			if lon > tlon {
				lon = tlon
			}
		} else {
			lon -= step
			if lon < tlon {
				lon = tlon
			}
		}

		update()
		time.Sleep(10 * time.Millisecond)
		go anim()
	}

	var (
		mouseStartX   int
		mouseStartY   int
		mouseStartLat float64
		mouseStartLon float64
		mouseDown     = false
	)

	canvasEl.Set("onmousedown", js.NewCallback(func(v []js.Value) {
		mouseStartX = v[0].Get("clientX").Int()
		mouseStartY = v[0].Get("clientY").Int()
		mouseStartLat = lat
		mouseStartLon = lon
		mouseDown = true
	}))

	canvasEl.Set("onmouseup", js.NewCallback(func(v []js.Value) {
		mouseDown = false
	}))

	canvasEl.Set("onmousemove", js.NewCallback(func(v []js.Value) {
		if !mouseDown {
			return
		}

		dx := mouseStartX - v[0].Get("clientX").Int()
		dy := mouseStartY - v[0].Get("clientY").Int()

		lat, lon = pichiwmap.Move(zoom, mouseStartLat, mouseStartLon, dx, dy)
		update()
	}))

	doc.Set("onkeyup", js.NewCallback(func(v []js.Value) {
		tlat = lat
		tlon = lon
	}))

	doc.Set("onkeydown", js.NewCallback(func(v []js.Value) {
		e := v[0]
		if e == js.Undefined() {
			e = js.Global().Get("window").Get("event")
		}

		if e == js.Undefined() {
			return
		}

		switch e.Get("keyCode").Int() {
		case 38: // up
			tlat += 0.001
		case 40: // down
			tlat -= 0.001
		case 37: // left
			tlon -= 0.001
		case 39: // right
			tlon += 0.001
		default:
			return
		}

		go anim()
	}))

	buttonEl.Set("onclick", js.NewCallback(func(v []js.Value) {
		v[0].Call("preventDefault")
		lastLat := lat
		lastLon := lon
		lastZoom := zoom

		var err error
		zoom, err = strconv.Atoi(zoomEl.Get("value").String())
		if err != nil {
			zoom = lastZoom
			updateControls()
			js.Global().Call("alert", "Invalid zoom (must be between 1 and 18)")
			return
		}

		lat, err = strconv.ParseFloat(latEl.Get("value").String(), 64)
		if err != nil {
			lat = lastLat
			updateControls()
			js.Global().Call("alert", "Invalid Lon Value")
			return
		}
		lon, err = strconv.ParseFloat(lonEl.Get("value").String(), 64)
		if err != nil {
			lon = lastLon
			updateControls()
			js.Global().Call("alert", "Invalid Lat Value")
			return
		}

		update()
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
