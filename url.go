package pichiwmap

import (
	"fmt"
	"net/url"
)

// NewOpenStreetMapURLer creates an OpenStreetMap
func NewOpenStreetMapURLer(baseURL *url.URL) *OpenStreetMapURLer {
	return &OpenStreetMapURLer{baseURL: baseURL}
}

// OpenStreetMapURLer calculates a URL based on OpenStreetMap's spec
// https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
type OpenStreetMapURLer struct {
	baseURL *url.URL
}

// URL calcualtes a URL from zoom, x, and y
func (u *OpenStreetMapURLer) URL(zoom, x, y int) *url.URL {
	mapURL := *u.baseURL
	mapURL.Path = fmt.Sprintf("%v/%v/%v.png", zoom, x, y)
	return &mapURL
}
