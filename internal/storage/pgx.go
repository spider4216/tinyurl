package storage

import (
	"bytes"
	"database/sql"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgxStorage struct {
	Con *sql.DB
}

func NewPgxStorage(con string) (*PgxStorage, error) {
	db, err := sql.Open(PostgresDriver, con)

	if err != nil {
		return nil, err
	}

	return &PgxStorage{Con: db}, nil
}

func (db *PgxStorage) Save(data []byte) error {
	return nil
}

func (db *PgxStorage) Load() (io.Reader, error) {
	var buf bytes.Buffer

	return &buf, nil
}

func (db *PgxStorage) Ping() error {
	return db.Con.Ping()
}
