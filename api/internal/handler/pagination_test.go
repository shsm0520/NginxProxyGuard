package handler

import "testing"

// The case that matters is "too large". Everything else already behaved; the
// ceiling was the one that answered with the default instead of the limit, and
// a caller has no way to notice — the response carries per_page, but a client
// that asked for 200 and got 20 has already thrown its request away by then.
func TestClampPerPage(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{"above the ceiling clamps to the ceiling", "200", MaxPerPage},
		{"far above the ceiling still clamps", "100000", MaxPerPage},
		{"exactly the ceiling is honoured", "100", MaxPerPage},
		{"an ordinary value is honoured", "50", 50},
		{"the minimum is honoured", "1", MinPerPage},

		// No request to honour — these fall back to the default rather than to
		// the ceiling, because none of them asked for anything.
		{"absent means default", "", DefaultPerPage},
		{"zero means default", "0", DefaultPerPage},
		{"negative means default", "-5", DefaultPerPage},
		{"unparseable means default", "all", DefaultPerPage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampPerPage(tt.raw, DefaultPerPage); got != tt.want {
				t.Errorf("clampPerPage(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}

// ParsePaginationParamsWithDefaults takes its own default, and must keep using
// it for the no-request cases while sharing the ceiling.
func TestClampPerPageHonoursACallerSuppliedDefault(t *testing.T) {
	const callerDefault = 500

	if got := clampPerPage("", callerDefault); got != callerDefault {
		t.Errorf("absent per_page = %d, want the caller's default %d", got, callerDefault)
	}
	// A caller-supplied default may legitimately exceed MaxPerPage — some
	// internal listings pass one — and that is not the caller's request being
	// clamped, so it passes through untouched.
	if got := clampPerPage("200", callerDefault); got != MaxPerPage {
		t.Errorf("per_page=200 = %d, want the ceiling %d", got, MaxPerPage)
	}
}
