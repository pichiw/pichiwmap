package pmwgl

import (
	"errors"
	"sync"
	"syscall/js"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pichiw/pichiwmap"
)

var ErrNoWebGL = errors.New("no webgl found")

func NewTileRenderer(canvasEl js.Value) (*TileRenderer, error) {
	cache, err := lru.New(150)
	if err != nil {
		return nil, err
	}

	gl, err := NewWebGL(canvasEl)
	if err != nil {
		return nil, err
	}

	program, err := gl.CreateProgramFromSource(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return nil, err
	}

	positionLocation := gl.GetAttribLocation(program, "a_position")
	texcoordLocation := gl.GetAttribLocation(program, "a_texcoord")

	matrixLocation := gl.GetUniformLocation(program, "u_matrix")
	textureLocation := gl.GetUniformLocation(program, "u_texture")

	positionBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, positionBuffer)
	positions := js.TypedArrayOf([]float32{
		0, 0,
		0, 1,
		1, 0,
		1, 0,
		0, 1,
		1, 1,
	})
	gl.BufferData(gl.ARRAY_BUFFER, positions, gl.STATIC_DRAW)

	texCoordBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, texCoordBuffer)
	texcoords := js.TypedArrayOf([]float32{
		0, 0,
		0, 1,
		1, 0,
		1, 0,
		0, 1,
		1, 1,
	})
	gl.BufferData(gl.ARRAY_BUFFER, texcoords, gl.STATIC_DRAW)

	t := &TileRenderer{
		gl:             gl,
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
	gl             *WebGL
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
	width = t.gl.Canvas().Get("width").Float()
	height = t.gl.Canvas().Get("height").Float()
	return
}

func (t *TileRenderer) updateGl() {
	cWidth, cHeight := t.Viewport()

	t.gl.Viewport(0, 0, cWidth, cHeight)

	t.gl.ClearColor(0, 0, 0, 0)
	t.gl.Clear(t.gl.COLOR_BUFFER_BIT)

	centreX := cWidth / 2
	centreY := cHeight / 2
	for _, td := range t.toDraw {
		t.drawImage(
			td.Texture,
			centreX-float64(td.DX),
			centreY-float64(td.DY),
			td.Scale,
		)
	}

}

func (t *TileRenderer) RenderTiles(tiles map[string]*pichiwmap.Tile) {
	for _, td := range t.toDraw {
		if _, ok := tiles[td.Texture.URL]; !ok {
			if td.Texture.Cancel() {
				t.cache.Remove(td.Texture.URL)
			}
		}
	}

	t.toDraw = nil

	for _, tile := range tiles {
		u := tile.URL.String()

		var txi *textureInfo
		v, ok := t.cache.Get(u)
		if ok {
			txi = v.(*textureInfo)
		} else {
			txi = t.loadImage(tile.URL.String(), t.imageLoadCallback)
			t.cache.Add(u, txi)
		}

		t.toDraw = append(t.toDraw, &drawInfo{
			Texture: txi,
			DX:      tile.DX,
			DY:      tile.DY,
			Scale:   tile.Scale,
		})
	}
	t.requestAnimationFrame()
}

func (t *TileRenderer) imageLoadCallback(txi *textureInfo) {
	t.requestAnimationFrame()
}

func (t *TileRenderer) requestAnimationFrame() {
	js.Global().Call("requestAnimationFrame", t.renderFrame)
}

func (t *TileRenderer) drawImage(tex *textureInfo, dstX, dstY, scale float64) {
	cwidth, cheight := t.Viewport()

	t.gl.BindTexture(t.gl.TEXTURE_2D, tex.Texture)
	t.gl.UseProgram(t.program)
	t.gl.BindBuffer(t.gl.ARRAY_BUFFER, t.positionBuffer)
	t.gl.EnableVertexAttribArray(t.position)
	t.gl.VertexAttribPointer(t.position, 2, t.gl.FLOAT, false, 0, 0)
	t.gl.BindBuffer(t.gl.ARRAY_BUFFER, t.texcoordBuffer)
	t.gl.EnableVertexAttribArray(t.texcoord)
	t.gl.VertexAttribPointer(t.texcoord, 2, t.gl.FLOAT, false, 0, 0)

	var matrix = t.gl.Orthographic(0, cwidth, cheight, 0, -1, 1)
	matrix = t.gl.Translate(matrix, dstX, dstY, 0)
	matrix = t.gl.Scale(matrix, float64(tex.Width)*scale, float64(tex.Height)*scale, 1)

	t.gl.UniformMatrix4fv(t.matrix, false, matrix)
	t.gl.Uniform1i(t.texture, 0)
	t.gl.DrawArrays(t.gl.TRIANGLES, 0, 6)
}

type textureInfo struct {
	m           sync.Mutex
	URL         string
	Width       int // we don't know the size until it loads
	Height      int
	Texture     js.Value
	Image       js.Value
	Loaded      bool
	Cancelled   bool
	Callback    js.Callback
	releaseOnce sync.Once
}

func (t *textureInfo) Release() {
	t.releaseOnce.Do(func() {
		//t.Callback.Release()
	})
}
func (t *textureInfo) Cancel() bool {
	t.m.Lock()
	defer t.m.Unlock()

	if t.Loaded || t.Cancelled {
		return false // Don't cancel if it's already loaded!
	}
	t.Release()
	t.Cancelled = true
	t.Image.Set("src", "")
	return true
}

var blankTexture js.TypedArray

func init() {
	bt := make([]uint8, pichiwmap.TileWidth*pichiwmap.TileHeight*4)

	for i := 0; i < len(bt); i += 4 {
		bt[i] = 0
		bt[i+1] = 0
		bt[i+2] = 0
		bt[i+3] = 30
	}

	blankTexture = js.TypedArrayOf(bt)
}

func (t *TileRenderer) loadImage(url string, onLoad func(txi *textureInfo)) *textureInfo {
	tex := t.gl.CreateTexture()
	t.gl.BindTexture(t.gl.TEXTURE_2D, tex)
	t.gl.TexImage2DColor(t.gl.TEXTURE_2D, 0, t.gl.RGBA, pichiwmap.TileWidth, pichiwmap.TileHeight, 0, t.gl.RGBA, t.gl.UNSIGNED_BYTE, blankTexture)
	t.gl.TexParameteri(t.gl.TEXTURE_2D, t.gl.TEXTURE_WRAP_S, t.gl.CLAMP_TO_EDGE)
	t.gl.TexParameteri(t.gl.TEXTURE_2D, t.gl.TEXTURE_WRAP_T, t.gl.CLAMP_TO_EDGE)
	t.gl.TexParameteri(t.gl.TEXTURE_2D, t.gl.TEXTURE_MIN_FILTER, t.gl.LINEAR)

	txi := &textureInfo{
		URL:     url,
		Width:   pichiwmap.TileWidth,
		Height:  pichiwmap.TileHeight,
		Texture: tex,
		Image:   js.Global().Get("Image").New(),
	}

	txi.Callback = js.NewEventCallback(0, func(event js.Value) {
		txi.m.Lock()
		defer txi.m.Unlock()

		txi.Release()
		txi.Loaded = true

		txi.Width = txi.Image.Get("width").Int()
		txi.Height = txi.Image.Get("height").Int()

		t.gl.BindTexture(t.gl.TEXTURE_2D, txi.Texture)
		t.gl.TexImage2DData(t.gl.TEXTURE_2D, 0, t.gl.RGBA, t.gl.RGBA, t.gl.UNSIGNED_BYTE, txi.Image)
		onLoad(txi)
	})

	txi.Image.Call("addEventListener", "load", txi.Callback)
	txi.Image.Set("crossOrigin", "")
	txi.Image.Set("src", url)
	return txi
}

type drawInfo struct {
	Texture *textureInfo
	DX      int
	DY      int
	Scale   float64
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
