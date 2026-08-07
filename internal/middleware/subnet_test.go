package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
	"github.com/spider4216/tinyurl/internal/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWithSubnet(t *testing.T) {
	zap, _ := zap.NewDevelopment()
	logger := zap.Sugar()
	cfg := &config.Config{
		TrustSubnet: "192.168.1.0/24",
	}
	store, err := storage.New(storage.MapDriver, cfg, logger)
	require.NoError(t, err)
	repo := repository.New(store)
	event := audit.NewAuditEvent(logger)
	service := service.New(repo, logger, event)

	m := New(logger, cfg, service)

	handler := m.WithSubnet(http.HandlerFunc(fakeHandler))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	reqBody := `{"version":"1.0"}`
	successBody := strings.Repeat("Hello World", 20)

	t.Run("subnet_ip_correct", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		buf.Write([]byte(reqBody))

		r := httptest.NewRequest("GET", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("X-Real-IP", "192.168.1.98")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Log("cannot close body", closeErr)
			}
		}()

		b, err := io.ReadAll(resp.Body)

		require.NoError(t, err)
		require.Equal(t, successBody, string(b))
	})

	t.Run("subnet_ip_not correct", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		buf.Write([]byte(reqBody))

		r := httptest.NewRequest("GET", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("X-Real-IP", "192.168.2.98")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				t.Log("cannot close body", closeErr)
			}
		}()
	})
}
