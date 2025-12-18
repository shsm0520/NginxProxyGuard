package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nginx-proxy-guard/internal/model"
)

// IPBanHistoryRepository handles database operations for IP ban history
type IPBanHistoryRepository struct {
	db *sql.DB
}

// NewIPBanHistoryRepository creates a new IPBanHistoryRepository
func NewIPBanHistoryRepository(db *sql.DB) *IPBanHistoryRepository {
	return &IPBanHistoryRepository{db: db}
}

// RecordBanEvent records a ban event in the history
func (r *IPBanHistoryRepository) RecordBanEvent(ctx context.Context, event *model.IPBanHistory) error {
	var metadataJSON interface{}
	if event.Metadata != nil && len(event.Metadata) > 0 {
		jsonBytes, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = jsonBytes
	}

	query := `
		INSERT INTO ip_ban_history (
			event_type, ip_address, proxy_host_id, domain_name, reason, source,
			ban_duration, expires_at, is_permanent, is_auto, fail_count,
			user_id, user_email, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		event.EventType,
		event.IPAddress,
		event.ProxyHostID,
		event.DomainName,
		event.Reason,
		event.Source,
		event.BanDuration,
		event.ExpiresAt,
		event.IsPermanent,
		event.IsAuto,
		event.FailCount,
		event.UserID,
		event.UserEmail,
		metadataJSON,
		now,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to insert ban history: %w", err)
	}

	event.CreatedAt = now
	return nil
}

// List retrieves ban history with pagination and filtering
func (r *IPBanHistoryRepository) List(ctx context.Context, filter *model.IPBanHistoryFilter) (*model.IPBanHistoryListResponse, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argIndex))
		args = append(args, filter.IPAddress)
		argIndex++
	}

	if filter.EventType != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIndex))
		args = append(args, filter.EventType)
		argIndex++
	}

	if filter.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", argIndex))
		args = append(args, filter.Source)
		argIndex++
	}

	if filter.ProxyHostID != "" {
		conditions = append(conditions, fmt.Sprintf("proxy_host_id = $%d", argIndex))
		args = append(args, filter.ProxyHostID)
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ip_ban_history %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count ban history: %w", err)
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	offset := (filter.Page - 1) * filter.PerPage
	totalPages := (total + filter.PerPage - 1) / filter.PerPage

	// Query with pagination
	query := fmt.Sprintf(`
		SELECT id, event_type, ip_address, proxy_host_id, domain_name, reason, source,
			   ban_duration, expires_at, is_permanent, is_auto, fail_count,
			   user_id, user_email, metadata, created_at
		FROM ip_ban_history
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ban history: %w", err)
	}
	defer rows.Close()

	var history []model.IPBanHistory
	for rows.Next() {
		var h model.IPBanHistory
		var proxyHostID, userID sql.NullString
		var domainName, reason, userEmail sql.NullString
		var banDuration, failCount sql.NullInt32
		var expiresAt sql.NullTime
		var metadataJSON []byte

		err := rows.Scan(
			&h.ID,
			&h.EventType,
			&h.IPAddress,
			&proxyHostID,
			&domainName,
			&reason,
			&h.Source,
			&banDuration,
			&expiresAt,
			&h.IsPermanent,
			&h.IsAuto,
			&failCount,
			&userID,
			&userEmail,
			&metadataJSON,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ban history: %w", err)
		}

		if proxyHostID.Valid {
			h.ProxyHostID = &proxyHostID.String
		}
		if userID.Valid {
			h.UserID = &userID.String
		}
		h.DomainName = domainName.String
		h.Reason = reason.String
		h.UserEmail = userEmail.String

		if banDuration.Valid {
			dur := int(banDuration.Int32)
			h.BanDuration = &dur
		}
		if failCount.Valid {
			fc := int(failCount.Int32)
			h.FailCount = &fc
		}
		if expiresAt.Valid {
			h.ExpiresAt = &expiresAt.Time
		}

		if len(metadataJSON) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
				h.Metadata = metadata
			}
		}

		history = append(history, h)
	}

	return &model.IPBanHistoryListResponse{
		Data:       history,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetByIP retrieves all ban history for a specific IP
func (r *IPBanHistoryRepository) GetByIP(ctx context.Context, ipAddress string, page, perPage int) (*model.IPBanHistoryListResponse, error) {
	return r.List(ctx, &model.IPBanHistoryFilter{
		IPAddress: ipAddress,
		Page:      page,
		PerPage:   perPage,
	})
}

// GetStats retrieves ban history statistics
func (r *IPBanHistoryRepository) GetStats(ctx context.Context) (*model.IPBanHistoryStats, error) {
	stats := &model.IPBanHistoryStats{
		BansBySource: make(map[string]int),
	}

	// Count total bans
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM ip_ban_history WHERE event_type = 'ban'",
	).Scan(&stats.TotalBans); err != nil {
		return nil, fmt.Errorf("failed to count bans: %w", err)
	}

	// Count total unbans
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM ip_ban_history WHERE event_type = 'unban'",
	).Scan(&stats.TotalUnbans); err != nil {
		return nil, fmt.Errorf("failed to count unbans: %w", err)
	}

	// Count active bans (from banned_ips table)
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM banned_ips WHERE expires_at IS NULL OR expires_at > NOW()",
	).Scan(&stats.ActiveBans); err != nil {
		return nil, fmt.Errorf("failed to count active bans: %w", err)
	}

	// Bans by source
	rows, err := r.db.QueryContext(ctx, `
		SELECT source, COUNT(*) as count
		FROM ip_ban_history
		WHERE event_type = 'ban'
		GROUP BY source
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query bans by source: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, fmt.Errorf("failed to scan bans by source: %w", err)
		}
		stats.BansBySource[source] = count
	}

	// Top banned IPs
	rows, err = r.db.QueryContext(ctx, `
		SELECT ip_address, COUNT(*) as ban_count
		FROM ip_ban_history
		WHERE event_type = 'ban'
		GROUP BY ip_address
		ORDER BY ban_count DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query top banned IPs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ipCount model.IPBanCount
		if err := rows.Scan(&ipCount.IPAddress, &ipCount.BanCount); err != nil {
			return nil, fmt.Errorf("failed to scan top banned IP: %w", err)
		}
		stats.TopBannedIPs = append(stats.TopBannedIPs, ipCount)
	}

	return stats, nil
}

// DeleteOldHistory deletes history older than the specified number of days
// Returns the number of deleted records
func (r *IPBanHistoryRepository) DeleteOldHistory(ctx context.Context, olderThanDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM ip_ban_history WHERE created_at < $1",
		cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old history: %w", err)
	}
	return result.RowsAffected()
}
