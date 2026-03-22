package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepateHandler() Handler {
	r := repository.New()
	s := service.New(r)

	return New(s)
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
		h := prepateHandler()

		h.GenerateId(w, r)

		res := w.Result()

		t.Run(tc.name, func(t *testing.T) {
			if tc.want.contentType != "" {
				assert.Equal(t, "plain/text", res.Header.Get("Content-Type"))

				body, err := io.ReadAll(res.Body)
				defer res.Body.Close()

				require.NoError(t, err)
				assert.NotEmpty(t, body)
			}

			if tc.want.status != 0 {
				assert.Equal(t, tc.want.status, res.StatusCode)
			}
		})
	}
}
