# pichiwmap - WebGL WASM Map in Go

_Requires a version of Go > 1.11beta1 (recommend build master from source)_

An attempt to write a 2D/3D WebAssembly + WebGL "[Slippy Map](https://wiki.openstreetmap.org/wiki/Slippy_Map)" for [OpenStreetMap](https://www.openstreetmap.org) data (and other tile servers with the same tile structure).

Right now this package is very experimental and most rendering logic is proof-of-concept in cmd/sample. 

## Done

- Basic map rendering sample (cmd/sample)
- Gesture / mouse panning and zooming
- Manual lat/lon/zoom entry
- Animtated keyboard panning
- Abstract the map logic from the rendering logic
  - Push all map logic into the "pichiwmap" package and a "pmwebgl" implementation of the renderer. 
  - Make UX friendly (`map, err := NewMap("divid")`)

## TODO

- Handle edge of the decidedly non-flat Earth.
- Implement concurrency protection (single thread != no race conditions :))
- Spike on vector tiles instead of (or in addition to) raster tiles. 
- Refinement of cache/loading (on-going)
- Markers, polygons, etc. 
- JavaScript hooks to allow non-wasm/go programmers to utilize the map

## FUTURE

- 3D Elevation/terrain (this is why I'm using WebGL instead of canvas) 
- 3D Tilt controls
- Leverage any future WASM / OpenGL native (non-javascript based) and other integrations
- _Much more! (scope is not yet well defined)_

## Shout-outs

_As early as Go+WASM is, I'm already standing on the shoulders of giants._

- [Brian Ketelsen](https://brianketelsen.com/) for the inspiration to try Go/WASM in the first place and much discussion of implementation
- [Richard Musiol](https://github.com/neelance) for making [GopherJS](https://github.com/gopherjs) and Go/WASM happen
- WebGL inspiration from https://github.com/stdiopt/gowasm-experiments
- WebGL is implemented with considerable help of [WebGL Fundamentals](https://webglfundamentals.org/). Particularly [Draw Image](https://webglfundamentals.org/webgl/lessons/webgl-2d-drawimage.html)
- Tile calculations, with the exception of layout ("[TileFromCenter](https://github.com/pichiw/pichiwmap/blob/master/tile.go#L36)"), are a direct port of https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames). 
- _and more to come..._
