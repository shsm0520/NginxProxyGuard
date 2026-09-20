package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"nginx-proxy-guard/internal/model"
	"nginx-proxy-guard/internal/service"
)

// Verify2FA completes login with 2FA code
func (h *AuthHandler) Verify2FA(c echo.Context) error {
	var req model.Verify2FARequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.TempToken == "" || req.TOTPCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Temporary token and TOTP code are required",
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	resp, err := h.authService.Verify2FA(c.Request().Context(), &req, ip)
	if err != nil {
		switch err {
		case service.ErrInvalidTempToken:
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid or expired temporary token",
			})
		case service.ErrInvalid2FACode:
			// Audited so a rejection leaves a trace. Without it the three
			// causes — rotated secret, clock drift, wrong code — are
			// indistinguishable to everyone, including us (#305).
			h.auditService.Log2FAFailed(c.Request().Context(), "", ip, userAgent, "login", "code rejected")
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid 2FA code",
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "2FA verification failed",
			})
		}
	}

	// Log successful login after 2FA verification
	if resp.User != nil {
		// For login, set user info directly since middleware hasn't set it yet
		ctx := c.Request().Context()
		ctx = context.WithValue(ctx, "user_id", resp.User.ID)
		ctx = context.WithValue(ctx, "username", resp.User.Username)
		ctx = context.WithValue(ctx, "client_ip", ip)
		ctx = context.WithValue(ctx, "user_agent", userAgent)
		h.auditService.LogUserLogin(ctx, resp.User.Username, ip, userAgent)
	}

	return c.JSON(http.StatusOK, resp)
}

// GetAccountInfo returns account information
func (h *AuthHandler) GetAccountInfo(c echo.Context) error {
	user, ok := getUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}

	info, err := h.authService.GetAccountInfo(c.Request().Context(), user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get account info",
		})
	}

	return c.JSON(http.StatusOK, info)
}

// Setup2FA initiates 2FA setup
func (h *AuthHandler) Setup2FA(c echo.Context) error {
	user, ok := getUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}

	resp, err := h.authService.Setup2FA(c.Request().Context(), user.ID)
	if err != nil {
		switch err {
		case service.Err2FAAlreadyEnabled:
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "2FA is already enabled",
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to setup 2FA",
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// Enable2FA enables 2FA after verifying code
func (h *AuthHandler) Enable2FA(c echo.Context) error {
	user, ok := getUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}

	var req model.Enable2FARequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.TOTPCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "TOTP code is required",
		})
	}

	backupCodes, err := h.authService.Enable2FA(c.Request().Context(), user.ID, &req)
	if err != nil {
		switch err {
		case service.Err2FAAlreadyEnabled:
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "2FA is already enabled",
			})
		case service.ErrSetupNotInitiated:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Start 2FA setup before enabling it",
			})
		case service.ErrInvalid2FACode:
			h.auditService.Log2FAFailed(c.Request().Context(), user.Username,
				c.RealIP(), c.Request().UserAgent(), "enable", "code rejected")
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid 2FA code",
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to enable 2FA",
			})
		}
	}

	// Log 2FA enabled
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.auditService.Log2FAEnabled(auditCtx, user.Username)

	// The only time the backup codes are ever returned in plaintext — the
	// database keeps bcrypt hashes. They come from the same write that enabled
	// 2FA, so what is shown here is what will actually work.
	return c.JSON(http.StatusOK, model.Enable2FAResponse{
		Message:     "2FA enabled successfully",
		BackupCodes: backupCodes,
	})
}

// Disable2FA disables 2FA
func (h *AuthHandler) Disable2FA(c echo.Context) error {
	user, ok := getUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}

	var req model.Disable2FARequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Password == "" || req.TOTPCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Password and TOTP code are required",
		})
	}

	err := h.authService.Disable2FA(c.Request().Context(), user.ID, &req)
	if err != nil {
		switch err {
		case service.Err2FANotEnabled:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "2FA is not enabled",
			})
		case service.ErrInvalidCredentials:
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid password",
			})
		case service.ErrInvalid2FACode:
			h.auditService.Log2FAFailed(c.Request().Context(), user.Username,
				c.RealIP(), c.Request().UserAgent(), "disable", "code rejected")
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid 2FA code",
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to disable 2FA",
			})
		}
	}

	// Log 2FA disabled
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.auditService.Log2FADisabled(auditCtx, user.Username)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA disabled successfully",
	})
}
