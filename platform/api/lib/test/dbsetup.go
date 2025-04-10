package test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"time"

	"log"

	"github.com/rasha-hantash/chariot-takehome/api/pkgs/postgres"

	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	pgC "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupDatabaseContainer() (*sql.DB, testcontainers.Container) {
	// Start a Docker container running PostgreSQL
	ctx := context.Background()
	dbName := "postgres"
	dbUser := "postgres"
	dbPassword := "postgres"
	pgContainer, err := pgC.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
		pgC.WithDatabase(dbName),
		pgC.WithUsername(dbUser),
		pgC.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(100*time.Second)),
	)
	if err != nil {
		panic(err)
	}

	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	db, err := postgres.NewDBClient(connString)
	if err != nil {
		log.Fatal(err)
	}
	driver, err := pg.WithInstance(db, &pg.Config{})
	if err != nil {
		log.Fatal(err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("No caller information")
	}

	dir := filepath.Dir(filename)
	migrationsDir := filepath.Join(dir, "../../../sql/migrations")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"postgres", driver)
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil {
		log.Fatal(err)
	}

	return db, pgContainer
}

func TeardownDatabaseContainer(container testcontainers.Container) error {
	ctx := context.Background()
	return container.Terminate(ctx)
}
