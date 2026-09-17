package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"nginx-proxy-guard/internal/model"
)

func TestListFilterBuildsAndedPredicates(t *testing.T) {
	repo, mock := newMockProxyHostRepo(t)
	enabled := true
	filter := model.ProxyHostListFilter{Tags: []string{"media", "family"}, Domain: "example.com", Upstream: "192.0.2.9", Enabled: &enabled}

	// Count and data queries share the WHERE; both must carry every predicate,
	// numbered in argument order after the search placeholder.
	where := regexp.QuoteMeta(`WHERE (array_to_string(domain_names, ',') ILIKE $1 OR forward_host ILIKE $1) AND tags @> $2::text[] AND regexp_replace(domain_names[1], '^[^.]+\.', '') = $3 AND forward_host = $4 AND enabled = $5`)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM proxy_hosts `+where).
		WithArgs("%web%", sqlmock.AnyArg(), "example.com", "192.0.2.9", true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`(?s)SELECT id,.*FROM proxy_hosts\s+`+where+`.*LIMIT \$6 OFFSET \$7`).
		WithArgs("%web%", sqlmock.AnyArg(), "example.com", "192.0.2.9", true, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	if _, _, err := repo.List(context.Background(), 1, 20, "web", "", "", filter); err != nil {
		t.Fatalf("List: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListWithoutFilterHasNoWhere(t *testing.T) {
	repo, mock := newMockProxyHostRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM proxy_hosts$`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`(?s)SELECT id,.*FROM proxy_hosts\s+ORDER BY is_favorite DESC, created_at DESC\s+LIMIT \$1 OFFSET \$2`).
		WithArgs(20, 0).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	if _, _, err := repo.List(context.Background(), 1, 20, "", "", "", model.ProxyHostListFilter{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGroupsRunsFourAggregatesAndNeverReturnsNilSlices(t *testing.T) {
	repo, mock := newMockProxyHostRepo(t)
	mock.ExpectQuery(`SELECT t, COUNT\(\*\) FROM proxy_hosts, unnest\(tags\) AS t GROUP BY t ORDER BY 2 DESC, 1 ASC`).
		WillReturnRows(sqlmock.NewRows([]string{"t", "count"}).AddRow("media", 4).AddRow("family", 2))
	mock.ExpectQuery(`SELECT regexp_replace\(domain_names\[1\], '\^\[\^\.\]\+\\\.', ''\), COUNT\(\*\) FROM proxy_hosts GROUP BY 1 ORDER BY 2 DESC, 1 ASC`).
		WillReturnRows(sqlmock.NewRows([]string{"d", "count"}).AddRow("example.com", 6))
	mock.ExpectQuery(`SELECT forward_host, COUNT\(\*\) FROM proxy_hosts GROUP BY 1 ORDER BY 2 DESC, 1 ASC`).
		WillReturnRows(sqlmock.NewRows([]string{"u", "count"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FILTER \(WHERE enabled\), COUNT\(\*\) FILTER \(WHERE NOT enabled\) FROM proxy_hosts`).
		WillReturnRows(sqlmock.NewRows([]string{"e", "d"}).AddRow(5, 1))

	g, err := repo.Groups(context.Background())
	if err != nil {
		t.Fatalf("Groups: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if len(g.Tags) != 2 || g.Tags[0].Name != "media" || g.Tags[0].Count != 4 {
		t.Fatalf("tags = %+v", g.Tags)
	}
	if g.Upstreams == nil {
		t.Fatal("empty groups must be [] not null")
	}
	if g.Status.Enabled != 5 || g.Status.Disabled != 1 {
		t.Fatalf("status = %+v", g.Status)
	}
}
