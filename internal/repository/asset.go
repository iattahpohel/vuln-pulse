package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atta/vulnpulse/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AssetRepository handles asset data operations
type AssetRepository struct {
	db *pgxpool.Pool
}

// NewAssetRepository creates a new asset repository
func NewAssetRepository(db *pgxpool.Pool) *AssetRepository {
	return &AssetRepository{db: db}
}

// Create creates a new asset
func (r *AssetRepository) Create(ctx context.Context, asset *domain.Asset) error {
	metadata, err := json.Marshal(asset.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO assets (id, tenant_id, name, type, version, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (tenant_id, name, version) DO UPDATE
		SET type = EXCLUDED.type, metadata = EXCLUDED.metadata, updated_at = NOW()
		RETURNING created_at, updated_at
	`
	err = r.db.QueryRow(ctx, query, asset.ID, asset.TenantID, asset.Name, asset.Type, asset.Version, metadata).
		Scan(&asset.CreatedAt, &asset.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}
	return nil
}

// GetByID retrieves an asset by ID
func (r *AssetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Asset, error) {
	query := `
		SELECT id, tenant_id, name, type, version, metadata, created_at, updated_at
		FROM assets
		WHERE id = $1
	`
	var asset domain.Asset
	var metadata []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&asset.ID, &asset.TenantID, &asset.Name, &asset.Type,
		&asset.Version, &metadata, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	if err := json.Unmarshal(metadata, &asset.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &asset, nil
}

// ListByTenant lists all assets for a tenant
func (r *AssetRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Asset, error) {
	query := `
		SELECT id, tenant_id, name, type, version, metadata, created_at, updated_at
		FROM assets
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}
	defer rows.Close()

	var assets []*domain.Asset
	for rows.Next() {
		var asset domain.Asset
		var metadata []byte
		err := rows.Scan(
			&asset.ID, &asset.TenantID, &asset.Name, &asset.Type,
			&asset.Version, &metadata, &asset.CreatedAt, &asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &asset.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		assets = append(assets, &asset)
	}

	return assets, nil
}

// Update updates an asset
func (r *AssetRepository) Update(ctx context.Context, asset *domain.Asset) error {
	metadata, err := json.Marshal(asset.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE assets
		SET name = $1, type = $2, version = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5 AND tenant_id = $6
	`
	result, err := r.db.Exec(ctx, query, asset.Name, asset.Type, asset.Version, metadata, asset.ID, asset.TenantID)
	if err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("asset not found")
	}

	return nil
}

// Delete deletes an asset
func (r *AssetRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	query := `DELETE FROM assets WHERE id = $1 AND tenant_id = $2`
	result, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("asset not found")
	}

	return nil
}
