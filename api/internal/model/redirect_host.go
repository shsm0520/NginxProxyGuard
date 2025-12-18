package model

import (
	"encoding/json"
	"time"
)

// RedirectHost represents a redirect-only virtual host
type RedirectHost struct {
	ID                string          `json:"id"`
	DomainNames       []string        `json:"domain_names"`
	ForwardScheme     string          `json:"forward_scheme"`      // auto, http, https
	ForwardDomainName string          `json:"forward_domain_name"` // Target domain
	ForwardPath       string          `json:"forward_path"`        // Optional path to append
	PreservePath      bool            `json:"preserve_path"`       // Keep original path
	RedirectCode      int             `json:"redirect_code"`       // 301, 302, 307, 308
	SSLEnabled        bool            `json:"ssl_enabled"`
	CertificateID     *string         `json:"certificate_id,omitempty"`
	SSLForceHTTPS     bool            `json:"ssl_force_https"`
	Enabled           bool            `json:"enabled"`
	BlockExploits     bool            `json:"block_exploits"`
	Meta              json.RawMessage `json:"meta,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// CreateRedirectHostRequest is the request to create a redirect host
type CreateRedirectHostRequest struct {
	DomainNames       []string        `json:"domain_names" validate:"required,min=1"`
	ForwardScheme     string          `json:"forward_scheme,omitempty"`      // default: auto
	ForwardDomainName string          `json:"forward_domain_name" validate:"required"`
	ForwardPath       string          `json:"forward_path,omitempty"`
	PreservePath      *bool           `json:"preserve_path,omitempty"`
	RedirectCode      int             `json:"redirect_code,omitempty"` // default: 301
	SSLEnabled        bool            `json:"ssl_enabled,omitempty"`
	CertificateID     *string         `json:"certificate_id,omitempty"`
	SSLForceHTTPS     *bool           `json:"ssl_force_https,omitempty"`
	Enabled           *bool           `json:"enabled,omitempty"`
	BlockExploits     bool            `json:"block_exploits,omitempty"`
	Meta              json.RawMessage `json:"meta,omitempty"`
}

// UpdateRedirectHostRequest is the request to update a redirect host
type UpdateRedirectHostRequest struct {
	DomainNames       []string        `json:"domain_names,omitempty"`
	ForwardScheme     *string         `json:"forward_scheme,omitempty"`
	ForwardDomainName *string         `json:"forward_domain_name,omitempty"`
	ForwardPath       *string         `json:"forward_path,omitempty"`
	PreservePath      *bool           `json:"preserve_path,omitempty"`
	RedirectCode      *int            `json:"redirect_code,omitempty"`
	SSLEnabled        *bool           `json:"ssl_enabled,omitempty"`
	CertificateID     *string         `json:"certificate_id,omitempty"`
	SSLForceHTTPS     *bool           `json:"ssl_force_https,omitempty"`
	Enabled           *bool           `json:"enabled,omitempty"`
	BlockExploits     *bool           `json:"block_exploits,omitempty"`
	Meta              json.RawMessage `json:"meta,omitempty"`
}

// RedirectHostListResponse is the response for listing redirect hosts
type RedirectHostListResponse struct {
	Data       []RedirectHost `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}
