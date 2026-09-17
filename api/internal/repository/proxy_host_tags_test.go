package repository

import (
	"context"
	"database/sql/driver"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"nginx-proxy-guard/internal/database"
	"nginx-proxy-guard/internal/model"
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

// --- nil slice vs NOT NULL ------------------------------------------------
//
// pq.StringArray(nil).Value() returns (nil, nil), which binds as SQL NULL, and
// an explicit NULL does not fall back to the column DEFAULT. proxy_hosts.tags
// is NOT NULL, so any write path that can reach the driver with a nil slice
// answers 500 with `pq: null value in column "tags" ... (23502)`. Create sees a
// nil slice on every request that omits the "tags" key — the UI, the clone
// path and every existing API client — and Update sees one whenever GetByID
// hands back a host whose tags never made it into the struct. The guard tests
// above are source inspection and cannot see this; these two look at the driver
// value actually bound.

// capturedArg records the driver value bound at its position and rejects nil,
// which is the exact shape of the bug.
type capturedArg struct {
	seen  bool
	value driver.Value
}

func (c *capturedArg) Match(v driver.Value) bool {
	c.seen = true
	c.value = v
	return v != nil
}

// assertEmptyArray fails unless an empty PostgreSQL array was bound. err is the
// error the repository call returned, reported only to explain a missed bind
// (a changed column count makes sqlmock reject the call before the match).
func (c *capturedArg) assertEmptyArray(t *testing.T, what string, err error) {
	t.Helper()
	if !c.seen {
		t.Fatalf("%s: no tags argument was bound (call returned %v) — if the column list moved, fix the placeholder count in this test", what, err)
	}
	if c.value == nil {
		t.Fatalf("%s: tags bound as nil, which reaches PostgreSQL as NULL and violates the NOT NULL constraint on proxy_hosts.tags (23502)", what)
	}
	got := ""
	switch v := c.value.(type) {
	case string:
		got = v
	case []byte:
		got = string(v)
	default:
		t.Fatalf("%s: tags bound as %T(%v), want the empty array literal", what, c.value, c.value)
	}
	if got != "{}" {
		t.Fatalf("%s: tags bound as %q, want %q", what, got, "{}")
	}
}

// bindsTagsLast expects n arguments of any value followed by the tags capture,
// so the assertion is about the tags column and nothing else.
func bindsTagsLast(n int, c *capturedArg) []driver.Value {
	args := make([]driver.Value, 0, n+1)
	for i := 0; i < n; i++ {
		args = append(args, sqlmock.AnyArg())
	}
	return append(args, c)
}

const testProxyHostID = "11111111-1111-1111-1111-111111111111"

// proxyHostRow is one row in the column order that Create's RETURNING and
// GetByID's SELECT both scan (the same 55 columns). tags is last, so a caller
// can pass nil to model a host whose tags never reached the struct.
func proxyHostRow(tags driver.Value) []driver.Value {
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	return []driver.Value{
		testProxyHostID,     // id
		"http",              // proxy_type
		"{app.example.com}", // domain_names
		"http",              // forward_scheme
		"192.0.2.9",         // forward_host
		nil,                 // forward_container_name
		nil,                 // forward_container_network
		int64(8080),         // forward_port
		"",                  // stream_listen_host
		int64(0),            // stream_listen_port
		"tcp",               // stream_protocol
		false,               // stream_ssl_preread
		false,               // stream_accept_proxy_protocol
		false,               // stream_send_proxy_protocol
		int64(0),            // stream_proxy_connect_timeout
		int64(0),            // stream_proxy_timeout
		false,               // ssl_enabled
		false,               // ssl_force_https
		false,               // ssl_http2
		false,               // ssl_http3
		nil,                 // certificate_id
		false,               // allow_websocket_upgrade
		false,               // cache_enabled
		false,               // cache_static_only
		"7d",                // cache_ttl
		false,               // block_exploits
		"",                  // block_exploits_exceptions
		nil,                 // custom_locations
		"",                  // advanced_config
		false,               // waf_enabled
		"blocking",          // waf_mode
		int64(1),            // waf_paranoia_level
		int64(5),            // waf_anomaly_threshold
		false,               // waf_use_global
		int64(0),            // proxy_connect_timeout
		int64(0),            // proxy_send_timeout
		int64(0),            // proxy_read_timeout
		"",                  // proxy_buffering
		"",                  // proxy_request_buffering
		"",                  // client_max_body_size
		"",                  // proxy_max_temp_file_size
		nil,                 // access_list_id
		true,                // enabled
		false,               // is_favorite
		"ok",                // config_status
		"",                  // config_error
		false,               // ddns_enabled
		nil,                 // ddns_provider_id
		false,               // ddns_proxied
		nil,                 // auth_provider_id
		"{}",                // auth_bypass_paths
		nil,                 // meta
		now,                 // created_at
		now,                 // updated_at
		tags,                // tags
	}
}

// Scans are positional, so the names only have to be the right count.
func proxyHostRowColumns() []string {
	cols := make([]string, len(proxyHostRow(nil)))
	for i := range cols {
		cols[i] = fmt.Sprintf("c%d", i)
	}
	return cols
}

func newMockProxyHostRepo(t *testing.T) (*ProxyHostRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &ProxyHostRepository{db: &database.DB{DB: db}}, mock
}

func TestCreateWithoutTagsBindsEmptyArrayNotNil(t *testing.T) {
	repo, mock := newMockProxyHostRepo(t)

	capture := &capturedArg{}
	mock.ExpectQuery(`INSERT INTO proxy_hosts`).
		WithArgs(bindsTagsLast(46, capture)...).
		WillReturnRows(sqlmock.NewRows(proxyHostRowColumns()).AddRow(proxyHostRow("{}")...))

	// No Tags field: the shape every pre-tags client, the UI and the clone path
	// still send.
	_, err := repo.Create(context.Background(), &model.CreateProxyHostRequest{
		DomainNames:   []string{"app.example.com"},
		ForwardScheme: "http",
		ForwardHost:   "192.0.2.9",
		ForwardPort:   8080,
	})
	capture.assertEmptyArray(t, "Create", err)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateWithNilExistingTagsBindsEmptyArrayNotNil(t *testing.T) {
	repo, mock := newMockProxyHostRepo(t)

	// A row with NULL tags leaves existing.Tags nil, the same state a cached
	// host deserialised from a blob written before the column existed arrives
	// in. Update must not pass that straight through to the driver.
	mock.ExpectQuery(`SELECT id, COALESCE\(proxy_type`).
		WithArgs(testProxyHostID).
		WillReturnRows(sqlmock.NewRows(proxyHostRowColumns()).AddRow(proxyHostRow(nil)...))

	capture := &capturedArg{}
	mock.ExpectQuery(`UPDATE proxy_hosts SET`).
		WithArgs(bindsTagsLast(47, capture)...).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)))

	// No Tags on the request either, so the merge leaves existing.Tags as it is.
	_, err := repo.Update(context.Background(), testProxyHostID, &model.UpdateProxyHostRequest{ForwardPort: 8081})
	capture.assertEmptyArray(t, "Update", err)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
