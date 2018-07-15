package pichiwmap

import (
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
	TileWidth  = 256
	TileHeight = 256
)

type Tile struct {
	DX    int
	DY    int
	Scale float64
	URL   *url.URL
}

type URLer interface {
	URL(zoom, x, y int) *url.URL
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
	dx := float64(pdx) / TileWidth
	dy := float64(pdy) / TileHeight

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
