package main

import (
	"log"
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

var (
	tvOK     = []byte{0xA1, 0xF1, 0xB3, 0x4C, 0xCE}
	tvLeft   = []byte{0xA1, 0xF1, 0xB3, 0x4C, 0x99}
	tvRight  = []byte{0xA1, 0xF1, 0xB3, 0x4C, 0xC1}
	tvUp     = []byte{0xA1, 0xF1, 0xB3, 0x4C, 0xCA}
	tvDown   = []byte{0xA1, 0xF1, 0xB3, 0x4C, 0xD2}
	tvBack   = []byte{0xA1, 0xF1, 0xB3, 0x4C, 0xC5}
	tvSleep  = []byte{0xA1, 0xF1, 0xB3, 0x4C, 0xDC}
	tvRed    = []byte{0xB3, 0x4C, 0x84}
	tvGreen  = []byte{0xB3, 0x4C, 0x89}
	tvYellow = []byte{0xB3, 0x4C, 0xD9}
	tvBlue   = []byte{0xB3, 0x4C, 0x96}
)

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
		data, err := ir.Read()
		if err != nil || len(data) != 3 {
			util.DelayMs(100)
			continue
		}

		if data[0] == tvRed[0] && data[1] == tvRed[1] && data[2] == tvRed[2] {
			go buz.Beep(1, 100)
			play()
			util.DelaySec(1)
		}

		if data[0] == tvBlue[0] && data[1] == tvBlue[1] && data[2] == tvBlue[2] {
			go buz.Beep(1, 100)
			sleep()
			util.DelaySec(1)
		}

		if playBtn.Pressed() {
			go buz.Beep(1, 100)
			play()
			util.DelaySec(1)
		}

		if sleepBtn.Pressed() {
			go buz.Beep(1, 100)
			sleep()
			util.DelaySec(1)
		}

		util.DelayMs(100)
	}

}

func play() {
	log.Printf("going to play tv")
	cmds := [][]byte{tvLeft, tvOK, tvBack, tvOK, tvDown, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvRight, tvDown, tvOK}
	secs := []int{2, 2, 15, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 5}
	for i, cmd := range cmds {
		log.Printf("delay %v sec", secs[i])
		util.DelaySec(time.Duration(secs[i]))
		if err := ir.Send(cmd); err != nil {
			go buz.Beep(1, 2000)
			log.Printf("error: %v", err)
			break
		}
		go buz.Beep(1, 100)
		log.Printf("sent %v\n", cmd)
	}
	go buz.Beep(3, 100)
	log.Printf("done")
}

func sleep() {
	log.Printf("going to sleep")
	cmds := [][]byte{tvBack, tvBack, tvBack, tvBack, tvLeft, tvOK, tvSleep}
	secs := []int{3, 3, 3, 3, 3, 3, 3}
	for i, cmd := range cmds {
		log.Printf("delay %v sec", secs[i])
		util.DelaySec(time.Duration(secs[i]))
		if err := ir.Send(cmd); err != nil {
			go buz.Beep(1, 2000)
			log.Printf("error: %v", err)
			break
		}
		go buz.Beep(1, 100)
		log.Printf("sent %v\n", cmd)
	}
	go buz.Beep(3, 100)
	log.Printf("done")
}
