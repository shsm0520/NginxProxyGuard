package scheduler

import (
	"context"
	"errors"
	"log"
	"time"

	"nginx-proxy-guard/internal/nginx"
)

// LogRotateScheduler runs logrotate daily for raw nginx logs
type LogRotateScheduler struct {
	stopCh chan struct{}
	nginx  *nginx.Manager
}

func NewLogRotateScheduler(nginxManager *nginx.Manager) *LogRotateScheduler {
	return &LogRotateScheduler{
		stopCh: make(chan struct{}),
		nginx:  nginxManager,
	}
}

func (s *LogRotateScheduler) Start() {
	log.Println("[LogRotateScheduler] Started (runs daily at midnight)")

	go func() {
		// Calculate time until next midnight
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		initialDelay := nextMidnight.Sub(now)

		// Wait until midnight
		select {
		case <-time.After(initialDelay):
		case <-s.stopCh:
			return
		}

		// Run initial rotation
		s.runLogrotate()

		// Then run daily
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runLogrotate()
			case <-s.stopCh:
				log.Println("[LogRotateScheduler] Stopped")
				return
			}
		}
	}()
}

func (s *LogRotateScheduler) Stop() {
	close(s.stopCh)
}

// runLogrotate delegates to the nginx manager, which owns every docker exec
// into the proxy container. The manual trigger (TriggerLogRotation handler)
// goes through the same call so the two paths cannot drift again (#301).
func (s *LogRotateScheduler) runLogrotate() {
	if s.nginx == nil {
		log.Println("[LogRotateScheduler] Nginx manager unavailable, skipping")
		return
	}

	if err := s.nginx.RotateLogs(context.Background()); err != nil {
		// Config absent means raw log files were never enabled — not a failure.
		if errors.Is(err, nginx.ErrLogrotateConfigMissing) {
			log.Println("[LogRotateScheduler] Logrotate config not found, skipping")
			return
		}
		log.Printf("[LogRotateScheduler] Logrotate failed: %v", err)
		return
	}

	log.Println("[LogRotateScheduler] Log rotation completed successfully")
}

// RunNow triggers immediate log rotation (for manual trigger)
func (s *LogRotateScheduler) RunNow() error {
	s.runLogrotate()
	return nil
}
