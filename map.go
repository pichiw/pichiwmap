package pichiwmap

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gowasm/gopherwasm/js"
)

type TileRenderer interface {
	RenderTiles(tiles []*Tile)
	Viewport() (width, height float64)
}

func New(baseURL *url.URL, tileRenderer TileRenderer) (*Map, error) {
	doc := js.Global().Get("document")
	body := doc.Get("body")

	m := &Map{
		tileRenderer: tileRenderer,
		lat:          49.8951,
		lon:          -97.1384,
		zoom:         15,
		step:         0.001,
		baseURL:      baseURL,

		body:     body,
		zoomEl:   doc.Call("getElementById", "zoom"),
		latEl:    doc.Call("getElementById", "latitude"),
		lonEl:    doc.Call("getElementById", "longitude"),
		buttonEl: doc.Call("getElementById", "updatePosition"),
		canvasEl: doc.Call("getElementById", "mycanvas"),
		width:    body.Get("clientWidth").Int(),
		height:   body.Get("clientHeight").Int(),
	}

	m.tlat = m.lat
	m.tlon = m.lon

	m.canvasEl.Set("width", m.width)
	m.canvasEl.Set("height", m.height)

	m.body.Call("addEventListener", "gesturechange", js.NewEventCallback(js.PreventDefault, func(event js.Value) {}), false)
	m.body.Call("addEventListener", "gesturestart", js.NewEventCallback(js.PreventDefault, func(event js.Value) {}), false)

	m.canvasEl.Call("addEventListener", "mousedown", js.NewEventCallback(js.PreventDefault, m.onMouseDown), false)
	m.canvasEl.Call("addEventListener", "mouseup", js.NewEventCallback(js.PreventDefault, m.onMouseUp), false)
	m.canvasEl.Call("addEventListener", "mousemove", js.NewEventCallback(js.PreventDefault, m.onMouseMove), false)
	m.canvasEl.Call("addEventListener", "touchstart", js.NewEventCallback(js.PreventDefault, m.onMouseDown), false)
	m.canvasEl.Call("addEventListener", "touchend", js.NewEventCallback(js.PreventDefault, m.onMouseUp), false)
	m.canvasEl.Call("addEventListener", "touchmove", js.NewEventCallback(js.PreventDefault, m.onMouseMove), false)

	doc.Call("addEventListener", "keyup", js.NewEventCallback(js.PreventDefault, m.onKeyUp), false)
	doc.Call("addEventListener", "keydown", js.NewEventCallback(js.PreventDefault, m.onKeyDown), false)

	m.buttonEl.Call("addEventListener", "click", js.NewEventCallback(js.PreventDefault, m.onUpdateClick), false)

	return m, nil
}

type Map struct {
	tileRenderer TileRenderer

	doc      js.Value
	canvasEl js.Value
	body     js.Value
	zoomEl   js.Value
	latEl    js.Value
	lonEl    js.Value
	buttonEl js.Value

	baseURL       *url.URL
	zoom          int
	width         int
	height        int
	lat           float64
	lon           float64
	tlat          float64
	tlon          float64
	step          float64
	mouseStartX   int
	mouseStartY   int
	mouseStartLat float64
	mouseStartLon float64
	mouseDown     bool
}

func (m *Map) anim() {
	if m.lat == m.tlat && m.lon == m.tlon {
		return
	}

	if m.tlat > m.lat {
		m.lat += m.step
		if m.lat > m.tlat {
			m.lat = m.tlat
		}
	} else {
		m.lat -= m.step
		if m.lat < m.tlat {
			m.lat = m.tlat
		}
	}

	if m.tlon > m.lon {
		m.lon += m.step
		if m.lon > m.tlon {
			m.lon = m.tlon
		}
	} else {
		m.lon -= m.step
		if m.lon < m.tlon {
			m.lon = m.tlon
		}
	}

	time.Sleep(100 * time.Millisecond)
	m.Update()
	go m.anim()
}

func (m *Map) updateControls() {
	m.zoomEl.Set("value", strconv.Itoa(m.zoom))
	m.latEl.Set("value", strconv.FormatFloat(m.lat, 'f', 6, 64))
	m.lonEl.Set("value", strconv.FormatFloat(m.lon, 'f', 6, 64))
}

func (m *Map) onKeyDown(event js.Value) {
	if event == js.Undefined() {
		event = js.Global().Get("window").Get("event")
	}

	if event == js.Undefined() {
		return
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

	go m.anim()
}

func (m *Map) onUpdateClick(event js.Value) {
	defer m.updateControls()
	lastLat := m.lat
	lastLon := m.lon
	lastZoom := m.zoom

	var err error
	m.zoom, err = strconv.Atoi(m.zoomEl.Get("value").String())
	if err != nil {
		m.zoom = lastZoom
		js.Global().Call("alert", "Invalid zoom (must be between 1 and 18)")
		return
	}

	m.lat, err = strconv.ParseFloat(m.latEl.Get("value").String(), 64)
	if err != nil {
		m.lat = lastLat
		js.Global().Call("alert", "Invalid m.Lon Value")
		return
	}
	m.lon, err = strconv.ParseFloat(m.lonEl.Get("value").String(), 64)
	if err != nil {
		m.lon = lastLon
		js.Global().Call("alert", "Invalid m.Lat Value")
		return
	}

	m.Update()
}

func (m *Map) onKeyUp(event js.Value) {
	m.tlat = m.lat
	m.tlon = m.lon
}

func (m *Map) onMouseDown(event js.Value) {
	if touches := event.Get("touches"); touches != js.Undefined() {
		event = touches.Index(0)
	}
	m.mouseStartX = event.Get("pageX").Int()
	m.mouseStartY = event.Get("pageY").Int()
	m.mouseStartLat = m.lat
	m.mouseStartLon = m.lon
	m.mouseDown = true
}

func (m *Map) onMouseUp(event js.Value) {
	m.mouseDown = false
}

func (m *Map) onMouseMove(event js.Value) {
	if !m.mouseDown {
		return
	}

	if touches := event.Get("touches"); touches != js.Undefined() {
		event = touches.Index(0)
	}
	dx := m.mouseStartX - event.Get("pageX").Int()
	dy := m.mouseStartY - event.Get("pageY").Int()

	m.lat, m.lon = Move(m.zoom, m.mouseStartLat, m.mouseStartLon, dx, dy)
	m.Update()
}

func (m *Map) Update() {
	width, height := m.tileRenderer.Viewport()
	tiles := TilesFromCenter(m.baseURL, m.zoom, m.lat, m.lon, int(width), int(height))
	m.tileRenderer.RenderTiles(tiles)
	m.updateControls()
}
