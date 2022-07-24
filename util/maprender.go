package util

import (
	"bytes"
	"image"

	"image/jpeg"

	sm "github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"
	"github.com/shanghuiyang/rpi-projects/util/geo"
)

type MapRender struct {
	ctx *sm.Context
}

func NewMapRender() *MapRender {
	return &MapRender{
		ctx: sm.NewContext(),
	}
}

func (m *MapRender) SetSize(width, height int) {
	m.ctx.SetSize(width, height)
}

func (m *MapRender) SetZoom(zoom int) {
	m.ctx.SetZoom(zoom)
}

func (m *MapRender) SetCenter(pt *geo.Point) {
	m.ctx.SetCenter(s2.LatLngFromDegrees(pt.Lat, pt.Lon))
}

func (m *MapRender) SetCache(cache sm.TileCache) {
	m.ctx.SetCache(cache)
}

func (m *MapRender) SetTileProvider(tileProvider *sm.TileProvider) {
	m.ctx.SetTileProvider(tileProvider)
}

func (m *MapRender) AddMarker(marker *sm.Marker) {
	m.ctx.AddObject(marker)
}

func (m *MapRender) ClearMarker() {
	m.ctx.ClearObjects()
}

func (m *MapRender) Render() ([]byte, error) {
	img, err := m.ctx.Render()
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *MapRender) RenderImg() (*image.Image, error) {
	img, err := m.ctx.Render()
	if err != nil {
		return nil, err
	}
	return &img, nil
}
