package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type DB struct {
	sqlDB *sqlx.DB
}

func New(dbPath string) (*DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("db path is empty")
	}

	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir %q: %w", dir, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("connect sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys pragma: %w", err)
	}

	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set busy_timeout pragma: %w", err)
	}

	if err := Migrate(dbPath); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DB{sqlDB: db}, nil
}

func (db *DB) Close() error {
	return db.sqlDB.Close()
}
