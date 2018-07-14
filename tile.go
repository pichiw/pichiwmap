package pichiwmap

import (
	"fmt"
	"math"
	"net/url"
)

const (
	RadToDeg  = 180 / math.Pi
	DegToRad  = math.Pi / 180
	RadToGrad = 200 / math.Pi
	GradToDeg = math.Pi / 200
)

var zooms = map[int]float64{}

func init() {
	for i := 0; i <= 18; i++ {
		zooms[i] = math.Pow(2, float64(i))
	}
}

const (
	tileWidth  = 256
	tileHeight = 256
)

type Tile struct {
	DX  int
	DY  int
	URL *url.URL
}

// TilesFromCenter gets the tiles required from the centre point
func TilesFromCenter(baseURL *url.URL, zoom int, lat, lon float64, canvasWidth, canvasHeight int) []*Tile {
	cx, cy := TileNum(zoom, lat, lon)

	tx := int(cx)
	ty := int(cy)

	px := float64(tx) - cx
	py := float64(ty) - cy

	dx := -int(px * tileWidth)
	dy := -int(py * tileHeight)

	center := &Tile{DX: dx, DY: dy, URL: URL(baseURL, zoom, tx, ty)}
	tiles := []*Tile{center}

	requiredWidth := int(math.Ceil(float64(canvasWidth)/tileWidth)) + 1
	requiredHeight := int(math.Ceil(float64(canvasHeight)/tileHeight)) + 1

	startWidth := (requiredWidth / 2) - requiredWidth
	startHeight := (requiredHeight / 2) - requiredHeight
	for cx := startWidth; cx < (requiredWidth - startWidth); cx++ {
		for cy := startHeight; cy < (requiredHeight - startHeight); cy++ {
			if cx == 0 && cy == 0 {
				continue
			}
			tiles = append(tiles, &Tile{
				URL: URL(baseURL, zoom, tx+cx, ty+cy),
				DX:  dx - (cx * tileWidth),
				DY:  dy - (cy * tileHeight),
			})
		}
	}

	return tiles
}

// TileNum returns the tile x and y and pixel offset from the zoom, lat, and lon
func TileNum(zoom int, lat, lon float64) (x, y float64) {
	latRad := lat * DegToRad
	n := zooms[zoom]
	x = (lon + 180.0) / 360.0 * n
	y = (1.0 - math.Log(math.Tan(latRad)+(1/math.Cos(latRad)))/math.Pi) / 2.0 * n
	return
}

// Move moves the lat and long by the delta pixels pdx and pdy
func Move(zoom int, lat, lon float64, pdx int, pdy int) (nlat, nlon float64) {
	xf, yf := TileNum(zoom, lat, lon)
	dx := float64(pdx) / tileWidth
	dy := float64(pdy) / tileHeight

	return latlonFromXY(zoom, xf+dx, yf+dy)
}

func latlonFromXY(zoom int, x, y float64) (lat, lon float64) {
	n := zooms[zoom]
	lon = x/n*360.0 - 180.0
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*y/n)))
	lat = latRad * RadToDeg
	return
}

// NW returns the northwest corner of the tile in lat/lon degrees
func NW(zoom, x, y int) (lat, lon float64) {
	n := zooms[zoom]
	lon = float64(x)/n*360.0 - 180.0
	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	lat = latRad * RadToDeg
	return
}

// URL generates a url for the tile from a base URL, zoom, lat, and lon
func URL(baseURL *url.URL, zoom, x, y int) *url.URL {
	mapURL := *baseURL
	mapURL.Path = fmt.Sprintf("%v/%v/%v.png", zoom, x, y)
	return &mapURL
}
