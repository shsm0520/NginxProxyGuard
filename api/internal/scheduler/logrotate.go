package scheduler

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"
)

// LogRotateScheduler runs logrotate daily for raw nginx logs
type LogRotateScheduler struct {
	stopCh         chan struct{}
	nginxContainer string
}

func NewLogRotateScheduler() *LogRotateScheduler {
	container := os.Getenv("NGINX_CONTAINER")
	if container == "" {
		container = "npg-proxy"
	}
	return &LogRotateScheduler{
		stopCh:         make(chan struct{}),
		nginxContainer: container,
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

func (s *LogRotateScheduler) runLogrotate() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Check if logrotate config exists
	configPath := "/etc/nginx/conf.d/.logrotate.conf"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Println("[LogRotateScheduler] Logrotate config not found, skipping")
		return
	}

	// Copy config to nginx container's logrotate.d and run logrotate
	cmd := exec.CommandContext(ctx, "docker", "exec", s.nginxContainer,
		"sh", "-c", "cp /etc/nginx/conf.d/.logrotate.conf /etc/logrotate.d/nginx-guard && logrotate -f /etc/logrotate.d/nginx-guard")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[LogRotateScheduler] Logrotate failed: %v - %s", err, string(output))
		return
	}

	log.Println("[LogRotateScheduler] Log rotation completed successfully")
}

// RunNow triggers immediate log rotation (for manual trigger)
func (s *LogRotateScheduler) RunNow() error {
	s.runLogrotate()
	return nil
}
