/*
Auto-Air opens the air-cleaner automatically when the pm2.5 >= 130.
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-projects/iot"
	"github.com/shanghuiyang/rpi-projects/util"
)

const (
	pinSG = 18
)

const (
	trigOnPM25  = 120
	trigOffPm25 = 100
)

const (
	onenetToken = "your_onenet_token"
	onenetAPI   = "http://api.heclouds.com/devices/540381180/datapoints"
)

var (
	autoair  *autoAir
	bool2int = map[bool]int{
		false: 0,
		true:  1,
	}
)

type pm25Response struct {
	PM25     uint16 `json:"pm25"`
	ErrorMsg string `json:"error_msg"`
}

type autoAir struct {
	serov   dev.Motor
	cloud   iot.Cloud
	state   bool        // true: turn on, false: turn off
	chClean chan uint16 // for turning on/off the air-cleaner
	chCloud chan uint16 // for pushing to iot cloud
}

func main() {
	sg := dev.NewSG90(pinSG)
	cfg := &iot.Config{
		Token: onenetToken,
		API:   onenetAPI,
	}
	cloud := iot.NewOnenet(cfg)

	autoair = newAutoAir(sg, cloud)
	util.WaitQuit(func() {
		autoair.stop()
	})
	autoair.start()
}

func newAutoAir(serov dev.Motor, cloud iot.Cloud) *autoAir {
	return &autoAir{
		serov:   serov,
		cloud:   cloud,
		state:   false,
		chClean: make(chan uint16, 4),
		chCloud: make(chan uint16, 4),
	}
}

func (a *autoAir) start() {
	log.Printf("[autoair]service starting")
	go a.serov.Roll(45)
	go a.clean()
	go a.push()
	a.detect()
}

func (a *autoAir) detect() {
	log.Printf("[autoair]detecting pm2.5")
	for {
		pm25, err := a.getPM25()
		if err != nil {
			log.Printf("[autoair]failed to get pm2.5, error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("[autoair]pm2.5: %v ug/m3", pm25)

		a.chClean <- pm25
		time.Sleep(60 * time.Second)
	}
}

func (a *autoAir) clean() {
	for pm25 := range a.chClean {
		hour := time.Now().Hour()
		if pm25 < 400 && (hour >= 20 || hour < 8) {
			// disable at 20:00-08:00
			log.Printf("[autoair]auto air-cleaner was disabled at 20:00-08:00")
			if a.state {
				a.off()
			}
			continue
		}

		if !a.state && pm25 >= trigOnPM25 {
			a.on()
			log.Printf("[autoair]air-cleaner was turned on")
			continue
		}
		if a.state && pm25 < trigOffPm25 {
			a.off()
			log.Printf("[autoair]air-cleaner was turned off")
			continue
		}
	}
}

// push state to cloud
func (a *autoAir) push() {
	for {
		time.Sleep(60 * time.Second)
		v := &iot.Value{
			Device: "air-cleaner",
			Value:  bool2int[a.state],
		}
		if err := a.cloud.Push(v); err != nil {
			log.Printf("[autoair]push: failed to push the state of air-cleaner to cloud, error: %v", err)
		}
	}
}

func (a *autoAir) getPM25() (uint16, error) {
	resp, err := http.Get("http://localhost:8000/pm25")
	if err != nil {
		return 0, fmt.Errorf("failed to get pm2.5 from sensers server, status: %v, err: %v", resp.Status, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read resp body, err: %v", err)
	}

	var pm25Resp pm25Response
	if err := json.Unmarshal(body, &pm25Resp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal resp, err: %v", err)
	}

	if pm25Resp.ErrorMsg != "" {
		return 0, fmt.Errorf("failed to get pm2.5 from sensers server, status: %v, err msg: %v", resp.Status, pm25Resp.ErrorMsg)
	}

	return pm25Resp.PM25, nil
}

func (a *autoAir) on() {
	a.serov.Roll(0)
	time.Sleep(1 * time.Second)
	a.serov.Roll(-45)
	a.state = true
}

func (a *autoAir) off() {
	a.serov.Roll(0)
	time.Sleep(1 * time.Second)
	a.serov.Roll(45)
	a.state = false
}

func (a *autoAir) stop() {
	a.serov.Roll(45)
}
