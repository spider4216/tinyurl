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
	"github.com/spider4216/tinyurl/internal/models"
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

func (s Service) StoreData(ctx context.Context, id string, val string, userId string) error {
	in := s.repo.MapInserData(id, val, userId)

	return s.repo.Insert(ctx, in)
}

func (s Service) StoreDataBatch(ctx context.Context, urls []UrlsIds) error {
	in := []models.InsertData{}

	for _, item := range urls {
		i := models.InsertData{
			Key:    item.Short,
			Value:  item.Origin,
			UserId: item.UserId,
		}

		in = append(in, i)
	}

	return s.repo.InsertBatch(ctx, in)
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
