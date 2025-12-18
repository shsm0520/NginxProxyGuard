package handler

import (
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Validation constants
const (
	// Pagination limits
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
	MinPerPage     = 1

	// String field length limits
	MaxNameLength        = 255
	MaxDescriptionLength = 2000
	MaxReasonLength      = 1000
	MaxURLLength         = 2048
	MaxDomainLength      = 253  // RFC 1035 max domain length
	MaxIPLength          = 45   // IPv6 max length
	MaxPathLength        = 4096 // URL path max length

	// Supported values
	LanguageKorean  = "ko"
	LanguageEnglish = "en"
)

// SupportedLanguages defines all supported language codes
var SupportedLanguages = []string{LanguageKorean, LanguageEnglish}

// domainRegex validates domain name format (RFC 1035)
var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

// ParsePaginationParams extracts and validates pagination parameters from request
func ParsePaginationParams(c echo.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	if page < DefaultPage {
		page = DefaultPage
	}

	perPage, _ = strconv.Atoi(c.QueryParam("per_page"))
	if perPage < MinPerPage || perPage > MaxPerPage {
		perPage = DefaultPerPage
	}

	return page, perPage
}

// ParsePaginationParamsWithDefaults extracts pagination with custom defaults
func ParsePaginationParamsWithDefaults(c echo.Context, defaultPerPage int) (page, perPage int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	if page < DefaultPage {
		page = DefaultPage
	}

	perPage, _ = strconv.Atoi(c.QueryParam("per_page"))
	if perPage < MinPerPage || perPage > MaxPerPage {
		perPage = defaultPerPage
	}

	return page, perPage
}

// ValidateStringLength checks if a string is within the maximum length
func ValidateStringLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return &ValidationError{
			Field:   fieldName,
			Message: "exceeds maximum length of " + strconv.Itoa(maxLength),
		}
	}
	return nil
}

// ValidateRequired checks if a required field is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: "is required",
		}
	}
	return nil
}

// ValidateDomainName validates a domain name format
func ValidateDomainName(domain string) bool {
	if len(domain) == 0 || len(domain) > MaxDomainLength {
		return false
	}
	// Allow wildcards
	if strings.HasPrefix(domain, "*.") {
		domain = domain[2:]
	}
	return domainRegex.MatchString(domain)
}

// ValidateHostnameOrIP validates that a string is either a valid hostname or IP address
func ValidateHostnameOrIP(host string) bool {
	if len(host) == 0 || len(host) > MaxDomainLength {
		return false
	}

	// Check if it's an IP address
	if net.ParseIP(host) != nil {
		return true
	}

	// Check if it's a valid hostname
	return ValidateDomainName(host)
}

// ValidateURL validates a URL format
func ValidateURL(urlStr string) bool {
	if len(urlStr) == 0 || len(urlStr) > MaxURLLength {
		return false
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Must have scheme and host
	return u.Scheme != "" && u.Host != ""
}

// ValidateLanguage checks if the language code is supported
func ValidateLanguage(lang string) bool {
	for _, supported := range SupportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Message
}

// ValidatePort checks if a port number is valid
func ValidatePort(port int) bool {
	return port >= 1 && port <= 65535
}

// SanitizeString trims whitespace and limits length
func SanitizeString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

// CalculateTotalPages calculates the total number of pages for pagination
func CalculateTotalPages(total, perPage int) int {
	if perPage <= 0 {
		perPage = DefaultPerPage
	}
	return (total + perPage - 1) / perPage
}

// GetHostDisplayName returns the first domain name of a host or a fallback value
func GetHostDisplayName(domainNames []string, fallback string) string {
	if len(domainNames) > 0 {
		return domainNames[0]
	}
	return fallback
}

// UserInfo holds extracted user information from context
type UserInfo struct {
	ID    *string
	Email string
}

// ExtractUserInfo extracts user ID and email from echo context
func ExtractUserInfo(c echo.Context) UserInfo {
	var info UserInfo

	if uid, ok := c.Get("user_id").(string); ok && uid != "" {
		info.ID = &uid
	}

	if email, ok := c.Get("username").(string); ok {
		info.Email = email
	}

	return info
}
