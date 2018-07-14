package main

import (
	"net/url"

	"syscall/js"

	"github.com/pichiw/pichiwmap"
	"github.com/pichiw/pichiwmap/pmwgl"
)

func main() {
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	body := doc.Get("body")
	width := body.Get("clientWidth").Int()
	height := body.Get("clientHeight").Int()
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	baseURL, err := url.Parse("https://a.tile.openstreetmap.org")
	if err != nil {
		panic(err)
	}

	tr, err := pmwgl.NewTileRenderer("mycanvas")
	if err != nil {
		panic(err)
	}

	m, err := pichiwmap.New(baseURL, tr)
	if err != nil {
		panic(err)
	}

	c := make(chan struct{}, 0)

	m.Update()

	<-c
}
