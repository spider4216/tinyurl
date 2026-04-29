package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spider4216/tinyurl/internal/repository"
)

func New(repo *repository.Repository) Service {
	return Service{
		repo: repo,
	}
}

type Service struct {
	repo *repository.Repository
}

func (s Service) GenerateId() string {
	key := make([]byte, 6)
	rand.Read(key)

	return base64.URLEncoding.EncodeToString(key)
}

func (s Service) StoreData(ctx context.Context, id string, val string) error {
	return s.repo.Insert(ctx, id, val)
}

func (s Service) StoreDataBatch(ctx context.Context, urls []UrlsIds) error {
	keyValues := map[string]string{}

	for _, item := range urls {
		keyValues[item.Short] = item.Origin
	}

	return s.repo.InsertBatch(ctx, keyValues)
}

func (s Service) GetUrl(ctx context.Context, k string) (string, error) {
	item, err := s.repo.Get(ctx, k)

	if err != nil {
		return "", err
	}

	itemMap := map[string]string{}

	err = json.Unmarshal([]byte(item), &itemMap)

	if err != nil {
		return "", err
	}

	url, ok := itemMap["original_url"]

	if !ok {
		return "", fmt.Errorf("cannot get original url from map")
	}

	return url, nil
}

func (s Service) GetShortByOrigin(ctx context.Context, v string) (string, error) {
	item, err := s.repo.GetByValue(ctx, v)

	if err != nil {
		return "", err
	}

	itemMap := map[string]string{}

	err = json.Unmarshal([]byte(item), &itemMap)

	if err != nil {
		return "", err
	}

	url, ok := itemMap["short_url"]

	if !ok {
		return "", fmt.Errorf("cannot get short url from map")
	}

	return url, nil
}

func (s Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s Service) IsErrAsDuplicate(err error) bool {
	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgerrcode.UniqueViolation
}
