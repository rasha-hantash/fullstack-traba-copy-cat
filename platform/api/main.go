package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log/slog"
	"time"
	"github.com/joho/godotenv"
	"github.com/caarlos0/env/v6"


	"context"
	"fmt"
	"net/http"
	"os"


	"github.com/go-chi/chi/v5"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/handler"
	// "github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/logger"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/lib/middleware"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/service"
	"github.com/rs/cors"
)


type DatabaseConfig struct {
	// todo update the connection string to be localhost, postgres, or whatever the host name is supposed to be 
	ConnString string `env:"CONN_STRING" envDefault:"postgresql://admin:your_password@localhost:5438/traba?sslmode=disable"`
	User       string `env:"DB_USER" envDefault:""`
	Port       string `env:"DB_PORT" envDefault:""`
	Host       string `env:"DB_HOST" envDefault:""`
	Region     string `env:"DB_REGION" envDefault:""`
	DBName     string `env:"DB_NAME" envDefault:""`
}

// type Auth

type Config struct {
	ServerPort         string `env:"PORT" envDefault:"8000"`
	Database           DatabaseConfig
	Mode               string `env:"MODE" envDefault:"local"`
}


// todo add logger later on
func main() {
	godotenv.Load(".env")
	var c Config
	err := env.Parse(&c)
	if err != nil {
		slog.Error("failed to parse default config", "error", err)
		os.Exit(1)
	}

	db, err := NewDBClient(c.Database.ConnString)
	if err != nil {
		slog.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}

	fmt.Println("Connected to database")

	svc := service.NewService(db)
	h := handler.NewHandler(svc)

	r := chi.NewRouter()

	// Middleware
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	r.Use(cors.New(cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-TOKEN"},
		AllowedOrigins: []string{
			"*",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			// "https://staging.getclaimclam.com",
			// "https://app.getclaimclam.com",
		},
		// Debug: true,
	}).Handler)
    r.Group(func(r chi.Router) {
        r.Use(middleware.EnsureValidToken())
        r.Get("/api/invoices", h.HandleFetchInvoices)
        r.Get("/api/user", h.HandleGetUser) 
		r.Get("/api/user-id", h.HandleGetUserId) 
    })

	r.Post("/hook/user", h.HandleCreateUser) // New endpoint for getting/creating user

	// todo catch the error here
	if err := http.ListenAndServe(":"+c.ServerPort, r); err != nil {
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


