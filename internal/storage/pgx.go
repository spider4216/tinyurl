package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const Table = "urls"

type PgxStorage struct {
	Con    *sql.DB
	logger *zap.SugaredLogger
}

type PGXIterator struct {
	rows *sql.Rows
}

func (i *PGXIterator) Next() bool {
	return i.rows.Next()
}

func (i *PGXIterator) Row() ([]byte, error) {

	var short string
	var origin string

	err := i.rows.Scan(&short, &origin)

	if err != nil {
		return nil, err
	}

	// Здесь приходится декларировать контракт
	// поскольку работаем с БД и нужно понимать поля
	tmp := map[string]string{
		"uuid":         short,
		"short_url":    short,
		"original_url": origin,
	}

	b, err := json.Marshal(tmp)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func (i *PGXIterator) Err() error {
	return i.rows.Err()
}

func (i *PGXIterator) Close() error {
	return i.rows.Close()
}

func NewPgxStorage(con string, logger *zap.SugaredLogger) (*PgxStorage, error) {
	db, err := sql.Open(PostgresDriver, con)

	if err != nil {
		return nil, err
	}

	return &PgxStorage{Con: db, logger: logger}, nil
}

func (db *PgxStorage) SaveBatch(ctx context.Context, data [][]byte) error {
	tx, err := db.Con.Begin()

	if err != nil {
		return err
	}

	for _, item := range data {
		line := map[string]string{}

		// Приходится делать unmarshal поскольку на уровне store нужно понимать
		// схему таблицы
		if err := json.Unmarshal(item, &line); err != nil {
			return err
		}

		short, ok := line["short_url"]

		if !ok {
			return fmt.Errorf("unrecognize columns")
		}

		origin, ok := line["original_url"]

		if !ok {
			return fmt.Errorf("unrecognize columns")
		}

		sql := "INSERT INTO urls (short, original) VALUES ($1, $2)"

		_, err := tx.ExecContext(ctx, sql, short, origin)

		if err != nil {
			if transErr := tx.Rollback(); transErr != nil {
				return transErr
			}

			return err
		}
	}

	return tx.Commit()
}

func (db *PgxStorage) Save(ctx context.Context, data []byte) error {
	vals := map[string]string{}

	// Приходится делать unmarshal поскольку на уровне store нужно понимать
	// схему таблицы
	if err := json.Unmarshal(data, &vals); err != nil {
		return err
	}

	short, ok := vals["short_url"]

	if !ok {
		return fmt.Errorf("unrecognize columns")
	}

	origin, ok := vals["original_url"]

	if !ok {
		return fmt.Errorf("unrecognize columns")
	}

	sql := "INSERT INTO urls (short, original) VALUES ($1, $2)"

	_, err := db.Con.ExecContext(ctx, sql, short, origin)

	return err
}

func (db *PgxStorage) Load(ctx context.Context) (Iterator, error) {
	rows, err := db.Con.QueryContext(ctx, "SELECT short, original FROM urls")

	if err != nil {
		return nil, err
	}

	return &PGXIterator{rows: rows}, nil
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
