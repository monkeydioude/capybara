package capybara

import (
	"log"
	"sync"
	"time"
)

const UPDATE_SERVICE_TIMER = 30 * time.Second

var mutex sync.Mutex

func UpdateServicesRoutine(h *Handler, path string, d time.Duration) {
	for {
		timer := time.NewTimer(d)
		<-timer.C

		c := NewConfig(path)
		mutex.Lock()
		h.services = c.Services
		mutex.Unlock()
	}
}

func Log(msg string) {
	log.Print(msg)
}
