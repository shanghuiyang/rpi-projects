package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/shanghuiyang/rpi-projects/util"
	"github.com/shanghuiyang/rpi-projects/util/geo"
)

const (
	ipPattern        = "((000.000.000.000))"
	selfDrivingState = "((selfdriving-state))"
	musicState       = "((music-state))"
	lightState       = "((light-state))"
	// volumePattern    = "((current-volume))"

	selfDrivingEnabled = "((selfdriving-enabled))"

	logHandlerTag = "handler"
)

var (
	ip          string
	pageContext []byte
)

func (s *service) loadHomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]load home page", logHandlerTag)
	rbuf := bytes.NewBuffer(pageContext)
	wbuf := bytes.NewBuffer([]byte{})
	// volume, err := util.GetVolume()
	// if err != nil {
	// 	log.Printf("[%v]failed to get volume, error: %v", logHandlerTag, err)
	// 	volume = 40
	// }
	disabled := false
	selfDriving := false
	if s.selfdriving != nil {
		selfDriving = s.selfdriving.InDrving()
	}

	if selfDriving {
		disabled = true
	}

	for {
		line, err := rbuf.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		sline := string(line)

		sline = strings.Replace(sline, ipPattern, ip, 1)
		// sline = strings.Replace(sline, volumePattern, fmt.Sprintf("%v", volume), 1)

		if strings.Contains(sline, selfDrivingState) {
			state := "unchecked"
			if selfDriving {
				state = "checked"
			}
			sline = strings.Replace(sline, selfDrivingState, state, 1)

			able := "enabled"
			if state == "unchecked" && disabled {
				able = "disabled"
			}
			sline = strings.Replace(sline, selfDrivingEnabled, able, 1)
		}

		if strings.Contains(sline, musicState) {
			state := "unchecked"
			if s.onMusic {
				state = "checked"
			}
			sline = strings.Replace(sline, musicState, state, 1)
		}

		if strings.Contains(sline, lightState) {
			state := "unchecked"
			if s.islighton {
				state = "checked"
			}
			sline = strings.Replace(sline, lightState, state, 1)
		}

		wbuf.Write([]byte(sline))
	}
	w.Write(wbuf.Bytes())
}

func (s *service) opHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]op", logHandlerTag)
	vars := mux.Vars(r)
	v, ok := vars["op"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid op: %v", vars["op"])
		return
	}
	op := Op(v)
	s.chOp <- op
}

func (s *service) turnAngleHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]turn in a angle(degree)", logHandlerTag)
	vars := mux.Vars(r)
	a, ok := vars["degree"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid angle: %v", vars["degree"])
		return
	}
	angle, err := strconv.ParseFloat(a, 64)
	if err != nil {
		log.Printf("[%v]invalid angle: %v", logHandlerTag, a)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid angle: %v", a)
		return
	}
	log.Printf("[%v]turn angle: %v", logHandlerTag, angle)
	s.car.Turn(angle)
}

func (s *service) turnDurationHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]turn in a duration(millisecond)", logHandlerTag)
	vars := mux.Vars(r)
	d, ok := vars["millisecond"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid angle: %v", vars["millisecond"])
		return
	}
	duration, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		log.Printf("[%v]invalid duration: %v", logHandlerTag, d)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid duration: %v", d)
		return
	}
	log.Printf("[%v]turn duration: %v", logHandlerTag, duration)
	s.car.TurnDuration(time.Duration(duration) * time.Millisecond)
}

func (s *service) lightOnHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]turn on the light", logHandlerTag)
	s.relay.On()
	s.islighton = true
}

func (s *service) lightOffHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]turn off the light", logHandlerTag)
	s.relay.Off()
	s.islighton = false
}

func (s *service) setVolumeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]set volume", logHandlerTag)
	vars := mux.Vars(r)
	v, err := strconv.Atoi(vars["v"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid volume: %v", vars["v"])
		return
	}
	if v < 0 || v > 100 { // volume should be 0~100%
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid volume: %v", vars["v"])
		return
	}

	log.Printf("[%v]set volume: %v%%", logHandlerTag, v)
	if err := util.SetVolume(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "server internal error: %v", err)
		return
	}
	util.PlayWav("current-volume.wav")
}

func (s *service) selfDrivingOnHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]self-driving on", logHandlerTag)
	if !s.cfg.SelfDriving.Enabled {
		log.Printf("[%v]self-driving was disabled", logHandlerTag)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("self-driving was disabled"))
		return
	}
	s.selfdriving.Start()
}

func (s *service) selfDrivingOffHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]self-driving off", logHandlerTag)
	if !s.cfg.SelfDriving.Enabled {
		log.Printf("[%v]self-driving was disabled", logHandlerTag)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("self-driving was disabled"))
		return
	}
	s.selfdriving.Stop()
}

func (s *service) selfNavOnHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]self-nav on", logHandlerTag)
	if !s.cfg.SelfNav.Enabled {
		log.Printf("[%v]self-nav was disabled", logHandlerTag)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("self-nav was disabled"))
		return
	}
	vars := mux.Vars(r)
	lat, err := strconv.ParseFloat(vars["lat"], 64)
	if err != nil {
		log.Printf("[%v]invalid lat: %v", logHandlerTag, vars["lat"])
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid lat: %v", vars["lat"])
		return
	}
	lon, err := strconv.ParseFloat(vars["lon"], 64)
	if err != nil {
		log.Printf("[%v]invalid lon: %v", logHandlerTag, vars["lon"])
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid lat: %v", vars["lon"])
		return
	}
	log.Printf("[%v]destination: %v, %v", logHandlerTag, lat, lon)

	if lat < -90 || lat > 90 {
		log.Printf("[%v]invalid lat: %v", logHandlerTag, lat)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid lat: %v", lat)
		return
	}

	if lon < -180 || lon > 180 {
		log.Printf("[%v]invalid lon: %v", logHandlerTag, lon)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid lon: %v", lon)
		return
	}

	dest := &geo.Point{
		Lat: lat,
		Lon: lon,
	}
	s.selfnav.Start(dest)
}

func (s *service) selfNavOffHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%v]self-nav off", logHandlerTag)
	if !s.cfg.SelfNav.Enabled {
		log.Printf("[%v]self-nav was disabled", logHandlerTag)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("self-nav was disabled"))
		return
	}
	s.selfnav.Stop()
}

func (s *service) musicOnHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[car]music on")
	if s.onMusic {
		return
	}
	s.onMusic = true
	util.PlayMp3("./music/*.mp3")
}

func (s *service) musicOffHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[car]music off")
	if !s.onMusic {
		return
	}
	s.onMusic = false
	util.StopMp3()
}
