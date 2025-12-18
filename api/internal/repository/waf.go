package repository

import (
	"context"
	"database/sql"
	"fmt"

	"nginx-proxy-guard/internal/database"
	"nginx-proxy-guard/internal/model"
)

type WAFRepository struct {
	db *database.DB
}

func NewWAFRepository(db *database.DB) *WAFRepository {
	return &WAFRepository{db: db}
}

// CreateExclusion creates a new WAF rule exclusion for a proxy host
func (r *WAFRepository) CreateExclusion(ctx context.Context, proxyHostID string, req *model.CreateWAFRuleExclusionRequest) (*model.WAFRuleExclusion, error) {
	query := `
		INSERT INTO waf_rule_exclusions (
			proxy_host_id, rule_id, rule_category, rule_description, reason
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, proxy_host_id, rule_id, rule_category, rule_description, reason, disabled_by, created_at
	`

	var exclusion model.WAFRuleExclusion
	var ruleCategory, ruleDescription, reason, disabledBy sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		proxyHostID,
		req.RuleID,
		req.RuleCategory,
		req.RuleDescription,
		req.Reason,
	).Scan(
		&exclusion.ID,
		&exclusion.ProxyHostID,
		&exclusion.RuleID,
		&ruleCategory,
		&ruleDescription,
		&reason,
		&disabledBy,
		&exclusion.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create WAF rule exclusion: %w", err)
	}

	if ruleCategory.Valid {
		exclusion.RuleCategory = ruleCategory.String
	}
	if ruleDescription.Valid {
		exclusion.RuleDescription = ruleDescription.String
	}
	if reason.Valid {
		exclusion.Reason = reason.String
	}
	if disabledBy.Valid {
		exclusion.DisabledBy = disabledBy.String
	}

	return &exclusion, nil
}

// DeleteExclusion removes a WAF rule exclusion (re-enables the rule)
func (r *WAFRepository) DeleteExclusion(ctx context.Context, proxyHostID string, ruleID int) error {
	query := `DELETE FROM waf_rule_exclusions WHERE proxy_host_id = $1 AND rule_id = $2`
	result, err := r.db.ExecContext(ctx, query, proxyHostID, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete WAF rule exclusion: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetExclusionsByProxyHost returns all rule exclusions for a specific proxy host
func (r *WAFRepository) GetExclusionsByProxyHost(ctx context.Context, proxyHostID string) ([]model.WAFRuleExclusion, error) {
	query := `
		SELECT id, proxy_host_id, rule_id, rule_category, rule_description, reason, disabled_by, created_at
		FROM waf_rule_exclusions
		WHERE proxy_host_id = $1
		ORDER BY rule_id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, proxyHostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAF rule exclusions: %w", err)
	}
	defer rows.Close()

	var exclusions []model.WAFRuleExclusion
	for rows.Next() {
		var exclusion model.WAFRuleExclusion
		var ruleCategory, ruleDescription, reason, disabledBy sql.NullString

		err := rows.Scan(
			&exclusion.ID,
			&exclusion.ProxyHostID,
			&exclusion.RuleID,
			&ruleCategory,
			&ruleDescription,
			&reason,
			&disabledBy,
			&exclusion.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WAF rule exclusion: %w", err)
		}

		if ruleCategory.Valid {
			exclusion.RuleCategory = ruleCategory.String
		}
		if ruleDescription.Valid {
			exclusion.RuleDescription = ruleDescription.String
		}
		if reason.Valid {
			exclusion.Reason = reason.String
		}
		if disabledBy.Valid {
			exclusion.DisabledBy = disabledBy.String
		}

		exclusions = append(exclusions, exclusion)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating WAF rule exclusions: %w", err)
	}

	return exclusions, nil
}

// GetExclusionByRuleID checks if a specific rule is excluded for a proxy host
func (r *WAFRepository) GetExclusionByRuleID(ctx context.Context, proxyHostID string, ruleID int) (*model.WAFRuleExclusion, error) {
	query := `
		SELECT id, proxy_host_id, rule_id, rule_category, rule_description, reason, disabled_by, created_at
		FROM waf_rule_exclusions
		WHERE proxy_host_id = $1 AND rule_id = $2
	`

	var exclusion model.WAFRuleExclusion
	var ruleCategory, ruleDescription, reason, disabledBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, proxyHostID, ruleID).Scan(
		&exclusion.ID,
		&exclusion.ProxyHostID,
		&exclusion.RuleID,
		&ruleCategory,
		&ruleDescription,
		&reason,
		&disabledBy,
		&exclusion.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get WAF rule exclusion: %w", err)
	}

	if ruleCategory.Valid {
		exclusion.RuleCategory = ruleCategory.String
	}
	if ruleDescription.Valid {
		exclusion.RuleDescription = ruleDescription.String
	}
	if reason.Valid {
		exclusion.Reason = reason.String
	}
	if disabledBy.Valid {
		exclusion.DisabledBy = disabledBy.String
	}

	return &exclusion, nil
}

// GetAllExclusions returns all rule exclusions grouped by proxy host
func (r *WAFRepository) GetAllExclusions(ctx context.Context) (map[string][]model.WAFRuleExclusion, error) {
	query := `
		SELECT id, proxy_host_id, rule_id, rule_category, rule_description, reason, disabled_by, created_at
		FROM waf_rule_exclusions
		ORDER BY proxy_host_id, rule_id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all WAF rule exclusions: %w", err)
	}
	defer rows.Close()

	exclusionsByHost := make(map[string][]model.WAFRuleExclusion)
	for rows.Next() {
		var exclusion model.WAFRuleExclusion
		var ruleCategory, ruleDescription, reason, disabledBy sql.NullString

		err := rows.Scan(
			&exclusion.ID,
			&exclusion.ProxyHostID,
			&exclusion.RuleID,
			&ruleCategory,
			&ruleDescription,
			&reason,
			&disabledBy,
			&exclusion.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WAF rule exclusion: %w", err)
		}

		if ruleCategory.Valid {
			exclusion.RuleCategory = ruleCategory.String
		}
		if ruleDescription.Valid {
			exclusion.RuleDescription = ruleDescription.String
		}
		if reason.Valid {
			exclusion.Reason = reason.String
		}
		if disabledBy.Valid {
			exclusion.DisabledBy = disabledBy.String
		}

		exclusionsByHost[exclusion.ProxyHostID] = append(exclusionsByHost[exclusion.ProxyHostID], exclusion)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating WAF rule exclusions: %w", err)
	}

	return exclusionsByHost, nil
}

// CountExclusionsByProxyHost returns the count of exclusions for each proxy host
func (r *WAFRepository) CountExclusionsByProxyHost(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT proxy_host_id, COUNT(*) as count
		FROM waf_rule_exclusions
		GROUP BY proxy_host_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count WAF rule exclusions: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var proxyHostID string
		var count int
		if err := rows.Scan(&proxyHostID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan exclusion count: %w", err)
		}
		counts[proxyHostID] = count
	}

	return counts, nil
}

// CreatePolicyHistory creates a new policy change history record
func (r *WAFRepository) CreatePolicyHistory(ctx context.Context, history *model.WAFPolicyHistory) error {
	query := `
		INSERT INTO waf_policy_history (
			proxy_host_id, rule_id, rule_category, rule_description, action, reason, changed_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		history.ProxyHostID,
		history.RuleID,
		history.RuleCategory,
		history.RuleDescription,
		history.Action,
		history.Reason,
		history.ChangedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create WAF policy history: %w", err)
	}

	return nil
}

// GetPolicyHistory returns policy change history for a specific proxy host
func (r *WAFRepository) GetPolicyHistory(ctx context.Context, proxyHostID string, limit int) ([]model.WAFPolicyHistory, error) {
	query := `
		SELECT id, proxy_host_id, rule_id, rule_category, rule_description, action, reason, changed_by, created_at
		FROM waf_policy_history
		WHERE proxy_host_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, proxyHostID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAF policy history: %w", err)
	}
	defer rows.Close()

	var history []model.WAFPolicyHistory
	for rows.Next() {
		var h model.WAFPolicyHistory
		var ruleCategory, ruleDescription, reason, changedBy sql.NullString

		err := rows.Scan(
			&h.ID,
			&h.ProxyHostID,
			&h.RuleID,
			&ruleCategory,
			&ruleDescription,
			&h.Action,
			&reason,
			&changedBy,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WAF policy history: %w", err)
		}

		if ruleCategory.Valid {
			h.RuleCategory = ruleCategory.String
		}
		if ruleDescription.Valid {
			h.RuleDescription = ruleDescription.String
		}
		if reason.Valid {
			h.Reason = reason.String
		}
		if changedBy.Valid {
			h.ChangedBy = changedBy.String
		}

		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating WAF policy history: %w", err)
	}

	return history, nil
}

// CountPolicyHistory returns the total count of policy history for a proxy host
func (r *WAFRepository) CountPolicyHistory(ctx context.Context, proxyHostID string) (int, error) {
	query := `SELECT COUNT(*) FROM waf_policy_history WHERE proxy_host_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, proxyHostID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count WAF policy history: %w", err)
	}

	return count, nil
}

// ============================================================================
// Global WAF Rule Exclusions
// ============================================================================

// CreateGlobalExclusion creates a new global WAF rule exclusion
func (r *WAFRepository) CreateGlobalExclusion(ctx context.Context, req *model.CreateGlobalWAFRuleExclusionRequest, username string) (*model.GlobalWAFRuleExclusion, error) {
	query := `
		INSERT INTO global_waf_rule_exclusions (
			rule_id, rule_category, rule_description, reason, disabled_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, rule_id, rule_category, rule_description, reason, disabled_by, created_at, updated_at
	`

	var exclusion model.GlobalWAFRuleExclusion
	var ruleCategory, ruleDescription, reason, disabledBy sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		req.RuleID,
		req.RuleCategory,
		req.RuleDescription,
		req.Reason,
		username,
	).Scan(
		&exclusion.ID,
		&exclusion.RuleID,
		&ruleCategory,
		&ruleDescription,
		&reason,
		&disabledBy,
		&exclusion.CreatedAt,
		&exclusion.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create global WAF rule exclusion: %w", err)
	}

	if ruleCategory.Valid {
		exclusion.RuleCategory = ruleCategory.String
	}
	if ruleDescription.Valid {
		exclusion.RuleDescription = ruleDescription.String
	}
	if reason.Valid {
		exclusion.Reason = reason.String
	}
	if disabledBy.Valid {
		exclusion.DisabledBy = disabledBy.String
	}

	return &exclusion, nil
}

// DeleteGlobalExclusion removes a global WAF rule exclusion (re-enables the rule globally)
func (r *WAFRepository) DeleteGlobalExclusion(ctx context.Context, ruleID int) error {
	query := `DELETE FROM global_waf_rule_exclusions WHERE rule_id = $1`
	result, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete global WAF rule exclusion: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetGlobalExclusions returns all global rule exclusions
func (r *WAFRepository) GetGlobalExclusions(ctx context.Context) ([]model.GlobalWAFRuleExclusion, error) {
	query := `
		SELECT id, rule_id, rule_category, rule_description, reason, disabled_by, created_at, updated_at
		FROM global_waf_rule_exclusions
		ORDER BY rule_id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get global WAF rule exclusions: %w", err)
	}
	defer rows.Close()

	var exclusions []model.GlobalWAFRuleExclusion
	for rows.Next() {
		var exclusion model.GlobalWAFRuleExclusion
		var ruleCategory, ruleDescription, reason, disabledBy sql.NullString

		err := rows.Scan(
			&exclusion.ID,
			&exclusion.RuleID,
			&ruleCategory,
			&ruleDescription,
			&reason,
			&disabledBy,
			&exclusion.CreatedAt,
			&exclusion.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan global WAF rule exclusion: %w", err)
		}

		if ruleCategory.Valid {
			exclusion.RuleCategory = ruleCategory.String
		}
		if ruleDescription.Valid {
			exclusion.RuleDescription = ruleDescription.String
		}
		if reason.Valid {
			exclusion.Reason = reason.String
		}
		if disabledBy.Valid {
			exclusion.DisabledBy = disabledBy.String
		}

		exclusions = append(exclusions, exclusion)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating global WAF rule exclusions: %w", err)
	}

	return exclusions, nil
}

// GetGlobalExclusionByRuleID checks if a specific rule is globally excluded
func (r *WAFRepository) GetGlobalExclusionByRuleID(ctx context.Context, ruleID int) (*model.GlobalWAFRuleExclusion, error) {
	query := `
		SELECT id, rule_id, rule_category, rule_description, reason, disabled_by, created_at, updated_at
		FROM global_waf_rule_exclusions
		WHERE rule_id = $1
	`

	var exclusion model.GlobalWAFRuleExclusion
	var ruleCategory, ruleDescription, reason, disabledBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, ruleID).Scan(
		&exclusion.ID,
		&exclusion.RuleID,
		&ruleCategory,
		&ruleDescription,
		&reason,
		&disabledBy,
		&exclusion.CreatedAt,
		&exclusion.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get global WAF rule exclusion: %w", err)
	}

	if ruleCategory.Valid {
		exclusion.RuleCategory = ruleCategory.String
	}
	if ruleDescription.Valid {
		exclusion.RuleDescription = ruleDescription.String
	}
	if reason.Valid {
		exclusion.Reason = reason.String
	}
	if disabledBy.Valid {
		exclusion.DisabledBy = disabledBy.String
	}

	return &exclusion, nil
}

// GetGlobalExclusionMap returns a map of rule_id -> exclusion for quick lookup
func (r *WAFRepository) GetGlobalExclusionMap(ctx context.Context) (map[int]*model.GlobalWAFRuleExclusion, error) {
	exclusions, err := r.GetGlobalExclusions(ctx)
	if err != nil {
		return nil, err
	}

	exclusionMap := make(map[int]*model.GlobalWAFRuleExclusion)
	for i := range exclusions {
		exclusionMap[exclusions[i].RuleID] = &exclusions[i]
	}

	return exclusionMap, nil
}

// CountGlobalExclusions returns the count of global exclusions
func (r *WAFRepository) CountGlobalExclusions(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM global_waf_rule_exclusions`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count global WAF rule exclusions: %w", err)
	}

	return count, nil
}

// CreateGlobalPolicyHistory creates a new global policy change history record
func (r *WAFRepository) CreateGlobalPolicyHistory(ctx context.Context, history *model.GlobalWAFPolicyHistory) error {
	query := `
		INSERT INTO global_waf_policy_history (
			rule_id, rule_category, rule_description, action, reason, changed_by
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		history.RuleID,
		history.RuleCategory,
		history.RuleDescription,
		history.Action,
		history.Reason,
		history.ChangedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create global WAF policy history: %w", err)
	}

	return nil
}

// GetGlobalPolicyHistory returns global policy change history
func (r *WAFRepository) GetGlobalPolicyHistory(ctx context.Context, limit int) ([]model.GlobalWAFPolicyHistory, error) {
	query := `
		SELECT id, rule_id, rule_category, rule_description, action, reason, changed_by, created_at
		FROM global_waf_policy_history
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get global WAF policy history: %w", err)
	}
	defer rows.Close()

	var history []model.GlobalWAFPolicyHistory
	for rows.Next() {
		var h model.GlobalWAFPolicyHistory
		var ruleCategory, ruleDescription, reason, changedBy sql.NullString

		err := rows.Scan(
			&h.ID,
			&h.RuleID,
			&ruleCategory,
			&ruleDescription,
			&h.Action,
			&reason,
			&changedBy,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan global WAF policy history: %w", err)
		}

		if ruleCategory.Valid {
			h.RuleCategory = ruleCategory.String
		}
		if ruleDescription.Valid {
			h.RuleDescription = ruleDescription.String
		}
		if reason.Valid {
			h.Reason = reason.String
		}
		if changedBy.Valid {
			h.ChangedBy = changedBy.String
		}

		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating global WAF policy history: %w", err)
	}

	return history, nil
}

// CountGlobalPolicyHistory returns the total count of global policy history
func (r *WAFRepository) CountGlobalPolicyHistory(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM global_waf_policy_history`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count global WAF policy history: %w", err)
	}

	return count, nil
}

// ============================================================================
// WAF Rule Snapshots (Versioning)
// ============================================================================

// CreateSnapshot creates a new WAF configuration snapshot
func (r *WAFRepository) CreateSnapshot(ctx context.Context, snapshot *model.WAFRuleSnapshot) (*model.WAFRuleSnapshot, error) {
	// Get next version number
	var nextVersion int
	versionQuery := `
		SELECT COALESCE(MAX(version_number), 0) + 1
		FROM waf_rule_snapshots
		WHERE ($1::uuid IS NULL AND proxy_host_id IS NULL) OR proxy_host_id = $1
	`
	err := r.db.QueryRowContext(ctx, versionQuery, snapshot.ProxyHostID).Scan(&nextVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get next version number: %w", err)
	}

	query := `
		INSERT INTO waf_rule_snapshots (
			proxy_host_id, version_number, snapshot_name, rule_engine, paranoia_level,
			anomaly_threshold, total_rules, disabled_rules, change_description, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err = r.db.QueryRowContext(ctx, query,
		snapshot.ProxyHostID,
		nextVersion,
		snapshot.SnapshotName,
		snapshot.RuleEngine,
		snapshot.ParanoiaLevel,
		snapshot.AnomalyThreshold,
		snapshot.TotalRules,
		snapshot.DisabledRules,
		snapshot.ChangeDescription,
		snapshot.CreatedBy,
	).Scan(&snapshot.ID, &snapshot.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create WAF snapshot: %w", err)
	}

	snapshot.VersionNumber = nextVersion
	return snapshot, nil
}

// CreateSnapshotDetail creates a snapshot detail record for a rule
func (r *WAFRepository) CreateSnapshotDetail(ctx context.Context, detail *model.WAFRuleSnapshotDetail) error {
	query := `
		INSERT INTO waf_rule_snapshot_details (
			snapshot_id, rule_id, rule_category, rule_description, is_disabled, reason
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	return r.db.QueryRowContext(ctx, query,
		detail.SnapshotID,
		detail.RuleID,
		detail.RuleCategory,
		detail.RuleDescription,
		detail.IsDisabled,
		detail.Reason,
	).Scan(&detail.ID)
}

// GetSnapshotByID retrieves a snapshot by ID with its details
func (r *WAFRepository) GetSnapshotByID(ctx context.Context, id string) (*model.WAFRuleSnapshot, error) {
	query := `
		SELECT id, proxy_host_id, version_number, snapshot_name, rule_engine, paranoia_level,
			anomaly_threshold, total_rules, disabled_rules, change_description, created_by, created_at
		FROM waf_rule_snapshots
		WHERE id = $1
	`

	var snapshot model.WAFRuleSnapshot
	var proxyHostID sql.NullString
	var snapshotName, ruleEngine, changeDesc, createdBy sql.NullString
	var paranoiaLevel, anomalyThreshold sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&snapshot.ID,
		&proxyHostID,
		&snapshot.VersionNumber,
		&snapshotName,
		&ruleEngine,
		&paranoiaLevel,
		&anomalyThreshold,
		&snapshot.TotalRules,
		&snapshot.DisabledRules,
		&changeDesc,
		&createdBy,
		&snapshot.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get WAF snapshot: %w", err)
	}

	if proxyHostID.Valid {
		snapshot.ProxyHostID = &proxyHostID.String
	}
	if snapshotName.Valid {
		snapshot.SnapshotName = snapshotName.String
	}
	if ruleEngine.Valid {
		snapshot.RuleEngine = ruleEngine.String
	}
	if paranoiaLevel.Valid {
		snapshot.ParanoiaLevel = int(paranoiaLevel.Int32)
	}
	if anomalyThreshold.Valid {
		snapshot.AnomalyThreshold = int(anomalyThreshold.Int32)
	}
	if changeDesc.Valid {
		snapshot.ChangeDescription = changeDesc.String
	}
	if createdBy.Valid {
		snapshot.CreatedBy = createdBy.String
	}

	return &snapshot, nil
}

// ListSnapshots returns snapshots for a proxy host (or global if proxyHostID is nil)
func (r *WAFRepository) ListSnapshots(ctx context.Context, proxyHostID *string, limit int) ([]model.WAFRuleSnapshot, error) {
	query := `
		SELECT id, proxy_host_id, version_number, snapshot_name, rule_engine, paranoia_level,
			anomaly_threshold, total_rules, disabled_rules, change_description, created_by, created_at
		FROM waf_rule_snapshots
		WHERE ($1::uuid IS NULL AND proxy_host_id IS NULL) OR proxy_host_id = $1
		ORDER BY version_number DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, proxyHostID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list WAF snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []model.WAFRuleSnapshot
	for rows.Next() {
		var snapshot model.WAFRuleSnapshot
		var hostID sql.NullString
		var snapshotName, ruleEngine, changeDesc, createdBy sql.NullString
		var paranoiaLevel, anomalyThreshold sql.NullInt32

		err := rows.Scan(
			&snapshot.ID,
			&hostID,
			&snapshot.VersionNumber,
			&snapshotName,
			&ruleEngine,
			&paranoiaLevel,
			&anomalyThreshold,
			&snapshot.TotalRules,
			&snapshot.DisabledRules,
			&changeDesc,
			&createdBy,
			&snapshot.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WAF snapshot: %w", err)
		}

		if hostID.Valid {
			snapshot.ProxyHostID = &hostID.String
		}
		if snapshotName.Valid {
			snapshot.SnapshotName = snapshotName.String
		}
		if ruleEngine.Valid {
			snapshot.RuleEngine = ruleEngine.String
		}
		if paranoiaLevel.Valid {
			snapshot.ParanoiaLevel = int(paranoiaLevel.Int32)
		}
		if anomalyThreshold.Valid {
			snapshot.AnomalyThreshold = int(anomalyThreshold.Int32)
		}
		if changeDesc.Valid {
			snapshot.ChangeDescription = changeDesc.String
		}
		if createdBy.Valid {
			snapshot.CreatedBy = createdBy.String
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// ============================================================================
// WAF Rule Change Events
// ============================================================================

// RecordRuleChangeEvent records a rule enable/disable event
func (r *WAFRepository) RecordRuleChangeEvent(ctx context.Context, event *model.WAFRuleChangeEvent) error {
	query := `
		INSERT INTO waf_rule_change_events (
			proxy_host_id, rule_id, action, rule_category, rule_description, reason, changed_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	return r.db.QueryRowContext(ctx, query,
		event.ProxyHostID,
		event.RuleID,
		event.Action,
		event.RuleCategory,
		event.RuleDescription,
		event.Reason,
		event.ChangedBy,
	).Scan(&event.ID, &event.CreatedAt)
}

// GetRuleChangeEvents returns change events for a proxy host
func (r *WAFRepository) GetRuleChangeEvents(ctx context.Context, proxyHostID *string, limit int) ([]model.WAFRuleChangeEvent, error) {
	query := `
		SELECT id, proxy_host_id, rule_id, action, rule_category, rule_description, reason, changed_by, created_at
		FROM waf_rule_change_events
		WHERE ($1::uuid IS NULL AND proxy_host_id IS NULL) OR proxy_host_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, proxyHostID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAF rule change events: %w", err)
	}
	defer rows.Close()

	var events []model.WAFRuleChangeEvent
	for rows.Next() {
		var event model.WAFRuleChangeEvent
		var hostID, ruleCategory, ruleDescription, reason, changedBy sql.NullString

		err := rows.Scan(
			&event.ID,
			&hostID,
			&event.RuleID,
			&event.Action,
			&ruleCategory,
			&ruleDescription,
			&reason,
			&changedBy,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WAF rule change event: %w", err)
		}

		if hostID.Valid {
			event.ProxyHostID = &hostID.String
		}
		if ruleCategory.Valid {
			event.RuleCategory = ruleCategory.String
		}
		if ruleDescription.Valid {
			event.RuleDescription = ruleDescription.String
		}
		if reason.Valid {
			event.Reason = reason.String
		}
		if changedBy.Valid {
			event.ChangedBy = changedBy.String
		}

		events = append(events, event)
	}

	return events, nil
}
