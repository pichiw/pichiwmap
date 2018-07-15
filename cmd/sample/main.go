package main

import (
	"net/url"
	"strconv"

	"syscall/js"

	"github.com/pichiw/pichiwmap"
	"github.com/pichiw/pichiwmap/pmwgl"
)

func main() {
	doc := js.Global().Get("document")
	divEl := doc.Call("getElementById", "mapid")

	body := doc.Get("body")

	divEl.Set("width", body.Get("clientWidth").Int())
	divEl.Set("height", body.Get("clientHeight").Int())

	zoomEl := doc.Call("getElementById", "zoom")
	latEl := doc.Call("getElementById", "latitude")
	lonEl := doc.Call("getElementById", "longitude")
	buttonEl := doc.Call("getElementById", "updatePosition")

	baseURL, err := url.Parse("https://a.tile.openstreetmap.org")
	if err != nil {
		panic(err)
	}

	events := pichiwmap.MapEvents{
		OnLatChanged: func(lat float64) {
			latEl.Set("value", strconv.FormatFloat(lat, 'f', 6, 64))
		},
		OnLonChanged: func(lon float64) {
			lonEl.Set("value", strconv.FormatFloat(lon, 'f', 6, 64))
		},
		OnZoomChanged: func(zoom float64) {
			zoomEl.Set("value", strconv.FormatFloat(zoom, 'f', 6, 64))
		},
	}

	m, err := pichiwmap.New(pichiwmap.NewOpenStreetMapURLer(baseURL), divEl, events)
	if err != nil {
		panic(err)
	}

	events.OnLatChanged(m.Lat())
	events.OnLonChanged(m.Lon())
	events.OnZoomChanged(m.Zoom())

	tr, err := pmwgl.NewTileRenderer(m.Canvas())
	if err != nil {
		panic(err)
	}

	m.AddTileRenderers(tr)

	buttonEl.Call("addEventListener", "click", js.NewEventCallback(js.PreventDefault, onUpdateClick(m, zoomEl, latEl, lonEl)), false)

	c := make(chan struct{}, 0)

	m.Update(pichiwmap.ZoomingZero)

	<-c
}

func onUpdateClick(m *pichiwmap.Map, zoomEl, latEl, lonEl js.Value) func(event js.Value) {
	return func(event js.Value) {

		zoom, err := strconv.ParseFloat(zoomEl.Get("value").String(), 64)
		if err != nil {
			js.Global().Call("alert", "Invalid zoom (must be between 1 and 18)")
			return
		}

		lat, err := strconv.ParseFloat(latEl.Get("value").String(), 64)
		if err != nil {
			js.Global().Call("alert", "Invalid m.Lon Value")
			return
		}
		lon, err := strconv.ParseFloat(lonEl.Get("value").String(), 64)
		if err != nil {
			js.Global().Call("alert", "Invalid m.Lat Value")
			return
		}

		m.SetPosition(zoom, lat, lon)
	}
}
