package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"nginx-proxy-guard/internal/config"
	"nginx-proxy-guard/internal/model"
	"nginx-proxy-guard/internal/service"
)

// GeoIP database status, manual and scheduled updates, and update history.
//
// Split out of system_settings.go for the same reason as the raw-log half: the
// file was past the 500-line handler ceiling. Pure move.

// GetGeoIPStatus returns the current GeoIP status
func (h *SystemSettingsHandler) GetGeoIPStatus(c echo.Context) error {
	settings, err := h.repo.Get(c.Request().Context())
	if err != nil {
		return directInternalError(c, err)
	}

	status := &model.GeoIPStatus{
		Enabled:         settings.GeoIPEnabled,
		LastUpdated:     settings.GeoIPLastUpdated,
		DatabaseVersion: settings.GeoIPDatabaseVersion,
	}

	// Check if GeoIP databases exist
	geoipPath := "/etc/nginx/geoip"
	countryDB := filepath.Join(geoipPath, "GeoLite2-Country.mmdb")
	asnDB := filepath.Join(geoipPath, "GeoLite2-ASN.mmdb")

	if info, err := os.Stat(countryDB); err == nil && info.Size() > config.MinGeoIPDatabaseSize {
		status.CountryDB = true
	}
	if info, err := os.Stat(asnDB); err == nil && info.Size() > config.MinGeoIPDatabaseSize {
		status.ASNDB = true
	}

	// Check .no-geoip flag
	noGeoipFlag := filepath.Join(geoipPath, ".no-geoip")
	if _, err := os.Stat(noGeoipFlag); err == nil {
		status.Status = config.StatusDisabled
	} else if status.CountryDB && status.ASNDB {
		status.Status = config.StatusOK
	} else if settings.GeoIPEnabled && settings.MaxmindLicenseKey != "" {
		status.Status = config.StatusError
		status.ErrorMessage = "GeoIP databases not found"
	} else {
		status.Status = config.StatusDisabled
	}

	// Calculate next update time
	if settings.GeoIPAutoUpdate && settings.GeoIPLastUpdated != nil {
		interval := parseInterval(settings.GeoIPUpdateInterval)
		nextUpdate := settings.GeoIPLastUpdated.Add(interval)
		status.NextUpdate = &nextUpdate
	}

	return c.JSON(http.StatusOK, status)
}

// CheckUpdate reports whether a newer NginxProxyGuard release is available (#190).
// Display + guidance only — NPG does not update itself. Pass ?force=true to bypass
// the cache (manual re-check). Always returns 200; fetch failures are conveyed via
// the check_failed flag so the UI degrades gracefully.
func (h *SystemSettingsHandler) CheckUpdate(c echo.Context) error {
	force := c.QueryParam("force") == "true"
	info := h.updateChecker.Check(c.Request().Context(), force)
	return c.JSON(http.StatusOK, info)
}

// UpdateGeoIPDatabases triggers a GeoIP database update
func (h *SystemSettingsHandler) UpdateGeoIPDatabases(c echo.Context) error {
	var req model.GeoIPUpdateRequest
	c.Bind(&req) // Ignore bind errors, force is optional

	ctx := c.Request().Context()

	// Get credentials from settings
	licenseKey, accountID, err := h.repo.GetGeoIPCredentials(ctx)
	if err != nil {
		return directInternalError(c, err)
	}

	if licenseKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "MaxMind license key not configured. Please set it in System Settings.",
		})
	}
	// geoipupdate 7.x requires a numeric Account ID too — reject early with a clear
	// message so a missing/invalid Account ID isn't mistaken for a bad license key. (#184)
	if !service.ValidMaxMindAccountID(accountID) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "MaxMind Account ID is missing or invalid. geoipupdate needs both a numeric Account ID and a License Key — set the Account ID in System Settings → GeoIP.",
		})
	}

	// Run GeoIP update in background using scheduler
	// Cloud provider seeding is now handled by the scheduler's RunUpdate
	go func() {
		bgCtx := context.Background()

		if h.geoipScheduler != nil {
			h.geoipScheduler.RunUpdate(bgCtx, model.GeoIPTriggerManual, licenseKey, accountID)
		} else {
			// Fallback to direct update if scheduler not available
			if err := h.runGeoIPUpdate(bgCtx, licenseKey, accountID); err != nil {
				log.Printf("[GeoIP] Update failed: %v", err)
			} else {
				h.repo.UpdateGeoIPStatus(bgCtx, time.Now(), "GeoLite2")
				if h.nginxManager != nil {
					h.nginxManager.ReloadNginx(bgCtx)
				}
			}

			// Seed cloud providers in fallback path
			if h.cloudProviderService != nil {
				if err := h.cloudProviderService.SeedDefaultProviders(bgCtx); err != nil {
					log.Printf("[GeoIP] Failed to seed cloud providers: %v", err)
				}
			}
		}
	}()

	// Audit log
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.audit.LogSettingsUpdate(auditCtx, "GeoIP 데이터베이스", map[string]interface{}{
		"action": "update_triggered",
	})

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"status":  "updating",
		"message": "GeoIP database update has been triggered. This may take a few minutes.",
	})
}

// GetGeoIPHistory returns the GeoIP update history
func (h *SystemSettingsHandler) GetGeoIPHistory(c echo.Context) error {
	if h.historyRepo == nil {
		return c.JSON(http.StatusOK, model.GeoIPUpdateHistoryResponse{
			Data:       []model.GeoIPUpdateHistory{},
			Total:      0,
			Page:       1,
			PerPage:    20,
			TotalPages: 1,
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	history, err := h.historyRepo.List(c.Request().Context(), page, perPage)
	if err != nil {
		return directInternalError(c, err)
	}

	return c.JSON(http.StatusOK, history)
}

// runGeoIPUpdate runs the geoipupdate command
func (h *SystemSettingsHandler) runGeoIPUpdate(ctx context.Context, licenseKey, accountID string) error {
	geoipPath := "/etc/nginx/geoip"
	confPath := filepath.Join(geoipPath, "GeoIP.conf")

	// Create GeoIP.conf
	conf := fmt.Sprintf(`# GeoIP.conf for Nginx Proxy Guard
# Auto-generated by API

AccountID %s
LicenseKey %s
EditionIDs GeoLite2-Country GeoLite2-ASN
DatabaseDirectory %s
`, accountID, licenseKey, geoipPath)

	if err := os.WriteFile(confPath, []byte(conf), 0600); err != nil {
		return fmt.Errorf("failed to write GeoIP.conf: %w", err)
	}

	// Run geoipupdate
	cmd := exec.CommandContext(ctx, "geoipupdate", "-f", confPath, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("geoipupdate failed: %s", service.ExtractGeoIPUpdateError(string(output), err))
	}

	// Update the symlink to use the enabled config
	enabledConf := filepath.Join(geoipPath, "geoip-enabled.conf")
	activeConf := filepath.Join(geoipPath, "geoip-active.conf")

	// Remove .no-geoip flag if exists
	os.Remove(filepath.Join(geoipPath, ".no-geoip"))

	// Update symlink
	os.Remove(activeConf)
	if err := os.Symlink(enabledConf, activeConf); err != nil {
		return fmt.Errorf("failed to update geoip symlink: %w", err)
	}

	return nil
}

// parseInterval parses interval string like "1d", "7d", "30d" to duration
func parseInterval(interval string) time.Duration {
	interval = strings.ToLower(strings.TrimSpace(interval))

	if strings.HasSuffix(interval, "d") {
		days := 7 // default
		fmt.Sscanf(interval, "%dd", &days)
		return time.Duration(days) * 24 * time.Hour
	}
	if strings.HasSuffix(interval, "h") {
		hours := 168 // default 7 days
		fmt.Sscanf(interval, "%dh", &hours)
		return time.Duration(hours) * time.Hour
	}

	// Default to 7 days
	return 7 * 24 * time.Hour
}
