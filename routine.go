package main

import (
	"log"
	"time"
)

const UPDATE_SERVICE_TIMER = 30 * time.Second

func updateServicesRoutine(h *handler, path string) {
	timer := time.NewTimer(UPDATE_SERVICE_TIMER)
	<-timer.C

	log.Println("[INFO] Updating services list")
	c := newConfig(path)
	h.s = c.Services
	updateServicesRoutine(h, path)
}
