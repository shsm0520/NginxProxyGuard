package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	// TOTP settings
	totpDigits   = 6
	totpPeriod   = 30 // seconds
	totpIssuer   = "Nginx-Proxy-Guard"
	secretLength = 20 // bytes, 160 bits
)

// GenerateTOTPSecret creates a new random TOTP secret
func GenerateTOTPSecret() (string, error) {
	secret := make([]byte, secretLength)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// GenerateQRCodeURL creates the otpauth:// URL for QR code generation
func GenerateQRCodeURL(secret, username string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		url.PathEscape(totpIssuer),
		url.PathEscape(username),
		secret,
		url.QueryEscape(totpIssuer),
		totpDigits,
		totpPeriod,
	)
}

// ValidateTOTPCode validates a TOTP code against a secret
// Uses constant-time comparison to prevent timing attacks
func ValidateTOTPCode(secret, code string) bool {
	// Allow 1 period before and after current time for clock skew
	currentTime := time.Now().Unix()

	for _, offset := range []int64{-1, 0, 1} {
		timestamp := currentTime + (offset * totpPeriod)
		expectedCode := generateTOTPCode(secret, timestamp)
		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(expectedCode), []byte(code)) == 1 {
			return true
		}
	}

	return false
}

// generateTOTPCode generates a TOTP code for a given timestamp
func generateTOTPCode(secret string, timestamp int64) string {
	// Decode base32 secret
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}

	// Calculate counter (number of periods since epoch)
	counter := uint64(timestamp / totpPeriod)

	// Convert counter to bytes (big endian)
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	// HMAC-SHA1
	mac := hmac.New(sha1.New, secretBytes)
	mac.Write(counterBytes)
	hash := mac.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate 6-digit code
	return fmt.Sprintf("%06d", code%1000000)
}

// GenerateBackupCodes creates a set of one-time backup codes
func GenerateBackupCodes(count int) ([]string, []string, error) {
	codes := make([]string, count)
	hashedCodes := make([]string, count)

	for i := 0; i < count; i++ {
		// Generate 8-character alphanumeric code
		bytes := make([]byte, 5)
		if _, err := rand.Read(bytes); err != nil {
			return nil, nil, err
		}
		code := strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes))[:8]
		codes[i] = code

		// Hash for storage
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, err
		}
		hashedCodes[i] = string(hash)
	}

	return codes, hashedCodes, nil
}

// ValidateBackupCode checks if a backup code is valid and returns remaining codes
func ValidateBackupCode(code string, hashedCodes []string) (bool, []string) {
	for i, hashedCode := range hashedCodes {
		if err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(strings.ToUpper(code))); err == nil {
			// Remove used code
			remaining := make([]string, 0, len(hashedCodes)-1)
			remaining = append(remaining, hashedCodes[:i]...)
			remaining = append(remaining, hashedCodes[i+1:]...)
			return true, remaining
		}
	}
	return false, hashedCodes
}
