package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"nginx-proxy-guard/internal/model"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// AuditLog represents a stored audit log entry
type AuditLog struct {
	ID           string          `json:"id"`
	UserID       *string         `json:"user_id,omitempty"`
	Username     string          `json:"username"`
	Action       string          `json:"action"`
	ResourceType *string         `json:"resource_type,omitempty"`
	ResourceID   *string         `json:"resource_id,omitempty"`
	ResourceName *string         `json:"resource_name,omitempty"`
	Details      json.RawMessage `json:"details,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// AuditLogFilter defines filter options for listing audit logs
type AuditLogFilter struct {
	UserID       string     `json:"user_id,omitempty"`
	Action       string     `json:"action,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	Search       string     `json:"search,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// Log creates an audit log entry
func (r *AuditLogRepository) Log(ctx context.Context, entry *model.AuditLogEntry) error {
	query := `
		INSERT INTO audit_logs (
			user_id, username, action, resource_type, resource_id, resource_name,
			details, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	var detailsBytes []byte
	if entry.Details != nil {
		detailsBytes, _ = json.Marshal(entry.Details)
	}

	// Get resource name from details if available
	var resourceName *string
	if entry.Details != nil {
		if name, ok := entry.Details["name"].(string); ok {
			resourceName = &name
		} else if name, ok := entry.Details["domain_names"].([]interface{}); ok && len(name) > 0 {
			if s, ok := name[0].(string); ok {
				resourceName = &s
			}
		} else if name, ok := entry.Details["token_name"].(string); ok {
			resourceName = &name
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		nullString(entry.UserID),
		entry.Username,
		entry.Action,
		nullString(entry.Resource),
		nullString(entry.ResourceID),
		resourceName,
		detailsBytes,
		nullString(entry.IPAddress),
		nullString(entry.UserAgent),
	)
	return err
}

// List retrieves audit logs with filters
func (r *AuditLogRepository) List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int, error) {
	// Build WHERE clause
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.UserID != "" {
		where += " AND user_id = $" + itoa(argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}

	if filter.Action != "" {
		where += " AND action ILIKE $" + itoa(argIdx)
		args = append(args, "%"+filter.Action+"%")
		argIdx++
	}

	if filter.ResourceType != "" {
		where += " AND resource_type = $" + itoa(argIdx)
		args = append(args, filter.ResourceType)
		argIdx++
	}

	if filter.Search != "" {
		where += " AND (username ILIKE $" + itoa(argIdx) + " OR action ILIKE $" + itoa(argIdx) + " OR resource_name ILIKE $" + itoa(argIdx) + ")"
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	if filter.StartTime != nil {
		where += " AND created_at >= $" + itoa(argIdx)
		args = append(args, *filter.StartTime)
		argIdx++
	}

	if filter.EndTime != nil {
		where += " AND created_at <= $" + itoa(argIdx)
		args = append(args, *filter.EndTime)
		argIdx++
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM audit_logs " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, user_id, username, action, resource_type, resource_id, resource_name,
		       details, ip_address, user_agent, created_at
		FROM audit_logs ` + where + `
		ORDER BY created_at DESC
		LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var details sql.NullString

		err := rows.Scan(
			&log.ID, &log.UserID, &log.Username, &log.Action,
			&log.ResourceType, &log.ResourceID, &log.ResourceName,
			&details, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if details.Valid {
			log.Details = json.RawMessage(details.String)
		}

		logs = append(logs, log)
	}

	return logs, total, rows.Err()
}

// GetActions returns distinct action types
func (r *AuditLogRepository) GetActions(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT action FROM audit_logs ORDER BY action`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []string
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// GetResourceTypes returns distinct resource types
func (r *AuditLogRepository) GetResourceTypes(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT resource_type FROM audit_logs WHERE resource_type IS NOT NULL ORDER BY resource_type`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		types = append(types, t)
	}

	return types, nil
}

// Cleanup removes old audit logs
func (r *AuditLogRepository) Cleanup(ctx context.Context, retentionDays int) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM audit_logs
		WHERE created_at < NOW() - ($1 || ' days')::INTERVAL
	`, retentionDays)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
