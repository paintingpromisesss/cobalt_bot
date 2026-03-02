package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func migrate(db *sql.DB) error {
	queries, err := loadMigrationQueries()
	if err != nil {
		return err
	}

	for _, query := range queries {
		if strings.TrimSpace(query.sql) == "" {
			continue
		}

		if _, err := db.Exec(query.sql); err != nil {
			return fmt.Errorf("run migration %q: %w", query.name, err)
		}
	}

	return nil
}

type migrationQuery struct {
	name string
	sql  string
}

func loadMigrationQueries() ([]migrationQuery, error) {
	files, err := loadMigrationFiles()
	if err != nil {
		return nil, err
	}

	queries := make([]migrationQuery, 0, len(files))
	for _, p := range files {
		b, err := os.ReadFile(filepath.Clean(p))
		if err != nil {
			return nil, fmt.Errorf("read migration file %q: %w", p, err)
		}

		queries = append(queries, migrationQuery{
			name: filepath.Base(p),
			sql:  string(b),
		})
	}

	return queries, nil
}

func loadMigrationFiles() ([]string, error) {
	candidates := []string{
		"migrations",
		filepath.Join("..", "..", "migrations"),
		filepath.Join("..", "..", "..", "migrations"),
		filepath.Join("/app", "migrations"),
	}

	for _, dir := range candidates {
		entries, err := os.ReadDir(filepath.Clean(dir))
		if err == nil {
			files := make([]string, 0, len(entries))
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
					continue
				}
				files = append(files, filepath.Join(dir, entry.Name()))
			}

			if len(files) == 0 {
				continue
			}

			sort.Strings(files)
			return files, nil
		}

		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read migrations dir %q: %w", dir, err)
		}
	}

	return nil, fmt.Errorf("migration dir with .sql files not found in candidates: %v", candidates)
}
