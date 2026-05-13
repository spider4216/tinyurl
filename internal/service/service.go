package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
	item, err := s.repo.Get(ctx, k)
	if err != nil {
		return "", err
	}

	itemMap := map[string]string{}

	err = json.Unmarshal([]byte(item), &itemMap)
	if err != nil {
		return "", err
	}

	s.logger.Debug("Url data ", itemMap)

	url, ok := itemMap["original_url"]

	if !ok {
		return "", fmt.Errorf("cannot get original url from map")
	}

	deleted, ok := itemMap["is_deleted"]

	if !ok {
		return "", fmt.Errorf("cannot get is_deleted from map")
	}

	b, err := strconv.ParseBool(deleted)
	if err != nil {
		return "", err
	}

	if b {
		return "", DeletedUrlError{}
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

// Подпись строки
func (s Service) SignVal(val string, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))

	if _, err := h.Write([]byte(val)); err != nil {
		return "", err
	}

	sign := h.Sum(nil)

	return hex.EncodeToString(sign), nil
}

func (s Service) ValidateSign(val string, key string, sig string) error {
	// Подписываем ключ
	m, err := s.SignVal(val, key)
	if err != nil {
		return err
	}

	// Декодим и получаем байты
	src, err := hex.DecodeString(m)
	if err != nil {
		return err
	}

	// Декодим подпись которую нужно проверить
	dst, err := hex.DecodeString(sig)
	if err != nil {
		return err
	}

	if hmac.Equal(src, dst) {
		return nil
	}

	return errors.New("invalid signature")
}
