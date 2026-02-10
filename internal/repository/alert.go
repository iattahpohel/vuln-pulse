package repository

import (
	"context"
	"fmt"

	"github.com/atta/vulnpulse/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AlertRepository handles alert data operations
type AlertRepository struct {
	db *pgxpool.Pool
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *pgxpool.Pool) *AlertRepository {
	return &AlertRepository{db: db}
}

// Upsert creates or updates an alert (idempotent)
func (r *AlertRepository) Upsert(ctx context.Context, alert *domain.Alert) error {
	query := `
		INSERT INTO alerts (id, tenant_id, asset_id, vulnerability_id, severity, status, matched_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (tenant_id, asset_id, vulnerability_id) DO UPDATE
		SET severity = EXCLUDED.severity, matched_at = EXCLUDED.matched_at
		WHERE alerts.status = 'resolved' OR alerts.status = 'false_positive'
		RETURNING matched_at
	`
	err := r.db.QueryRow(ctx, query, alert.ID, alert.TenantID, alert.AssetID,
		alert.VulnerabilityID, alert.Severity, domain.AlertStatusOpen).
		Scan(&alert.MatchedAt)

	// If no rows affected (conflict but didn't update), still return success
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("failed to upsert alert: %w", err)
	}

	return nil
}

// GetByID retrieves an alert by ID
func (r *AlertRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, tenant_id, asset_id, vulnerability_id, severity, status, matched_at, resolved_at
		FROM alerts
		WHERE id = $1 AND tenant_id = $2
	`
	var alert domain.Alert
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&alert.ID, &alert.TenantID, &alert.AssetID, &alert.VulnerabilityID,
		&alert.Severity, &alert.Status, &alert.MatchedAt, &alert.ResolvedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	return &alert, nil
}

// ListByTenant lists all alerts for a tenant with optional filtering
func (r *AlertRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*domain.Alert, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, tenant_id, asset_id, vulnerability_id, severity, status, matched_at, resolved_at
			FROM alerts
			WHERE tenant_id = $1 AND status = $2
			ORDER BY matched_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{tenantID, status, limit, offset}
	} else {
		query = `
			SELECT id, tenant_id, asset_id, vulnerability_id, severity, status, matched_at, resolved_at
			FROM alerts
			WHERE tenant_id = $1
			ORDER BY matched_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{tenantID, limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		var alert domain.Alert
		err := rows.Scan(
			&alert.ID, &alert.TenantID, &alert.AssetID, &alert.VulnerabilityID,
			&alert.Severity, &alert.Status, &alert.MatchedAt, &alert.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

// UpdateStatus updates the status of an alert
func (r *AlertRepository) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status string) error {
	query := `
		UPDATE alerts
		SET status = $1, resolved_at = 
			CASE WHEN $1 IN ('resolved', 'false_positive') THEN NOW() ELSE NULL END,
			matched_at = CASE WHEN $1 = 'open' THEN NOW() ELSE matched_at END
		WHERE id = $2 AND tenant_id = $3
	`
	result, err := r.db.Exec(ctx, query, status, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update alert status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}
