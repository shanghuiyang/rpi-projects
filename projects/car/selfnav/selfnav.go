package selfnav

import "github.com/shanghuiyang/rpi-projects/util/geo"

type SelfNav interface {
	Start(dest *geo.Point)
	Stop()
	InNaving() bool
}
