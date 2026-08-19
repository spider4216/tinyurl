// Пакет storage содержит различные драйверы для работы с хранилищем.
//
// Список драйверов
//   - file - драйвер для хранения данных в файле
//   - map  - в сервисе исторически сложилось, что наименование драйвера на самом деле является slice
//   - pgx - драйвер для хранения данных в PostgreSQL
package storage

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/models"
)

// Наименование драйверов
const (
	FileDriver     = "file" // FileDriver хранилищем является файл.
	MapDriver      = "map"  // MapDriver хранилищем является Slice.
	PostgresDriver = "pgx"  // PostgresDriver хранилищем является БД POstgreSQL.
)

// Storage - основной интерфейс хранилища.
// Чтобы реализовать новое хранилище, необходимо реализовать все методы указанные в интерфейсе.
type Storage interface {
	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error

	// Source возвращает инкапсулированное хранилище (источник).
	Source() any

	// StoreName возвращает наименование хранилища.
	StoreName() string

	// DeleteBatch удаляет из хранилища множество данных (пачки).
	DeleteBatch(ctx context.Context, ids []string, userId string) error

	// GetByUserId извлекает из хранилища данные по идентификатору пользователя.
	GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error)

	// GetByOrigin извлекает сокращенный URL по оригинальному.
	GetByOrigin(ctx context.Context, origin string) (*models.UrlItem, error)

	// GetByShort извлекает из хранилища оригинальный URL по сокращенному.
	GetByShort(ctx context.Context, origin string) (*models.UrlItem, error)

	// CreateUrl создает в хранилище запись по сокращенному URL.
	CreateUrl(ctx context.Context, data models.InsertData) error

	// CreateUrls создает в хранилище множество сокращенных url (пачки).
	CreateUrls(ctx context.Context, data []models.InsertData) error

	// CountUsers возвращает кол-во униникальных пользователей.
	CountUsers(ctx context.Context) (int, error)

	// CountUrls возвращает общее кол-во сокращенных URL.
	CountUrls(ctx context.Context) (int, error)
}

// New конструктор хранилища, создает конкретное хранилище по идентификатору.
func New(driver string, cfg *config.Config, logger *zap.SugaredLogger) (Storage, error) {
	switch driver {
	case FileDriver:
		fileStore, err := NewFileStorage(cfg.FileStorePath, logger)
		if err != nil {
			return nil, err
		}

		return fileStore, nil
	case MapDriver:
		mapStore := NewMapStorage(logger)
		return mapStore, nil
	case PostgresDriver:
		pgxStore, err := NewPgxStorage(cfg.DbDsn, logger)
		if err != nil {
			return nil, err
		}

		return pgxStore, nil
	}

	return nil, fmt.Errorf("unsupported driver %s", driver)
}
