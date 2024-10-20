package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log/slog"
	"time"

	"github.com/caarlos0/env/v6"


	"context"
	"fmt"
	"net/http"
	"os"


	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/handler"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/api/service"
)


type DatabaseConfig struct {
	// todo update the connection string to be localhost, postgres, or whatever the host name is supposed to be 
	ConnString string `env:"CONN_STRING" envDefault:"postgresql://postgres:postgres@localhost:5438/?sslmode=disable"`
	User       string `env:"DB_USER" envDefault:""`
	Port       string `env:"DB_PORT" envDefault:""`
	Host       string `env:"DB_HOST" envDefault:""`
	Region     string `env:"DB_REGION" envDefault:""`
	DBName     string `env:"DB_NAME" envDefault:""`
}

type Config struct {
	ServerPort         string `env:"PORT" envDefault:"8000"`
	Database           DatabaseConfig
	Mode               string `env:"MODE" envDefault:"local"`
}


// todo add logger later on
func main() {
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
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// todo middleware handler for protected route 
	// r.Middlewares().Handler()

	// Public routes
	// r.Get("/", publicRoute)
	r.Get("/api/invoices", h.HandleFetchInvoices)
	r.Post("/api/create-user", h.HandleCreateUser)

	// // Protected routes
	// r.Group(func(r chi.Router) {
		
	// })

	// todo catch the error here
	http.ListenAndServe(":8000", r)

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


