package storage

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const Table = "urls"

type PgxStorage struct {
	Con *sql.DB
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
	i.rows.Close()
	return nil
}

func NewPgxStorage(con string) (*PgxStorage, error) {
	db, err := sql.Open(PostgresDriver, con)

	if err != nil {
		return nil, err
	}

	return &PgxStorage{Con: db}, nil
}

func (db *PgxStorage) Save(data []byte) error {
	vals := map[string]string{}

	// Приходится делать unmarshal поскольку на уровне драйвера нужно понимать
	// схему таблицы
	if err := json.Unmarshal(data, &vals); err != nil {
		return err
	}

	sql := "INSERT INTO urls (short, original) VALUES ($1, $2)"

	_, err := db.Con.Exec(sql, vals["short_url"], vals["original_url"])

	log.Println("Error here", err)

	return err
}

func (db *PgxStorage) Load() (Iterator, error) {
	rows, err := db.Con.Query("SELECT short, original FROM urls")

	if err != nil {
		return nil, err
	}

	return &PGXIterator{rows: rows}, nil
}

func (db *PgxStorage) Ping() error {
	return db.Con.Ping()
}
