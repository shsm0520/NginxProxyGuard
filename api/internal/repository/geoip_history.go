package repository

import (
	"context"
	"database/sql"
	"time"

	"nginx-proxy-guard/internal/model"
)

type GeoIPHistoryRepository struct {
	db *sql.DB
}

func NewGeoIPHistoryRepository(db *sql.DB) *GeoIPHistoryRepository {
	return &GeoIPHistoryRepository{db: db}
}

// Create creates a new update history record
func (r *GeoIPHistoryRepository) Create(ctx context.Context, triggerType model.GeoIPTriggerType) (*model.GeoIPUpdateHistory, error) {
	query := `
		INSERT INTO geoip_update_history (status, trigger_type, started_at)
		VALUES ($1, $2, NOW())
		RETURNING id, status, trigger_type, started_at, completed_at, duration_ms,
		          database_version, country_db_size, asn_db_size, error_message, created_at
	`

	var h model.GeoIPUpdateHistory
	var completedAt sql.NullTime
	var durationMs sql.NullInt32
	var countrySize, asnSize sql.NullInt64
	var dbVersion, errMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, model.GeoIPUpdateStatusRunning, triggerType).Scan(
		&h.ID, &h.Status, &h.TriggerType, &h.StartedAt, &completedAt, &durationMs,
		&dbVersion, &countrySize, &asnSize, &errMsg, &h.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		h.CompletedAt = &completedAt.Time
	}
	if durationMs.Valid {
		d := int(durationMs.Int32)
		h.DurationMs = &d
	}
	if countrySize.Valid {
		h.CountryDBSize = &countrySize.Int64
	}
	if asnSize.Valid {
		h.ASNDBSize = &asnSize.Int64
	}
	h.DatabaseVersion = dbVersion.String
	h.ErrorMessage = errMsg.String

	return &h, nil
}

// UpdateSuccess marks the update as successful
func (r *GeoIPHistoryRepository) UpdateSuccess(ctx context.Context, id string, dbVersion string, countrySize, asnSize int64) error {
	query := `
		UPDATE geoip_update_history
		SET status = $1,
		    completed_at = NOW(),
		    duration_ms = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER * 1000,
		    database_version = $2,
		    country_db_size = $3,
		    asn_db_size = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, model.GeoIPUpdateStatusSuccess, dbVersion, countrySize, asnSize, id)
	return err
}

// UpdateFailed marks the update as failed
func (r *GeoIPHistoryRepository) UpdateFailed(ctx context.Context, id string, errMsg string) error {
	query := `
		UPDATE geoip_update_history
		SET status = $1,
		    completed_at = NOW(),
		    duration_ms = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER * 1000,
		    error_message = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, model.GeoIPUpdateStatusFailed, errMsg, id)
	return err
}

// List returns paginated update history
func (r *GeoIPHistoryRepository) List(ctx context.Context, page, perPage int) (*model.GeoIPUpdateHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM geoip_update_history").Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * perPage
	query := `
		SELECT id, status, trigger_type, started_at, completed_at, duration_ms,
		       database_version, country_db_size, asn_db_size, error_message, created_at
		FROM geoip_update_history
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.GeoIPUpdateHistory
	for rows.Next() {
		var h model.GeoIPUpdateHistory
		var completedAt sql.NullTime
		var durationMs sql.NullInt32
		var countrySize, asnSize sql.NullInt64
		var dbVersion, errMsg sql.NullString

		err := rows.Scan(
			&h.ID, &h.Status, &h.TriggerType, &h.StartedAt, &completedAt, &durationMs,
			&dbVersion, &countrySize, &asnSize, &errMsg, &h.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			h.CompletedAt = &completedAt.Time
		}
		if durationMs.Valid {
			d := int(durationMs.Int32)
			h.DurationMs = &d
		}
		if countrySize.Valid {
			h.CountryDBSize = &countrySize.Int64
		}
		if asnSize.Valid {
			h.ASNDBSize = &asnSize.Int64
		}
		h.DatabaseVersion = dbVersion.String
		h.ErrorMessage = errMsg.String

		history = append(history, h)
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &model.GeoIPUpdateHistoryResponse{
		Data:       history,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetLatestSuccess returns the most recent successful update
func (r *GeoIPHistoryRepository) GetLatestSuccess(ctx context.Context) (*model.GeoIPUpdateHistory, error) {
	query := `
		SELECT id, status, trigger_type, started_at, completed_at, duration_ms,
		       database_version, country_db_size, asn_db_size, error_message, created_at
		FROM geoip_update_history
		WHERE status = $1
		ORDER BY completed_at DESC
		LIMIT 1
	`

	var h model.GeoIPUpdateHistory
	var completedAt sql.NullTime
	var durationMs sql.NullInt32
	var countrySize, asnSize sql.NullInt64
	var dbVersion, errMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, model.GeoIPUpdateStatusSuccess).Scan(
		&h.ID, &h.Status, &h.TriggerType, &h.StartedAt, &completedAt, &durationMs,
		&dbVersion, &countrySize, &asnSize, &errMsg, &h.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		h.CompletedAt = &completedAt.Time
	}
	if durationMs.Valid {
		d := int(durationMs.Int32)
		h.DurationMs = &d
	}
	if countrySize.Valid {
		h.CountryDBSize = &countrySize.Int64
	}
	if asnSize.Valid {
		h.ASNDBSize = &asnSize.Int64
	}
	h.DatabaseVersion = dbVersion.String
	h.ErrorMessage = errMsg.String

	return &h, nil
}

// CleanupOld removes records older than the specified duration
func (r *GeoIPHistoryRepository) CleanupOld(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := r.db.ExecContext(ctx, "DELETE FROM geoip_update_history WHERE created_at < $1", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
