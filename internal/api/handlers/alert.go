package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/atta/vulnpulse/internal/api/middleware"
	"github.com/atta/vulnpulse/internal/domain"
	"github.com/atta/vulnpulse/internal/repository"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AlertHandler handles alert endpoints
type AlertHandler struct {
	repo *repository.AlertRepository
	log  *logger.Logger
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(repo *repository.AlertRepository, log *logger.Logger) *AlertHandler {
	return &AlertHandler{
		repo: repo,
		log:  log,
	}
}

type UpdateAlertRequest struct {
	Status string `json:"status"`
}

// List lists all alerts for the tenant
func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}

	alerts, err := h.repo.ListByTenant(r.Context(), claims.TenantID, status, limit, offset)
	if err != nil {
		h.log.Error("failed to list alerts", "error", err)
		http.Error(w, "failed to list alerts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
		"limit":  limit,
		"offset": offset,
	})
}

// Get retrieves an alert by ID
func (h *AlertHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid alert ID", http.StatusBadRequest)
		return
	}

	alert, err := h.repo.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		http.Error(w, "alert not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

// UpdateStatus updates the status of an alert
func (h *AlertHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid alert ID", http.StatusBadRequest)
		return
	}

	var req UpdateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := []string{
		domain.AlertStatusOpen,
		domain.AlertStatusAcknowledged,
		domain.AlertStatusResolved,
		domain.AlertStatusFalsePositive,
	}
	valid := false
	for _, s := range validStatuses {
		if req.Status == s {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), id, claims.TenantID, req.Status); err != nil {
		h.log.Error("failed to update alert status", "error", err)
		http.Error(w, "failed to update alert", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     id,
		"status": req.Status,
	})
}
