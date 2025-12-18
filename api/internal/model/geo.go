package model

import "time"

// GeoRestriction represents geo-blocking settings for a proxy host
type GeoRestriction struct {
	ID              string    `json:"id"`
	ProxyHostID     string    `json:"proxy_host_id"`
	Mode            string    `json:"mode"`              // "whitelist" or "blacklist"
	Countries       []string  `json:"countries"`         // ISO 3166-1 alpha-2 country codes
	AllowedIPs      []string  `json:"allowed_ips"`       // IPs/CIDRs that bypass geo restrictions
	AllowPrivateIPs bool      `json:"allow_private_ips"` // Allow private IP ranges (10.x, 172.16-31.x, 192.168.x)
	AllowSearchBots bool      `json:"allow_search_bots"` // Allow search engine bots (Googlebot, Bingbot, etc.)
	Enabled         bool      `json:"enabled"`
	ChallengeMode   bool      `json:"challenge_mode"`    // Show CAPTCHA instead of blocking
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateGeoRestrictionRequest is the request to create/update geo restriction
type CreateGeoRestrictionRequest struct {
	Mode            string   `json:"mode" validate:"required,oneof=whitelist blacklist"`
	Countries       []string `json:"countries" validate:"required,min=1"`
	AllowedIPs      []string `json:"allowed_ips,omitempty"`       // IPs/CIDRs that bypass geo restrictions
	AllowPrivateIPs *bool    `json:"allow_private_ips,omitempty"` // Allow private IP ranges
	AllowSearchBots *bool    `json:"allow_search_bots,omitempty"` // Allow search engine bots
	Enabled         *bool    `json:"enabled,omitempty"`
	ChallengeMode   *bool    `json:"challenge_mode,omitempty"`    // Show CAPTCHA instead of blocking
}

// UpdateGeoRestrictionRequest is the request to update geo restriction
type UpdateGeoRestrictionRequest struct {
	Mode            *string  `json:"mode,omitempty"`
	Countries       []string `json:"countries,omitempty"`
	AllowedIPs      []string `json:"allowed_ips,omitempty"`       // IPs/CIDRs that bypass geo restrictions
	AllowPrivateIPs *bool    `json:"allow_private_ips,omitempty"` // Allow private IP ranges
	AllowSearchBots *bool    `json:"allow_search_bots,omitempty"` // Allow search engine bots
	Enabled         *bool    `json:"enabled,omitempty"`
	ChallengeMode   *bool    `json:"challenge_mode,omitempty"`    // Show CAPTCHA instead of blocking
}

// Common country codes for reference
var CommonCountryCodes = map[string]string{
	"US": "United States",
	"CN": "China",
	"RU": "Russia",
	"KR": "South Korea",
	"JP": "Japan",
	"DE": "Germany",
	"FR": "France",
	"GB": "United Kingdom",
	"IN": "India",
	"BR": "Brazil",
	"CA": "Canada",
	"AU": "Australia",
	"NL": "Netherlands",
	"IT": "Italy",
	"ES": "Spain",
	"PL": "Poland",
	"UA": "Ukraine",
	"ID": "Indonesia",
	"TW": "Taiwan",
	"HK": "Hong Kong",
	"SG": "Singapore",
	"VN": "Vietnam",
	"TH": "Thailand",
	"PH": "Philippines",
	"MY": "Malaysia",
	"IR": "Iran",
	"IQ": "Iraq",
	"KP": "North Korea",
}
