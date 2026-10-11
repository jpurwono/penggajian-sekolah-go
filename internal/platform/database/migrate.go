package database

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"gorm.io/gorm"

	"penggajian/server/migrations"
)

// Migrate menjalankan semua file SQL di migrations/ secara berurutan.
// Versi yang sudah diterapkan dicatat di tabel schema_migrations.
func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    VARCHAR(50) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`).Error; err != nil {
		return fmt.Errorf("buat schema_migrations: %w", err)
	}

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var count int64
		if err := db.Raw(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, name).Scan(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		content, err := migrations.FS.ReadFile(name)
		if err != nil {
			return err
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(content)).Error; err != nil {
				return err
			}
			return tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, name).Error
		})
		if err != nil {
			return fmt.Errorf("migrasi %s gagal: %w", name, err)
		}
		log.Printf("[db] migrasi diterapkan: %s", name)
	}

	return nil
}
