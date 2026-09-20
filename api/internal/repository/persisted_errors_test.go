package repository

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"nginx-proxy-guard/internal/database"
)

// What actually has to hold is not that ScrubDriverText works — that is tested
// where it lives — but that the value reaching the column went through it. So
// assert on the bound argument, which is the only thing that decides what the
// row ends up holding.
//
// Reverting persistedErrorText() at any of these call sites fails this test.

// A representative rendering from lib/pq: message, SQLSTATE, and the statement
// position it appends when the driver kept the query.
const driverError = `failed to reload nginx: pq: invalid input syntax for type uuid: "abc" (22P02) at position 1:63`

// capture is a sqlmock argument matcher that always matches and records the
// string it was handed, so the assertion can be about the value that reached
// the driver rather than about a value the test constructed itself.
type capture struct{ into *string }

func (c capture) Match(v driver.Value) bool {
	switch s := v.(type) {
	case string:
		*c.into = s
	case []byte:
		*c.into = string(s)
	default:
		*c.into = fmt.Sprint(v)
	}
	return true
}

func assertScrubbed(t *testing.T, got string) {
	t.Helper()
	if strings.Contains(got, database.DriverTextPrefix) || strings.Contains(got, "22P02") {
		t.Errorf("driver text was persisted verbatim: %q", got)
	}
	if !strings.HasPrefix(got, "failed to reload nginx") {
		t.Errorf("the scrub took our own wrapper with it, leaving nothing locatable: %q", got)
	}
}

func TestUpdateConfigStatusScrubsDriverTextBeforePersisting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	repo := &ProxyHostRepository{db: &database.DB{DB: db}}

	var bound string
	mock.ExpectExec(`UPDATE proxy_hosts SET config_status`).
		WithArgs("error", capture{&bound}, testProxyHostID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateConfigStatus(context.Background(), testProxyHostID, "error", driverError); err != nil {
		t.Fatalf("UpdateConfigStatus: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	assertScrubbed(t, bound)
}

func TestBackupUpdateStatusScrubsDriverTextBeforePersisting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	repo := NewBackupRepository(db)

	var bound string
	mock.ExpectExec(`UPDATE backups`).
		WithArgs(testProxyHostID, "failed", capture{&bound}).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateStatus(context.Background(), testProxyHostID, "failed", driverError); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	assertScrubbed(t, bound)
}

// An ordinary failure is the common case and must survive byte for byte: an
// nginx -t message is the single most useful thing a host row can carry, and a
// scrub that quietly ate it would be a worse bug than the leak it fixed.
func TestNonDriverErrorsArePersistedVerbatim(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	repo := &ProxyHostRepository{db: &database.DB{DB: db}}

	const nginxError = `nginx: [emerg] duplicate location "/" in /etc/nginx/conf.d/host_1.conf:42`

	var bound string
	mock.ExpectExec(`UPDATE proxy_hosts SET config_status`).
		WithArgs("error", capture{&bound}, testProxyHostID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateConfigStatus(context.Background(), testProxyHostID, "error", nginxError); err != nil {
		t.Fatalf("UpdateConfigStatus: %v", err)
	}
	if bound != nginxError {
		t.Errorf("non-driver error was altered on the way to the column:\n got: %q\nwant: %q", bound, nginxError)
	}
}

func TestPersistedErrorPtrKeepsNilDistinctFromEmpty(t *testing.T) {
	if got := persistedErrorPtr(nil); got != nil {
		t.Errorf("nil became %q — a certificate with no error would start reporting a blank one", *got)
	}

	in := driverError
	got := persistedErrorPtr(&in)
	if got == nil {
		t.Fatal("a real message became nil")
	}
	assertScrubbed(t, *got)
	if in != driverError {
		t.Errorf("the caller's string was mutated in place: %q", in)
	}
}
