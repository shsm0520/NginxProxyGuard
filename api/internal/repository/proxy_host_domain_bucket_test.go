package repository

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

// The domain bucket started life as a regexp_replace that stripped the first
// label of domain_names[1] unconditionally. That filed an apex host
// (example.com) under "com" — a useless group — and yielded a NULL bucket for a
// row with no domain name, which fails rows.Scan into a string ("converting NULL
// to string is unsupported") and would turn GET /proxy-hosts/groups into a 500
// for every caller. sqlmock cannot evaluate SQL, so the cases are pinned against
// a real planner.
func openDomainBucketTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("NPG_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NPG_TEST_DATABASE_URL not set — skipping DB-backed domain bucket test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestProxyHostDomainBucketExpr(t *testing.T) {
	db := openDomainBucketTestDB(t)

	cases := []struct {
		name       string
		domain     string // the row's domain_names[1]; "" means an empty array
		want       string
		wantBucket bool
	}{
		{name: "three labels strip the first", domain: "app.example.com", want: "example.com", wantBucket: true},
		{name: "four labels strip only the first", domain: "a.b.example.com", want: "b.example.com", wantBucket: true},
		{name: "apex is its own bucket", domain: "example.com", want: "example.com", wantBucket: true},
		{name: "two label test domain is its own bucket", domain: "apexcheck.test", want: "apexcheck.test", wantBucket: true},
		{name: "single label has no parent domain", domain: "streamlistener"},
		{name: "empty string has no parent domain", domain: ""},
	}
	// A domainless row (a stream listener, or a malformed row) must bucket to
	// NULL rather than crash the aggregate.
	query := `SELECT ` + proxyHostDomainBucketExpr + ` FROM (SELECT $1::text[] AS domain_names) t`
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arr := "{}"
			if tc.domain != "" {
				arr = `{"` + tc.domain + `"}`
			}
			var bucket sql.NullString
			if err := db.QueryRow(query, arr).Scan(&bucket); err != nil {
				t.Fatalf("evaluate bucket: %v", err)
			}
			if !tc.wantBucket {
				if bucket.Valid {
					t.Fatalf("domain %q must have no bucket, got %q", tc.domain, bucket.String)
				}
				return
			}
			if !bucket.Valid || bucket.String != tc.want {
				t.Fatalf("domain %q bucketed to %v, want %q", tc.domain, bucket, tc.want)
			}
		})
	}
}

// The aggregate the handler actually runs, over a fixed row set: a NULL bucket
// must be dropped before it reaches the string Scan, and the apex host must be
// counted in its own bucket alongside its subdomain.
func TestProxyHostDomainGroupQueryDropsRowsWithoutAParentDomain(t *testing.T) {
	db := openDomainBucketTestDB(t)

	rows := `FROM (VALUES ('{"app.example.com"}'::text[]), ('{"example.com"}'::text[]), ('{"streamlistener"}'::text[]), ('{""}'::text[]), ('{}'::text[]), (NULL::text[])) AS proxy_hosts(domain_names)`
	query := strings.Replace(proxyHostDomainGroupQuery, "FROM proxy_hosts", rows, 1)
	if query == proxyHostDomainGroupQuery {
		t.Fatal("domain group query no longer reads FROM proxy_hosts; update this test")
	}

	got := map[string]int{}
	result, err := db.Query(query)
	if err != nil {
		t.Fatalf("run domain group query: %v", err)
	}
	defer result.Close()
	for result.Next() {
		// Same binding as groupCounts: a NULL bucket fails here.
		var name string
		var count int
		if err := result.Scan(&name, &count); err != nil {
			t.Fatalf("scan bucket: %v", err)
		}
		got[name] = count
	}
	if err := result.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	if len(got) != 1 || got["example.com"] != 2 {
		t.Fatalf("buckets = %v, want only example.com=2 (apex + subdomain, domainless rows dropped)", got)
	}
}

// One expression, two callers. A second copy in either file is the bug this
// pair of files exists to prevent: the filter would stop matching the groups.
func TestProxyHostDomainBucketExprIsSharedNotCopied(t *testing.T) {
	for _, f := range []string{"proxy_host_groups.go", "proxy_host_queries.go"} {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		s := string(src)
		if !strings.Contains(s, "proxyHostDomainBucketExpr") {
			t.Errorf("%s buckets hosts by domain but does not use proxyHostDomainBucketExpr", f)
		}
		// The raw regexp belongs to the constant's definition only.
		if n := strings.Count(s, "regexp_replace(domain_names[1]"); n > 0 && f != "proxy_host_groups.go" {
			t.Errorf("%s inlines the bucket SQL %d time(s); use proxyHostDomainBucketExpr", f, n)
		}
	}
}
