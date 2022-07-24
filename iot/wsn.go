package iot

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/shanghuiyang/rpi-projects/util"
)

// Wsn is the implement of Cloud
type Wsn struct {
	token string
	api   string
}

// NewWsn ...
func NewWsn(cfg *Config) *Wsn {
	return &Wsn{
		token: cfg.Token,
		api:   cfg.API,
	}
}

// Push ...
func (w *Wsn) Push(v *Value) error {
	var formData url.Values
	api := w.api
	if v.Device == "gps" {
		api = strings.Replace(w.api, "numerical", "gps", -1)
		pt, ok := v.Value.(*util.Point)
		if !ok {
			return fmt.Errorf("failed to convert value to point")
		}
		formData = url.Values{
			"ak":    {w.token},
			"id":    {v.Device},
			"lat":   {fmt.Sprintf("%v", pt.Lat)},
			"lng":   {fmt.Sprintf("%v", pt.Lon)},
			"speed": {"30"},
		}
	} else {
		formData = url.Values{
			"ak":    {w.token},
			"id":    {v.Device},
			"value": {fmt.Sprintf("%v", v.Value)},
		}
	}

	resp, err := http.PostForm(api, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Get ...
func (w *Wsn) Get(params map[string]interface{}) ([]byte, error) {
	return nil, fmt.Errorf("not implement")
}
