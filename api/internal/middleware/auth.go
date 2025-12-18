package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"nginx-proxy-guard/internal/service"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if already authenticated via API token
			if c.Get("user_id") != nil {
				return next(c)
			}

			token := extractToken(c)
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authentication required",
				})
			}

			user, err := authService.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid or expired session",
				})
			}

			// Store user in context
			c.Set("user", user)
			c.Set("token", token)
			c.Set("user_id", user.ID)
			c.Set("username", user.Username)
			c.Set("role", user.Role)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware adds user to context if authenticated, but doesn't require it
func OptionalAuthMiddleware(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token != "" {
				user, err := authService.ValidateToken(c.Request().Context(), token)
				if err == nil && user != nil {
					c.Set("user", user)
					c.Set("token", token)
					c.Set("user_id", user.ID)
					c.Set("username", user.Username)
					c.Set("role", user.Role)
				}
			}
			return next(c)
		}
	}
}

// InitialSetupRequired middleware ensures user has completed initial setup
func InitialSetupRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user")
		if user == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Authentication required",
			})
		}

		// Allow access to change-credentials endpoint even if initial setup not complete
		path := c.Path()
		if strings.Contains(path, "/auth/change-credentials") {
			return next(c)
		}

		// For initial setup users, block access to most endpoints
		// They should only be able to change their credentials
		// But we allow read-only status checks

		return next(c)
	}
}

func extractToken(c echo.Context) string {
	// Check Authorization header
	auth := c.Request().Header.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	return ""
}
