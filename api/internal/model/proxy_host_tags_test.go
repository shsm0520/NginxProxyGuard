package model

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	cases := []struct {
		name    string
		in      []string
		want    []string
		wantErr string
	}{
		{name: "nil is empty not nil", in: nil, want: []string{}},
		{name: "trims lowercases dedupes", in: []string{" Media ", "media", "FAMILY"}, want: []string{"media", "family"}},
		{name: "drops empties", in: []string{"", "  ", "prod"}, want: []string{"prod"}},
		{name: "allows dot underscore hyphen", in: []string{"home.lab", "my_app", "tier-1"}, want: []string{"home.lab", "my_app", "tier-1"}},
		{name: "rejects leading punctuation", in: []string{"-bad"}, wantErr: `tag "-bad"`},
		{name: "rejects spaces inside", in: []string{"two words"}, wantErr: `tag "two words"`},
		{name: "rejects non-ascii", in: []string{"미디어"}, wantErr: "tag"},
		{name: "rejects over 32 chars", in: []string{strings.Repeat("a", 33)}, wantErr: "longer than 32"},
		{name: "accepts exactly 32 chars", in: []string{strings.Repeat("a", 32)}, want: []string{strings.Repeat("a", 32)}},
		{name: "rejects more than 10", in: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}, wantErr: "at most 10 tags"},
		{name: "dupes do not count toward the cap", in: []string{"a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "b"}, want: []string{"a", "b"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeTags(tc.in)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("error must wrap ErrInvalidInput so the handler answers 400, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("result must be a non-nil slice so it stores as '{}'")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
