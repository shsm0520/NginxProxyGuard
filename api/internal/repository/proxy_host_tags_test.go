package repository

import (
	"os"
	"strings"
	"testing"
)

// Every SELECT that materialises a model.ProxyHost must also read tags, and
// every matching Scan must bind &host.Tags — otherwise that code path returns
// a host with empty tags while the row has some. Eight sites today; this test
// stays green only if all of them agree.
func TestProxyHostReadsIncludeTags(t *testing.T) {
	for _, file := range []string{"proxy_host.go", "proxy_host_queries.go", "proxy_host_favorites.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		s := string(src)
		selects := strings.Count(s, "meta, created_at, updated_at")
		selectsWithTags := strings.Count(s, "meta, created_at, updated_at, COALESCE(tags, '{}') as tags")
		if selects != selectsWithTags {
			t.Errorf("%s: %d SELECT column lists end in updated_at but only %d also read tags", file, selects, selectsWithTags)
		}
		scans := strings.Count(s, "&host.UpdatedAt,")
		tagScans := strings.Count(s, "&host.Tags,")
		if scans != tagScans {
			t.Errorf("%s: %d Scans bind &host.UpdatedAt but %d bind &host.Tags", file, scans, tagScans)
		}
	}
}

func TestProxyHostWritesIncludeTags(t *testing.T) {
	src, err := os.ReadFile("proxy_host.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	for _, want := range []string{
		"auth_provider_id, auth_bypass_paths, waf_use_global, tags\n", // INSERT column list
		"$46, $47)",           // INSERT placeholders
		"pq.Array(req.Tags),", // INSERT argument
		"waf_use_global = $47,\n\t\t\ttags = $48\n",               // UPDATE SET
		"pq.Array(existing.Tags),",                                // UPDATE argument
		"if req.Tags != nil {\n\t\texisting.Tags = req.Tags\n\t}", // UPDATE merge
	} {
		if !strings.Contains(s, want) {
			t.Errorf("proxy_host.go is missing %q", want)
		}
	}
}
