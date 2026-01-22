// backend/cmd/api/main.go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djryanj/k8s-rbactory-backend/internal/handlers"
	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/middleware"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		slog.Error("Failed to create Kubernetes client", "error", err)
		os.Exit(1)
	}
	slog.Info("Kubernetes client initialized successfully")

	// Initialize handlers
	rbacHandler := handlers.NewRBACHandler(k8sClient, logger)
	clusterHandler := handlers.NewClusterHandler(k8sClient, logger)
	resourceHandler := handlers.NewResourceHandler(k8sClient, logger) //
	// Setup router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Swagger documentation
	r.HandleFunc("/swagger", handlers.SwaggerUIHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/swagger.json", handlers.SwaggerJSONHandler).Methods(http.MethodGet)

	// Redirect root to swagger UI
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusMovedPermanently)
	}).Methods(http.MethodGet)

	// Health check
	api.HandleFunc("/healthz", handlers.SimpleHealthHandler).Methods(http.MethodGet)

	// Cluster info
	api.HandleFunc("/cluster/info", clusterHandler.GetClusterInfo).Methods(http.MethodGet)
	api.HandleFunc("/namespaces", clusterHandler.ListNamespaces).Methods(http.MethodGet)

	// Resource counts endpoint
	api.HandleFunc("/counts", rbacHandler.GetCounts).Methods(http.MethodGet)

	// RBAC routes
	api.HandleFunc("/roles", rbacHandler.ListRoles).Methods(http.MethodGet)
	api.HandleFunc("/roles/{namespace}/{name}", rbacHandler.GetRole).Methods(http.MethodGet)
	api.HandleFunc("/clusterroles", rbacHandler.ListClusterRoles).Methods(http.MethodGet)
	api.HandleFunc("/clusterroles/{name}", rbacHandler.GetClusterRole).Methods(http.MethodGet)
	api.HandleFunc("/rolebindings", rbacHandler.ListRoleBindings).Methods(http.MethodGet)
	api.HandleFunc("/clusterrolebindings", rbacHandler.ListClusterRoleBindings).Methods(http.MethodGet)
	api.HandleFunc("/relationships/{kind}/{namespace}/{name}", rbacHandler.GetRelationships).Methods(http.MethodGet)
	api.HandleFunc("/relationships/{kind}/{name}", rbacHandler.GetRelationships).Methods(http.MethodGet)
	api.HandleFunc("/principals", rbacHandler.ListPrincipals).Methods(http.MethodGet)

	// Kubernetes Resource routes
	api.HandleFunc("/resources", resourceHandler.ListKubernetesResources).Methods(http.MethodGet)
	api.HandleFunc("/resources/types", resourceHandler.ListResourceTypes).Methods(http.MethodGet)
	api.HandleFunc("/resources/{type}/{namespace}/{name}/access", resourceHandler.GetResourceAccess).Methods(http.MethodGet)
	api.HandleFunc("/resources/{type}/{name}/access", resourceHandler.GetResourceAccess).Methods(http.MethodGet)

	// Apply middleware in order
	handler := middleware.Recovery(logger)(r)                     // 1. Catch panics
	handler = middleware.RequestID(handler)                       // 2. Add request ID
	handler = middleware.Logging(logger)(handler)                 // 3. Log requests
	handler = middleware.RateLimit(logger, 100, 200)(handler)     // 4. Rate limiting (100 req/sec, burst 200)
	handler = middleware.CORS().Handler(handler)                  // 5. Handle CORS
	handler = middleware.Compression(logger)(handler)             // 6. Compress responses
	handler = middleware.Timeout(logger, 60*time.Second)(handler) // 7. Enforce timeout

	// Configure server with security settings
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   120 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
		// Add read header timeout to prevent slowloris attacks
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server",
			"port", port,
			"read_timeout", srv.ReadTimeout,
			"write_timeout", srv.WriteTimeout,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}
