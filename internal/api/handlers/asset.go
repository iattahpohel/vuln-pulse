package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/atta/vulnpulse/internal/api/middleware"
	"github.com/atta/vulnpulse/internal/domain"
	"github.com/atta/vulnpulse/internal/repository"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/atta/vulnpulse/pkg/queue"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AssetHandler handles asset endpoints
type AssetHandler struct {
	repo  *repository.AssetRepository
	queue *queue.Client
	log   *logger.Logger
}

// NewAssetHandler creates a new asset handler
func NewAssetHandler(repo *repository.AssetRepository, queueClient *queue.Client, log *logger.Logger) *AssetHandler {
	return &AssetHandler{
		repo:  repo,
		queue: queueClient,
		log:   log,
	}
}

type CreateAssetRequest struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Version  *string                `json:"version,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Create creates a new asset
func (h *AssetHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)

	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	asset := &domain.Asset{
		ID:       uuid.New(),
		TenantID: claims.TenantID,
		Name:     req.Name,
		Type:     req.Type,
		Version:  req.Version,
		Metadata: req.Metadata,
	}

	if err := h.repo.Create(r.Context(), asset); err != nil {
		h.log.Error("failed to create asset", "error", err)
		http.Error(w, "failed to create asset", http.StatusInternalServerError)
		return
	}

	// Publish asset.changed job
	h.publishAssetChanged(r.Context(), claims.TenantID, asset.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(asset)
}

// List lists all assets for the tenant
func (h *AssetHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)

	assets, err := h.repo.ListByTenant(r.Context(), claims.TenantID)
	if err != nil {
		h.log.Error("failed to list assets", "error", err)
		http.Error(w, "failed to list assets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assets": assets,
		"count":  len(assets),
	})
}

// Get retrieves an asset by ID
func (h *AssetHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid asset ID", http.StatusBadRequest)
		return
	}

	asset, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "asset not found", http.StatusNotFound)
		return
	}

	// Verify tenant ownership
	if asset.TenantID != claims.TenantID && claims.Role != domain.RoleAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(asset)
}

// Update updates an asset
func (h *AssetHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid asset ID", http.StatusBadRequest)
		return
	}

	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	asset := &domain.Asset{
		ID:       id,
		TenantID: claims.TenantID,
		Name:     req.Name,
		Type:     req.Type,
		Version:  req.Version,
		Metadata: req.Metadata,
	}

	if err := h.repo.Update(r.Context(), asset); err != nil {
		h.log.Error("failed to update asset", "error", err)
		http.Error(w, "failed to update asset", http.StatusInternalServerError)
		return
	}

	// Publish asset.changed job
	h.publishAssetChanged(r.Context(), claims.TenantID, asset.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(asset)
}

// Delete deletes an asset
func (h *AssetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid asset ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id, claims.TenantID); err != nil {
		h.log.Error("failed to delete asset", "error", err)
		http.Error(w, "failed to delete asset", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AssetHandler) publishAssetChanged(ctx context.Context, tenantID, assetID uuid.UUID) {
	msg := queue.Message{
		Type:     "asset.changed",
		TenantID: tenantID.String(),
		Payload: map[string]interface{}{
			"asset_id": assetID.String(),
		},
	}
	if err := h.queue.Publish(ctx, "vulnpulse-jobs", msg); err != nil {
		h.log.Error("failed to publish asset.changed event", "error", err)
	}
}
