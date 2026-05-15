package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/repository"
	"go.uber.org/zap"
)

func New(repo *repository.Repository, logger *zap.SugaredLogger) Service {
	return Service{
		repo:   repo,
		logger: logger,
	}
}

type Service struct {
	repo   *repository.Repository
	logger *zap.SugaredLogger
}

type DeletedUrlError struct{}

func (e DeletedUrlError) Error() string {
	return "Url was deleted"
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

func (s Service) DeleteBatch(ctx context.Context, ids []string, userId string) error {
	return s.repo.DeleteByIds(ctx, ids, userId)
}

func (s Service) DeleteBatchAsync(ctx context.Context, ids []string, userId string) {
	// канал с данными
	inputCh := s.GenerateChunk(ids, 2, userId)

	// получаем слайс каналов
	channels := s.FanOutDeleteBatch(ctx, inputCh)

	// а теперь объединяем каналы в один
	resCh := s.FanInDeleteBatch(channels)

	// логируем результаты расчетов из канала
	for res := range resCh {
		s.logger.Debug("Result chank delete received")

		if res.Err != nil {
			s.logger.Error("Chank delete error ", zap.Error(res.Err), res.IDs)
		}
	}
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
	item, err := s.repo.GetByShort(ctx, k)

	if err != nil {
		return "", err
	}

	if item.IsDeleted {
		return "", DeletedUrlError{}
	}

	return item.OriginarUrl, nil
}

func (s Service) GetShortByOrigin(ctx context.Context, v string) (string, error) {
	item, err := s.repo.GetByOrigin(ctx, v)

	if err != nil {
		return "", err
	}

	return item.ShortUrl, nil
}

func (s Service) GetUrlsByUserId(ctx context.Context, userId string, baseUrl string) ([]models.UrlItem, error) {
	items, err := s.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].ShortUrl = fmt.Sprintf("%s/%s", baseUrl, items[i].ShortUrl)
	}

	return items, nil
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
