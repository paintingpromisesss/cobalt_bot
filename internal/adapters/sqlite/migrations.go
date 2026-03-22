package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	migrations "github.com/paintingpromisesss/cobalt_bot/migrations"
)

func Migrate(dbPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	migrationDB, err := sqlx.ConnectContext(ctx, "sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("connect sqlite db for migrations: %w", err)
	}
	defer func() {
		_ = migrationDB.Close()
	}()

	migrationDB.SetMaxOpenConns(1)

	if _, err := migrationDB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign keys pragma for migrations: %w", err)
	}

	if _, err := migrationDB.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		return fmt.Errorf("set busy_timeout pragma for migrations: %w", err)
	}

	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("init embedded migration source: %w", err)
	}

	dbDriver, err := sqlite.WithInstance(migrationDB.DB, &sqlite.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("init migration db driver: %w", err)
	}

	migrateInstance, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", dbDriver)
	if err != nil {
		return fmt.Errorf("init migrate instance: %w", err)
	}

	defer func() {
		_, _ = migrateInstance.Close()
	}()

	if err := migrateInstance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
