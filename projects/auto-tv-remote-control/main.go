package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-projects/util"
)

const (
	devName     = "/dev/ttyAMA0"
	baud        = 9600
	pinPlayBtn  = 21
	pinSleepBtn = 20
	pinBuz      = 26
)

const (
	OK     = "OK"
	LEFT   = "LEFT"
	RIGHT  = "RIGHT"
	UP     = "UP"
	DOWN   = "DOWN"
	BACK   = "BACK"
	SLEEP  = "SLEEP"
	RED    = "RED"
	GREEN  = "GREEN"
	YELLOW = "YELLOW"
	BLUE   = "BLUE"
)

type code []byte

var codes = map[string]code{
	OK:     {0xA1, 0xF1, 0xB3, 0x4C, 0xCE},
	LEFT:   {0xA1, 0xF1, 0xB3, 0x4C, 0x99},
	RIGHT:  {0xA1, 0xF1, 0xB3, 0x4C, 0xC1},
	UP:     {0xA1, 0xF1, 0xB3, 0x4C, 0xCA},
	DOWN:   {0xA1, 0xF1, 0xB3, 0x4C, 0xD2},
	BACK:   {0xA1, 0xF1, 0xB3, 0x4C, 0xC5},
	SLEEP:  {0xA1, 0xF1, 0xB3, 0x4C, 0xDC},
	RED:    {0xB3, 0x4C, 0x84},
	GREEN:  {0xB3, 0x4C, 0x89},
	YELLOW: {0xB3, 0x4C, 0xD9},
	BLUE:   {0xB3, 0x4C, 0x96},
}

var (
	ir       *dev.IRCoder
	playBtn  = dev.NewButtonImp(pinPlayBtn)
	sleepBtn = dev.NewButtonImp(pinSleepBtn)
	buz      = dev.NewBuzzerImp(pinBuz, dev.High)
)

func main() {
	var err error
	ir, err = dev.NewIRCoder(devName, baud)
	if err != nil {
		log.Fatalf("new ircoder error: %v", err)
	}
	defer ir.Close()

	log.Printf("get ready for working")
	for {
		recv, err := ir.Read()
		if err != nil || len(recv) != 3 {
			util.DelayMs(100)
			continue
		}

		redBtn := codes[RED]
		redBtnPressed := recv[0] == redBtn[0] && recv[1] == redBtn[1] && recv[2] == redBtn[2]
		if redBtnPressed {
			log.Printf("red button was pressed")
			go buz.Beep(1, 100)
			play()
			util.DelaySec(1)
		}

		blueBtn := codes[BLUE]
		blueBtnPressed := recv[0] == blueBtn[0] && recv[1] == blueBtn[1] && recv[2] == blueBtn[2]
		if blueBtnPressed {
			log.Printf("blue button was pressed")
			go buz.Beep(1, 100)
			sleep()
			util.DelaySec(1)
		}

		if playBtn.Pressed() {
			log.Printf("play button was pressed")
			go buz.Beep(1, 100)
			play()
			util.DelaySec(1)
		}

		if sleepBtn.Pressed() {
			log.Printf("sleep button was pressed")
			go buz.Beep(1, 100)
			sleep()
			util.DelaySec(1)
		}

		util.DelayMs(100)
	}

}

func play() {
	log.Printf("going to play tv")
	ops := []string{LEFT, OK, BACK, OK, DOWN, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, RIGHT, DOWN, OK}
	secs := []int{2, 2, 15, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 5}
	for i, op := range ops {
		log.Printf("delay %v sec", secs[i])
		util.DelaySec(time.Duration(secs[i]))
		code := codes[op]
		if err := ir.Send(code); err != nil {
			go buz.Beep(1, 2000)
			log.Printf("error: %v", err)
			break
		}
		go buz.Beep(1, 100)
		log.Printf("sent: [%v], code: %v\n", op, code)
	}
	go buz.Beep(3, 100)
	log.Printf("done")
}

func sleep() {
	log.Printf("going to sleep")
	ops := []string{BACK, BACK, BACK, BACK, LEFT, OK, SLEEP}
	secs := []int{3, 3, 3, 3, 3, 3, 3}
	for i, op := range ops {
		log.Printf("delay %v sec", secs[i])
		util.DelaySec(time.Duration(secs[i]))
		code := codes[op]
		if err := ir.Send(code); err != nil {
			go buz.Beep(1, 2000)
			log.Printf("error: %v", err)
			break
		}
		go buz.Beep(1, 100)
		log.Printf("sent: [%v], code: %v\n", op, code)
	}
	go buz.Beep(3, 100)
	log.Printf("done")
}

func (c code) String() string {
	var s []string
	for _, b := range c {
		s = append(s, fmt.Sprintf("0x%X", b))
	}
	return strings.Join(s, ", ")
}
