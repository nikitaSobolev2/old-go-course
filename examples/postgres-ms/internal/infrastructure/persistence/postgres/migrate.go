package postgres // golang-migrate с драйвером БД postgres (pgx через database/sql)

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres" // не путать с именем пакета этого файла
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB, migrationsDir string) error {
	driver, err := postgresmigrate.WithInstance(db, &postgresmigrate.Config{}) // обёртка для migrate
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(migrationsDir) // абсолютный путь к SQL
	if err != nil {
		return err
	}
	fileURL := "file://" + filepath.ToSlash(abs)
	m, err := migrate.NewWithDatabaseInstance(fileURL, "postgres", driver) // имя БД для migrate — "postgres"
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
