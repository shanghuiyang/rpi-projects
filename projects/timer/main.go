package main

import (
	"log"

	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-projects/util"
)

const (
	configJSON = "config.json"
)

func main() {
	cfg, err := loadConfig(configJSON)
	if err != nil {
		log.Panicf("load alarm clock config error: %v", err)
	}

	buz := dev.NewBuzzerImp(cfg.Button, dev.High)
	btn := dev.NewButtonImp(cfg.Buzzer)
	timer := newTimer(buz, btn, cfg.Cronjobs)
	util.WaitQuit(func() {
		timer.stop()
	})
	timer.start()
	select {}
}
