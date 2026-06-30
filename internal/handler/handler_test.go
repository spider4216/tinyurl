package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/logger"
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
	"github.com/spider4216/tinyurl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepateHandler(store storage.Storage) Handler {
	conf, err := config.New()
	if err != nil {
		panic("cannot create config")
	}

	r := repository.New(store)
	logger, err := logger.InitZap("debug")
	event := audit.NewAuditEvent()
	s := service.New(r, logger, event)
	if err != nil {
		panic("cannot prepare handler")
	}

	return New(conf, logger, s)
}

func TestGetShortenUrl(t *testing.T) {
	type want struct {
		contentType string
		status      int
	}

	cases := []struct {
		name   string
		method string
		urlTo  string
		urlSrc string
		userId string
		want   want
	}{
		{
			name:   "Case #1 Positive",
			method: http.MethodPost,
			urlTo:  "/api/shorten",
			urlSrc: "http://mysite.loc/",
			userId: "qwerty1",
			want: want{
				contentType: "application/json",
				status:      http.StatusCreated,
			},
		},
	}

	cfg, err := config.New()
	require.NoError(t, err)
	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	for _, tc := range cases {
		req := models.ShortenReq{
			Url: tc.urlSrc,
		}

		reqJson, err := json.Marshal(req)
		assert.NoError(t, err)

		r := httptest.NewRequest(tc.method, tc.urlTo, bytes.NewBuffer(reqJson))
		w := httptest.NewRecorder()
		store, err := storage.New(storage.MapDriver, cfg, logger)
		require.NoError(t, err)

		h := prepateHandler(store)

		repo := repository.New(store)
		event := audit.NewAuditEvent()
		service := service.New(repo, logger, event)

		ctx := service.SetUserIdToCtx(r.Context(), tc.userId)
		ctx = service.SetIsSignValidToCtx(ctx, true)

		r = r.WithContext(ctx)

		h.GetShortenUrl(w, r)

		res := w.Result()

		t.Run(tc.name, func(t *testing.T) {
			if tc.want.contentType != "" {
				assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))

				body, err := io.ReadAll(res.Body)
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("Error closing: %s", err.Error())
					}
				}()

				require.NoError(t, err)
				assert.NotEmpty(t, body)

				resp := models.ShortenResp{}

				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Result)
			}

			if tc.want.status != 0 {
				assert.Equal(t, tc.want.status, res.StatusCode)
			}
		})
	}
}

func BenchmarkGenerateId(b *testing.B) {
	cfg, err := config.New()

	if err != nil {
		b.Fatal(err)
	}

	logger, err := logger.InitZap("info")

	if err != nil {
		b.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("http://mysite.loc/")))
	w := httptest.NewRecorder()
	store, err := storage.New(storage.MapDriver, cfg, logger)

	if err != nil {
		b.Fatal(err)
	}

	h := prepateHandler(store)

	repo := repository.New(store)
	event := audit.NewAuditEvent()
	service := service.New(repo, logger, event)

	ctx := service.SetUserIdToCtx(r.Context(), "qwerty999")
	ctx = service.SetIsSignValidToCtx(ctx, true)

	r = r.WithContext(ctx)

	// Сбрасываем таймер
	b.ResetTimer()

	b.Run("GenerateId", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			h.GenerateId(w, r)
			res := w.Result()

			if res.StatusCode != http.StatusCreated {
				b.Fatal("created status fail")
			}
		}
	})
}

func TestGenerateId(t *testing.T) {
	type want struct {
		contentType string
		status      int
	}

	cases := []struct {
		name   string
		method string
		urlTo  string
		urlSrc string
		userId string
		want   want
	}{
		{
			name:   "Case #1 Positive",
			method: http.MethodPost,
			urlTo:  "/",
			urlSrc: "http://mysite.loc/",
			userId: "qwerty6",
			want: want{
				contentType: "plain/text",
				status:      http.StatusCreated,
			},
		},
	}

	cfg, err := config.New()
	require.NoError(t, err)

	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	for _, tc := range cases {
		r := httptest.NewRequest(tc.method, tc.urlTo, bytes.NewBuffer([]byte(tc.urlSrc)))
		w := httptest.NewRecorder()
		store, err := storage.New(storage.MapDriver, cfg, logger)
		require.NoError(t, err)

		h := prepateHandler(store)

		repo := repository.New(store)
		event := audit.NewAuditEvent()
		service := service.New(repo, logger, event)

		ctx := service.SetUserIdToCtx(r.Context(), tc.userId)
		ctx = service.SetIsSignValidToCtx(ctx, true)

		r = r.WithContext(ctx)

		h.GenerateId(w, r)

		res := w.Result()

		t.Run(tc.name, func(t *testing.T) {
			if tc.want.contentType != "" {
				assert.Equal(t, "plain/text", res.Header.Get("Content-Type"))

				body, err := io.ReadAll(res.Body)
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("Error closing: %s", err.Error())
					}
				}()

				require.NoError(t, err)
				assert.NotEmpty(t, body)
			}

			if tc.want.status != 0 {
				assert.Equal(t, tc.want.status, res.StatusCode)
			}
		})
	}
}

func TestUrls(t *testing.T) {
	type want struct {
		contentType string
		status      int
	}

	cases := []struct {
		name   string
		method string
		userId string
		want   want
	}{
		{
			name:   "Case #1 Method return cookie",
			method: http.MethodGet,
			userId: "qwerty4",
			want: want{
				status: http.StatusNoContent,
			},
		},
	}

	cfg, err := config.New()
	require.NoError(t, err)

	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	for _, tc := range cases {
		store, err := storage.New(storage.MapDriver, cfg, logger)
		require.NoError(t, err)

		r := httptest.NewRequest(tc.method, "/api/user/urls", nil)
		w := httptest.NewRecorder()

		h := prepateHandler(store)

		repo := repository.New(store)
		event := audit.NewAuditEvent()
		service := service.New(repo, logger, event)

		ctx := service.SetUserIdToCtx(r.Context(), tc.userId)
		ctx = service.SetIsSignValidToCtx(ctx, true)

		r = r.WithContext(ctx)

		h.Urls(w, r)

		t.Run(tc.name, func(t *testing.T) {
			resp := w.Result()

			if tc.want.contentType != "" {
				assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
			}

			assert.Equal(t, tc.want.status, resp.StatusCode)
		})
	}
}

func TestGetUrl(t *testing.T) {
	type want struct {
		contentType string
		status      int
		url         string
	}

	cases := []struct {
		name   string
		method string
		urlSrc string
		id     string
		userId string
		want   want
	}{
		{
			name:   "Case #1 Positive",
			method: http.MethodGet,
			urlSrc: "http://mysite.loc/",
			id:     "QdYD7GY5",
			userId: "qwerty777",
			want: want{
				contentType: "plain/text",
				status:      http.StatusTemporaryRedirect,
				url:         "http://mysite.loc/",
			},
		},
		{
			name:   "Case #2 Not Found",
			method: http.MethodGet,
			userId: "qwerty778",
			want: want{
				status: http.StatusNotFound,
			},
		},
	}

	cfg, err := config.New()
	require.NoError(t, err)

	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	for _, tc := range cases {
		store, err := storage.New(storage.MapDriver, cfg, logger)
		require.NoError(t, err)

		if tc.urlSrc != "" {
			m := models.InsertData{
				Key:   tc.id,
				Value: tc.urlSrc,
			}

			require.NoError(t, err)

			err = store.CreateUrl(context.Background(), m)
			require.NoError(t, err)
		}

		h := prepateHandler(store)

		t.Run(tc.name, func(t *testing.T) {
			// использую результат короткого отправляю запрос на GET
			r := httptest.NewRequest(tc.method, fmt.Sprintf("/%s", tc.id), nil)
			r.SetPathValue("id", tc.id)

			repo := repository.New(store)
			event := audit.NewAuditEvent()
			service := service.New(repo, logger, event)

			ctx := service.SetUserIdToCtx(r.Context(), tc.userId)
			ctx = service.SetIsSignValidToCtx(ctx, true)

			r = r.WithContext(ctx)

			w := httptest.NewRecorder()

			h.GetUrl(w, r)
			respGet := w.Result()
			loc := respGet.Header.Get("Location")

			if tc.want.contentType != "" {
				assert.Equal(t, tc.want.contentType, respGet.Header.Get("Content-Type"))
			}

			if tc.want.url != "" {
				assert.NotEmpty(t, loc)
				assert.Equal(t, tc.want.url, loc)
			}

			assert.Equal(t, tc.want.status, respGet.StatusCode)
		})
	}
}

func BenchmarkGetUrls(b *testing.B) {
	cfg, err := config.New()

	if err != nil {
		b.Fatal(err)
	}

	logger, err := logger.InitZap("debug")

	if err != nil {
		b.Fatal(err)
	}

	store, err := storage.New(storage.MapDriver, cfg, logger)

	if err != nil {
		b.Fatal(err)
	}

	type urlsSrc struct {
		CorrelationId string `json:"correlation_id"`
		OriginalUrl   string `json:"original_url"`
	}

	urls := []urlsSrc{
		{
			CorrelationId: "abc1",
			OriginalUrl:   "http://mysite.loc",
		},
		{
			CorrelationId: "abc2",
			OriginalUrl:   "http://mysite2.loc",
		},
	}

	h := prepateHandler(store)

	body, err := json.Marshal(urls)

	if err != nil {
		b.Fatal(err)
	}

	repo := repository.New(store)
	event := audit.NewAuditEvent()
	service := service.New(repo, logger, event)

	// Сбрасываем таймер
	b.ResetTimer()

	b.Run("GetUrlsBatch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(body))
			ctx := service.SetUserIdToCtx(r.Context(), "qwerty12123123")
			ctx = service.SetIsSignValidToCtx(ctx, true)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			h.GetShortenUrls(w, r)
			res := w.Result()

			if res.StatusCode != http.StatusCreated {
				b.Fatal("created status fail")
			}
		}
	})
}

func TestGetUrls(t *testing.T) {
	type urlsSrc struct {
		CorrelationId string `json:"correlation_id"`
		OriginalUrl   string `json:"original_url"`
	}

	type want struct {
		contentType string
		status      int
		ids         []string
	}

	cases := []struct {
		name   string
		method string
		urlSrc []urlsSrc
		userId string
		want   want
	}{
		{
			name:   "Case #1 Positive",
			method: http.MethodPost,
			urlSrc: []urlsSrc{
				{
					CorrelationId: "abc1",
					OriginalUrl:   "http://mysite.loc",
				},
				{
					CorrelationId: "abc2",
					OriginalUrl:   "http://mysite2.loc",
				},
			},
			userId: "qwerty12",
			want: want{
				contentType: "application/json",
				status:      http.StatusCreated,
				ids:         []string{"abc1", "abc2"},
			},
		},
	}

	cfg, err := config.New()
	require.NoError(t, err)

	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	for _, tc := range cases {
		store, err := storage.New(storage.MapDriver, cfg, logger)
		require.NoError(t, err)

		h := prepateHandler(store)

		body, err := json.Marshal(tc.urlSrc)
		require.NoError(t, err)

		r := httptest.NewRequest(tc.method, "/api/shorten/batch", bytes.NewBuffer(body))
		repo := repository.New(store)
		event := audit.NewAuditEvent()
		service := service.New(repo, logger, event)

		ctx := service.SetUserIdToCtx(r.Context(), tc.userId)
		ctx = service.SetIsSignValidToCtx(ctx, true)
		r = r.WithContext(ctx)

		w := httptest.NewRecorder()

		h.GetShortenUrls(w, r)

		res := w.Result()

		t.Run(tc.name, func(t *testing.T) {
			if tc.want.contentType != "" {
				assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))

				body, err := io.ReadAll(res.Body)
				defer func() {
					if err := res.Body.Close(); err != nil {
						log.Printf("Error closing: %s", err.Error())
					}
				}()

				require.NoError(t, err)
				assert.NotEmpty(t, body)

				items := []map[string]string{}
				err = json.Unmarshal(body, &items)

				t.Log(items)

				require.NoError(t, err)

				actualIds := []string{}

				for _, item := range items {
					corId, ok := item["correlation_id"]
					assert.True(t, ok)
					actualIds = append(actualIds, corId)
				}

				assert.Equal(t, tc.want.ids, actualIds)
			}

			if tc.want.status != 0 {
				assert.Equal(t, tc.want.status, res.StatusCode)
			}
		})
	}
}
