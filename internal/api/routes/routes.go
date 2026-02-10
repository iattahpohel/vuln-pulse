package routes

import (
	"net/http"

	"github.com/atta/vulnpulse/internal/api/handlers"
	"github.com/atta/vulnpulse/internal/api/middleware"
	"github.com/atta/vulnpulse/pkg/auth"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	authHandler *handlers.AuthHandler,
	assetHandler *handlers.AssetHandler,
	vulnHandler *handlers.VulnerabilityHandler,
	alertHandler *handlers.AlertHandler,
	authService *auth.Service,
	log *logger.Logger,
) *mux.Router {
	router := mux.NewRouter()

	// Middleware chain
	router.Use(middleware.LoggingMiddleware(log))
	router.Use(middleware.CORSMiddleware())

	// Public routes
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods(http.MethodPost, http.MethodOptions)

	// Protected routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.AuthMiddleware(authService, log))

	// Auth routes
	api.Handle("/auth/register",
		middleware.RBACMiddleware(auth.PermissionCreateTenant, log)(http.HandlerFunc(authHandler.Register)),
	).Methods(http.MethodPost, http.MethodOptions)

	// Asset routes
	assetRoutes := api.PathPrefix("/assets").Subrouter()
	assetRoutes.Handle("",
		middleware.RBACMiddleware(auth.PermissionReadAsset, log)(http.HandlerFunc(assetHandler.List)),
	).Methods(http.MethodGet)
	assetRoutes.Handle("",
		middleware.RBACMiddleware(auth.PermissionCreateAsset, log)(http.HandlerFunc(assetHandler.Create)),
	).Methods(http.MethodPost)
	assetRoutes.Handle("/{id}",
		middleware.RBACMiddleware(auth.PermissionReadAsset, log)(http.HandlerFunc(assetHandler.Get)),
	).Methods(http.MethodGet)
	assetRoutes.Handle("/{id}",
		middleware.RBACMiddleware(auth.PermissionUpdateAsset, log)(http.HandlerFunc(assetHandler.Update)),
	).Methods(http.MethodPut)
	assetRoutes.Handle("/{id}",
		middleware.RBACMiddleware(auth.PermissionDeleteAsset, log)(http.HandlerFunc(assetHandler.Delete)),
	).Methods(http.MethodDelete)

	// Vulnerability routes
	vulnRoutes := api.PathPrefix("/vulnerabilities").Subrouter()
	vulnRoutes.Handle("",
		middleware.RBACMiddleware(auth.PermissionReadVuln, log)(http.HandlerFunc(vulnHandler.List)),
	).Methods(http.MethodGet)
	vulnRoutes.Handle("",
		middleware.RBACMiddleware(auth.PermissionIngestVuln, log)(http.HandlerFunc(vulnHandler.Ingest)),
	).Methods(http.MethodPost)
	vulnRoutes.Handle("/{id}",
		middleware.RBACMiddleware(auth.PermissionReadVuln, log)(http.HandlerFunc(vulnHandler.Get)),
	).Methods(http.MethodGet)

	// Alert routes
	alertRoutes := api.PathPrefix("/alerts").Subrouter()
	alertRoutes.Handle("",
		middleware.RBACMiddleware(auth.PermissionReadAlert, log)(http.HandlerFunc(alertHandler.List)),
	).Methods(http.MethodGet)
	alertRoutes.Handle("/{id}",
		middleware.RBACMiddleware(auth.PermissionReadAlert, log)(http.HandlerFunc(alertHandler.Get)),
	).Methods(http.MethodGet)
	alertRoutes.Handle("/{id}",
		middleware.RBACMiddleware(auth.PermissionUpdateAlert, log)(http.HandlerFunc(alertHandler.UpdateStatus)),
	).Methods(http.MethodPatch)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods(http.MethodGet)

	return router
}
