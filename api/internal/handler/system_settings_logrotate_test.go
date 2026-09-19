package handler

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// The rotated-name scheme gained a time component so that "Rotate now" stops
// colliding with the day's scheduled rotation (#301). logrotate prunes dateext
// archives by globbing the pattern its dateformat implies, so archives written
// under the old 8-digit scheme would fall outside `rotate N` and stay on disk
// for good. They are renamed into the new scheme instead of being deleted —
// the operator keeps their history and logrotate resumes counting it.
func TestLegacyRotatedNamesAreMigratedIntoTheTimestampedScheme(t *testing.T) {
	dir := t.TempDir()
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	write("access_raw.log")                    // the live log: never touched
	write("access_raw.log-20260918.gz")        // legacy, compressed
	write("access_raw.log-20260919")           // legacy, delaycompress' newest
	write("error_raw.log-20260919.gz")         // legacy, other log in the stanza
	write("access_raw.log-20260920-031500.gz") // already migrated
	write("access.log")                        // unrelated
	write("nginx-error.log-20260919")          // not one of ours

	migrateLegacyRotatedNamesIn(dir)

	want := []string{
		"access.log",
		"access_raw.log",
		"access_raw.log-20260918-000000.gz",
		"access_raw.log-20260919-000000",
		"access_raw.log-20260920-031500.gz",
		"error_raw.log-20260919-000000.gz",
		"nginx-error.log-20260919",
	}
	if got := names(t, dir); !equal(got, want) {
		t.Fatalf("after migration:\n got %v\nwant %v", got, want)
	}

	// Sorting is what logrotate's retention depends on, and a renamed archive
	// has to land before the same day's later rotations, not after them.
	got := names(t, dir)
	if !sort.StringsAreSorted(got) {
		t.Errorf("migrated names do not sort: %v", got)
	}

	// Running again must be a no-op: a name that carries a time no longer
	// matches, so nothing is renamed twice.
	migrateLegacyRotatedNamesIn(dir)
	if again := names(t, dir); !equal(again, want) {
		t.Fatalf("second run was not a no-op:\n got %v\nwant %v", again, want)
	}
}

func TestMigrationNeverOverwritesAnExistingArchive(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "access_raw.log-20260919"), []byte("legacy"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "access_raw.log-20260919-000000"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	migrateLegacyRotatedNamesIn(dir)

	body, err := os.ReadFile(filepath.Join(dir, "access_raw.log-20260919-000000"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "existing" {
		t.Errorf("migration clobbered an existing archive: %q", body)
	}
	if _, err := os.Stat(filepath.Join(dir, "access_raw.log-20260919")); err != nil {
		t.Errorf("the legacy file should have been left in place: %v", err)
	}
}

func names(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
