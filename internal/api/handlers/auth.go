package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/atta/vulnpulse/internal/domain"
	"github.com/atta/vulnpulse/internal/repository"
	"github.com/atta/vulnpulse/pkg/auth"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/google/uuid"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo    *repository.UserRepository
	authService *auth.Service
	log         *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository, authService *auth.Service, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		authService: authService,
		log:         log,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	TenantID uuid.UUID `json:"tenant_id"`
}

type RegisterRequest struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     string    `json:"role"`
	TenantID uuid.UUID `json:"tenant_id"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("failed to decode login request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		h.log.Warn("user not found", "email", req.Email)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		h.log.Warn("invalid password", "user_id", user.ID)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		h.log.Error("failed to generate token", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			Role:     user.Role,
			TenantID: user.TenantID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Register handles user registration (admin only)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("failed to decode register request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != domain.RoleAdmin && req.Role != domain.RoleAnalyst && req.Role != domain.RoleViewer {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.log.Error("failed to hash password", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     req.TenantID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		h.log.Error("failed to create user", "error", err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"tenant_id":  user.TenantID,
		"created_at": user.CreatedAt,
	})
}
