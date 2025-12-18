package model

import "time"

// WAFRuleExclusion represents a disabled WAF rule for a specific proxy host
type WAFRuleExclusion struct {
	ID              string    `json:"id"`
	ProxyHostID     string    `json:"proxy_host_id"`
	RuleID          int       `json:"rule_id"`
	RuleCategory    string    `json:"rule_category,omitempty"`
	RuleDescription string    `json:"rule_description,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	DisabledBy      string    `json:"disabled_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// WAFRule represents an OWASP CRS rule
type WAFRule struct {
	ID               int                     `json:"id"`
	Category         string                  `json:"category"`
	Description      string                  `json:"description,omitempty"`
	Severity         string                  `json:"severity,omitempty"`
	Tags             []string                `json:"tags,omitempty"`
	Enabled          bool                    `json:"enabled"`                     // Whether this rule is enabled for the current host
	GloballyDisabled bool                    `json:"globally_disabled"`           // Whether this rule is disabled globally
	Exclusion        *WAFRuleExclusion       `json:"exclusion,omitempty"`         // Host-specific exclusion details if rule is disabled
	GlobalExclusion  *GlobalWAFRuleExclusion `json:"global_exclusion,omitempty"`  // Global exclusion details if rule is globally disabled
}

// WAFRuleCategory represents a group of related WAF rules
type WAFRuleCategory struct {
	ID          string    `json:"id"`          // e.g., "941" for XSS
	Name        string    `json:"name"`        // e.g., "XSS"
	Description string    `json:"description"` // e.g., "Cross-Site Scripting Attack Detection"
	FileName    string    `json:"file_name"`   // e.g., "REQUEST-941-APPLICATION-ATTACK-XSS.conf"
	RuleCount   int       `json:"rule_count"`
	Rules       []WAFRule `json:"rules,omitempty"`
}

// WAFHostConfig represents the WAF configuration for a specific host
type WAFHostConfig struct {
	ProxyHostID   string             `json:"proxy_host_id"`
	ProxyHostName string             `json:"proxy_host_name"` // First domain name
	WAFEnabled    bool               `json:"waf_enabled"`
	WAFMode       string             `json:"waf_mode"` // "blocking" or "detection"
	Exclusions    []WAFRuleExclusion `json:"exclusions"`
	ExclusionCount int               `json:"exclusion_count"`
}

// CreateWAFRuleExclusionRequest is the request to disable a rule for a host
type CreateWAFRuleExclusionRequest struct {
	RuleID          int    `json:"rule_id" validate:"required,min=100000,max=999999"`
	RuleCategory    string `json:"rule_category,omitempty"`
	RuleDescription string `json:"rule_description,omitempty"`
	Reason          string `json:"reason,omitempty"`
}

// WAFRulesResponse is the response for listing WAF rules
type WAFRulesResponse struct {
	Categories []WAFRuleCategory `json:"categories"`
	TotalRules int               `json:"total_rules"`
}

// WAFHostConfigListResponse is the response for listing all hosts' WAF configs
type WAFHostConfigListResponse struct {
	Hosts []WAFHostConfig `json:"hosts"`
	Total int             `json:"total"`
}

// WAFPolicyHistory represents a WAF policy change record
type WAFPolicyHistory struct {
	ID              string    `json:"id"`
	ProxyHostID     string    `json:"proxy_host_id"`
	RuleID          int       `json:"rule_id"`
	RuleCategory    string    `json:"rule_category,omitempty"`
	RuleDescription string    `json:"rule_description,omitempty"`
	Action          string    `json:"action"` // "disabled" or "enabled"
	Reason          string    `json:"reason,omitempty"`
	ChangedBy       string    `json:"changed_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// WAFPolicyHistoryResponse is the response for listing policy history
type WAFPolicyHistoryResponse struct {
	History []WAFPolicyHistory `json:"history"`
	Total   int                `json:"total"`
}

// GlobalWAFRuleExclusion represents a globally disabled WAF rule (applies to all hosts)
type GlobalWAFRuleExclusion struct {
	ID              string    `json:"id"`
	RuleID          int       `json:"rule_id"`
	RuleCategory    string    `json:"rule_category,omitempty"`
	RuleDescription string    `json:"rule_description,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	DisabledBy      string    `json:"disabled_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GlobalWAFPolicyHistory represents a global WAF policy change record
type GlobalWAFPolicyHistory struct {
	ID              string    `json:"id"`
	RuleID          int       `json:"rule_id"`
	RuleCategory    string    `json:"rule_category,omitempty"`
	RuleDescription string    `json:"rule_description,omitempty"`
	Action          string    `json:"action"` // "disabled" or "enabled"
	Reason          string    `json:"reason,omitempty"`
	ChangedBy       string    `json:"changed_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateGlobalWAFRuleExclusionRequest is the request to disable a rule globally
type CreateGlobalWAFRuleExclusionRequest struct {
	RuleID          int    `json:"rule_id" validate:"required,min=100000,max=999999"`
	RuleCategory    string `json:"rule_category,omitempty"`
	RuleDescription string `json:"rule_description,omitempty"`
	Reason          string `json:"reason,omitempty"`
}

// GlobalWAFRulesResponse is the response for listing global WAF rules status
type GlobalWAFRulesResponse struct {
	Categories       []WAFRuleCategory        `json:"categories"`
	TotalRules       int                      `json:"total_rules"`
	GlobalExclusions []GlobalWAFRuleExclusion `json:"global_exclusions"`
}

// GlobalWAFPolicyHistoryResponse is the response for listing global policy history
type GlobalWAFPolicyHistoryResponse struct {
	History []GlobalWAFPolicyHistory `json:"history"`
	Total   int                      `json:"total"`
}

// WAFRuleSnapshot represents a complete snapshot of WAF configuration state
type WAFRuleSnapshot struct {
	ID                string                    `json:"id"`
	ProxyHostID       *string                   `json:"proxy_host_id,omitempty"` // NULL for global snapshots
	VersionNumber     int                       `json:"version_number"`
	SnapshotName      string                    `json:"snapshot_name,omitempty"`
	RuleEngine        string                    `json:"rule_engine,omitempty"`
	ParanoiaLevel     int                       `json:"paranoia_level,omitempty"`
	AnomalyThreshold  int                       `json:"anomaly_threshold,omitempty"`
	TotalRules        int                       `json:"total_rules"`
	DisabledRules     int                       `json:"disabled_rules"`
	ChangeDescription string                    `json:"change_description,omitempty"`
	CreatedBy         string                    `json:"created_by,omitempty"`
	CreatedAt         time.Time                 `json:"created_at"`
	Details           []WAFRuleSnapshotDetail   `json:"details,omitempty"`
}

// WAFRuleSnapshotDetail represents a single rule's state within a snapshot
type WAFRuleSnapshotDetail struct {
	ID              string `json:"id"`
	SnapshotID      string `json:"snapshot_id"`
	RuleID          string `json:"rule_id"`
	RuleCategory    string `json:"rule_category,omitempty"`
	RuleDescription string `json:"rule_description,omitempty"`
	IsDisabled      bool   `json:"is_disabled"`
	Reason          string `json:"reason,omitempty"`
}

// WAFRuleChangeEvent represents an individual rule enable/disable event
type WAFRuleChangeEvent struct {
	ID              string    `json:"id"`
	ProxyHostID     *string   `json:"proxy_host_id,omitempty"`
	RuleID          string    `json:"rule_id"`
	Action          string    `json:"action"` // "enabled" or "disabled"
	RuleCategory    string    `json:"rule_category,omitempty"`
	RuleDescription string    `json:"rule_description,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	ChangedBy       string    `json:"changed_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateWAFSnapshotRequest is the request to create a new WAF snapshot
type CreateWAFSnapshotRequest struct {
	SnapshotName      string `json:"snapshot_name,omitempty"`
	ChangeDescription string `json:"change_description,omitempty"`
}

// WAFSnapshotListResponse is the response for listing WAF snapshots
type WAFSnapshotListResponse struct {
	Snapshots []WAFRuleSnapshot `json:"snapshots"`
	Total     int               `json:"total"`
}

// WAFRuleChangeEventListResponse is the response for listing rule change events
type WAFRuleChangeEventListResponse struct {
	Events []WAFRuleChangeEvent `json:"events"`
	Total  int                  `json:"total"`
}
