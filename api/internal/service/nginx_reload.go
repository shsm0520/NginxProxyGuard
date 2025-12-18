package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// NginxReloader provides debounced nginx reload functionality
// Multiple reload requests within the debounce window are coalesced into a single reload
type NginxReloader struct {
	nginx        NginxManager
	debounceTime time.Duration
	mu           sync.Mutex
	pending      bool
	timer        *time.Timer
	lastReload   time.Time
	reloadCount  int64
}

// NewNginxReloader creates a new debounced nginx reloader
func NewNginxReloader(nginx NginxManager, debounceTime time.Duration) *NginxReloader {
	if debounceTime == 0 {
		debounceTime = 2 * time.Second // Default 2 second debounce
	}
	return &NginxReloader{
		nginx:        nginx,
		debounceTime: debounceTime,
	}
}

// RequestReload queues a reload request
// Returns immediately - the actual reload happens after the debounce period
func (r *NginxReloader) RequestReload(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pending = true

	// Cancel existing timer
	if r.timer != nil {
		r.timer.Stop()
	}

	// Start new timer
	r.timer = time.AfterFunc(r.debounceTime, func() {
		r.executeReload()
	})

	log.Printf("[NginxReloader] Reload queued, will execute in %v", r.debounceTime)
}

// RequestReloadImmediate performs an immediate reload, bypassing debounce
func (r *NginxReloader) RequestReloadImmediate(ctx context.Context) error {
	r.mu.Lock()
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	r.pending = false
	r.mu.Unlock()

	return r.nginx.ReloadNginx(ctx)
}

// executeReload performs the actual nginx reload
func (r *NginxReloader) executeReload() {
	r.mu.Lock()
	if !r.pending {
		r.mu.Unlock()
		return
	}
	r.pending = false
	r.timer = nil
	r.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("[NginxReloader] Executing debounced nginx reload")
	if err := r.nginx.ReloadNginx(ctx); err != nil {
		log.Printf("[NginxReloader] Reload failed: %v", err)
	} else {
		r.mu.Lock()
		r.lastReload = time.Now()
		r.reloadCount++
		r.mu.Unlock()
		log.Println("[NginxReloader] Reload completed successfully")
	}
}

// IsPending returns whether a reload is pending
func (r *NginxReloader) IsPending() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.pending
}

// GetStats returns reload statistics
func (r *NginxReloader) GetStats() map[string]interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	return map[string]interface{}{
		"pending":      r.pending,
		"last_reload":  r.lastReload,
		"reload_count": r.reloadCount,
	}
}

// Flush forces any pending reload to execute immediately
func (r *NginxReloader) Flush(ctx context.Context) error {
	r.mu.Lock()
	pending := r.pending
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	r.pending = false
	r.mu.Unlock()

	if pending {
		log.Println("[NginxReloader] Flushing pending reload")
		return r.nginx.ReloadNginx(ctx)
	}
	return nil
}
