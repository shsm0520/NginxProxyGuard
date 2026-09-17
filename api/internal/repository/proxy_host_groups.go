package repository

import (
	"context"
	"fmt"

	"nginx-proxy-guard/internal/model"
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
	if err := r.groupCounts(ctx, `SELECT regexp_replace(domain_names[1], '^[^.]+\.', ''), COUNT(*) FROM proxy_hosts GROUP BY 1 ORDER BY 2 DESC, 1 ASC`, &out.Domains); err != nil {
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
