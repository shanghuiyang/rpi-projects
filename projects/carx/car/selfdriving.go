package car

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shanghuiyang/kalman1d"
	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-projects/util"
)

const (
	forward  operator = "forward"
	backward operator = "backward"
	left     operator = "left"
	right    operator = "right"
	stop     operator = "stop"
	turn     operator = "turn"
	scan     operator = "scan"

	logSelfDrivingTag = "selfdriving"
)

const (
	trueV      = -30.0 // cm/s (negative means distance is decreasing)
	initPosVar = 200.0
	initVelVar = 2000.0
	sigmaZ     = 7.0   // cm, measurement noise std-dev
	sigmaA     = 100.0 // cm/s^2, process noise std-dev (acceleration)
	gateNSigma = 3.5   // typical values: 3~4
)

var (
	scanningAngles = []float64{-90, -75, -60, -45, -30, -15, 0, 15, 30, 45, 60, 75, 90}
	aheadAngles    = []float64{0, -20, 0, 20}
)

type operator string

type SelfDriving interface {
	Start()
	Stop()
	InDrving() bool
}

type SelfDrivingImp struct {
	car       Car
	dmeter    dev.DistanceMeter
	servo     dev.ServoMotor
	indriving bool
}

func NewSelfDrivingImp(c Car, d dev.DistanceMeter, servo dev.ServoMotor) *SelfDrivingImp {
	servo.Roll(0)
	return &SelfDrivingImp{
		car:       c,
		dmeter:    d,
		servo:     servo,
		indriving: false,
	}
}

func (s *SelfDrivingImp) Start() {
	if s.indriving {
		return
	}

	s.indriving = true

	var (
		fwd       bool
		retry     int
		mindAngle float64
		maxdAngle float64
		mind      float64
		maxd      float64
		op        = forward
		chOp      = make(chan operator, 4)
	)

	for s.indriving {
		select {
		case p := <-chOp:
			op = p
			for len(chOp) > 0 {
				log.Printf("[%v]skip op: %v", logSelfDrivingTag, <-chOp)
			}
		default:
			// 	do nothing
		}
		log.Printf("[%v]op: %v", logSelfDrivingTag, op)

		switch op {
		case backward:
			fwd = false
			s.car.Stop()
			util.DelayMs(20)
			s.car.Backward()
			util.DelayMs(500)
			chOp <- stop
			continue
		case stop:
			fwd = false
			s.car.Stop()
			util.DelayMs(20)
			chOp <- scan
			continue
		case scan:
			fwd = false
			mind, maxd, mindAngle, maxdAngle = s.lookingForWay()
			log.Printf("[%v]mind=%.0f, maxd=%.0f, mindAngle=%v, maxdAngle=%v", logSelfDrivingTag, mind, maxd, mindAngle, maxdAngle)
			if mind < 10 && mindAngle != 90 && mindAngle != -90 && retry < 4 {
				chOp <- backward
				retry++
				continue
			}
			chOp <- turn
			retry = 0
		case turn:
			fwd = false
			s.car.Turn(maxdAngle)
			util.DelayMs(150)
			chOp <- forward
			continue
		case forward:
			if !fwd {
				s.car.Forward()
				fwd = true
				go s.lookingForObs(chOp)
			}
			util.DelayMs(50)
			continue
		}
	}
	s.car.Stop()
	util.DelaySec(1)
	close(chOp)
}

func (s *SelfDrivingImp) InDrving() bool {
	return s.indriving
}

func (s *SelfDrivingImp) Stop() {
	s.indriving = false
}

// lookingForWay looks for geting the min & max distance, and their corresponding angles
// mind: the min distance
// maxd: the max distance
// mindAngle: the angle correspond to the mind
// maxdAngle: the angle correspond to the maxd
func (s *SelfDrivingImp) lookingForWay() (mind, maxd, mindAngle, maxdAngle float64) {
	mind = 9999
	maxd = -9999
	for _, ang := range scanningAngles {
		s.servo.Roll(ang)
		util.DelayMs(150)
		d, err := s.dmeter.Dist()
		for i := 0; err != nil && i < 3; i++ {
			util.DelayMs(100)
			d, err = s.dmeter.Dist()
		}
		if err != nil {
			continue
		}
		log.Printf("[%v]scan: angle=%v, dist=%.0f", logSelfDrivingTag, ang, d)
		if d < mind {
			mind = d
			mindAngle = ang
		}
		if d > maxd {
			maxd = d
			maxdAngle = ang
		}
	}
	s.servo.Roll(0)
	util.DelayMs(50)
	return
}

func (s *SelfDrivingImp) lookingForObs(chOp chan operator) {

	cfg := &kalman1d.Config{
		InitX:            0,
		InitV:            trueV,
		InitPosVar:       initPosVar,
		InitVelVar:       initVelVar,
		SigmaA:           sigmaA,
		SigmaZ:           sigmaZ,
		GateNSigma:       gateNSigma,
		OutlierInflation: 100,
	}

	var kFilter *kalman1d.Filter

	now := time.Now().Unix()
	createdFile := true
	f, err := os.OpenFile(fmt.Sprintf("dist_%v.csv", now), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("[%v]create file error: %v", logSelfDrivingTag, err)
		createdFile = true
	}
	f.WriteString("raw,est\n")

	defer func() {
		s.car.Speed(50)
	}()

	for s.indriving {
		for _, angle := range aheadAngles {
			s.servo.Roll(angle)
			util.DelayMs(150)
			d, err := s.dmeter.Dist()
			for i := 0; err != nil && i < 3; i++ {
				log.Printf("[%v]get distance error: %v", logSelfDrivingTag, err)
				util.DelayMs(150)
				d, err = s.dmeter.Dist()
			}
			if err != nil {
				continue
			}

			if kFilter == nil {
				cfg.InitX = d
				kFilter = kalman1d.NewFilter(cfg)
			}

			est := kFilter.Update(d, 0.155)
			log.Printf("[%v]raw distance=%.2f, estimated=%.2f", logSelfDrivingTag, d, est)
			if createdFile {
				f.WriteString(fmt.Sprintf("%.2f,%.2f\n", d, est))
			}

			switch {
			case est < 70:
				s.car.Speed(35)
			case est < 65:
				s.car.Speed(30)
			case est < 60:
				s.car.Speed(25)
			case est < 55:
				s.car.Speed(20)
			case est < 50:
				s.car.Speed(20)
			}

			if est < 20 {
				chOp <- backward
				return
			}
			if est < 40 {
				chOp <- stop
				return
			}
		}
	}
}
