package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "octa-byte-ai/internal/config"
    "octa-byte-ai/internal/database"
    "octa-byte-ai/internal/handlers"
    "octa-byte-ai/pkg/logger"

    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Initialize logger
    log := logger.New()
    defer log.Sync()

    // Load configuration
    cfg := config.Load()
    log.Info("Configuration loaded", "port", cfg.Port, "db_host", cfg.Database.Host)

    // Initialize database
    db, err := database.NewPostgresDB(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database", "error", err)
    }
    defer db.Close()

    log.Info("Database connection established")

    // Skip migrations for local testing
    log.Info("Skipping automatic migrations - table should be created manually")

    // Initialize handlers
    h := handlers.New(db, log)

    // Setup routes
    router := mux.NewRouter()
    
    // Health endpoints
    router.HandleFunc("/health", h.Health).Methods("GET")
    router.HandleFunc("/ready", h.Ready).Methods("GET")
    router.HandleFunc("/metrics", promhttp.Handler().ServeHTTP).Methods("GET")

    // API routes
    api := router.PathPrefix("/api").Subrouter()
    api.Use(h.LoggingMiddleware)
    api.Use(h.MetricsMiddleware)
    api.Use(h.RateLimitMiddleware)

    // User endpoints
    api.HandleFunc("/users", h.GetUsers).Methods("GET")
    api.HandleFunc("/users", h.CreateUser).Methods("POST")
    api.HandleFunc("/users/{id}", h.GetUser).Methods("GET")

    // Create server
    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server in goroutine
    go func() {
        log.Info("Server starting", "port", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server failed to start", "error", err)
        }
    }()

    // Wait for interrupt signal to gracefully shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("Server shutting down...")

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown", "error", err)
    }

    log.Info("Server exited")
}
