package migrations

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Поскольку GitLab Pipeline по тестированию инкрементов
// не содержит функционал запуска миграций и по условию
// задания "Сервису нужно самостоятельно создать все
// необходимые таблицы в базе данных", следовательно придется
// запустить миграции при старте приложения

//go:embed *.sql
var FS embed.FS

func Run(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		return err
	}

	d, err := iofs.New(FS, ".")

	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
