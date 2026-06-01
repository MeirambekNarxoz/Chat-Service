package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func resolveMigrationsPath() (string, error) {
	var candidates []string

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "migrations"),
			filepath.Join(wd, "..", "migrations"),
			filepath.Join(wd, "..", "..", "migrations"),
		)
	}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "migrations"),
			filepath.Join(dir, "..", "migrations"),
		)
	}

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			return filepath.Abs(c)
		}
	}

	return "", fmt.Errorf("migrations directory not found (run from chat-service root or set working directory)")
}

func RunMigrations(dsn string) {
	dir, err := resolveMigrationsPath()
	if err != nil {
		log.Printf("Migration path err: %v", err)
		return
	}

	sourceURL := "file://" + filepath.ToSlash(dir)
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		log.Printf("Migration init err: %v", err)
		return
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Migration failed: %v", err)
	} else {
		log.Printf("Chat Service: migrations applied from %s", dir)
	}
}
