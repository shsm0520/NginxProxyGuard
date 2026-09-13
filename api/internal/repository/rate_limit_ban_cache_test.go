package repository

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type fakeActiveBanCache struct {
	deleted []string
}

func (f *fakeActiveBanCache) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}

func (f *fakeActiveBanCache) Get(_ context.Context, _ string, _ interface{}) error {
	return assertCacheMiss{}
}

func (f *fakeActiveBanCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

type assertCacheMiss struct{}

func (assertCacheMiss) Error() string { return "cache miss" }

func assertDeletedKeys(t *testing.T, got []string, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("deleted cache keys = %#v, want %#v", got, want)
	}
}

func TestBanAutoIPPreservesAutoMetadataAndInvalidatesGlobalBanCache(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	cache := &fakeActiveBanCache{}
	repo := &RateLimitRepository{db: db, cache: cache}
	now := time.Date(2026, 9, 11, 15, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)

	// One statement, no DELETE: the global partial unique index is the arbiter.
	// A separate DELETE would reopen the window where a failed INSERT lifts an
	// existing ban, so the absence of ExpectExec here is the assertion.
	mock.ExpectQuery(`INSERT INTO banned_ips.*ON CONFLICT \(ip_address\) WHERE proxy_host_id IS NULL DO UPDATE SET`).
		WithArgs(nil, "2001:db8::123", "WAF threshold exceeded", 7, sqlmock.AnyArg(), false, true).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "proxy_host_id", "ip_address", "reason", "fail_count", "banned_at", "expires_at", "is_permanent", "is_auto_banned", "created_at",
		}).AddRow("ban-1", nil, "2001:db8::123", "WAF threshold exceeded", 7, now, expires, false, true, now))

	ban, err := repo.BanAutoIP(ctx, nil, "2001:db8::123", "WAF threshold exceeded", 3600, 7)
	if err != nil {
		t.Fatalf("BanAutoIP: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
	if !ban.IsAutoBanned {
		t.Fatal("BanAutoIP must mark the row as auto-banned")
	}
	if ban.FailCount != 7 {
		t.Fatalf("FailCount = %d, want 7", ban.FailCount)
	}
	assertDeletedKeys(t, cache.deleted, []string{"banned_ips:active:global"})
}

// The host scope needs its own arbiter: the two scopes live on two different
// partial unique indexes, and Postgres infers one arbiter per statement.
func TestBanIPHostScopedUsesHostArbiterAndInvalidatesBothKeys(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	cache := &fakeActiveBanCache{}
	repo := &RateLimitRepository{db: db, cache: cache}
	now := time.Date(2026, 9, 14, 15, 0, 0, 0, time.UTC)
	hostID := "0f1e2d3c-4b5a-6978-8796-a5b4c3d2e1f0"

	mock.ExpectQuery(`INSERT INTO banned_ips.*ON CONFLICT \(ip_address, proxy_host_id\) WHERE proxy_host_id IS NOT NULL DO UPDATE SET`).
		WithArgs(&hostID, "192.0.2.10", "manual", 1, sqlmock.AnyArg(), true, false).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "proxy_host_id", "ip_address", "reason", "fail_count", "banned_at", "expires_at", "is_permanent", "is_auto_banned", "created_at",
		}).AddRow("ban-2", hostID, "192.0.2.10", "manual", 1, now, nil, true, false, now))

	ban, err := repo.BanIP(ctx, &hostID, "192.0.2.10", "manual", 0)
	if err != nil {
		t.Fatalf("BanIP: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
	if ban.IsAutoBanned {
		t.Fatal("a manual ban must not be marked auto-banned")
	}
	// invalidateBans always drops the global key too: a host-scoped write can
	// shadow a global ban, and config generation consumes both lists.
	assertDeletedKeys(t, cache.deleted, []string{
		"banned_ips:active:global",
		"banned_ips:active:host:" + hostID,
	})
}

func TestUnbanIPByAddressInvalidatesEveryAffectedHostBanCache(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	cache := &fakeActiveBanCache{}
	repo := &RateLimitRepository{db: db, cache: cache}

	mock.ExpectQuery(`DELETE FROM banned_ips WHERE ip_address = \$1 RETURNING proxy_host_id`).
		WithArgs("203.0.113.77").
		WillReturnRows(sqlmock.NewRows([]string{"proxy_host_id"}).
			AddRow("host-a").
			AddRow(nil).
			AddRow("host-b"))

	if err := repo.UnbanIPByAddress(ctx, "203.0.113.77"); err != nil {
		t.Fatalf("UnbanIPByAddress: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}

	assertDeletedKeys(t, cache.deleted, []string{
		"banned_ips:active:global",
		"banned_ips:active:host:host-a",
		"banned_ips:active:host:host-b",
	})
}

func TestUnbanIPsByIDsInvalidatesEveryAffectedHostBanCache(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	cache := &fakeActiveBanCache{}
	repo := &RateLimitRepository{db: db, cache: cache}

	mock.ExpectQuery(`DELETE FROM banned_ips WHERE id = ANY\(\$1::uuid\[\]\) RETURNING proxy_host_id`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"proxy_host_id"}).
			AddRow("host-a").
			AddRow("host-b").
			AddRow(nil))

	deleted, err := repo.UnbanIPsByIDs(ctx, []string{
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
	})
	if err != nil {
		t.Fatalf("UnbanIPsByIDs: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("deleted = %d, want 3", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}

	assertDeletedKeys(t, cache.deleted, []string{
		"banned_ips:active:global",
		"banned_ips:active:host:host-a",
		"banned_ips:active:host:host-b",
	})
}
