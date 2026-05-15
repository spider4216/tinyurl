package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type (
	UserIdKey      string
	IsSignValidKey string
)

const (
	UserKey      UserIdKey      = "user_id"
	ValidSignKey IsSignValidKey = "is_sign_valid"
)

func (m Middleware) WithAuth(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("user_id")
		var userId string
		var sign string

		if err != nil {
			// Если ошибка и она не связана с ErrNoCookie, то останавливаемся
			if err != http.ErrNoCookie {
				m.logger.Error("something went wrong while getting cookie", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			m.logger.Debug("Cookie not found, generate new one")

			// Если куки просто нет, то генерируем новый User ID
			userId = uuid.NewString()

			// Подписываем новый идентификатор
			sign, err = m.signVal(userId, m.cfg.SignKey)
			if err != nil {
				m.logger.Error("cannot sign cookie", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Устанавливаем куку
			c := m.createCookie(userId, sign)
			http.SetCookie(w, &c)

			m.logger.Debug("Set user_id and sign validate result to ctx ", userId, true)

			// Устанавливаем user ID в контекст
			ctx := context.WithValue(r.Context(), UserKey, userId)
			ctx = context.WithValue(ctx, ValidSignKey, true)

			// создаем новый request с обновленным контекстом
			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
			return
		}

		m.logger.Debug("Cookie found")

		// Кука существует, извлекаем из нее ID пользователя
		sig := authCookie.Value

		// Получаем отдельно ID и отдельно подпись
		parts := strings.Split(sig, ".")

		if len(parts) < 2 {
			m.logger.Error("Something went wrong while getting cookies")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userId = parts[0]
		sign = parts[1]

		m.logger.Debug("Validate cookie...")

		isValid := false

		// Валидируем куку
		if err := m.validateSign(userId, m.cfg.SignKey, sign); err == nil {
			isValid = true
		}

		m.logger.Debug("Cookie validation result ", isValid)

		// Устанавливаем user ID в контекст
		ctx := context.WithValue(r.Context(), UserKey, userId)
		// Поскольку по требованию инкрементов, некоторые эндпоинты должны
		// пропускать невалидность, устанавливаем флаг валидности куки
		// в контекст и делегируем поведение невалидности обработчикам
		ctx = context.WithValue(ctx, ValidSignKey, isValid)

		// создаем новый request с обновленным контекстом
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(logFn)
}

// Подпись строки
func (m Middleware) signVal(val string, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))

	if _, err := h.Write([]byte(val)); err != nil {
		return "", err
	}

	sign := h.Sum(nil)

	return hex.EncodeToString(sign), nil
}

func (m Middleware) createCookie(userId string, sign string) http.Cookie {
	return http.Cookie{
		Name:     "user_id",
		Value:    userId + "." + sign,
		Path:     "/",
		Expires:  time.Now().Add(m.cfg.CookieTTL),
		HttpOnly: true,                 // Protects against XSS
		Secure:   true,                 // Only sent over HTTPS
		SameSite: http.SameSiteLaxMode, // CSRF protection
	}
}

func (m Middleware) validateSign(val string, key string, sig string) error {
	// Подписываем ключ
	f, err := m.signVal(val, key)
	if err != nil {
		return err
	}

	// Декодим и получаем байты
	src, err := hex.DecodeString(f)
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
