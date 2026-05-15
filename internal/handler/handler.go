package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/middleware"
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/service"
	"go.uber.org/zap"
)

func New(conf *config.Config, logger *zap.SugaredLogger, service service.Service) Handler {
	return Handler{
		conf:    conf,
		service: service,
		logger:  logger,
	}
}

type Handler struct {
	conf    *config.Config
	service service.Service
	logger  *zap.SugaredLogger
}

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

	isCookieValid, ok := ctx.Value(middleware.IsSignValidKey).(bool)

	if !ok {
		h.logger.Error("cannot conver user id to string")
		return
	}

	if !isCookieValid {
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

	userId, ok := ctx.Value(middleware.UserIdKey).(string)

	if !ok {
		h.logger.Error("cannot conver user id to string")
		return
	}

	go h.service.DeleteBatchAsync(ctx, req, userId)

	h.logger.Debug("Accepted OK")
	w.WriteHeader(http.StatusAccepted)
}

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

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	userId, ok := ctx.Value(middleware.UserIdKey).(string)

	if !ok {
		h.logger.Error("cannot convert user id to string")
		return
	}

	urls := h.service.MapForMapUrlIds(req, userId)

	if err := h.service.StoreDataBatch(ctx, urls); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	resp := h.MapGenUrlsResp(urls, h.conf.BaseUrl)

	respJson, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("cannot marshall", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(respJson); err != nil {
		h.logger.Fatalln("failed to write response", zap.Error(err))
	}
}

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

	req := models.ShortenReq{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	id := h.service.GenerateId()

	userId, ok := ctx.Value(middleware.UserIdKey).(string)

	if !ok {
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

	full := fmt.Sprintf("%s/%s", h.conf.BaseUrl, id)

	resp := models.ShortenResp{
		Result: full,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("cannot marshall", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(respJson); err != nil {
		h.logger.Fatalln("failed to write response", zap.Error(err))
	}
}

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

	userId, ok := ctx.Value(middleware.UserIdKey).(string)

	if !ok {
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

	full := fmt.Sprintf("%s/%s", h.conf.BaseUrl, id)

	w.Header().Set("Content-Type", "plain/text")

	w.WriteHeader(status)

	if _, err := w.Write([]byte(full)); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

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

	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h Handler) Urls(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.conf.CtxTimeout)

	defer cancel()

	isValid, ok := ctx.Value(middleware.IsSignValidKey).(bool)

	if !ok {
		h.logger.Error("cannot conver is valid cookie to bool")
		return
	}

	// Если кука пришла, то нужно ее провалидировать
	if !isValid {
		h.logger.Error("Unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, ok := ctx.Value(middleware.UserIdKey).(string)

	if !ok {
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

func (h Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), h.conf.CtxTimeout)

	defer cancel()

	if err := h.service.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Cannot ping store", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Ping store OK")
}
