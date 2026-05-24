package car

import (
	"log"
	"time"

	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-projects/util"
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
	TurnDuration(d time.Duration)
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

	yaw, _, _, err := c.acc.Angles()
	if err != nil {
		log.Printf("[car]failed to get angles from gy-25, error: %v", err)
		return
	}

	turnf()
	retry := 0
	ang := 0.0
	coeff := 2.0
	for ang < angle {
		yaw2, _, _, err := c.acc.Angles()
		if err != nil {
			log.Printf("[car]failed to get angles from gy-25, error: %v", err)
			if retry < 3 {
				retry++
				continue
			}
			break
		}
		ang = coeff * util.IncludedAngle(yaw, yaw2)
		log.Printf("target angle: %.2f, start angle: %.2f, current angle: %.2f, turned angle: %.2f", angle, yaw, yaw2, ang)
	}
	c.Stop()
}

func (c *CarImp) TurnWithoutSensor(angle float64) {
	turnf := c.Right
	x := angle
	if angle < 0 {
		turnf = c.Left
		x *= (-1)
	}

	var y float64
	switch {
	case x < 5:
		return
	case x <= 45:
		y = 0.001481*x*x*x - 0.1778*x*x + 11*x - 4.019e-14
	case x <= 360:
		y = -1.922e-07*x*x*x*x + 0.0001511*x*x*x - 0.04016*x*x + 8.541*x - 53.57
	default:
		log.Printf("[car]turn, invalid angle: %v > 360", angle)
		return
	}

	dt := time.Duration(y) * time.Millisecond

	log.Printf("[car]turn, angle: %v, duration: %v", angle, dt)
	turnf()
	time.Sleep(dt)
	c.Stop()
}

func (c *CarImp) TurnDuration(d time.Duration) {
	turnf := c.Right
	if d < 0 {
		turnf = c.Left
		d *= (-1)
	}

	log.Printf("[car]turn, duration: %v", d)
	turnf()
	time.Sleep(d)
	c.Stop()
}
