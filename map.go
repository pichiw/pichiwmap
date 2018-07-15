package pichiwmap

import (
	"math"
	"time"

	"github.com/gowasm/gopherwasm/js"
)

type (
	// OnLatChanged is fired whent he latitude changes
	OnLatChanged func(lat float64)
	// OnLonChanged is fired when the longitude changes
	OnLonChanged func(lon float64)
	// OnZoomChanged is fired when the zoom changes
	OnZoomChanged func(zoom float64)
)

// MapEvents wraps all map events into a convenient type
type MapEvents struct {
	OnLatChanged  OnLatChanged
	OnLonChanged  OnLonChanged
	OnZoomChanged OnZoomChanged
}

func (e MapEvents) wrapEmpty() MapEvents {
	if e.OnLatChanged == nil {
		e.OnLatChanged = func(float64) {}
	}
	if e.OnLonChanged == nil {
		e.OnLonChanged = func(float64) {}
	}
	if e.OnZoomChanged == nil {
		e.OnZoomChanged = func(float64) {}
	}
	return e
}

// TileRenderer is anything that can render tiles
type TileRenderer interface {
	RenderTiles(zoom int, tiles map[string]*Tile)
}

// New creates a new map at the specified dib
func New(urlEr URLer, divEl js.Value, events MapEvents) (*Map, error) {
	doc := js.Global().Get("document")

	viewport := doc.Call("createElement", "canvas")

	onResize := func(event js.Value) {
		width := divEl.Get("offsetWidth").Int()
		height := divEl.Get("offsetHeight").Int()
		viewport.Set("width", width)
		viewport.Set("height", height)
	}

	onResize(js.Null())

	divEl.Call("appendChild", viewport)

	m := &Map{
		events:   events.wrapEmpty(),
		lat:      49.8951,
		lon:      -97.1384,
		zoom:     15,
		zoomStep: 0.1,
		step:     0.001,
		urlEr:    urlEr,
		viewport: viewport,
		maxZoom:  18,
		minZoom:  0,
	}

	window := js.Global().Get("window")
	window.Call("addEventListener", "resize", js.NewEventCallback(0, func(event js.Value) {
		onResize(event)
		m.Update(ZoomingZero)
	}))

	m.tlat = m.Lat()
	m.tlon = m.Lon()

	m.viewport.Call("addEventListener", "mousedown", js.NewEventCallback(js.PreventDefault, m.onMouseDown), false)
	m.viewport.Call("addEventListener", "mouseup", js.NewEventCallback(js.PreventDefault, m.onMouseUp), false)
	m.viewport.Call("addEventListener", "mousemove", js.NewEventCallback(js.PreventDefault, m.onMouseMove), false)

	m.viewport.Call("addEventListener", "touchstart", js.NewEventCallback(js.PreventDefault, m.onTouchStart), false)
	m.viewport.Call("addEventListener", "touchend", js.NewEventCallback(js.PreventDefault, m.onTouchEnd), false)
	m.viewport.Call("addEventListener", "touchmove", js.NewEventCallback(js.PreventDefault, m.onTouchMove), false)

	doc.Call("addEventListener", "keyup", js.NewEventCallback(js.PreventDefault, m.onKeyUp), false)
	doc.Call("addEventListener", "keydown", js.NewEventCallback(js.PreventDefault, m.onKeyDown), false)
	doc.Call("addEventListener", "wheel", js.NewEventCallback(js.PreventDefault, m.wheel), false)

	return m, nil
}

// Map represents a map
type Map struct {
	tileRenderers []TileRenderer

	events MapEvents

	doc      js.Value
	viewport js.Value

	urlEr          URLer
	zoom           float64
	zoomStep       float64
	lat            float64
	lon            float64
	tlat           float64
	tlon           float64
	step           float64
	pinchZoomStart float64
	pinchDelta     float64
	pinchDown      bool
	mouseStartX    int
	mouseStartY    int
	mouseStartLat  float64
	mouseStartLon  float64
	mouseDown      bool
	arrowDown      bool
	maxZoom        float64
	minZoom        float64
}

// Zoom returns the current zoom
func (m *Map) Zoom() float64 {
	return m.zoom
}

func (m *Map) setZoom(zoom float64) {
	if m.zoom == zoom {
		return
	}
	m.zoom = zoom
	m.events.OnZoomChanged(zoom)
}

// Lat returns the current latitude
func (m *Map) Lat() float64 {
	return m.lat
}

func (m *Map) setLat(lat float64) {
	if m.lat == lat {
		return
	}
	m.lat = lat
	m.events.OnLatChanged(lat)
}

// Lon returns the current longitude
func (m *Map) Lon() float64 {
	return m.lon
}

func (m *Map) setLon(lon float64) {
	if m.lon == lon {
		return
	}
	m.lon = lon
	m.events.OnLonChanged(lon)
}

// SetPosition sets the current zoom, latitude, and longitude of the map
func (m *Map) SetPosition(zoom, lat, lon float64) {
	if zoom < m.minZoom || zoom > m.maxZoom {
		return
	}
	if zoom == m.zoom && lat == m.lat && lon == m.lon {
		return
	}
	zooming := ZoomingZero
	if zoom > m.zoom {
		zooming = ZoomingIn
	} else if zoom < m.zoom {
		zooming = ZoomingOut
	}
	m.setZoom(zoom)
	m.setLat(lat)
	m.setLon(lon)
	m.Update(zooming)
}

// AddTileRenderers adds tile renderers to the map
func (m *Map) AddTileRenderers(tr ...TileRenderer) {
	m.tileRenderers = append(m.tileRenderers, tr...)
}

// Canvas returns the canvas the map is being drawn onto
func (m *Map) Canvas() js.Value {
	return m.viewport
}

func (m *Map) anim() {
	if m.Lat() == m.tlat && m.Lon() == m.tlon {
		return
	}

	var newLat float64
	var newLon float64

	if m.tlat > m.Lat() {
		newLat = m.Lat() + m.step
		if newLat > m.tlat {
			newLat = m.tlat
		}
	} else {
		newLat = m.Lat() - m.step
		if newLat < m.tlat {
			newLat = m.tlat
		}
	}

	if m.tlon > m.Lon() {
		newLon = m.Lon() + m.step
		if newLon > m.tlon {
			newLon = m.tlon
		}
	} else {
		newLon = m.Lon() - m.step
		if newLon < m.tlon {
			newLon = m.tlon
		}
	}

	m.SetPosition(m.zoom, newLat, newLon)
	time.Sleep(100 * time.Millisecond)
	go m.anim()
}

func (m *Map) wheel(event js.Value) {
	var delta float64
	if event.Get("deltaY").Int() < 0 {
		delta = m.zoomStep
	} else {
		delta = -m.zoomStep
	}

	m.SetPosition(m.zoom+delta, m.lat, m.lon)
}

func (m *Map) onKeyDown(event js.Value) {
	if event == js.Undefined() {
		event = js.Global().Get("window").Get("event")
	}

	if event == js.Undefined() {
		return
	}

	if !m.arrowDown {
		m.tlat = m.lat
		m.tlon = m.lon
	}
	switch event.Get("keyCode").Int() {
	case 38: // up
		m.tlat += 0.005
	case 40: // down
		m.tlat -= 0.005
	case 37: // left
		m.tlon -= 0.005
	case 39: // right
		m.tlon += 0.005
	default:
		return
	}

	m.arrowDown = true

	go m.anim()
}

func (m *Map) onKeyUp(event js.Value) {
	m.arrowDown = false
}

func pinchDelta(touches js.Value) float64 {
	if touches.Length() <= 1 {
		return 0
	}
	t1 := touches.Index(0)
	t2 := touches.Index(1)

	x1 := t1.Get("pageX").Float()
	y1 := t1.Get("pageY").Float()

	x2 := t2.Get("pageX").Float()
	y2 := t2.Get("pageY").Float()

	lx := math.Abs(x1 - x2)
	ly := math.Abs(y1 - y2)

	return math.Sqrt(math.Pow(lx, 2) + math.Pow(ly, 2))
}

func (m *Map) onTouchStart(event js.Value) {
	touches := event.Get("touches")
	if touches == js.Undefined() {
		return
	}

	if touches.Length() == 1 {
		m.onMouseDown(touches.Index(0))
		return
	}

	if touches.Length() != 2 {
		return
	}

	m.pinchZoomStart = m.zoom
	m.pinchDelta = pinchDelta(touches)
	m.pinchDown = true
}

func (m *Map) onTouchEnd(event js.Value) {
	m.pinchDown = false
}

func (m *Map) onTouchMove(event js.Value) {
	touches := event.Get("touches")
	if touches == js.Undefined() {
		return
	}

	if touches.Length() == 1 {
		m.onMouseMove(touches.Index(0))
		return
	}

	if !m.pinchDown || touches.Length() != 2 {
		return
	}

	npd := pinchDelta(touches)
	zoom := (m.pinchDelta - npd) / 100
	m.SetPosition(m.pinchZoomStart-zoom, m.Lat(), m.Lon())
}

func (m *Map) onMouseDown(event js.Value) {
	m.mouseStartX = event.Get("pageX").Int()
	m.mouseStartY = event.Get("pageY").Int()
	m.mouseStartLat = m.Lat()
	m.mouseStartLon = m.Lon()
	m.mouseDown = true
}

func (m *Map) onMouseUp(event js.Value) {
	m.mouseDown = false
	m.pinchDown = false
}

func (m *Map) onMouseMove(event js.Value) {
	if !m.mouseDown {
		return
	}

	dx := m.mouseStartX - event.Get("pageX").Int()
	dy := m.mouseStartY - event.Get("pageY").Int()

	lat, lon := Move(m.Zoom(), m.mouseStartLat, m.mouseStartLon, dx, dy)
	m.SetPosition(m.Zoom(), lat, lon)
}

func scale(zoom float64) float64 {
	iz := int(zoom)
	return 1 + (0.5 + (zoom - float64(iz)))
}

// TilesFromCenter gets the tiles required from the current centre point
func (m *Map) TilesFromCenter(zoom float64, canvasWidth, canvasHeight int) map[string]*Tile {
	cx, cy := TileNum(int(zoom), m.lat, m.lon)

	tx := int(cx)
	ty := int(cy)

	px := float64(tx) - cx
	py := float64(ty) - cy

	scale := scale(zoom)

	dx := -int(px * TileWidth * scale)
	dy := -int(py * TileHeight * scale)

	tiles := map[string]*Tile{}

	requiredWidth := int(math.Ceil(float64(canvasWidth)/(TileWidth*scale))) + 2
	requiredHeight := int(math.Ceil(float64(canvasHeight)/(TileHeight*scale))) + 2

	startX := -(requiredWidth / 2)
	startY := -(requiredHeight / 2)
	endX := startX + requiredWidth
	endY := startY + requiredHeight
	for cx := startX; cx < endX; cx++ {
		for cy := startY; cy < endY; cy++ {
			t := &Tile{
				URL:   m.urlEr.URL(int(zoom), cx+tx, cy+ty),
				DX:    dx - (cx * int(TileWidth*scale)),
				DY:    dy - (cy * int(TileHeight*scale)),
				Scale: scale,
				Zoom:  int(zoom),
			}
			tiles[t.URL.String()] = t
		}
	}

	return tiles
}

// Zooming specifies whether or not the map is currently zooming
type Zooming byte

// Possible values for zooming
const (
	ZoomingZero Zooming = iota
	ZoomingIn
	ZoomingOut
)

// Update will repaint the map
func (m *Map) Update(zooming Zooming) {
	zoomStart := m.zoom
	zoomEnd := m.zoom

	switch zooming {
	case ZoomingIn:
		zoomStart--
	case ZoomingOut:
		zoomEnd++
	}

	width := m.viewport.Get("width").Int()
	height := m.viewport.Get("height").Int()
	tiles := map[string]*Tile{}
	for zoom := zoomStart; zoom <= zoomEnd; zoom++ {
		ztiles := m.TilesFromCenter(zoom, int(width), int(height))

		for k, v := range ztiles {
			tiles[k] = v
		}
	}

	for _, r := range m.tileRenderers {
		r.RenderTiles(int(m.zoom), tiles)
	}
}
