package server

import (
	"io/ioutil"
	"log"

	"github.com/shanghuiyang/astar/tilemap"
	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-projects/projects/carx/car"
	"github.com/shanghuiyang/rpi-projects/util"
)

const (
	forward  Op = "forward"
	backward Op = "backward"
	left     Op = "left"
	right    Op = "right"
	stop     Op = "stop"
	beep     Op = "beep"

	chSize                  = 8
	defaultVolume           = 40
	defaultSpeed            = 30
	defaultHost             = ":8080"
	defaultVideoHost        = ":8081"
	defaultTrackingVideoURL = ":8082/tracking"
)

type Op string

func init() {
	var err error
	pageContext, err = ioutil.ReadFile("home.html")
	if err != nil {
		log.Panicf("failed to load home page, error: %v", err)
	}

	ip = util.GetIP()
	if ip == "" {
		log.Panicf("failed to get ip address")
	}
}

type service struct {
	cfg        *Config
	car        car.Car
	led        dev.Led
	relay      dev.Relay
	ledBlinked bool
	islighton  bool
	onMusic    bool
	chOp       chan Op

	selfdriving car.SelfDriving
	selfnav     car.SelfNav
}

func newService(cfg *Config) (*service, error) {
	if cfg.Speed == 0 {
		cfg.Speed = defaultSpeed
	}
	if cfg.Volume == 0 {
		cfg.Volume = defaultVolume
	}
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.VideoHost == "" {
		cfg.Host = defaultVideoHost
	}
	if cfg.SelfTracking.VideoURL == "" {
		cfg.SelfTracking.VideoURL = defaultTrackingVideoURL
	}

	l298n := dev.NewL298N(
		cfg.L298N.IN1Pin,
		cfg.L298N.IN2Pin,
		cfg.L298N.IN3Pin,
		cfg.L298N.IN4Pin,
		cfg.L298N.ENAPin,
		cfg.L298N.ENBPin,
	)

	motorL := dev.NewDCMotor(l298n.MotorA)
	motorR := dev.NewDCMotor(l298n.MotorB)
	buz := dev.NewBuzzerImp(cfg.BuzzerPin, dev.High)
	relay := dev.NewRelayImp(cfg.RelayPin)
	led := dev.NewLedImp(cfg.LedPin)
	sg90 := dev.NewSG90(cfg.SG90DataPin)
	us100, err := dev.NewUS100GPIO(cfg.US100.TrigPin, cfg.US100.EchoPin)
	if err != nil {
		log.Panicf("[%v]new us100 error: %v", logTag, err)
	}
	gy25, err := dev.NewGY25(cfg.GY25.Dev, cfg.GY25.Baud)
	if err != nil {
		log.Panicf("[%v]new gy-25 error: %v", logTag, err)
	}

	carimp := car.NewCarImp(motorL, motorR, gy25, buz)
	carimp.Speed(cfg.Speed)

	s := &service{
		cfg:        cfg,
		car:        carimp,
		led:        led,
		relay:      relay,
		ledBlinked: true,
		islighton:  false,
		onMusic:    false,
		chOp:       make(chan Op, chSize),
	}

	if cfg.SelfDriving.Enabled {
		s.selfdriving = car.NewSelfDrivingImp(carimp, us100, sg90)
	}

	if cfg.SelfNav.Enabled {
		gps, err := dev.NewNeo6mGPS(cfg.SelfNav.GPSConfig.Dev, cfg.SelfNav.GPSConfig.Baud)
		if err != nil {
			log.Panicf("[%v]failed to create gps, error: %v", logTag, err)
		}
		data, err := ioutil.ReadFile(cfg.SelfNav.TileMapConfig.MapFile)
		if err != nil {
			log.Panicf("[%v]failed to read map file: %v, errror: %v", logTag, cfg.SelfNav.TileMapConfig.MapFile, err)
		}
		m := tilemap.BuildFromStr(string(data))
		s.selfnav = car.NewSelfNavImp(carimp, gps, m, cfg.SelfNav.TileMapConfig.Box, cfg.SelfNav.TileMapConfig.GridSize)
	}

	// if err := util.SetVolume(cfg.Volume); err != nil {
	// 	log.Panicf("[%v]failed to create tracker, error: %v", logTag, err)
	// }

	return s, nil
}

func (s *service) start() error {
	go s.operate()
	go s.ledBlink()
	return nil
}

func (s *service) shutdown() error {
	s.ledBlinked = false
	close(s.chOp)
	s.car.Stop()
	s.led.Off()
	s.relay.Off()
	return nil
}

func (s *service) operate() {
	for op := range s.chOp {
		log.Printf("[car]op: %v", op)
		s.car.Speed(s.cfg.Speed)
		switch op {
		case forward:
			s.car.Forward()
		case backward:
			s.car.Backward()
		case left:
			s.car.Left()
		case right:
			s.car.Right()
		case stop:
			s.car.Stop()
		case beep:
			go s.car.Beep(3, 100)
		default:
			log.Printf("[car]invalid op")
		}
	}
}

func (s *service) ledBlink() {
	for s.ledBlinked {
		s.led.Blink(1, 1000)
	}
}
