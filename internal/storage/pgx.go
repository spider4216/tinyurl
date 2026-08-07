package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/models"
)

const Table = "urls"

type recordPgx struct {
	ShortUrl    string `json:"short_url"`
	OriginarUrl string `json:"original_url"`
	IsDeleted   bool   `json:"deleted_at"`
	UserId      string `json:"user_id"`
}

// PgxStorage хранилище где данные складываются в БД PostgreSQL.
type PgxStorage struct {
	Con    *sql.DB
	logger *zap.SugaredLogger
}

// NewPgxStorage создание хранилища с БД PostgreSQL.
func NewPgxStorage(con string, logger *zap.SugaredLogger) (*PgxStorage, error) {
	db, err := sql.Open(PostgresDriver, con)
	if err != nil {
		return nil, err
	}

	return &PgxStorage{Con: db, logger: logger}, nil
}

func (db *PgxStorage) Ping(ctx context.Context) error {
	return db.Con.PingContext(ctx)
}

func (db *PgxStorage) Source() any {
	return db.Con
}

func (db *PgxStorage) StoreName() string {
	return PostgresDriver
}

func (db *PgxStorage) DeleteBatch(ctx context.Context, ids []string, userId string) error {
	sql := "UPDATE urls SET is_deleted=TRUE WHERE short = ANY($1) AND user_id = $2"

	_, err := db.Con.ExecContext(ctx, sql, ids, userId)

	return err
}

func (db *PgxStorage) GetByUserId(ctx context.Context, userId string) ([]models.UrlItem, error) {
	sql := "SELECT short, original, user_id, is_deleted FROM urls WHERE user_id = $1"

	rows, err := db.Con.QueryContext(ctx, sql, userId)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			db.logger.Warn("Cannot close rows", zap.Error(err))
		}
	}()

	var items []models.UrlItem

	for rows.Next() {
		var item recordPgx

		if err := rows.Scan(
			&item.ShortUrl,
			&item.OriginarUrl,
			&item.UserId,
			&item.IsDeleted,
		); err != nil {
			return nil, err
		}

		items = append(items, models.UrlItem{
			OriginarUrl: item.OriginarUrl,
			ShortUrl:    item.ShortUrl,
			UserId:      item.UserId,
			IsDeleted:   item.IsDeleted,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (db *PgxStorage) GetByOrigin(ctx context.Context, origin string) (*models.UrlItem, error) {
	sql := "SELECT short, original, user_id, is_deleted FROM urls WHERE original = $1"
	row := db.Con.QueryRowContext(ctx, sql, origin)

	var item recordPgx

	if err := row.Scan(
		&item.ShortUrl,
		&item.OriginarUrl,
		&item.UserId,
		&item.IsDeleted,
	); err != nil {
		return nil, err
	}

	return &models.UrlItem{
		OriginarUrl: item.OriginarUrl,
		ShortUrl:    item.ShortUrl,
		UserId:      item.UserId,
		IsDeleted:   item.IsDeleted,
	}, nil
}

func (db *PgxStorage) GetByShort(ctx context.Context, short string) (*models.UrlItem, error) {
	sql := "SELECT short, original, user_id, is_deleted FROM urls WHERE short = $1"
	row := db.Con.QueryRowContext(ctx, sql, short)

	var item recordPgx

	if err := row.Scan(
		&item.ShortUrl,
		&item.OriginarUrl,
		&item.UserId,
		&item.IsDeleted,
	); err != nil {
		return nil, err
	}

	return &models.UrlItem{
		OriginarUrl: item.OriginarUrl,
		ShortUrl:    item.ShortUrl,
		UserId:      item.UserId,
		IsDeleted:   item.IsDeleted,
	}, nil
}

// CountUsers количество пользователей в сервисе
func (db *PgxStorage) CountUsers(ctx context.Context) (int, error) {
	sql := "SELECT COUNT(DISTINCT user_id) from urls where is_deleted=false"

	row := db.Con.QueryRowContext(ctx, sql)

	var count int

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// CountUrls количество сокращённых URL в сервисе
func (db *PgxStorage) CountUrls(ctx context.Context) (int, error) {
	sql := "SELECT COUNT(*) from urls where is_deleted=false"

	row := db.Con.QueryRowContext(ctx, sql)

	var count int

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (db *PgxStorage) CreateUrl(ctx context.Context, data models.InsertData) error {
	sql := "INSERT INTO urls (short, original, user_id) VALUES ($1, $2, $3)"

	_, err := db.Con.ExecContext(ctx, sql, data.Key, data.Value, data.UserId)

	return err
}

func (db *PgxStorage) CreateUrls(ctx context.Context, data []models.InsertData) error {
	tx, err := db.Con.Begin()
	if err != nil {
		return err
	}

	for _, item := range data {
		sql := "INSERT INTO urls (short, original, user_id) VALUES ($1, $2, $3)"

		_, err := tx.ExecContext(ctx, sql, item.Key, item.Value, item.UserId)
		if err != nil {
			if transErr := tx.Rollback(); transErr != nil {
				return transErr
			}

			return err
		}
	}

	return tx.Commit()
}
