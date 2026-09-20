package model

import "time"

type User struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	PasswordHash   string     `json:"-"` // Never expose
	Role           string     `json:"role"`
	// RoleID points at the roles table and is the authoritative permission
	// source; Role above stays as the legacy 'admin'|'user' marker the CLI
	// recovery path still keys on. (#222)
	RoleID             *string `json:"role_id,omitempty"`
	MustChangePassword bool    `json:"must_change_password"`
	Language       string     `json:"language"`
	FontFamily     string     `json:"font_family"`
	IsInitialSetup bool       `json:"is_initial_setup"`
	TOTPEnabled    bool       `json:"totp_enabled"`
	TOTPSecret     string     `json:"-"` // Never expose
	TOTPVerifiedAt *time.Time `json:"-"`
	BackupCodes    []string   `json:"-"` // Never expose
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP    string     `json:"last_login_ip,omitempty"`
	LoginCount     int        `json:"login_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	TOTPCode string `json:"totp_code,omitempty"` // Required if 2FA enabled
}

type LoginResponse struct {
	Token          string `json:"token,omitempty"`
	User           *User  `json:"user,omitempty"`
	IsInitialSetup bool   `json:"is_initial_setup"`
	Requires2FA    bool   `json:"requires_2fa,omitempty"`
	TempToken      string `json:"temp_token,omitempty"` // Temporary token for 2FA verification
}

type Verify2FARequest struct {
	TempToken string `json:"temp_token" validate:"required"`
	TOTPCode  string `json:"totp_code" validate:"required"`
}

type ChangeCredentialsRequest struct {
	CurrentPassword    string `json:"current_password" validate:"required"`
	NewUsername        string `json:"new_username" validate:"required,min=3"`
	NewPassword        string `json:"new_password" validate:"required,min=8"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required"`
	// NewEmail is optional and kept as-is when blank. Setting it here matters
	// because SSO links an identity to an account by verified email, and an
	// account left on the synthesised <username>@localhost can never match a
	// real identity provider. (#240)
	NewEmail string `json:"new_email,omitempty"`
}

// SetUserEmailRequest changes the address an account is linked by.
//
// Deliberately admin-only (user:write). SSO linking matches a verified IdP
// email to a local account and then syncs the role from the provider's groups,
// so a user able to set their own address could claim an administrator's IdP
// address before that administrator first signs in — and inherit the role when
// they do. Assigning roles is already a user:write power, so gating the address
// the same way adds no new authority. (#240)
type SetUserEmailRequest struct {
	Email string `json:"email" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword    string `json:"current_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=8"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required"`
}

type AuthSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginAttempt struct {
	ID          string    `json:"id"`
	IPAddress   string    `json:"ip_address"`
	Username    string    `json:"username"`
	Success     bool      `json:"success"`
	AttemptedAt time.Time `json:"attempted_at"`
}

type AuthStatus struct {
	Authenticated  bool  `json:"authenticated"`
	IsInitialSetup bool  `json:"is_initial_setup"`
	User           *User `json:"user,omitempty"`
}

// 2FA Setup
// Setup2FAResponse carries only what is needed to pair an authenticator.
//
// Backup codes deliberately do NOT live here. They used to, and that made the
// displayed set and the stored hashes two different things the moment a second
// setup call happened: the secret is kept across calls (so the QR already
// scanned keeps working), but a second call rewrote the hashes, leaving the
// first screen holding ten codes the database had never seen. They are issued
// by Enable2FA instead, where "displayed" and "stored" are the same write.
type Setup2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// Enable2FAResponse returns the backup codes minted for this enrolment. This is
// the only time they are ever shown in plaintext — only bcrypt hashes are kept.
type Enable2FAResponse struct {
	Message     string   `json:"message"`
	BackupCodes []string `json:"backup_codes"`
}

type Enable2FARequest struct {
	TOTPCode string `json:"totp_code" validate:"required"`
}

type Disable2FARequest struct {
	Password string `json:"password" validate:"required"`
	TOTPCode string `json:"totp_code" validate:"required"`
}

// Account settings
type AccountInfo struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Role        string     `json:"role"`
	Language    string     `json:"language"`
	FontFamily  string     `json:"font_family"`
	TOTPEnabled bool       `json:"totp_enabled"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP string     `json:"last_login_ip,omitempty"`
	LoginCount  int        `json:"login_count"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Language settings
type LanguageRequest struct {
	Language string `json:"language" validate:"required,oneof=ko en"`
}

type LanguageResponse struct {
	Language string `json:"language"`
}

// Font settings
type FontFamilyRequest struct {
	FontFamily string `json:"font_family" validate:"required"`
}

type FontFamilyResponse struct {
	FontFamily string `json:"font_family"`
}

// Username change
type ChangeUsernameRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewUsername     string `json:"new_username" validate:"required,min=3"`
}
