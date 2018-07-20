package pmwgl

import (
	"math"
	"sync"
	"syscall/js"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pichiw/pichiwmap"
)

var unitSquare = []float32{
	0, 0,
	0, 1,
	1, 0,
	0, 1,
	1, 1,
	1, 0,
}

var tileSquare = []float32{
	0, 0, 0,
	0, pichiwmap.TileHeight, 0,
	pichiwmap.TileWidth, 0, 0,

	0, pichiwmap.TileHeight, 0,
	pichiwmap.TileWidth, pichiwmap.TileHeight, 0,
	pichiwmap.TileWidth, 0, 0,
}

var fov = 60 * math.Pi / 180

// NewTileRenderer creates a new tile renderer
func NewTileRenderer(canvasEl js.Value) (*TileRenderer, error) {
	cache, err := lru.New(150)
	if err != nil {
		return nil, err
	}

	gl, err := NewWebGL(canvasEl)
	if err != nil {
		return nil, err
	}

	program, err := gl.CreateProgramFromSource(tileVertexShaderSource, tileFragmentShaderSource)
	if err != nil {
		return nil, err
	}

	positionLocation := gl.GetAttribLocation(program, "a_position")
	texcoordLocation := gl.GetAttribLocation(program, "a_texcoord")

	matrixLocation := gl.GetUniformLocation(program, "u_matrix")
	textureLocation := gl.GetUniformLocation(program, "u_texture")

	markerProgram, err := gl.CreateProgramFromSource(markerVertexShaderSource, markerFragmentShaderSource)
	if err != nil {
		return nil, err
	}

	markerPosition := gl.GetAttribLocation(markerProgram, "a_position")
	markerMatrix := gl.GetUniformLocation(markerProgram, "u_matrix")

	// Marker vertex buffer
	markerBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ArrayBuffer, markerBuffer)
	gl.BufferData(gl.ArrayBuffer, js.TypedArrayOf(unitCircle), gl.StaticDraw)

	// Unit square vertex buffer
	squareBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ArrayBuffer, squareBuffer)
	gl.BufferData(gl.ArrayBuffer, js.TypedArrayOf(tileSquare), gl.StaticDraw)

	// Texture buffer
	texCoordBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ArrayBuffer, texCoordBuffer)
	gl.BufferData(gl.ArrayBuffer, js.TypedArrayOf(unitSquare), gl.StaticDraw)

	t := &TileRenderer{
		gl:             gl,
		program:        program,
		position:       positionLocation,
		squareBuffer:   squareBuffer,
		texcoord:       texcoordLocation,
		texcoordBuffer: texCoordBuffer,
		markerProgram:  markerProgram,
		markerPosition: markerPosition,
		markerMatrix:   markerMatrix,
		markerBuffer:   markerBuffer,
		matrix:         matrixLocation,
		texture:        textureLocation,
		cache:          cache,
	}

	t.renderFrame = js.NewCallback(func(args []js.Value) { t.updateGl() })

	return t, nil
}

// TileRenderer will render tiles onto a canvas using webgl
type TileRenderer struct {
	gl             *WebGL
	program        js.Value
	position       js.Value
	squareBuffer   js.Value
	texcoord       js.Value
	texcoordBuffer js.Value
	matrix         js.Value

	markerProgram  js.Value
	markerPosition js.Value
	markerMatrix   js.Value
	markerBuffer   js.Value
	markerTexture  js.Value

	texture     js.Value
	zoom        float64
	lat         float64
	lon         float64
	toDraw      []*drawInfo
	cache       *lru.Cache
	renderFrame js.Callback
}

// Viewport returns the current width and height of the tile renderer's viewport
func (t *TileRenderer) Viewport() (width, height float64) {
	width = t.gl.Canvas().Get("width").Float()
	height = t.gl.Canvas().Get("height").Float()
	return
}

var up = Coord{X: 0, Y: -1, Z: 0}

func (t *TileRenderer) updateGl() {
	cWidth, cHeight := t.Viewport()
	t.gl.Viewport(0, 0, cWidth, cHeight)

	//t.gl.Enable(t.gl.CullFace)
	t.gl.Enable(t.gl.DepthTest)

	t.gl.Clear(t.gl.ColorBufferBit | t.gl.DepthBufferBit)

	aspect := cWidth / cHeight

	projection := Perspective(float32(fov), float32(aspect), 1, 2000)

	x, y := pichiwmap.TileNum(int(t.zoom), t.lat, t.lon)

	x *= pichiwmap.TileWidth
	y *= pichiwmap.TileHeight

	cameraPosition := Identity().XRotate(math.Pi / 4).TransformVector(Coord{X: 0, Y: 0, Z: -500})

	cameraPosition.X += float32(x)
	cameraPosition.Y += float32(y)

	target := Coord{X: float32(x), Y: float32(y), Z: 0}

	camera := LookAt(cameraPosition, target, up)

	view := camera.Inverse()
	viewProjection := projection.Multiply(view)

	for _, td := range t.toDraw {
		x, y := pichiwmap.TileNum(int(t.zoom), td.DX, td.DY)

		t.drawImage(
			viewProjection,
			td.Texture,
			float32(x*pichiwmap.TileWidth),
			float32(y*pichiwmap.TileHeight),
		)
	}
}

// RenderTiles will render the given tiles at the current zoom level
func (t *TileRenderer) RenderTiles(zoom, lat, lon float64, tiles map[string]*pichiwmap.Tile) {
	t.zoom = zoom
	t.lat = lat
	t.lon = lon
	// Cancel any loads that are no longer necessary
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

		if tile.Zoom == int(zoom) {
			t.toDraw = append(t.toDraw, &drawInfo{
				Texture: txi,
				DX:      tile.Lat,
				DY:      tile.Lon,
			})
		}
	}
	t.requestAnimationFrame()
}

func (t *TileRenderer) imageLoadCallback(txi *textureInfo) {
	t.requestAnimationFrame()
}

func (t *TileRenderer) requestAnimationFrame() {
	js.Global().Call("requestAnimationFrame", t.renderFrame)
}

func (t *TileRenderer) drawImage(viewProjection Matrix4, tex *textureInfo, dstX, dstY float32) {
	t.gl.UseProgram(t.program)

	t.gl.EnableVertexAttribArray(t.position)
	t.gl.BindBuffer(t.gl.ArrayBuffer, t.squareBuffer)
	t.gl.VertexAttribPointer(t.position, 3, t.gl.Float, false, 0, 0)

	t.gl.EnableVertexAttribArray(t.texcoord)
	t.gl.BindBuffer(t.gl.ArrayBuffer, t.texcoordBuffer)
	t.gl.VertexAttribPointer(t.texcoord, 2, t.gl.Float, false, 0, 0)

	matrix := viewProjection.Translate(dstX, dstY, 0)

	t.gl.BindTexture(t.gl.Texture2D, tex.Texture)
	t.gl.Uniform1i(t.texture, 0)
	t.gl.UniformMatrix4fv(t.matrix, false, matrix)
	t.gl.DrawArrays(t.gl.Triangles, 0, 6)
}

// func (t *TileRenderer) drawMarker(dstX, dstY, scale float32) {
// 	cwidth, cheight := t.Viewport()

// 	t.gl.UseProgram(t.markerProgram)

// 	t.gl.BindBuffer(t.gl.ArrayBuffer, t.markerBuffer)
// 	t.gl.EnableVertexAttribArray(t.markerPosition)
// 	t.gl.VertexAttribPointer(t.markerPosition, 3, t.gl.Float, false, 0, 0)

// 	matrix := Orthographic(0, float32(cwidth), float32(cheight), 0, -1, 1)
// 	matrix = matrix.Translate(dstX, dstY, 0)

// 	matrix = matrix.Scale(20, 20, 20)

// 	t.gl.UniformMatrix4fv(t.markerMatrix, false, matrix)
// 	t.gl.DrawArrays(t.gl.TriangleFan, 0, unitCirclePoints)
// }

type textureInfo struct {
	m         sync.Mutex
	URL       string
	Width     int // we don't know the size until it loads
	Height    int
	Texture   js.Value
	Image     js.Value
	Loaded    bool
	Cancelled bool
}

func (t *textureInfo) Cancel() bool {
	t.m.Lock()
	defer t.m.Unlock()

	if t.Loaded || t.Cancelled {
		return false // Don't cancel if it's already loaded!
	}
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

func powerOfTwo(v int) bool {
	return (v & (v - 1)) == 0
}

func (t *TileRenderer) loadImage(url string, onLoad func(txi *textureInfo)) *textureInfo {
	tex := t.gl.CreateTexture()
	t.gl.BindTexture(t.gl.Texture2D, tex)
	t.gl.TexImage2DColor(t.gl.Texture2D, 0, t.gl.RGBA, pichiwmap.TileWidth, pichiwmap.TileHeight, 0, t.gl.RGBA, t.gl.UnsignedByte, blankTexture)
	t.gl.GenerateMipmap(t.gl.Texture2D)

	txi := &textureInfo{
		URL:     url,
		Width:   pichiwmap.TileWidth,
		Height:  pichiwmap.TileHeight,
		Texture: tex,
		Image:   js.Global().Get("Image").New(),
	}

	txi.Image.Call("addEventListener", "load", js.NewEventCallback(0, func(event js.Value) {
		txi.m.Lock()
		defer txi.m.Unlock()

		txi.Loaded = true

		txi.Width = txi.Image.Get("width").Int()
		txi.Height = txi.Image.Get("height").Int()

		t.gl.BindTexture(t.gl.Texture2D, txi.Texture)
		t.gl.TexImage2DData(t.gl.Texture2D, 0, t.gl.RGBA, t.gl.RGBA, t.gl.UnsignedByte, txi.Image)
		if powerOfTwo(txi.Width) && powerOfTwo(txi.Height) {
			t.gl.GenerateMipmap(t.gl.Texture2D)
		} else {
			t.gl.TexParameteri(t.gl.Texture2D, t.gl.TextureWrapS, t.gl.ClampToEdge)
			t.gl.TexParameteri(t.gl.Texture2D, t.gl.TextureWrapT, t.gl.ClampToEdge)
			t.gl.TexParameteri(t.gl.Texture2D, t.gl.TextureMinFilter, t.gl.Linear)
		}
		onLoad(txi)
	}))
	txi.Image.Set("crossOrigin", "")
	txi.Image.Set("src", url)
	return txi
}

type drawInfo struct {
	Texture *textureInfo
	DX      float64
	DY      float64
}

const tileVertexShaderSource = `
attribute vec4 a_position;
attribute vec2 a_texcoord;
 
uniform mat4 u_matrix;
 
varying vec2 v_texcoord;
 
void main() {
   gl_Position = u_matrix * a_position;
   v_texcoord = a_texcoord;
}
`

const tileFragmentShaderSource = `
precision mediump float;
 
varying vec2 v_texcoord;
 
uniform sampler2D u_texture;
 
void main() {
   gl_FragColor = texture2D(u_texture, v_texcoord);
}
`

const markerVertexShaderSource = `
attribute vec4 a_position;
uniform mat4 u_matrix;

void main() {
   gl_Position = u_matrix * a_position;
}
`

const markerFragmentShaderSource = `
precision mediump float;

void main() {
   gl_FragColor = vec4(1.0, 0, 0, 0.5);
}
`
