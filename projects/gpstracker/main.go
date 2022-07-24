package main

import (
	"log"

	"github.com/shanghuiyang/rpi-projects/projects/gpstracker/lbs"
)

const configJSON = "config.json"

func main() {
	cfg, err := lbs.LoadConfig(configJSON)
	if err != nil {
		log.Fatalf("failed to load config, error: %v", err)
		panic(err)
	}
	lbs.Start(cfg)
}
