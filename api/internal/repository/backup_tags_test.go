package repository

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestBackupTagsOrEmpty(t *testing.T) {
	if got := backupTagsOrEmpty(nil); got == nil || len(got) != 0 {
		t.Fatalf("a pre-tags backup (no key) must restore as [] not nil, got %#v", got)
	}
	if got := backupTagsOrEmpty([]string{" Media ", "media"}); !reflect.DeepEqual(got, []string{"media"}) {
		t.Fatalf("got %v", got)
	}
	if got := backupTagsOrEmpty([]string{"ok", "-bad"}); len(got) != 0 {
		t.Fatalf("unreadable tags must be dropped for that host, not abort the restore; got %v", got)
	}
}

// The export SELECT and import INSERT are long hand-maintained lists; pin the
// tag column in both so a future column can't silently drop it.
func TestBackupSQLCarriesTags(t *testing.T) {
	exp, _ := os.ReadFile("backup_export_proxy.go")
	if !strings.Contains(string(exp), "COALESCE(tags, '{}') as tags") || !strings.Contains(string(exp), "pq.Array(&ph.Tags)") {
		t.Error("backup_export_proxy.go does not read tags")
	}
	imp, _ := os.ReadFile("backup_import_proxy.go")
	if !strings.Contains(string(imp), "waf_use_global, tags)") || !strings.Contains(string(imp), "$53, $54)") || !strings.Contains(string(imp), "pq.Array(backupTagsOrEmpty(ph.ProxyHost.Tags))") {
		t.Error("backup_import_proxy.go does not write tags")
	}
}
