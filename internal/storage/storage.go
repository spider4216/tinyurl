package storage

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/models"
)

const (
	FileDriver     = "file"
	MapDriver      = "map"
	PostgresDriver = "pgx"
)

type Storage interface {
	Ping(ctx context.Context) error
	Source() any
	StoreName() string
	DeleteBatch(ctx context.Context, ids []string, userId string) error
	GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error)
	GetByOrigin(ctx context.Context, origin string) (*models.UrlItem, error)
	GetByShort(ctx context.Context, origin string) (*models.UrlItem, error)
	CreateUrl(ctx context.Context, data models.InsertData) error
	CreateUrls(ctx context.Context, data []models.InsertData) error
}

type Iterator interface {
	Next() bool
	Row() ([]byte, error)
	Err() error
	Close() error
}

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
