package model

import (
	"time"
)

// CloudProvider represents a cloud service provider with its IP ranges
type CloudProvider struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Region       string     `json:"region"`
	Description  string     `json:"description,omitempty"`
	IPRanges     []string   `json:"ip_ranges"`
	IPRangesURL  string     `json:"ip_ranges_url,omitempty"`
	LastUpdated  *time.Time `json:"last_updated,omitempty"`
	Enabled      bool       `json:"enabled"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CloudProviderSummary is a lightweight version for listing
type CloudProviderSummary struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Description string `json:"description,omitempty"`
	IPCount     int    `json:"ip_count"` // Number of IP ranges
	Enabled     bool   `json:"enabled"`
}

// CloudProvidersByRegion groups providers by region
type CloudProvidersByRegion struct {
	US []CloudProviderSummary `json:"us"`
	EU []CloudProviderSummary `json:"eu"`
	CN []CloudProviderSummary `json:"cn"`
	KR []CloudProviderSummary `json:"kr"`
}

// UpdateCloudProviderRequest for updating cloud provider
type UpdateCloudProviderRequest struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	IPRanges    *[]string `json:"ip_ranges,omitempty"`
	IPRangesURL *string   `json:"ip_ranges_url,omitempty"`
	Enabled     *bool     `json:"enabled,omitempty"`
}

// CreateCloudProviderRequest for creating a custom cloud provider
type CreateCloudProviderRequest struct {
	Name        string   `json:"name" validate:"required"`
	Slug        string   `json:"slug" validate:"required"`
	Region      string   `json:"region" validate:"required,oneof=us eu cn kr other"`
	Description string   `json:"description,omitempty"`
	IPRanges    []string `json:"ip_ranges" validate:"required,min=1"`
	IPRangesURL string   `json:"ip_ranges_url,omitempty"`
}

// BlockedCloudProvidersRequest for updating blocked providers on a proxy host
type BlockedCloudProvidersRequest struct {
	BlockedProviders []string `json:"blocked_providers"` // List of provider slugs
	ChallengeMode    bool     `json:"challenge_mode"`    // Show challenge instead of blocking
}

// RegionInfo provides metadata about regions
var RegionInfo = map[string]struct {
	Name    string
	Flag    string
	Country string
}{
	"us":    {Name: "United States", Flag: "🇺🇸", Country: "US"},
	"eu":    {Name: "Europe", Flag: "🇪🇺", Country: "EU"},
	"cn":    {Name: "China", Flag: "🇨🇳", Country: "CN"},
	"kr":    {Name: "South Korea", Flag: "🇰🇷", Country: "KR"},
	"other": {Name: "Other", Flag: "🌍", Country: ""},
}
