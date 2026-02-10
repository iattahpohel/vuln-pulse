package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/atta/vulnpulse/internal/domain"
	"github.com/atta/vulnpulse/internal/repository"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/atta/vulnpulse/pkg/queue"
	"github.com/google/uuid"
)

// MatchService handles vulnerability-to-asset matching
type MatchService struct {
	vulnRepo  *repository.VulnerabilityRepository
	assetRepo *repository.AssetRepository
	alertRepo *repository.AlertRepository
	queue     *queue.Client
	log       *logger.Logger
}

// NewMatchService creates a new match service
func NewMatchService(
	vulnRepo *repository.VulnerabilityRepository,
	assetRepo *repository.AssetRepository,
	alertRepo *repository.AlertRepository,
	queueClient *queue.Client,
	log *logger.Logger,
) *MatchService {
	return &MatchService{
		vulnRepo:  vulnRepo,
		assetRepo: assetRepo,
		alertRepo: alertRepo,
		queue:     queueClient,
		log:       log,
	}
}

// MatchVulnerability matches a vulnerability against all assets across all tenants
func (s *MatchService) MatchVulnerability(ctx context.Context, vulnID uuid.UUID) error {
	vuln, err := s.vulnRepo.GetByID(ctx, vulnID)
	if err != nil {
		return fmt.Errorf("failed to get vulnerability: %w", err)
	}

	s.log.Info("matching vulnerability", "cve_id", vuln.CveID, "severity", vuln.Severity)

	// For now, we'll match against all assets (in production, you'd query tenants first)
	// This is simplified - ideally you'd iterate over tenants and their assets
	tenants := s.getTenantIDs(ctx)

	matchCount := 0
	for _, tenantID := range tenants {
		assets, err := s.assetRepo.ListByTenant(ctx, tenantID)
		if err != nil {
			s.log.Error("failed to list assets for tenant", "tenant_id", tenantID, "error", err)
			continue
		}

		for _, asset := range assets {
			if s.isAssetAffected(asset, vuln) {
				alert := &domain.Alert{
					ID:              uuid.New(),
					TenantID:        asset.TenantID,
					AssetID:         asset.ID,
					VulnerabilityID: vuln.ID,
					Severity:        vuln.Severity,
					Status:          domain.AlertStatusOpen,
				}

				if err := s.alertRepo.Upsert(ctx, alert); err != nil {
					s.log.Error("failed to create alert", "error", err)
					continue
				}

				matchCount++

				// Publish webhook dispatch event
				s.publishWebhookEvent(ctx, tenantID, domain.EventAlertCreated, map[string]interface{}{
					"alert_id":         alert.ID.String(),
					"asset_id":         asset.ID.String(),
					"vulnerability_id": vuln.ID.String(),
					"cve_id":           vuln.CveID,
					"severity":         vuln.Severity,
				})
			}
		}
	}

	s.log.Info("vulnerability matching completed", "cve_id", vuln.CveID, "matches", matchCount)
	return nil
}

// MatchAsset re-evaluates all vulnerabilities for a specific asset
func (s *MatchService) MatchAsset(ctx context.Context, assetID uuid.UUID) error {
	asset, err := s.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %w", err)
	}

	s.log.Info("matching asset", "asset_id", asset.ID, "name", asset.Name)

	// Get all vulnerabilities (in production, paginate this)
	vulns, err := s.vulnRepo.List(ctx, "", 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to list vulnerabilities: %w", err)
	}

	matchCount := 0
	for _, vuln := range vulns {
		if s.isAssetAffected(asset, vuln) {
			alert := &domain.Alert{
				ID:              uuid.New(),
				TenantID:        asset.TenantID,
				AssetID:         asset.ID,
				VulnerabilityID: vuln.ID,
				Severity:        vuln.Severity,
				Status:          domain.AlertStatusOpen,
			}

			if err := s.alertRepo.Upsert(ctx, alert); err != nil {
				s.log.Error("failed to create alert", "error", err)
				continue
			}

			matchCount++

			// Publish webhook dispatch event
			s.publishWebhookEvent(ctx, asset.TenantID, domain.EventAlertCreated, map[string]interface{}{
				"alert_id":         alert.ID.String(),
				"asset_id":         asset.ID.String(),
				"vulnerability_id": vuln.ID.String(),
				"cve_id":           vuln.CveID,
				"severity":         vuln.Severity,
			})
		}
	}

	s.log.Info("asset matching completed", "asset_id", asset.ID, "matches", matchCount)
	return nil
}

// isAssetAffected checks if an asset is affected by a vulnerability
func (s *MatchService) isAssetAffected(asset *domain.Asset, vuln *domain.Vulnerability) bool {
	// Simple string matching logic
	// In production, use CPE matching or more sophisticated logic
	for _, affectedProduct := range vuln.AffectedProducts {
		// Case-insensitive name match
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(affectedProduct.Name)) {
			// Check version if specified
			if asset.Version != nil && len(affectedProduct.Versions) > 0 {
				for _, version := range affectedProduct.Versions {
					if *asset.Version == version || version == "*" {
						return true
					}
				}
			} else if asset.Version == nil && len(affectedProduct.Versions) == 0 {
				// Match if neither has version
				return true
			} else if len(affectedProduct.Versions) > 0 {
				// Check for wildcard
				for _, version := range affectedProduct.Versions {
					if version == "*" {
						return true
					}
				}
			}
		}
	}

	return false
}

// getTenantIDs is a placeholder - in production, query the tenants table
func (s *MatchService) getTenantIDs(ctx context.Context) []uuid.UUID {
	// For demo purposes, returning empty slice
	// In real implementation: SELECT DISTINCT tenant_id FROM assets
	return []uuid.UUID{}
}

func (s *MatchService) publishWebhookEvent(ctx context.Context, tenantID uuid.UUID, eventType string, payload map[string]interface{}) {
	msg := queue.Message{
		Type:     "webhook.dispatch",
		TenantID: tenantID.String(),
		Payload: map[string]interface{}{
			"event_type": eventType,
			"event_data": payload,
		},
	}
	if err := s.queue.Publish(ctx, "vulnpulse-jobs", msg); err != nil {
		s.log.Error("failed to publish webhook event", "error", err)
	}
}
