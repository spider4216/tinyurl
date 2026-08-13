package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/pool"
	"github.com/spider4216/tinyurl/internal/service"
)

// Handler основной обработчик запросов.
type Handler struct {
	conf         *config.Config
	service      service.Service
	logger       *zap.SugaredLogger
	delSemaphore chan struct{}
	reqPool      pool.ReqPools
}

// New конструктор обработчика. Зависим от:
// - конфигурации.
// - логгера.
// - сервиса.
func New(conf *config.Config, logger *zap.SugaredLogger, service service.Service, reqPool pool.ReqPools) Handler {
	return Handler{
		conf:         conf,
		service:      service,
		logger:       logger,
		delSemaphore: make(chan struct{}, conf.DeleteMaxPool),
		reqPool:      reqPool,
	}
}

// DeleteUrls удаление множества сокращенных URL.
func (h Handler) DeleteUrls(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.conf.MaxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := []string{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	ctx := r.Context()

	if !h.service.IsSignValidFromCtx(ctx) {
		// Если  токен не валидный, нет смысла ходить в БД и удалять
		// записи, поскольку таковых не будет
		h.logger.Error("Unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(req) <= 0 {
		h.logger.Error("Empty ids")
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	ctx = context.WithoutCancel(ctx)

	userId := h.service.GetUserIdFromCtx(ctx)

	if userId == "" {
		h.logger.Error("cannot conver user id to string")
		return
	}

	// Реализация паттерна Семафора, ограничиваем
	// кол-во задач на удаление
	select {
	case h.delSemaphore <- struct{}{}:
		go func() {
			defer func() {
				<-h.delSemaphore
				h.logger.Debug("Release task for delete. Left: ", len(h.delSemaphore))
			}()
			h.logger.Debug("Push task for delete. Left: ", len(h.delSemaphore))
			h.service.DeleteBatchAsync(ctx, req, userId)
		}()
	default:
		// Поскольку эндпоинт должен возвращать сразу же HTTP 202
		// Если превышен лимит запросов, то выводим ошибку
		// HTTP 429
		h.logger.Error("Too many tasks for delete. Try again later")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	// По требованию к задаче, эндпоинт должен синхронно возвращать
	// 202 Accepted OK
	h.logger.Debug("Accepted OK")
	w.WriteHeader(http.StatusAccepted)
}

// GetShortenUrls сокращение множества URL.
func (h Handler) GetShortenUrls(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	r.Body = http.MaxBytesReader(w, r.Body, h.conf.MaxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := []models.ShortenBatchReq{}

	if umErr := json.Unmarshal(body, &req); umErr != nil {
		h.logger.Error("unmarshall error", zap.Error(umErr))
		return
	}

	userId := h.service.GetUserIdFromCtx(ctx)

	if userId == "" {
		h.logger.Error("cannot convert user id to string")
		return
	}

	urls := h.service.MapForMapUrlIds(req, userId)

	if storeErr := h.service.StoreDataBatch(ctx, urls); storeErr != nil {
		h.logger.Error("unmarshall error", zap.Error(storeErr))
		return
	}

	resp := h.MapGenUrlsResp(urls, h.conf.BaseUrl)

	respJson, mErr := json.Marshal(resp)
	if mErr != nil {
		h.logger.Error("cannot marshall", zap.Error(mErr))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(respJson); err != nil {
		h.logger.Fatalln("failed to write response", zap.Error(err))
	}
}

// GetShortenUrl сокращение URL.
func (h Handler) GetShortenUrl(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	r.Body = http.MaxBytesReader(w, r.Body, h.conf.MaxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := h.reqPool.ShortenReq.Get()
	defer h.reqPool.ShortenReq.Put(req)

	if umErr := json.Unmarshal(body, &req); umErr != nil {
		h.logger.Error("unmarshall error", zap.Error(umErr))
		return
	}

	id := h.service.GenerateId()

	userId := h.service.GetUserIdFromCtx(ctx)

	if userId == "" {
		h.logger.Error("cannot convert user id to string")
		return
	}

	err = h.service.StoreData(ctx, id, string(req.Url), userId)

	if err != nil && !h.service.IsErrAsDuplicate(err) {
		h.logger.Error("store error", zap.Error(err))
		return
	}

	status := http.StatusCreated

	// Если дубликат, то тогда извлекаем по значению
	if err != nil && h.service.IsErrAsDuplicate(err) {
		h.logger.Debug("Duplicate, try getting exist reccord")
		status = http.StatusConflict

		id, err = h.service.GetShortByOrigin(ctx, req.Url)
		if err != nil {
			h.logger.Error("store error", zap.Error(err))
			return
		}
	}

	full := h.conf.BaseUrl + "/" + id

	resp := models.ShortenResp{
		Result: full,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("cannot marshall", zap.Error(err))
		return
	}

	h.service.AuditNotify(audit.ShortenAction, userId, string(req.Url))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(respJson); err != nil {
		h.logger.Fatalln("failed to write response", zap.Error(err))
	}
}

// GenerateId сокращение URL.
func (h Handler) GenerateId(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	r.Body = http.MaxBytesReader(w, r.Body, h.conf.MaxBodySize)

	url, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := h.service.GenerateId()

	userId := h.service.GetUserIdFromCtx(ctx)

	if userId == "" {
		h.logger.Error("cannot conver user id to string")
		return
	}

	err = h.service.StoreData(ctx, id, string(url), userId)

	if err != nil && !h.service.IsErrAsDuplicate(err) {
		h.logger.Error("store error", zap.Error(err))
		return
	}

	status := http.StatusCreated

	// Если дубликат, то тогда извлекаем по значению
	if err != nil && h.service.IsErrAsDuplicate(err) {
		h.logger.Debug("Duplicate, try getting exist reccord")
		status = http.StatusConflict

		id, err = h.service.GetShortByOrigin(ctx, string(url))
		if err != nil {
			h.logger.Error("store error", zap.Error(err))
			return
		}
	}

	full := h.conf.BaseUrl + "/" + id

	h.service.AuditNotify(audit.ShortenAction, userId, string(url))

	w.Header().Set("Content-Type", "plain/text")

	w.WriteHeader(status)

	if _, err := w.Write([]byte(full)); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// GetUrl получение сокращенного URL.
func (h Handler) GetUrl(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	id := r.PathValue("id")

	url, err := h.service.GetUrl(ctx, id)
	var deletedError service.DeletedUrlError

	if errors.As(err, &deletedError) {
		h.logger.Error("url was deleted")
		w.WriteHeader(http.StatusGone)
		return
	}

	if err != nil {
		h.logger.Error("get data error", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)

		if _, err := w.Write([]byte("Url not found")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	if url == "" {
		w.WriteHeader(http.StatusNotFound)

		if _, err := w.Write([]byte("Url not found")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	userId := h.service.GetUserIdFromCtx(ctx)

	h.service.AuditNotify(audit.FollowAction, userId, url)

	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Urls получение всех сокращенных URL пользователя.
func (h Handler) Urls(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	// Если кука пришла, то нужно ее провалидировать
	if !h.service.IsSignValidFromCtx(ctx) {
		h.logger.Error("Unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId := h.service.GetUserIdFromCtx(ctx)

	if userId == "" {
		h.logger.Error("cannot conver user id to string")
		return
	}

	urls, err := h.service.GetUrlsByUserId(ctx, userId, h.conf.BaseUrl)
	if err != nil {
		h.logger.Error("get data error", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.logger.Debug("Got urls: ", urls)

	if len(urls) <= 0 {
		h.logger.Debug("No items for user ", userId)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	d, err := json.Marshal(urls)
	if err != nil {
		h.logger.Error("marshal response error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(d); err != nil {
		h.logger.Fatalln("failed to write response", zap.Error(err))
	}
}

// Ping проверка доступности источника данных.
func (h Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	if err := h.service.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Cannot ping store", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Ping store OK")
}

// Stat внутренний эндпоинт показывающий статистику по сервису.
func (h Handler) Stat(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	users, err := h.service.CountUsers(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Cannot count users", zap.Error(err))
		return
	}

	urls, err := h.service.CountUrls(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Cannot count urls", zap.Error(err))
		return
	}

	resp := models.StatResp{
		UrlsCount:  urls,
		UsersCount: users,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Cannot prepare response", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("failed to write response", zap.Error(err))
	}
}
