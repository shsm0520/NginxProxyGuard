package repository

import (
	"context"
	"fmt"

	"nginx-proxy-guard/internal/model"
)

// proxyHostDomainBucketExpr is the one definition of a host's "parent domain"
// bucket. The /groups aggregate and the ?domain= list filter must use the same
// expression — if they drift, clicking a group returns zero rows.
//
//	app.example.com  (3+ labels)      -> example.com     (strip the first label)
//	example.com      (exactly 2)      -> example.com     (an apex host is its own
//	                                                      bucket; stripping
//	                                                      unconditionally filed it
//	                                                      under "com")
//	listener / "" / NULL (0-1 labels) -> NULL            (a stream listener name or
//	                                                      a domainless row has no
//	                                                      parent domain; NULL is
//	                                                      excluded from the
//	                                                      aggregate and never
//	                                                      matches the filter)
//
// NULL matters twice: the aggregate scans into a string, so a NULL bucket would
// fail with "converting NULL to string is unsupported" and turn GET
// /proxy-hosts/groups into a 500 for every caller.
const proxyHostDomainBucketExpr = `CASE WHEN cardinality(string_to_array(domain_names[1], '.')) > 2 THEN regexp_replace(domain_names[1], '^[^.]+\.', '') WHEN cardinality(string_to_array(domain_names[1], '.')) = 2 THEN domain_names[1] END`

// proxyHostDomainGroupQuery buckets every host by parent domain, dropping the
// rows that have none. Package-level so the DB-backed test runs this exact
// statement.
var proxyHostDomainGroupQuery = fmt.Sprintf(
	`SELECT bucket, COUNT(*) FROM (SELECT %s AS bucket FROM proxy_hosts) b WHERE bucket IS NOT NULL AND bucket <> '' GROUP BY 1 ORDER BY 2 DESC, 1 ASC`,
	proxyHostDomainBucketExpr,
)

// Groups returns the buckets the list's filter panel is drawn from. Four
// small aggregates; tag and domain cardinality is bounded by the host count,
// so there is no pagination. Slices start non-nil so the JSON is [] not null.
func (r *ProxyHostRepository) Groups(ctx context.Context) (*model.ProxyHostGroups, error) {
	out := &model.ProxyHostGroups{
		Tags:      []model.ProxyHostGroupCount{},
		Domains:   []model.ProxyHostGroupCount{},
		Upstreams: []model.ProxyHostGroupCount{},
	}
	if err := r.groupCounts(ctx, `SELECT t, COUNT(*) FROM proxy_hosts, unnest(tags) AS t GROUP BY t ORDER BY 2 DESC, 1 ASC`, &out.Tags); err != nil {
		return nil, fmt.Errorf("group by tag: %w", err)
	}
	if err := r.groupCounts(ctx, proxyHostDomainGroupQuery, &out.Domains); err != nil {
		return nil, fmt.Errorf("group by domain: %w", err)
	}
	if err := r.groupCounts(ctx, `SELECT forward_host, COUNT(*) FROM proxy_hosts GROUP BY 1 ORDER BY 2 DESC, 1 ASC`, &out.Upstreams); err != nil {
		return nil, fmt.Errorf("group by upstream: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER (WHERE enabled), COUNT(*) FILTER (WHERE NOT enabled) FROM proxy_hosts`).
		Scan(&out.Status.Enabled, &out.Status.Disabled); err != nil {
		return nil, fmt.Errorf("group by status: %w", err)
	}
	return out, nil
}

func (r *ProxyHostRepository) groupCounts(ctx context.Context, query string, dst *[]model.ProxyHostGroupCount) error {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var c model.ProxyHostGroupCount
		if err := rows.Scan(&c.Name, &c.Count); err != nil {
			return err
		}
		*dst = append(*dst, c)
	}
	return rows.Err()
}
