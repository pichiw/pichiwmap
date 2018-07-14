package pmwgl

import (
	"errors"
	"syscall/js"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pichiw/pichiwmap"
)

var ErrNoWebGL = errors.New("no webgl found")

func NewTileRenderer(canvasID string) (*TileRenderer, error) {
	cache, err := lru.New(1000)
	if err != nil {
		return nil, err
	}

	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")

	gl := canvasEl.Call("getContext", "webgl")
	if gl == js.Undefined() {
		gl = canvasEl.Call("getContext", "experimental-webgl")
	}
	if gl == js.Undefined() {
		return nil, ErrNoWebGL
	}

	program, err := createProgram(gl, vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return nil, err
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

	t := &TileRenderer{
		gl:             gl,
		m4:             js.Global().Get("m4"),
		program:        program,
		position:       positionLocation,
		positionBuffer: positionBuffer,
		texcoord:       texcoordLocation,
		texcoordBuffer: texCoordBuffer,
		matrix:         matrixLocation,
		texture:        textureLocation,
		cache:          cache,
	}

	t.renderFrame = js.NewCallback(func(args []js.Value) { t.updateGl() })

	return t, nil
}

type TileRenderer struct {
	gl             js.Value
	m4             js.Value
	program        js.Value
	position       js.Value
	positionBuffer js.Value
	texcoord       js.Value
	texcoordBuffer js.Value
	matrix         js.Value
	texture        js.Value
	toDraw         []*drawInfo
	cache          *lru.Cache
	renderFrame    js.Callback
}

func (t *TileRenderer) Viewport() (width, height float64) {
	width = t.gl.Get("canvas").Get("width").Float()
	height = t.gl.Get("canvas").Get("height").Float()
	return
}

func (t *TileRenderer) updateGl() {
	cWidth, cHeight := t.Viewport()

	t.gl.Call("viewport", 0, 0, cWidth, cHeight)

	t.gl.Call("clearColor", 0, 0, 0, 0)
	t.gl.Call("clear", t.gl.Get("COLOR_BUFFER_BIT"))

	centreX := cWidth / 2
	centreY := cHeight / 2
	for _, td := range t.toDraw {
		t.drawImage(
			td.Texture,
			centreX-float64(td.DX),
			centreY-float64(td.DY),
		)
	}

}

func (t *TileRenderer) RenderTiles(tiles []*pichiwmap.Tile) {
	t.toDraw = nil

	for _, tile := range tiles {
		currentTile := tile
		u := currentTile.URL.String()
		v, ok := t.cache.Get(u)
		if ok {
			txi := v.(*textureInfo)
			t.toDraw = append(t.toDraw, &drawInfo{
				Texture: txi,
				DX:      currentTile.DX,
				DY:      currentTile.DY,
			})
		} else {
			txi := t.loadImage(currentTile.URL.String(), func(txi *textureInfo) {
				t.toDraw = append(t.toDraw, &drawInfo{
					Texture: txi,
					DX:      currentTile.DX,
					DY:      currentTile.DY,
				})
				t.cache.Add(u, txi)
				js.Global().Call("requestAnimationFrame", t.renderFrame)
			})
			t.cache.Add(u, txi)
		}
	}
	js.Global().Call("requestAnimationFrame", t.renderFrame)
}

func (t *TileRenderer) drawImage(tex *textureInfo, dstX, dstY float64) {
	cwidth, cheight := t.Viewport()

	t.gl.Call("bindTexture", t.gl.Get("TEXTURE_2D"), tex.Texture)

	t.gl.Call("useProgram", t.program)

	t.gl.Call("bindBuffer", t.gl.Get("ARRAY_BUFFER"), t.positionBuffer)
	t.gl.Call("enableVertexAttribArray", t.position)
	t.gl.Call("vertexAttribPointer", t.position, 2, t.gl.Get("FLOAT"), false, 0, 0)
	t.gl.Call("bindBuffer", t.gl.Get("ARRAY_BUFFER"), t.texcoordBuffer)
	t.gl.Call("enableVertexAttribArray", t.texcoord)
	t.gl.Call("vertexAttribPointer", t.texcoord, 2, t.gl.Get("FLOAT"), false, 0, 0)

	var matrix = t.m4.Call("orthographic", 0, cwidth, cheight, 0, -1, 1)
	matrix = t.m4.Call("translate", matrix, dstX, dstY, 0)
	matrix = t.m4.Call("scale", matrix, tex.Width, tex.Height, 1)

	t.gl.Call("uniformMatrix4fv", t.matrix, false, matrix)
	t.gl.Call("uniform1i", t.texture, 0)
	t.gl.Call("drawArrays", t.gl.Get("TRIANGLES"), 0, 6)
}

type textureInfo struct {
	Width   int // we don't know the size until it loads
	Height  int
	Texture js.Value
}

func (t *TileRenderer) loadImage(url string, onLoad func(txi *textureInfo)) *textureInfo {
	tex := t.gl.Call("createTexture")
	t.gl.Call("bindTexture", t.gl.Get("TEXTURE_2D"), tex)
	t.gl.Call("texImage2D", t.gl.Get("TEXTURE_2D"), 0, t.gl.Get("RGBA"), 1, 1, 0, t.gl.Get("RGBA"), t.gl.Get("UNSIGNED_BYTE"), js.TypedArrayOf([]uint8{0, 0, 255, 255}))
	t.gl.Call("texParameteri", t.gl.Get("TEXTURE_2D"), t.gl.Get("TEXTURE_WRAP_S"), t.gl.Get("CLAMP_TO_EDGE"))
	t.gl.Call("texParameteri", t.gl.Get("TEXTURE_2D"), t.gl.Get("TEXTURE_WRAP_T"), t.gl.Get("CLAMP_TO_EDGE"))
	t.gl.Call("texParameteri", t.gl.Get("TEXTURE_2D"), t.gl.Get("TEXTURE_MIN_FILTER"), t.gl.Get("LINEAR"))

	txi := &textureInfo{
		Width:   1,
		Height:  1,
		Texture: tex,
	}

	img := js.Global().Get("Image").New()
	img.Call("addEventListener", "load", js.NewCallback(func(args []js.Value) {
		txi.Width = img.Get("width").Int()
		txi.Height = img.Get("height").Int()

		t.gl.Call("bindTexture", t.gl.Get("TEXTURE_2D"), txi.Texture)
		t.gl.Call("texImage2D", t.gl.Get("TEXTURE_2D"), 0, t.gl.Get("RGBA"), t.gl.Get("RGBA"), t.gl.Get("UNSIGNED_BYTE"), img)
		onLoad(txi)
	}))
	img.Set("crossOrigin", "")
	img.Set("src", url)
	return txi
}

type drawInfo struct {
	Texture *textureInfo
	DX      int
	DY      int
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
