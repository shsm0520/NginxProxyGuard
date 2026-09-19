package handler

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"nginx-proxy-guard/internal/service"
)

// LogFileInfo represents information about a log file
type LogFileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModifiedAt   time.Time `json:"modified_at"`
	IsCompressed bool      `json:"is_compressed"`
	LogType      string    `json:"log_type"` // access, error
}

// LogFilesResponse represents the response for log files listing
type LogFilesResponse struct {
	Files      []LogFileInfo `json:"files"`
	TotalSize  int64         `json:"total_size"`
	TotalCount int           `json:"total_count"`
	RawLogEnabled bool       `json:"raw_log_enabled"`
}

// ListLogFiles returns a list of nginx log files
func (h *SystemSettingsHandler) ListLogFiles(c echo.Context) error {
	settings, err := h.repo.Get(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	response := LogFilesResponse{
		Files:         []LogFileInfo{},
		RawLogEnabled: settings.RawLogEnabled,
	}

	// Read log files from directory
	entries, err := os.ReadDir(nginxLogsPath)
	if err != nil {
		// If directory doesn't exist or is not readable, return empty
		return c.JSON(http.StatusOK, response)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip symlinks (stdout/stderr)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if it's a symlink
		fullPath := filepath.Join(nginxLogsPath, name)
		if linkInfo, err := os.Lstat(fullPath); err == nil {
			if linkInfo.Mode()&os.ModeSymlink != 0 {
				// Skip symlinks to /dev/stdout or /dev/stderr
				continue
			}
		}

		// Determine log type
		logType := "unknown"
		if strings.HasPrefix(name, "access") {
			logType = "access"
		} else if strings.HasPrefix(name, "error") {
			logType = "error"
		}

		// Check if compressed
		isCompressed := strings.HasSuffix(name, ".gz") || strings.HasSuffix(name, ".bz2")

		fileInfo := LogFileInfo{
			Name:         name,
			Size:         info.Size(),
			ModifiedAt:   info.ModTime(),
			IsCompressed: isCompressed,
			LogType:      logType,
		}

		response.Files = append(response.Files, fileInfo)
		response.TotalSize += info.Size()
	}

	// Sort by modification time (newest first)
	sort.Slice(response.Files, func(i, j int) bool {
		return response.Files[i].ModifiedAt.After(response.Files[j].ModifiedAt)
	})

	response.TotalCount = len(response.Files)

	return c.JSON(http.StatusOK, response)
}

// DownloadLogFile downloads a specific log file
func (h *SystemSettingsHandler) DownloadLogFile(c echo.Context) error {
	filename := c.Param("filename")

	// Validate filename (prevent path traversal)
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid filename"})
	}

	filePath := filepath.Join(nginxLogsPath, filename)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Don't allow downloading symlinks
	if linkInfo, err := os.Lstat(filePath); err == nil {
		if linkInfo.Mode()&os.ModeSymlink != 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot download symlink"})
		}
	}

	// Audit log
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.audit.LogSettingsUpdate(auditCtx, "로그 파일", map[string]interface{}{
		"action":   "download",
		"filename": filename,
	})

	c.Response().Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	return c.Attachment(filePath, filename)
}

// DeleteLogFile deletes a specific log file
func (h *SystemSettingsHandler) DeleteLogFile(c echo.Context) error {
	filename := c.Param("filename")

	// Validate filename (prevent path traversal)
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid filename"})
	}

	// Don't allow deleting main log files (access.log, error.log)
	if filename == "access.log" || filename == "error.log" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot delete active log file"})
	}

	filePath := filepath.Join(nginxLogsPath, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Don't allow deleting symlinks
	if linkInfo, err := os.Lstat(filePath); err == nil {
		if linkInfo.Mode()&os.ModeSymlink != 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot delete symlink"})
		}
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Audit log
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.audit.LogSettingsUpdate(auditCtx, "로그 파일", map[string]interface{}{
		"action":   "delete",
		"filename": filename,
	})

	return c.NoContent(http.StatusNoContent)
}

// ViewLogFile returns the last N lines of a log file (for preview)
func (h *SystemSettingsHandler) ViewLogFile(c echo.Context) error {
	filename := c.Param("filename")

	// Validate filename
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid filename"})
	}

	// Get line count parameter
	lines := 100
	if linesParam := c.QueryParam("lines"); linesParam != "" {
		if n, err := strconv.Atoi(linesParam); err == nil && n > 0 && n <= 1000 {
			lines = n
		}
	}

	filePath := filepath.Join(nginxLogsPath, filename)

	// Check if file is compressed
	if strings.HasSuffix(filename, ".gz") {
		// For compressed files, use zcat
		cmd := exec.Command("zcat", filePath)
		tailCmd := exec.Command("tail", "-n", strconv.Itoa(lines))

		pipe, err := cmd.StdoutPipe()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		tailCmd.Stdin = pipe

		cmd.Start()
		output, err := tailCmd.Output()
		cmd.Wait()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to read compressed file"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"filename": filename,
			"lines":    lines,
			"content":  string(output),
		})
	}

	// For regular files, read directly
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer file.Close()

	// Read last N lines using tail command
	cmd := exec.Command("tail", "-n", strconv.Itoa(lines), filePath)
	output, err := cmd.Output()
	if err != nil {
		// Fallback: read the whole file if tail fails
		content, err := io.ReadAll(file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		output = content
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"filename": filename,
		"lines":    lines,
		"content":  string(output),
	})
}

// TriggerLogRotation manually triggers log rotation
func (h *SystemSettingsHandler) TriggerLogRotation(c echo.Context) error {
	// Get settings
	settings, err := h.repo.Get(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if !settings.RawLogEnabled {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "raw log files are not enabled",
		})
	}

	// Generate logrotate config first
	if err := h.generateLogrotateConfig(settings); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Rotate through the nginx manager: logrotate lives in the nginx container,
	// not here (#301). The scheduler uses the same call so the two cannot drift.
	if h.nginxManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "nginx manager is not available",
		})
	}

	if err := h.nginxManager.RotateLogs(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "logrotate failed",
			"details": err.Error(),
		})
	}

	// Audit log
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.audit.LogSettingsUpdate(auditCtx, "로그 파일", map[string]interface{}{
		"action": "rotate",
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "completed",
		"message": "Log rotation completed successfully",
	})
}

// GetSystemLogConfig returns the current system log configuration
func (h *SystemSettingsHandler) GetSystemLogConfig(c echo.Context) error {
	if h.dockerLogCollector == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Docker log collector is not enabled"})
	}

	config := h.dockerLogCollector.GetConfig()
	return c.JSON(http.StatusOK, config)
}

// UpdateSystemLogConfig updates system log configuration
func (h *SystemSettingsHandler) UpdateSystemLogConfig(c echo.Context) error {
	if h.dockerLogCollector == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Docker log collector is not enabled"})
	}

	var config service.SystemLogConfig
	if err := c.Bind(&config); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.dockerLogCollector.UpdateConfig(config); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update config: " + err.Error()})
	}

	// Audit log
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.audit.LogSettingsUpdate(auditCtx, "시스템 로그 설정", map[string]interface{}{
		"action": "update",
	})

	return c.JSON(http.StatusOK, config)
}
