package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/shanghuiyang/rpi-projects/iot"
)

const (
	maxRetry = 10
)

const (
	onenetToken = "your_onenet_token"
	onenetAPI   = "http://api.heclouds.com/devices/540381180/datapoints"
)

func main() {
	oneCfg := &iot.Config{
		Token: onenetToken,
		API:   onenetAPI,
	}
	cloud := iot.NewOnenet(oneCfg)
	if cloud == nil {
		log.Printf("[ip]failed to new OneNet iot cloud")
		return
	}

	n := 0
	for ; n < maxRetry; n++ {
		time.Sleep(10 * time.Second)
		ip, err := getIP()
		if err != nil {
			log.Printf("[ip]failed to get ip address, error: %v", err)
			log.Printf("[ip]retry %v...", n+1)
			continue
		}
		log.Printf("[ip]ip: %v", ip)

		v := &iot.Value{
			Device: "ip",
			Value:  ip,
		}
		if err := cloud.Push(v); err != nil {
			log.Printf("[ip]failed to push ip address to cloud, error: %v", err)
			log.Printf("[ip]retry %v...", n+1)
			continue
		}
		break
	}
	if n >= maxRetry {
		log.Printf("[ip]failed to get ip address")
		return
	}
	log.Printf("[ip]success")
}

func getIP() (string, error) {
	cmd := exec.Command("hostname", "-I")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	items := strings.Split(string(out), " ")
	if len(items) == 0 {
		return "", fmt.Errorf("failed to exec hostname")
	}
	return items[0], nil
}
