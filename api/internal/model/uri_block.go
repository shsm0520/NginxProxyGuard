package model

import "time"

// URIMatchType represents the type of URI pattern matching
type URIMatchType string

const (
	URIMatchExact  URIMatchType = "exact"  // Exact match (nginx: location =)
	URIMatchPrefix URIMatchType = "prefix" // Prefix match (nginx: location ^~)
	URIMatchRegex  URIMatchType = "regex"  // Regex match (nginx: location ~*)
)

// URIBlockRule represents a single URI blocking rule
type URIBlockRule struct {
	ID          string       `json:"id"`
	Pattern     string       `json:"pattern"`
	MatchType   URIMatchType `json:"match_type"`
	Description string       `json:"description,omitempty"`
	Enabled     bool         `json:"enabled"`
}

// URIBlock represents URI blocking settings for a proxy host
type URIBlock struct {
	ID              string         `json:"id"`
	ProxyHostID     string         `json:"proxy_host_id"`
	Enabled         bool           `json:"enabled"`
	Rules           []URIBlockRule `json:"rules"`
	ExceptionIPs    []string       `json:"exception_ips"`
	AllowPrivateIPs bool           `json:"allow_private_ips"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// CreateURIBlockRequest is the request to create/update URI block settings
type CreateURIBlockRequest struct {
	Enabled         *bool          `json:"enabled,omitempty"`
	Rules           []URIBlockRule `json:"rules,omitempty"`
	ExceptionIPs    []string       `json:"exception_ips,omitempty"`
	AllowPrivateIPs *bool          `json:"allow_private_ips,omitempty"`
}

// AddURIBlockRuleRequest is the request to add a single rule (from log viewer)
type AddURIBlockRuleRequest struct {
	Pattern     string       `json:"pattern" validate:"required"`
	MatchType   URIMatchType `json:"match_type" validate:"required,oneof=exact prefix regex"`
	Description string       `json:"description,omitempty"`
	Enabled     *bool        `json:"enabled,omitempty"`
}

// GlobalURIBlock represents global URI blocking settings (applied to all hosts)
type GlobalURIBlock struct {
	ID              string         `json:"id"`
	Enabled         bool           `json:"enabled"`
	Rules           []URIBlockRule `json:"rules"`
	ExceptionIPs    []string       `json:"exception_ips"`
	AllowPrivateIPs bool           `json:"allow_private_ips"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// CreateGlobalURIBlockRequest is the request to create/update global URI block settings
type CreateGlobalURIBlockRequest struct {
	Enabled         *bool          `json:"enabled,omitempty"`
	Rules           []URIBlockRule `json:"rules,omitempty"`
	ExceptionIPs    []string       `json:"exception_ips,omitempty"`
	AllowPrivateIPs *bool          `json:"allow_private_ips,omitempty"`
}

// Common URI patterns for quick blocking
var CommonURIPatterns = []struct {
	Pattern     string
	MatchType   URIMatchType
	Description string
}{
	{"/wp-admin", URIMatchPrefix, "WordPress Admin"},
	{"/wp-login.php", URIMatchExact, "WordPress Login"},
	{"/xmlrpc.php", URIMatchExact, "WordPress XML-RPC"},
	{"/wp-config.php", URIMatchExact, "WordPress Config"},
	{"/.env", URIMatchExact, "Environment File"},
	{"/.git", URIMatchPrefix, "Git Directory"},
	{"/phpmyadmin", URIMatchPrefix, "phpMyAdmin"},
	{"/adminer", URIMatchPrefix, "Adminer"},
	{`\.php$`, URIMatchRegex, "PHP Files"},
	{`\.(sql|bak|backup|old|orig|swp)$`, URIMatchRegex, "Backup Files"},
	{"/cgi-bin", URIMatchPrefix, "CGI-BIN Directory"},
	{"/administrator", URIMatchPrefix, "Joomla Admin"},
}
