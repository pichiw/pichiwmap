package pichiwmap

import (
	"fmt"
	"net/url"
)

func NewOpenStreetMapURLer(baseURL *url.URL) *OpenStreetMapURLer {
	return &OpenStreetMapURLer{baseURL: baseURL}
}

type OpenStreetMapURLer struct {
	baseURL *url.URL
}

func (u *OpenStreetMapURLer) URL(zoom, x, y int) *url.URL {
	mapURL := *u.baseURL
	mapURL.Path = fmt.Sprintf("%v/%v/%v.png", zoom, x, y)
	return &mapURL
}
