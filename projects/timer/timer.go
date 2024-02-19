package main

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/shanghuiyang/rpi-devices/dev"
	"github.com/shanghuiyang/rpi-projects/util"
)

type timer struct {
	buz  dev.Buzzer
	btn  dev.Button
	jobs []*cronjob
}

func newTimer(buz dev.Buzzer, btn dev.Button, jobs []*cronjob) *timer {
	return &timer{
		buz:  buz,
		btn:  btn,
		jobs: jobs,
	}
}

func (t *timer) start() {
	for i, job := range t.jobs {
		c := cron.New()
		c.AddFunc(job.Schedule, t.doJob)
		c.Start()
		t.jobs[i].cron = c
		log.Printf("cron job started: %v", job.Name)
	}
}

func (t *timer) doJob() {
	trigTime := time.Now()
	alerting := true
	go func() {
		for alerting {
			t.buz.Beep(1, 200)
		}
	}()

	for {
		if t.btn.Pressed() {
			alerting = false
			break
		}
		timeout := time.Since(trigTime).Seconds() > 60
		if timeout && alerting {
			log.Printf("cron job alert timeout, stop alert")
			alerting = false
			break
		}
		util.DelayMs(100)
	}
}

func (t *timer) stop() {
	t.buz.Off()
	for _, job := range t.jobs {
		job.cron.Stop()
		log.Printf("cron job stopped: %v", job.Name)
	}
}
