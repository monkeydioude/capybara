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
		if err := reloadServices(h, path); err != nil {
			slog.Error("Config reload failed, keeping the current services", "path", path, "error", err.Error())
		}
	}
}

// reloadServices loads path and swaps the routing table; on any error the table is untouched.
func reloadServices(h *Handler, path string) error {
	c, err := LoadConfig(path)
	if err != nil {
		return err
	}
	h.SetServices(c.Services)
	return nil
}

func Log(msg string) {
	log.Print(msg)
}
