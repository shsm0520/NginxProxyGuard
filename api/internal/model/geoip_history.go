package model

import "time"

// GeoIPUpdateStatus represents the status of a GeoIP update
type GeoIPUpdateStatus string

const (
	GeoIPUpdateStatusPending GeoIPUpdateStatus = "pending"
	GeoIPUpdateStatusRunning GeoIPUpdateStatus = "running"
	GeoIPUpdateStatusSuccess GeoIPUpdateStatus = "success"
	GeoIPUpdateStatusFailed  GeoIPUpdateStatus = "failed"
)

// GeoIPTriggerType represents how the update was triggered
type GeoIPTriggerType string

const (
	GeoIPTriggerManual GeoIPTriggerType = "manual"
	GeoIPTriggerAuto   GeoIPTriggerType = "auto"
)

// GeoIPUpdateHistory represents a GeoIP database update record
type GeoIPUpdateHistory struct {
	ID              string            `json:"id" db:"id"`
	Status          GeoIPUpdateStatus `json:"status" db:"status"`
	TriggerType     GeoIPTriggerType  `json:"trigger_type" db:"trigger_type"`
	StartedAt       time.Time         `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty" db:"completed_at"`
	DurationMs      *int              `json:"duration_ms,omitempty" db:"duration_ms"`
	DatabaseVersion string            `json:"database_version,omitempty" db:"database_version"`
	CountryDBSize   *int64            `json:"country_db_size,omitempty" db:"country_db_size"`
	ASNDBSize       *int64            `json:"asn_db_size,omitempty" db:"asn_db_size"`
	ErrorMessage    string            `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}

// GeoIPUpdateHistoryResponse is the API response for update history list
type GeoIPUpdateHistoryResponse struct {
	Data       []GeoIPUpdateHistory `json:"data"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}
