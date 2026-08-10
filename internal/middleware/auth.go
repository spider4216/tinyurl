package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WithAuth прослойка для аутинтификации и авторизации
func (m Middleware) WithAuth(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		if m.cfg.SignKey == "" {
			m.logger.Errorf("cannot find sign key. Set and try again.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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
			sign, err = m.service.SignVal(userId, m.cfg.SignKey)
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
			ctx := m.service.SetUserIdToCtx(r.Context(), userId)
			ctx = m.service.SetIsSignValidToCtx(ctx, true)

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
		if err := m.service.ValidateSign(userId, m.cfg.SignKey, sign); err == nil {
			isValid = true
		}

		m.logger.Debug("Cookie validation result ", isValid)

		// Устанавливаем user ID в контекст
		ctx := m.service.SetUserIdToCtx(r.Context(), userId)
		// Поскольку по требованию инкрементов, некоторые эндпоинты должны
		// пропускать невалидность, устанавливаем флаг валидности куки
		// в контекст и делегируем поведение невалидности обработчикам
		ctx = m.service.SetIsSignValidToCtx(ctx, isValid)

		// создаем новый request с обновленным контекстом
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(logFn)
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
