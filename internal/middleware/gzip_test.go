package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGzip(t *testing.T) {
	zap, _ := zap.NewDevelopment()
	logger := zap.Sugar()

	m := New(logger, &config.Config{})

	handler := m.Gzip(http.HandlerFunc(fakeHandler))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	reqBody := `{"version":"1.0"}`
	successBody := strings.Repeat("Hello World", 20)

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zp := gzip.NewWriter(buf)
		_, err := zp.Write([]byte(reqBody))
		require.NoError(t, err)
		err = zp.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")
		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Log("cannot close body", err)
			}
		}()

		b, err := io.ReadAll(resp.Body)

		require.NoError(t, err)
		require.Equal(t, successBody, string(b))
	})

	t.Run("accept_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(reqBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Log("cannot close body")
			}
		}()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, successBody, string(b))
	})
}

func fakeHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(strings.Repeat("Hello World", 20))); err != nil {
		log.Println("cannot write in handler")
	}
}
