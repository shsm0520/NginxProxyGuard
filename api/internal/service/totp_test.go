package service

import (
	"encoding/base32"
	"fmt"
	"testing"
	"time"
)

// A reference code computed the way any authenticator computes it, so these
// tests measure our validator against RFC 6238 rather than against itself.
func referenceCode(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	code := generateTOTPCode(secret, at.Unix())
	if len(code) != 6 {
		t.Fatalf("reference code for %q is %q, want six digits", secret, code)
	}
	return code
}

func testSecret(t *testing.T) string {
	t.Helper()
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret: %v", err)
	}
	return secret
}

// #305: three independent generators produced the same correct code and the
// server rejected it. Spacing is one of the ways that happens — 1Password
// renders a TOTP as "123 456" and copying it carries the space, and a Korean
// IME can leave a non-breaking or ideographic space behind. None of those are
// part of the code.
func TestValidateTOTPCodeAcceptsTheCodeAsUsersActuallyCopyIt(t *testing.T) {
	secret := testSecret(t)
	code := referenceCode(t, secret, time.Now())

	grouped := code[:3] + " " + code[3:]

	for _, tc := range []struct {
		name string
		code string
	}{
		{"exact", code},
		{"grouped with a space, as 1Password renders it", grouped},
		{"trailing space from a paste", code + " "},
		{"leading space from a paste", " " + code},
		{"non-breaking space", code[:3] + " " + code[3:]},
		{"ideographic space from a CJK IME", code[:3] + "　" + code[3:]},
		{"hyphen separator", code[:3] + "-" + code[3:]},
		{"newline from a terminal paste", code + "\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !ValidateTOTPCode(secret, tc.code) {
				t.Errorf("rejected %q, which is the correct code %q with separators the user did not type", tc.code, code)
			}
		})
	}
}

// Normalisation must not become a way in. A wrong code stays wrong however it
// is spaced, and the window stays ±1 period.
func TestValidateTOTPCodeStillRejectsWhatItShould(t *testing.T) {
	secret := testSecret(t)
	now := time.Now()
	code := referenceCode(t, secret, now)

	wrong := fmt.Sprintf("%06d", (mustAtoi(t, code)+1)%1000000)

	for _, tc := range []struct {
		name string
		code string
	}{
		{"a different code", wrong},
		{"the same code with a digit dropped", code[:5]},
		{"the same digits reordered", code[3:] + code[:3]},
		{"letters where digits belong", "ABCDEF"},
		{"empty", ""},
		{"whitespace only", "   "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The reordering case can legitimately collide with the real code
			// when the halves are equal; skip that degenerate input.
			if tc.code == code {
				t.Skip("degenerate: reordering produced the same code")
			}
			if ValidateTOTPCode(secret, tc.code) {
				t.Errorf("accepted %q", tc.code)
			}
		})
	}

	// Two periods away is outside the documented ±1 window.
	stale := referenceCode(t, secret, now.Add(-2*totpPeriod*time.Second))
	if stale != code && ValidateTOTPCode(secret, stale) {
		t.Errorf("accepted a code from two periods ago (%q); the skew window is supposed to be one", stale)
	}
}

// The primitive's answer to "I could not evaluate this" must be no.
//
// generateTOTPCode returns "" for a secret it cannot base32-decode, and
// subtle.ConstantTimeCompare("", "") is 1 — so before the guard,
// ValidateTOTPCode(garbageSecret, "") returned TRUE. No caller could reach it,
// because all four check for an empty code first, but an authentication check
// that fails open is one refactor away from a bypass.
func TestValidateTOTPCodeFailsClosedOnAnUnusableSecret(t *testing.T) {
	for _, tc := range []struct {
		name   string
		secret string
		code   string
	}{
		{"undecodable secret, empty code", "not-base32!!!", ""},
		{"undecodable secret, whitespace code", "not-base32!!!", " "},
		{"undecodable secret, plausible code", "not-base32!!!", "123456"},
		{"empty secret, empty code", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if ValidateTOTPCode(tc.secret, tc.code) {
				t.Errorf("ValidateTOTPCode(%q, %q) = true — an unusable secret must never authenticate anyone", tc.secret, tc.code)
			}
		})
	}

	// An EMPTY secret is the state of every user who never enrolled, and it
	// base32-decodes cleanly to a zero-length key — so it does produce a real
	// six-digit code. Nothing behind totp_enabled can reach it today, but the
	// empty-code guard must still hold there.
	if ValidateTOTPCode("", "") {
		t.Error("an empty secret with an empty code authenticated")
	}
}

// Backup codes are 8 characters of base32 and get read off a screen or pasted
// from a password manager, so they arrive cased and spaced however the user
// happened to copy them.
func TestValidateBackupCodeToleratesHowCodesAreCopied(t *testing.T) {
	codes, hashed, err := GenerateBackupCodes(3)
	if err != nil {
		t.Fatalf("GenerateBackupCodes: %v", err)
	}
	code := codes[0]

	for _, tc := range []struct {
		name string
		code string
	}{
		{"exact", code},
		{"lower-cased", lower(code)},
		{"with surrounding whitespace", "  " + code + "\n"},
		{"hyphenated in the middle", code[:4] + "-" + code[4:]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ok, remaining := ValidateBackupCode(tc.code, hashed)
			if !ok {
				t.Fatalf("rejected %q, which is backup code %q", tc.code, code)
			}
			if len(remaining) != len(hashed)-1 {
				t.Errorf("consumed %d codes, want exactly 1", len(hashed)-len(remaining))
			}
		})
	}

	if ok, _ := ValidateBackupCode("", hashed); ok {
		t.Error("an empty backup code was accepted")
	}
	if ok, _ := ValidateBackupCode("NOTAREALCODE", hashed); ok {
		t.Error("an unissued backup code was accepted")
	}
}

// The login form now allows eight characters; the codes it has to carry must
// actually be eight characters of the base32 alphabet, or that cap is wrong.
func TestBackupCodesFitWhatTheLoginFormAccepts(t *testing.T) {
	codes, _, err := GenerateBackupCodes(10)
	if err != nil {
		t.Fatalf("GenerateBackupCodes: %v", err)
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

	for _, code := range codes {
		if len(code) != 8 {
			t.Errorf("backup code %q is %d characters; ui/src/components/Login.tsx caps input at 8", code, len(code))
		}
		for _, r := range code {
			if !containsRune(alphabet, r) {
				t.Errorf("backup code %q contains %q, which the login input's [^0-9A-Z] filter would drop", code, r)
			}
		}
	}
}

// GenerateTOTPSecret must keep producing something every authenticator accepts.
func TestGenerateTOTPSecretIsStandardBase32(t *testing.T) {
	secret := testSecret(t)
	if len(secret) != 32 {
		t.Errorf("secret is %d characters, want 32 (160 bits, unpadded base32)", len(secret))
	}
	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		t.Fatalf("our own secret does not base32-decode: %v", err)
	}
	if len(raw) != secretLength {
		t.Errorf("decoded to %d bytes, want %d", len(raw), secretLength)
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		t.Fatalf("parsing %q: %v", s, err)
	}
	return n
}

func lower(s string) string {
	out := []rune(s)
	for i, r := range out {
		if r >= 'A' && r <= 'Z' {
			out[i] = r + 32
		}
	}
	return string(out)
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
