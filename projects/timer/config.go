package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/robfig/cron/v3"
)

type config struct {
	Buzzer   uint8      `json:"buzzer"`
	Button   uint8      `json:"button"`
	Cronjobs []*cronjob `json:"cronjobs"`
}

type cronjob struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	cron      *cron.Cron
}

func loadConfig(file string) (*config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
