package capybara

import (
	"context"
	"log"
	"log/slog"
	"time"
)

const UPDATE_SERVICE_TIMER = 30 * time.Second

// UpdateServicesRoutine re-reads the config every d and swaps the routing table. A config
// that cannot be read, parsed or validated is logged and the current table stays in place.
func UpdateServicesRoutine(ctx context.Context, h *Handler, path string, d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		c, err := LoadConfig(path)
		if err != nil {
			slog.Error("Config reload failed, keeping the current services", "path", path, "error", err.Error())
			continue
		}
		h.SetServices(c.Services)
	}
}

func Log(msg string) {
	log.Print(msg)
}
