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
