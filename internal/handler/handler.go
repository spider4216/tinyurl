package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spider4216/tinyurl/internal/config"
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
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)

		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	lr := io.LimitReader(r.Body, h.conf.MaxBodySize)

	body, err := io.ReadAll(lr)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		return
	}

	req := []string{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	userId, needSetCookie, sign, err := h.authCookie(r)
	if err != nil {
		h.logger.Error("something went wrong while getting cookie", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если кука пришла, то нужно ее провалидировать
	if !needSetCookie {
		h.logger.Debug("Validate user...")
		if err := h.service.ValidateSign(userId, h.conf.SignKey, sign); err != nil {
			// Если  токен не валидный, нет смысла ходить в БД и удалять
			// записи, поскольку таковых не будет
			h.logger.Error("Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if len(req) <= 0 {
		h.logger.Error("Empty ids")
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	ctx := context.WithoutCancel(context.Background())

	go h.service.DeleteBatchAsync(ctx, req, userId)

	h.logger.Debug("Accepted OK")
	w.WriteHeader(http.StatusAccepted)
}

func (h Handler) GetShortenUrls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)

		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.conf.CtxTimeout)

	defer cancel()

	lr := io.LimitReader(r.Body, h.conf.MaxBodySize)

	body, err := io.ReadAll(lr)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		return
	}

	req := []models.ShortenBatchReq{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	userId, needSetCookie, sign, err := h.authCookie(r)
	if err != nil {
		h.logger.Error("something went wrong while getting cookie", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если кука пришла, то нужно ее провалидировать
	if !needSetCookie {
		h.logger.Debug("Validate user...")
		if err := h.service.ValidateSign(userId, h.conf.SignKey, sign); err != nil {
			// По требованию если токен не валидный, устанавливаем новый
			// Делаем новый userId
			userId = uuid.NewString()
			// Устанавливаем новую куку
			needSetCookie = true
		}
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

	if needSetCookie {
		h.logger.Debug("Set cookie")

		uidSign, err := h.service.SignVal(userId, h.conf.SignKey)
		if err != nil {
			h.logger.Error("cannot sign user id")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cookie := h.createCookie(userId, uidSign)

		http.SetCookie(w, &cookie)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(respJson); err != nil {
		h.logger.Fatalln("failed to write response", zap.Error(err))
	}
}

func (h Handler) GetShortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)

		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.conf.CtxTimeout)

	defer cancel()

	lr := io.LimitReader(r.Body, h.conf.MaxBodySize)

	body, err := io.ReadAll(lr)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		return
	}

	req := models.ShortenReq{}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("unmarshall error", zap.Error(err))
		return
	}

	id := h.service.GenerateId()

	userId, needSetCookie, sign, err := h.authCookie(r)
	if err != nil {
		h.logger.Error("something went wrong while getting cookie", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если кука пришла, то нужно ее провалидировать
	if !needSetCookie {
		h.logger.Debug("validate user...")
		if err := h.service.ValidateSign(userId, h.conf.SignKey, sign); err != nil {
			// По требованию если токен не валидный, устанавливаем новый
			// Делаем новый userId
			userId = uuid.NewString()
			// Устанавливаем новую куку
			needSetCookie = true
		}
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

	if needSetCookie {
		h.logger.Debug("Set cookie")

		uidSign, err := h.service.SignVal(userId, h.conf.SignKey)
		if err != nil {
			h.logger.Error("cannot sign user id")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cookie := h.createCookie(userId, uidSign)

		http.SetCookie(w, &cookie)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

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

	userId, needSetCookie, sign, err := h.authCookie(r)
	if err != nil {
		h.logger.Error("something went wrong while getting cookie", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если кука пришла, то нужно ее провалидировать
	if !needSetCookie {
		h.logger.Debug("Validate user...")
		if err := h.service.ValidateSign(userId, h.conf.SignKey, sign); err != nil {
			// По требованию если токен не валидный, устанавливаем новый
			// Делаем новый userId
			userId = uuid.NewString()
			// Устанавливаем новую куку
			needSetCookie = true
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.conf.CtxTimeout)

	defer cancel()

	lr := io.LimitReader(r.Body, h.conf.MaxBodySize)

	url, err := io.ReadAll(lr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		if _, err := w.Write([]byte("cannot read body")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	id := h.service.GenerateId()

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

	if needSetCookie {
		h.logger.Debug("Set cookie")
		uidSign, err := h.service.SignVal(userId, h.conf.SignKey)
		if err != nil {
			h.logger.Error("cannot sign user id")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cookie := h.createCookie(userId, uidSign)

		http.SetCookie(w, &cookie)
	}

	w.WriteHeader(status)

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

	userId, needSetCookie, sign, err := h.authCookie(r)
	if err != nil {
		h.logger.Error("something went wrong while getting cookie", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если кука пришла, то нужно ее провалидировать
	if !needSetCookie {
		h.logger.Debug("Validate user...")
		if err := h.service.ValidateSign(userId, h.conf.SignKey, sign); err != nil {
			// По требованию если токен не валидный, устанавливаем новый
			// Делаем новый userId
			userId = uuid.NewString()
			// Устанавливаем новую куку
			needSetCookie = true
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.conf.CtxTimeout)

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

	if needSetCookie {
		h.logger.Debug("Set cookie")
		uidSign, err := h.service.SignVal(userId, h.conf.SignKey)
		if err != nil {
			h.logger.Error("cannot sign user id")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cookie := h.createCookie(userId, uidSign)

		http.SetCookie(w, &cookie)
	}

	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h Handler) Urls(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), h.conf.CtxTimeout)

	defer cancel()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		if _, err := w.Write([]byte("Method not allowed")); err != nil {
			h.logger.Error("failed to write response", zap.Error(err))
		}

		return
	}

	userId, needSetCookie, sign, err := h.authCookie(r)
	if err != nil {
		h.logger.Error("something went wrong while getting cookie", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.Debug("extracted user id is ", userId)

	// Если кука пришла, то нужно ее провалидировать
	if !needSetCookie {
		h.logger.Debug("Validate user...")

		if err := h.service.ValidateSign(userId, h.conf.SignKey, sign); err != nil {
			h.logger.Error("Unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
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
		// По требованию если нету urls, то выдаем новую куку
		userId = uuid.NewString()
		// Устанавливаем новую куку
		h.logger.Debug("Set cookie")
		uidSign, err := h.service.SignVal(userId, h.conf.SignKey)
		if err != nil {
			h.logger.Error("cannot sign user id")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cookie := h.createCookie(userId, uidSign)

		http.SetCookie(w, &cookie)

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

func (h Handler) authCookie(r *http.Request) (userId string, needSet bool, sign string, err error) {
	authCookie, err := r.Cookie("user_id")

	if err != nil {
		// Если ошибка и она не связана с ErrNoCookie, то останавливаемся
		if err != http.ErrNoCookie {
			h.logger.Error("something went wrong while getting cookie", zap.Error(err))
			return
		}

		// Если куки просто нет, то генерируем новый User ID
		userId = uuid.NewString()
		// Перед позитивным ответом, нужно установить новую куку
		needSet = true

		// Подписываем новый идентификатор
		sign, err = h.service.SignVal(userId, h.conf.SignKey)
		if err != nil {
			return
		}
	} else {
		// Кука существует, извлекаем из нее ID пользователя
		sig := authCookie.Value

		// Получаем отдельно ID и отдельно подпись
		parts := strings.Split(sig, ".")

		if len(parts) < 2 {
			err = errors.New("cannot get cookie parts")
			return
		}

		userId = parts[0]
		sign = parts[1]
	}

	return
}

func (h Handler) createCookie(userId string, sign string) http.Cookie {
	return http.Cookie{
		Name:     "user_id",
		Value:    userId + "." + sign,
		Path:     "/",
		Expires:  time.Now().Add(h.conf.CookieTTL),
		HttpOnly: true,                 // Protects against XSS
		Secure:   true,                 // Only sent over HTTPS
		SameSite: http.SameSiteLaxMode, // CSRF protection
	}
}
