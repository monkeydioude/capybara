package capybara

import (
	"log"
	"time"
)

const UPDATE_SERVICE_TIMER = 30 * time.Second

func UpdateServicesRoutine(h *Handler, path string, d time.Duration) {
	timer := time.NewTimer(d)
	<-timer.C

	log.Println("[INFO] Updating services list")
	c := NewConfig(path)
	h.services = c.Services
	UpdateServicesRoutine(h, path, d)
}

func Log(msg string) {
	log.Print(msg)
}
