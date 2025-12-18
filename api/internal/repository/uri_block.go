package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"nginx-proxy-guard/internal/database"
	"nginx-proxy-guard/internal/model"
)

type URIBlockRepository struct {
	db *database.DB
}

func NewURIBlockRepository(db *database.DB) *URIBlockRepository {
	return &URIBlockRepository{db: db}
}

func (r *URIBlockRepository) GetByProxyHostID(ctx context.Context, proxyHostID string) (*model.URIBlock, error) {
	var ub model.URIBlock
	var rulesJSON []byte
	var exceptionIPs pq.StringArray
	var allowPrivateIPs sql.NullBool

	err := r.db.QueryRowContext(ctx, `
		SELECT id, proxy_host_id, enabled, rules, COALESCE(exception_ips, '{}'),
		       COALESCE(allow_private_ips, true), created_at, updated_at
		FROM uri_blocks WHERE proxy_host_id = $1
	`, proxyHostID).Scan(
		&ub.ID, &ub.ProxyHostID, &ub.Enabled, &rulesJSON, &exceptionIPs,
		&allowPrivateIPs, &ub.CreatedAt, &ub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse rules JSON
	if len(rulesJSON) > 0 {
		if err := json.Unmarshal(rulesJSON, &ub.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}
	}
	if ub.Rules == nil {
		ub.Rules = []model.URIBlockRule{}
	}

	ub.ExceptionIPs = []string(exceptionIPs)
	ub.AllowPrivateIPs = allowPrivateIPs.Bool
	return &ub, nil
}

func (r *URIBlockRepository) Upsert(ctx context.Context, proxyHostID string, req *model.CreateURIBlockRequest) (*model.URIBlock, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	allowPrivateIPs := true
	if req.AllowPrivateIPs != nil {
		allowPrivateIPs = *req.AllowPrivateIPs
	}

	rules := req.Rules
	if rules == nil {
		rules = []model.URIBlockRule{}
	}

	// Ensure all rules have proper UUIDs (replace empty or temp-* IDs)
	for i := range rules {
		if rules[i].ID == "" || strings.HasPrefix(rules[i].ID, "temp-") {
			rules[i].ID = uuid.New().String()
		}
	}

	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rules: %w", err)
	}

	exceptionIPs := req.ExceptionIPs
	if exceptionIPs == nil {
		exceptionIPs = []string{}
	}

	var id string
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO uri_blocks (proxy_host_id, enabled, rules, exception_ips, allow_private_ips)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (proxy_host_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			rules = EXCLUDED.rules,
			exception_ips = EXCLUDED.exception_ips,
			allow_private_ips = EXCLUDED.allow_private_ips,
			updated_at = NOW()
		RETURNING id
	`, proxyHostID, enabled, rulesJSON, pq.Array(exceptionIPs), allowPrivateIPs).Scan(&id)
	if err != nil {
		return nil, err
	}

	return r.GetByProxyHostID(ctx, proxyHostID)
}

// AddRule adds a single rule to existing URI block settings (for quick add from log viewer)
func (r *URIBlockRepository) AddRule(ctx context.Context, proxyHostID string, req *model.AddURIBlockRuleRequest) (*model.URIBlock, error) {
	existing, err := r.GetByProxyHostID(ctx, proxyHostID)
	if err != nil {
		return nil, err
	}

	// Create new rule
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	newRule := model.URIBlockRule{
		ID:          uuid.New().String(),
		Pattern:     req.Pattern,
		MatchType:   req.MatchType,
		Description: req.Description,
		Enabled:     enabled,
	}

	if existing == nil {
		// Create new URI block with this rule
		return r.Upsert(ctx, proxyHostID, &model.CreateURIBlockRequest{
			Enabled: &enabled,
			Rules:   []model.URIBlockRule{newRule},
		})
	}

	// Check if pattern already exists
	for _, rule := range existing.Rules {
		if rule.Pattern == req.Pattern && rule.MatchType == req.MatchType {
			return existing, nil // Pattern already exists
		}
	}

	// Add new rule
	existing.Rules = append(existing.Rules, newRule)

	rulesJSON, err := json.Marshal(existing.Rules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rules: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE uri_blocks SET rules = $1, updated_at = NOW() WHERE proxy_host_id = $2
	`, rulesJSON, proxyHostID)
	if err != nil {
		return nil, err
	}

	return r.GetByProxyHostID(ctx, proxyHostID)
}

// RemoveRule removes a single rule by ID
func (r *URIBlockRepository) RemoveRule(ctx context.Context, proxyHostID string, ruleID string) error {
	existing, err := r.GetByProxyHostID(ctx, proxyHostID)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	// Filter out the rule
	var newRules []model.URIBlockRule
	for _, rule := range existing.Rules {
		if rule.ID != ruleID {
			newRules = append(newRules, rule)
		}
	}

	if newRules == nil {
		newRules = []model.URIBlockRule{}
	}

	rulesJSON, err := json.Marshal(newRules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE uri_blocks SET rules = $1, updated_at = NOW() WHERE proxy_host_id = $2
	`, rulesJSON, proxyHostID)
	return err
}

func (r *URIBlockRepository) Delete(ctx context.Context, proxyHostID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM uri_blocks WHERE proxy_host_id = $1`, proxyHostID)
	return err
}

// URIBlockWithHost includes host information for listing
type URIBlockWithHost struct {
	model.URIBlock
	DomainNames []string `json:"domain_names"`
	HostEnabled bool     `json:"host_enabled"`
}

// ListAll returns all URI blocks with host information
func (r *URIBlockRepository) ListAll(ctx context.Context) ([]URIBlockWithHost, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ub.id, ub.proxy_host_id, ub.enabled, ub.rules,
		       COALESCE(ub.exception_ips, '{}'), COALESCE(ub.allow_private_ips, true),
		       ub.created_at, ub.updated_at,
		       ph.domain_names, ph.enabled as host_enabled
		FROM uri_blocks ub
		JOIN proxy_hosts ph ON ub.proxy_host_id = ph.id
		ORDER BY ph.domain_names[1]
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []URIBlockWithHost
	for rows.Next() {
		var ub URIBlockWithHost
		var rulesJSON []byte
		var exceptionIPs pq.StringArray
		var allowPrivateIPs sql.NullBool
		var domainNames pq.StringArray

		if err := rows.Scan(
			&ub.ID, &ub.ProxyHostID, &ub.Enabled, &rulesJSON,
			&exceptionIPs, &allowPrivateIPs, &ub.CreatedAt, &ub.UpdatedAt,
			&domainNames, &ub.HostEnabled,
		); err != nil {
			return nil, err
		}

		if len(rulesJSON) > 0 {
			if err := json.Unmarshal(rulesJSON, &ub.Rules); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
			}
		}
		if ub.Rules == nil {
			ub.Rules = []model.URIBlockRule{}
		}

		ub.ExceptionIPs = []string(exceptionIPs)
		ub.AllowPrivateIPs = allowPrivateIPs.Bool
		ub.DomainNames = []string(domainNames)
		result = append(result, ub)
	}

	if result == nil {
		result = []URIBlockWithHost{}
	}
	return result, rows.Err()
}

// ============================================================================
// Global URI Block Repository Functions
// ============================================================================

// GetGlobalURIBlock returns the global URI block settings
func (r *URIBlockRepository) GetGlobalURIBlock(ctx context.Context) (*model.GlobalURIBlock, error) {
	var ub model.GlobalURIBlock
	var rulesJSON []byte
	var exceptionIPs pq.StringArray

	err := r.db.QueryRowContext(ctx, `
		SELECT id, enabled, rules, COALESCE(exception_ips, '{}'),
		       COALESCE(allow_private_ips, true), created_at, updated_at
		FROM global_uri_blocks
		LIMIT 1
	`).Scan(
		&ub.ID, &ub.Enabled, &rulesJSON, &exceptionIPs,
		&ub.AllowPrivateIPs, &ub.CreatedAt, &ub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(rulesJSON) > 0 {
		if err := json.Unmarshal(rulesJSON, &ub.Rules); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}
	}
	if ub.Rules == nil {
		ub.Rules = []model.URIBlockRule{}
	}

	ub.ExceptionIPs = []string(exceptionIPs)
	return &ub, nil
}

// UpsertGlobalURIBlock creates or updates the global URI block settings
func (r *URIBlockRepository) UpsertGlobalURIBlock(ctx context.Context, req *model.CreateGlobalURIBlockRequest) (*model.GlobalURIBlock, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	allowPrivateIPs := true
	if req.AllowPrivateIPs != nil {
		allowPrivateIPs = *req.AllowPrivateIPs
	}

	rules := req.Rules
	if rules == nil {
		rules = []model.URIBlockRule{}
	}

	// Ensure all rules have proper UUIDs (replace empty or temp-* IDs)
	for i := range rules {
		if rules[i].ID == "" || strings.HasPrefix(rules[i].ID, "temp-") {
			rules[i].ID = uuid.New().String()
		}
	}

	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rules: %w", err)
	}

	exceptionIPs := req.ExceptionIPs
	if exceptionIPs == nil {
		exceptionIPs = []string{}
	}

	// Check if global block exists
	existing, _ := r.GetGlobalURIBlock(ctx)
	if existing == nil {
		// Insert new
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO global_uri_blocks (enabled, rules, exception_ips, allow_private_ips)
			VALUES ($1, $2, $3, $4)
		`, enabled, rulesJSON, pq.Array(exceptionIPs), allowPrivateIPs)
	} else {
		// Update existing
		_, err = r.db.ExecContext(ctx, `
			UPDATE global_uri_blocks SET
				enabled = $1,
				rules = $2,
				exception_ips = $3,
				allow_private_ips = $4,
				updated_at = NOW()
			WHERE id = $5
		`, enabled, rulesJSON, pq.Array(exceptionIPs), allowPrivateIPs, existing.ID)
	}

	if err != nil {
		return nil, err
	}

	return r.GetGlobalURIBlock(ctx)
}

// AddGlobalRule adds a single rule to the global URI block settings
func (r *URIBlockRepository) AddGlobalRule(ctx context.Context, req *model.AddURIBlockRuleRequest) (*model.GlobalURIBlock, error) {
	existing, err := r.GetGlobalURIBlock(ctx)
	if err != nil {
		return nil, err
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	newRule := model.URIBlockRule{
		ID:          uuid.New().String(),
		Pattern:     req.Pattern,
		MatchType:   req.MatchType,
		Description: req.Description,
		Enabled:     enabled,
	}

	if existing == nil {
		// Create new global URI block with this rule
		return r.UpsertGlobalURIBlock(ctx, &model.CreateGlobalURIBlockRequest{
			Enabled: &enabled,
			Rules:   []model.URIBlockRule{newRule},
		})
	}

	// Check if pattern already exists
	for _, rule := range existing.Rules {
		if rule.Pattern == req.Pattern && rule.MatchType == req.MatchType {
			return existing, nil
		}
	}

	// Add new rule
	existing.Rules = append(existing.Rules, newRule)

	rulesJSON, err := json.Marshal(existing.Rules)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rules: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE global_uri_blocks SET rules = $1, updated_at = NOW() WHERE id = $2
	`, rulesJSON, existing.ID)
	if err != nil {
		return nil, err
	}

	return r.GetGlobalURIBlock(ctx)
}

// RemoveGlobalRule removes a single rule from global URI block settings
func (r *URIBlockRepository) RemoveGlobalRule(ctx context.Context, ruleID string) error {
	existing, err := r.GetGlobalURIBlock(ctx)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	// Filter out the rule
	var newRules []model.URIBlockRule
	for _, rule := range existing.Rules {
		if rule.ID != ruleID {
			newRules = append(newRules, rule)
		}
	}

	if newRules == nil {
		newRules = []model.URIBlockRule{}
	}

	rulesJSON, err := json.Marshal(newRules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE global_uri_blocks SET rules = $1, updated_at = NOW() WHERE id = $2
	`, rulesJSON, existing.ID)
	return err
}
