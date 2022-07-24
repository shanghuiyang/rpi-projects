package selftracking

import (
	"image/color"
	"log"

	"github.com/shanghuiyang/rpi-projects/projects/car/car"
	"github.com/shanghuiyang/rpi-projects/cv"

	"github.com/shanghuiyang/rpi-projects/util"
	"gocv.io/x/gocv"
)

const (
	logTag = "selftracking"
)

type SelfTrackingImp struct {
	car        car.Car
	tracker    *cv.Tracker
	streamer   *util.Streamer
	inTracking bool
}

func NewSelfTrackingImp(c car.Car, t *cv.Tracker, s *util.Streamer) *SelfTrackingImp {
	return &SelfTrackingImp{
		car:        c,
		tracker:    t,
		streamer:   s,
		inTracking: false,
	}
}

func (s *SelfTrackingImp) Start(chImg chan *gocv.Mat) {
	if s.inTracking {
		return
	}

	s.inTracking = true

	rcolor := color.RGBA{G: 255, A: 255}
	firstTime := true // saw the ball at the first time
	for s.inTracking {
		util.DelayMs(200)

		img, ok := <-chImg
		if !ok {
			s.inTracking = false
			return
		}

		ok, rect := s.tracker.Locate(img)
		if ok {
			gocv.Rectangle(img, *rect, rcolor, 2)
		}
		buf, err := gocv.IMEncode(".jpg", *img)
		if err != nil {
			log.Printf("[%v]failed to encode image, err: %v", logTag, err)
			continue
		}
		s.streamer.Push(buf)

		if !ok {
			log.Printf("[%v]ball not found", logTag)
			firstTime = true
			continue
		}

		if rect.Max.Y > 480 {
			s.car.Stop()
			s.car.Beep(1, 300)
			continue
		}
		if firstTime {
			go s.car.Beep(1, 100)
		}
		firstTime = false
		x, y := s.tracker.MiddleXY(rect)
		log.Printf("[%v]found a ball at: (%v, %v)", logTag, x, y)
		if x < 200 {
			log.Printf("[%v]turn left to the ball", logTag)
			s.car.Left()
			util.DelayMs(100)
			s.car.Stop()
			continue
		}
		if x > 400 {
			log.Printf("[%v]turn right to the ball", logTag)
			s.car.Right()
			util.DelayMs(100)
			s.car.Stop()
			continue
		}
		log.Printf("[%v]forward to the ball", logTag)
		s.car.Forward()
		util.DelayMs(100)
		s.car.Stop()

	}
}

func (s *SelfTrackingImp) InTracking() bool {
	return s.inTracking
}

func (s *SelfTrackingImp) Stop() {
	s.inTracking = false
}
