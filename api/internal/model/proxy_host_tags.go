package model

import (
	"fmt"
	"regexp"
	"strings"
)

// Tags are free-form labels an operator puts on a proxy host so the host list
// can be narrowed by intent ("media", "family", "prod"). They never reach an
// nginx config, so a tag-only change must not regenerate or reload anything.
const (
	MaxTagsPerHost = 10
	MaxTagLength   = 32
)

// tagRegex: lowercase alphanumerics plus '.', '_' and '-', starting with a
// letter or digit. Kept ASCII-only so tags are safe in URLs and query strings
// without escaping.
var tagRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// NormalizeTags trims, lowercases, drops empties and duplicates (first
// occurrence wins), then validates length, charset and count. A nil or
// all-empty input yields an empty, non-nil slice so callers can store '{}'
// without a nil check. Every error wraps ErrInvalidInput so the handler layer
// answers 400 rather than 500 — this repo has no struct validator, so this is
// the only gate.
func NormalizeTags(in []string) ([]string, error) {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, raw := range in {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			continue
		}
		if len(tag) > MaxTagLength {
			return nil, fmt.Errorf("%w: tag %q is longer than %d characters", ErrInvalidInput, tag, MaxTagLength)
		}
		if !tagRegex.MatchString(tag) {
			return nil, fmt.Errorf("%w: tag %q may only contain a-z, 0-9, '.', '_' and '-' and must start with a letter or digit", ErrInvalidInput, tag)
		}
		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) > MaxTagsPerHost {
		return nil, fmt.Errorf("%w: at most %d tags per host (got %d)", ErrInvalidInput, MaxTagsPerHost, len(out))
	}
	return out, nil
}
