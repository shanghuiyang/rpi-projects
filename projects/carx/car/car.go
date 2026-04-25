package car

import (
	"log"
	"time"

	"github.com/shanghuiyang/rpi-devices/dev"
)

const defaultSpeed = 30

// Car ...
type Car interface {
	Forward()
	Backward()
	Left()
	Right()
	Stop()
	Speed(speed uint32)
	Beep(n int, interval int)
	Turn(angle float64)
}

type CarImp struct {
	motorL dev.Motor
	motorR dev.Motor
	acc    dev.Accelerometer
	buzzer dev.Buzzer
}

func NewCarImp(motorL, motorR dev.Motor, acc dev.Accelerometer, buz dev.Buzzer) *CarImp {
	c := &CarImp{
		motorL: motorL,
		motorR: motorR,
		acc:    acc,
		buzzer: buz,
	}
	c.Speed(defaultSpeed)
	return c
}

func (c *CarImp) Forward() {
	c.motorL.Forward()
	c.motorR.Forward()
}

func (c *CarImp) Backward() {
	c.motorL.Backward()
	c.motorR.Backward()
}

func (c *CarImp) Left() {
	c.motorL.Backward()
	c.motorR.Forward()
}

func (c *CarImp) Right() {
	c.motorL.Forward()
	c.motorR.Backward()
}

func (c *CarImp) Stop() {
	c.motorL.Stop()
	c.motorR.Stop()
}

func (c *CarImp) Speed(speed uint32) {
	c.motorL.SetSpeed(speed)
	c.motorR.SetSpeed(speed)
}

func (c *CarImp) Beep(n int, interval int) {
	c.buzzer.Beep(n, interval)
}

func (c *CarImp) Turn(angle float64) {
	turnf := c.Right
	if angle < 0 {
		turnf = c.Left
		angle *= (-1)
	}

	var dt time.Duration
	switch {
	case angle <= 15:
		dt = time.Millisecond * 200
	case angle <= 30:
		dt = time.Millisecond * 300
	case angle <= 45:
		dt = time.Millisecond * 400
	case angle <= 60:
		dt = time.Millisecond * 500
	case angle <= 75:
		dt = time.Millisecond * 600
	case angle <= 90:
		dt = time.Millisecond * 700
	default:
		dt = time.Millisecond * 400
	}
	log.Printf("[car]turn, angle: %v, time: %v", angle, dt)
	turnf()
	time.Sleep(dt)
	c.Stop()
}
