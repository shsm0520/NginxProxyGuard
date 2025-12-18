package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"nginx-proxy-guard/internal/model"

	"github.com/lib/pq"
)

type APITokenRepository struct {
	db *sql.DB
}

func NewAPITokenRepository(db *sql.DB) *APITokenRepository {
	return &APITokenRepository{db: db}
}

// Create creates a new API token
func (r *APITokenRepository) Create(ctx context.Context, token *model.APIToken) error {
	query := `
		INSERT INTO api_tokens (
			user_id, name, token_hash, token_prefix,
			permissions, allowed_ips, rate_limit, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	permBytes, _ := json.Marshal(token.Permissions)

	return r.db.QueryRowContext(ctx, query,
		token.UserID,
		token.Name,
		token.TokenHash,
		token.TokenPrefix,
		permBytes,
		pq.Array(token.AllowedIPs),
		token.RateLimit,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)
}

// GetByID retrieves a token by ID
func (r *APITokenRepository) GetByID(ctx context.Context, id string) (*model.APIToken, error) {
	query := `
		SELECT t.id, t.user_id, t.name, t.token_hash, t.token_prefix,
		       t.permissions, t.allowed_ips, t.rate_limit, t.expires_at,
		       t.last_used_at, t.last_used_ip, t.use_count, t.is_active,
		       t.revoked_at, t.revoked_reason, t.created_at, t.updated_at,
		       u.username
		FROM api_tokens t
		JOIN users u ON t.user_id = u.id
		WHERE t.id = $1
	`

	token := &model.APIToken{}
	var permBytes []byte
	var revokedReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&token.ID, &token.UserID, &token.Name, &token.TokenHash, &token.TokenPrefix,
		&permBytes, &token.AllowedIPs, &token.RateLimit, &token.ExpiresAt,
		&token.LastUsedAt, &token.LastUsedIP, &token.UseCount, &token.IsActive,
		&token.RevokedAt, &revokedReason, &token.CreatedAt, &token.UpdatedAt,
		&token.Username,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}

	json.Unmarshal(permBytes, &token.Permissions)
	if revokedReason.Valid {
		token.RevokedReason = &revokedReason.String
	}

	return token, nil
}

// GetByHash retrieves a token by its hash (for authentication)
func (r *APITokenRepository) GetByHash(ctx context.Context, hash string) (*model.APIToken, error) {
	query := `
		SELECT t.id, t.user_id, t.name, t.token_hash, t.token_prefix,
		       t.permissions, t.allowed_ips, t.rate_limit, t.expires_at,
		       t.last_used_at, t.last_used_ip, t.use_count, t.is_active,
		       t.revoked_at, t.revoked_reason, t.created_at, t.updated_at,
		       u.username
		FROM api_tokens t
		JOIN users u ON t.user_id = u.id
		WHERE t.token_hash = $1 AND t.is_active = true
	`

	token := &model.APIToken{}
	var permBytes []byte
	var revokedReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&token.ID, &token.UserID, &token.Name, &token.TokenHash, &token.TokenPrefix,
		&permBytes, &token.AllowedIPs, &token.RateLimit, &token.ExpiresAt,
		&token.LastUsedAt, &token.LastUsedIP, &token.UseCount, &token.IsActive,
		&token.RevokedAt, &revokedReason, &token.CreatedAt, &token.UpdatedAt,
		&token.Username,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API token by hash: %w", err)
	}

	json.Unmarshal(permBytes, &token.Permissions)
	if revokedReason.Valid {
		token.RevokedReason = &revokedReason.String
	}

	return token, nil
}

// ListByUser lists all tokens for a user
func (r *APITokenRepository) ListByUser(ctx context.Context, userID string) ([]*model.APIToken, error) {
	query := `
		SELECT t.id, t.user_id, t.name, t.token_hash, t.token_prefix,
		       t.permissions, t.allowed_ips, t.rate_limit, t.expires_at,
		       t.last_used_at, t.last_used_ip, t.use_count, t.is_active,
		       t.revoked_at, t.revoked_reason, t.created_at, t.updated_at,
		       u.username
		FROM api_tokens t
		JOIN users u ON t.user_id = u.id
		WHERE t.user_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*model.APIToken
	for rows.Next() {
		token := &model.APIToken{}
		var permBytes []byte
		var revokedReason sql.NullString

		err := rows.Scan(
			&token.ID, &token.UserID, &token.Name, &token.TokenHash, &token.TokenPrefix,
			&permBytes, &token.AllowedIPs, &token.RateLimit, &token.ExpiresAt,
			&token.LastUsedAt, &token.LastUsedIP, &token.UseCount, &token.IsActive,
			&token.RevokedAt, &revokedReason, &token.CreatedAt, &token.UpdatedAt,
			&token.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API token: %w", err)
		}

		json.Unmarshal(permBytes, &token.Permissions)
		if revokedReason.Valid {
			token.RevokedReason = &revokedReason.String
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// ListAll lists all tokens (admin only)
func (r *APITokenRepository) ListAll(ctx context.Context) ([]*model.APIToken, error) {
	query := `
		SELECT t.id, t.user_id, t.name, t.token_hash, t.token_prefix,
		       t.permissions, t.allowed_ips, t.rate_limit, t.expires_at,
		       t.last_used_at, t.last_used_ip, t.use_count, t.is_active,
		       t.revoked_at, t.revoked_reason, t.created_at, t.updated_at,
		       u.username
		FROM api_tokens t
		JOIN users u ON t.user_id = u.id
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all API tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*model.APIToken
	for rows.Next() {
		token := &model.APIToken{}
		var permBytes []byte
		var revokedReason sql.NullString

		err := rows.Scan(
			&token.ID, &token.UserID, &token.Name, &token.TokenHash, &token.TokenPrefix,
			&permBytes, &token.AllowedIPs, &token.RateLimit, &token.ExpiresAt,
			&token.LastUsedAt, &token.LastUsedIP, &token.UseCount, &token.IsActive,
			&token.RevokedAt, &revokedReason, &token.CreatedAt, &token.UpdatedAt,
			&token.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API token: %w", err)
		}

		json.Unmarshal(permBytes, &token.Permissions)
		if revokedReason.Valid {
			token.RevokedReason = &revokedReason.String
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Update updates a token
func (r *APITokenRepository) Update(ctx context.Context, id string, req *model.UpdateAPITokenRequest) error {
	// Build dynamic query
	query := "UPDATE api_tokens SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIndex)
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Permissions != nil {
		permBytes, _ := json.Marshal(req.Permissions)
		query += fmt.Sprintf(", permissions = $%d", argIndex)
		args = append(args, permBytes)
		argIndex++
	}
	if req.AllowedIPs != nil {
		query += fmt.Sprintf(", allowed_ips = $%d", argIndex)
		args = append(args, pq.Array(req.AllowedIPs))
		argIndex++
	}
	if req.RateLimit != nil {
		query += fmt.Sprintf(", rate_limit = $%d", argIndex)
		args = append(args, *req.RateLimit)
		argIndex++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIndex)
		args = append(args, *req.IsActive)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// Revoke revokes a token
func (r *APITokenRepository) Revoke(ctx context.Context, id, reason string) error {
	query := `
		UPDATE api_tokens
		SET is_active = false, revoked_at = NOW(), revoked_reason = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, reason)
	return err
}

// Delete permanently deletes a token
func (r *APITokenRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM api_tokens WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateLastUsed updates the last used timestamp and IP
func (r *APITokenRepository) UpdateLastUsed(ctx context.Context, id, ip string) error {
	query := `
		UPDATE api_tokens
		SET last_used_at = NOW(), last_used_ip = $2, use_count = use_count + 1
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, ip)
	return err
}

// LogUsage logs API token usage
func (r *APITokenRepository) LogUsage(ctx context.Context, usage *model.APITokenUsage) error {
	query := `
		INSERT INTO api_token_usage (
			token_id, endpoint, method, status_code, client_ip,
			user_agent, request_body_size, response_time_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		usage.TokenID, usage.Endpoint, usage.Method, usage.StatusCode,
		usage.ClientIP, usage.UserAgent, usage.RequestBodySize, usage.ResponseTimeMs,
	)
	return err
}

// GetUsageStats returns usage statistics for a token
func (r *APITokenRepository) GetUsageStats(ctx context.Context, tokenID string, since time.Time) ([]model.APITokenUsage, error) {
	query := `
		SELECT id, token_id, endpoint, method, status_code, client_ip,
		       user_agent, request_body_size, response_time_ms, created_at
		FROM api_token_usage
		WHERE token_id = $1 AND created_at >= $2
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, tokenID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []model.APITokenUsage
	for rows.Next() {
		var u model.APITokenUsage
		err := rows.Scan(
			&u.ID, &u.TokenID, &u.Endpoint, &u.Method, &u.StatusCode,
			&u.ClientIP, &u.UserAgent, &u.RequestBodySize, &u.ResponseTimeMs, &u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		usages = append(usages, u)
	}

	return usages, nil
}

// CleanupExpiredTokens deactivates expired tokens
func (r *APITokenRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	query := `
		UPDATE api_tokens
		SET is_active = false, revoked_at = NOW(), revoked_reason = 'expired'
		WHERE is_active = true AND expires_at IS NOT NULL AND expires_at < NOW()
	`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanupOldUsageLogs removes usage logs older than the specified duration
func (r *APITokenRepository) CleanupOldUsageLogs(ctx context.Context, before time.Time) (int64, error) {
	query := "DELETE FROM api_token_usage WHERE created_at < $1"
	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
