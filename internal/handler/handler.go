package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/service"
	"go.uber.org/zap"
)

const maxBodySize = 2 * 1024

func New(conf config.Config, logger *zap.SugaredLogger, service service.Service) Handler {
	return Handler{
		conf:    conf,
		service: service,
		logger:  logger,
	}
}

type Handler struct {
	conf    config.Config
	service service.Service
	logger  *zap.SugaredLogger
}

func (h Handler) GetShortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)

		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	lr := io.LimitReader(r.Body, maxBodySize)

	body, err := io.ReadAll(lr)

	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Warn("failed to close request body", zap.Error(err))
		}
	}()

	req := models.ShortenReq{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	id := h.service.GenerateId()
	h.service.StoreData(id, string(req.Url))

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
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(respJson); err != nil {
		h.logger.Fatalln("failed to write response", zap.Error(err))
	}
}

func (h Handler) GenerateId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)

		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	lr := io.LimitReader(r.Body, maxBodySize)

	url, err := io.ReadAll(lr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		if _, err := w.Write([]byte("cannot read body")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("failed to close request body", zap.Error(err))
		}
	}()

	id := h.service.GenerateId()
	h.service.StoreData(id, string(url))

	full := fmt.Sprintf("%s/%s", h.conf.BaseUrl, id)

	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(full)); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

func (h Handler) GetUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	id := r.PathValue("id")

	url := h.service.GetUrl(id)

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
