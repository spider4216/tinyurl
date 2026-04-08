package handler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepateHandler(store map[string]string) Handler {
	conf := config.New("", "")
	r := repository.New(store)
	s := service.New(r)

	return New(conf, s)
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
		want   want
	}{
		{
			name:   "Case #1 Positive",
			method: http.MethodPost,
			urlTo:  "/",
			urlSrc: "http://mysite.loc/",
			want: want{
				contentType: "plain/text",
				status:      http.StatusCreated,
			},
		},
		{
			name:   "Case #2 Method Not Allowed",
			method: http.MethodGet,
			urlTo:  "/",
			urlSrc: "http://mysite.loc/",
			want: want{
				status: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tc := range cases {
		r := httptest.NewRequest(tc.method, tc.urlTo, bytes.NewBuffer([]byte(tc.urlSrc)))
		w := httptest.NewRecorder()
		store := map[string]string{}
		h := prepateHandler(store)

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
		want   want
	}{
		{
			name:   "Case #1 Positive",
			method: http.MethodGet,
			urlSrc: "http://mysite.loc/",
			id:     "QdYD7GY5",
			want: want{
				contentType: "plain/text",
				status:      http.StatusTemporaryRedirect,
				url:         "http://mysite.loc/",
			},
		},
		{
			name:   "Case #2 Not Found",
			method: http.MethodGet,
			want: want{
				status: http.StatusNotFound,
			},
		},
	}

	for _, tc := range cases {
		store := map[string]string{}

		if tc.urlSrc != "" {
			store = map[string]string{
				tc.id: tc.urlSrc,
			}
		}

		h := prepateHandler(store)

		t.Run(tc.name, func(t *testing.T) {
			// использую результат короткого отправляю запрос на GET
			r := httptest.NewRequest(tc.method, fmt.Sprintf("/%s", tc.id), nil)
			r.SetPathValue("id", tc.id)
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
