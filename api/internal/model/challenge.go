package model

import (
	"time"
)

// ChallengeConfig represents CAPTCHA challenge configuration
type ChallengeConfig struct {
	ID           string    `json:"id"`
	ProxyHostID  *string   `json:"proxy_host_id,omitempty"` // NULL = global config
	Enabled      bool      `json:"enabled"`
	ChallengeType string   `json:"challenge_type"` // recaptcha_v2, recaptcha_v3, hcaptcha, turnstile
	SiteKey      string    `json:"site_key"`
	SecretKey    string    `json:"secret_key,omitempty"` // Not returned in API responses
	TokenValidity int      `json:"token_validity"`       // seconds
	MinScore     float64   `json:"min_score"`            // for reCAPTCHA v3
	ApplyTo      string    `json:"apply_to"`             // bot_filter, geo_restriction, both
	PageTitle    string    `json:"page_title"`
	PageMessage  string    `json:"page_message"`
	Theme        string    `json:"theme"` // light, dark
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ChallengeConfigRequest for creating/updating challenge config
type ChallengeConfigRequest struct {
	Enabled       *bool    `json:"enabled,omitempty"`
	ChallengeType *string  `json:"challenge_type,omitempty"`
	SiteKey       *string  `json:"site_key,omitempty"`
	SecretKey     *string  `json:"secret_key,omitempty"`
	TokenValidity *int     `json:"token_validity,omitempty"`
	MinScore      *float64 `json:"min_score,omitempty"`
	ApplyTo       *string  `json:"apply_to,omitempty"`
	PageTitle     *string  `json:"page_title,omitempty"`
	PageMessage   *string  `json:"page_message,omitempty"`
	Theme         *string  `json:"theme,omitempty"`
}

// ChallengeConfigResponse for API responses (without secret key)
type ChallengeConfigResponse struct {
	ID            string    `json:"id"`
	ProxyHostID   *string   `json:"proxy_host_id,omitempty"`
	Enabled       bool      `json:"enabled"`
	ChallengeType string    `json:"challenge_type"`
	SiteKey       string    `json:"site_key"`
	HasSecretKey  bool      `json:"has_secret_key"` // true if secret key is configured
	TokenValidity int       `json:"token_validity"`
	MinScore      float64   `json:"min_score"`
	ApplyTo       string    `json:"apply_to"`
	PageTitle     string    `json:"page_title"`
	PageMessage   string    `json:"page_message"`
	Theme         string    `json:"theme"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts ChallengeConfig to response (without secret)
func (c *ChallengeConfig) ToResponse() ChallengeConfigResponse {
	return ChallengeConfigResponse{
		ID:            c.ID,
		ProxyHostID:   c.ProxyHostID,
		Enabled:       c.Enabled,
		ChallengeType: c.ChallengeType,
		SiteKey:       c.SiteKey,
		HasSecretKey:  c.SecretKey != "",
		TokenValidity: c.TokenValidity,
		MinScore:      c.MinScore,
		ApplyTo:       c.ApplyTo,
		PageTitle:     c.PageTitle,
		PageMessage:   c.PageMessage,
		Theme:         c.Theme,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// ChallengeToken represents an issued challenge token
type ChallengeToken struct {
	ID              string     `json:"id"`
	ProxyHostID     *string    `json:"proxy_host_id,omitempty"`
	TokenHash       string     `json:"-"` // Never expose
	ClientIP        string     `json:"client_ip"`
	UserAgent       string     `json:"user_agent,omitempty"`
	ChallengeReason string     `json:"challenge_reason,omitempty"`
	IssuedAt        time.Time  `json:"issued_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	UseCount        int        `json:"use_count"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	Revoked         bool       `json:"revoked"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	RevokedReason   string     `json:"revoked_reason,omitempty"`
}

// IsValid checks if token is still valid
func (t *ChallengeToken) IsValid() bool {
	if t.Revoked {
		return false
	}
	return time.Now().Before(t.ExpiresAt)
}

// ChallengeLog represents a challenge event log
type ChallengeLog struct {
	ID            string    `json:"id"`
	ProxyHostID   *string   `json:"proxy_host_id,omitempty"`
	ClientIP      string    `json:"client_ip"`
	UserAgent     string    `json:"user_agent,omitempty"`
	Result        string    `json:"result"` // presented, passed, failed, expired
	TriggerReason string    `json:"trigger_reason,omitempty"`
	CaptchaScore  *float64  `json:"captcha_score,omitempty"`
	SolveTime     *int      `json:"solve_time,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// VerifyCaptchaRequest for CAPTCHA verification
type VerifyCaptchaRequest struct {
	Token         string `json:"token"`          // reCAPTCHA response token
	ProxyHostID   string `json:"proxy_host_id"`  // Which proxy host triggered
	ChallengeReason string `json:"challenge_reason,omitempty"` // bot_filter or geo_restriction
}

// VerifyCaptchaResponse after successful verification
type VerifyCaptchaResponse struct {
	Success     bool      `json:"success"`
	Token       string    `json:"token,omitempty"`       // Challenge bypass token
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	ExpiresIn   int       `json:"expires_in,omitempty"`  // seconds
	Error       string    `json:"error,omitempty"`
	Score       float64   `json:"score,omitempty"`       // reCAPTCHA v3 score
}

// ValidateTokenRequest for validating challenge token
type ValidateTokenRequest struct {
	Token       string `json:"token"`
	ProxyHostID string `json:"proxy_host_id"`
	ClientIP    string `json:"client_ip"`
}

// ValidateTokenResponse
type ValidateTokenResponse struct {
	Valid       bool      `json:"valid"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Reason      string    `json:"reason,omitempty"` // Why challenge was required
}

// ChallengeStats for dashboard
type ChallengeStats struct {
	TotalChallenges    int     `json:"total_challenges"`
	PassedChallenges   int     `json:"passed_challenges"`
	FailedChallenges   int     `json:"failed_challenges"`
	ActiveTokens       int     `json:"active_tokens"`
	AverageScore       float64 `json:"average_score,omitempty"`
	AverageSolveTime   float64 `json:"average_solve_time,omitempty"`
}

// reCAPTCHA verification response from Google
type RecaptchaVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score,omitempty"`        // v3 only
	Action      string    `json:"action,omitempty"`       // v3 only
	ChallengeTs string    `json:"challenge_ts,omitempty"`
	Hostname    string    `json:"hostname,omitempty"`
	ErrorCodes  []string  `json:"error-codes,omitempty"`
}

// Cloudflare Turnstile verification response
type TurnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Action      string   `json:"action,omitempty"`
	CData       string   `json:"cdata,omitempty"`
}
