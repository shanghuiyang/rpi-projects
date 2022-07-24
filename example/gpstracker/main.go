package main

import (
	"bytes"
	"log"
	"os"
	"time"

	"image/color"
	"image/png"

	sm "github.com/flopp/go-staticmaps"
	"github.com/golang/geo/s2"
	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-devices/util"
)

const (
	timeFormat  = "2006-01-02T15:04:05"
	devName     = "/dev/ttyAMA0"
	baud        = 9600
	streamerURL = ":8082/map"

	cacheDir          = ".cache/maptiles"
	bingSatelliteTile = "bing-satellite"
	osmOutdoorTile    = "osm-outdoor"
	isOnline          = false
	tileWidth         = 256
	tileHeight        = 256
	zoom              = 18
)

var latlons = [][]float64{
	{39.956767, 116.447697},
	{39.956777, 116.447698},
}

func main() {
	gps, err := dev.NewNeo6mGPS(devName, baud)
	// gps, err := dev.NewGPSSimulator(latlons)
	if err != nil {
		log.Printf("failed to create gps, error: %v", err)
		return
	}
	defer gps.Close()

	logfile := time.Now().Format(timeFormat) + ".csv"
	logger, err := util.NewGPSLogger(logfile)
	if err != nil {
		log.Printf("failed to create log file, error: %v", err)
		return
	}
	logger.Printf("timestamp,lat,lon\n")

	streamer, err := util.NewStreamer(streamerURL)
	if err != nil {
		log.Printf("failed to create streamer, error: %v", err)
		return
	}

	ctx := sm.NewContext()
	tc := sm.NewTileCache(cacheDir, os.ModePerm)
	tp := newLocalTileProvider(bingSatelliteTile)
	ctx.SetCache(tc)
	ctx.SetTileProvider(tp)
	ctx.SetOnline(isOnline)
	ctx.SetSize(tileWidth, tileHeight)
	ctx.SetZoom(zoom)

	for {
		time.Sleep(1 * time.Second)
		lat, lon, err := gps.Loc()
		if err != nil {
			log.Printf("failed to get location, error: %v", err)
			continue
		}
		logger.Printf("%v,%.6f,%.6f\n", time.Now().Format(timeFormat), lat, lon)
		log.Printf("%v, %v", lat, lon)

		marker := sm.NewMarker(
			s2.LatLngFromDegrees(lat, lon),
			color.RGBA{0xff, 0, 0, 0xff},
			8.0,
		)
		ctx.ClearObjects()
		ctx.AddObject(marker)
		img, err := ctx.Render()
		if err != nil {
			log.Printf("failed to render map: %v", err)
			continue
		}

		buf := &bytes.Buffer{}
		if err := png.Encode(buf, img); err != nil {
			log.Printf("failed to encode image, error: %v", err)
			continue
		}
		streamer.Push(buf.Bytes())
	}
}

func newLocalTileProvider(name string) *sm.TileProvider {
	return &sm.TileProvider{
		Name:           name,
		Attribution:    "",
		IgnoreNotFound: true,
		TileSize:       256,
		URLPattern:     "",
		Shards:         []string{},
	}
}
