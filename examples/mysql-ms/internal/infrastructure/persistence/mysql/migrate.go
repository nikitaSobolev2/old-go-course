package mysql // применение SQL-миграций из каталога через golang-migrate

import (
	"database/sql"  // пул соединений с MySQL
	"errors"        // ErrNoChange
	"fmt"           // обёртка ошибки migrate
	"path/filepath" // абсолютный путь к файлам миграций

	"github.com/golang-migrate/migrate/v4"                             // оркестратор версий схемы
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql" // драйвер «mysql» для migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"               // схема URL file:// для чтения .sql с диска
)

// RunMigrations applies SQL files from migrationsDir (absolute path) using the same *sql.DB connection.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	driver, err := mysqlmigrate.WithInstance(db, &mysqlmigrate.Config{}) // обёртка *sql.DB для migrate
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(migrationsDir) // migrate требует стабильный абсолютный путь
	if err != nil {
		return err
	}
	fileURL := "file://" + filepath.ToSlash(abs) // URL источника миграций для Windows/Unix
	m, err := migrate.NewWithDatabaseInstance(fileURL, "mysql", driver) // связываем файлы и БД
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }() // закрываем источник (не закрывает db)
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) { // применяем все неприменённые up
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
