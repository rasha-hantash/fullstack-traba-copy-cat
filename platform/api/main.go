package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log/slog"
	"time"
	"context"
	"fmt"
	"net/http"
	"os"


	"github.com/go-chi/chi/v5"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/handler"
	// "github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/logger"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/middleware"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/service"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/config"
	"github.com/rs/cors"


)


type Auth0Config struct {
	Auth0Secret string `env:"AUTH0_SECRET"`
	Auth0Domain string `env:"AUTH0_DOMAIN"`
	Auth0BaseURL string `env:"AUTH0_BASE_URL"`
	Auth0IssuerBaseURL string `env:"AUTH0_ISSUER_BASE_URL"`
	Auth0ClientID string `env:"AUTH0_CLIENT_ID"`
	Auth0ClientSecret string `env:"AUTH0_CLIENT_SECRET"`
	Auth0RoleID string `env:"AUTH0_ROLE_ID"`
	Auth0Audience string `env:"AUTH0_AUDIENCE"`
	Auth0HookSecret string `env:"AUTH_HOOK_SECRET"`
}

type DatabaseConfig struct {
	// todo update the connection string to be localhost, postgres, or whatever the host name is supposed to be 
	ConnString string `env:"CONN_STRING"`
}

// type Auth

type Config struct {
	ServerPort         string `json:"PORT"`
	DBConnString string `json:"CONN_STRING"`
	Auth0Secret string `json:"AUTH0_SECRET"`
	Auth0Domain string `json:"AUTH0_DOMAIN"`
	Auth0BaseURL string `json:"AUTH0_BASE_URL"`
	Auth0IssuerBaseURL string `json:"AUTH0_ISSUER_BASE_URL"`
	Auth0ClientID string `json:"AUTH0_CLIENT_ID"`
	Auth0ClientSecret string `json:"AUTH0_CLIENT_SECRET"`
	Auth0RoleID string `json:"AUTH0_ROLE_ID"`
	Auth0Audience string `json:"AUTH0_AUDIENCE"`
	Auth0HookSecret string `json:"AUTH_HOOK_SECRET"`
}


// todo add logger later on
func main() {
	ctx := context.Background()
	cfg , err := config.LoadConfig(ctx)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := NewDBClient(cfg.DBConnString)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}

	fmt.Println("Connected to database")

	svc := service.NewService(db)
	// todo: look more into why it is more appropriate to pass in pointers vs values
	h := handler.NewHandler(svc, cfg)
	r := chi.NewRouter()

	// Middleware
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Optional: Check critical dependencies
		// Example: Check database connection
		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("unhealthy: database unreachable"))
			return
		}
	
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	r.Post("/hook/user", h.HandleCreateUser) // New endpoint for getting/creating user

	r.Use(cors.New(cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-TOKEN"},
		AllowedOrigins: []string{
			"*",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"https://traba-staging.fs0ciety.dev",
			// "https://app.getclaimclam.com",
		},
		// Debug: true,
	}).Handler)
    r.Group(func(r chi.Router) {
        r.Use(middleware.EnsureValidToken(cfg))
        r.Get("/api/invoices", h.HandleFetchInvoices)
        r.Get("/api/user", h.HandleGetUser)
    })

	slog.InfoContext(ctx, "starting server", "port", cfg.ServerPort)
	// todo catch the error here
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}


// NewDBClient creates a new database client
func NewDBClient(psqlConnStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", psqlConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	// db.SetMaxOpenConns(25)
	// db.SetMaxIdleConns(25)
	// db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("postgres connection success")
	return db, nil
}




