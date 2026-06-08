package database

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"booknest/backend/migrations"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(64) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`).Error; err != nil {
		return err
	}
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return db.Transaction(func(tx *gorm.DB) error {
		for _, name := range files {
			var exists int64
			if err := tx.Table("schema_migrations").Where("version = ?", name).Count(&exists).Error; err != nil {
				return err
			}
			if exists > 0 {
				continue
			}
			b, err := fs.ReadFile(migrations.FS, name)
			if err != nil {
				return fmt.Errorf("read migration %s: %w", name, err)
			}
			if err := tx.Exec(string(b)).Error; err != nil {
				return fmt.Errorf("apply migration %s: %w", name, err)
			}
			if err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, name).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
