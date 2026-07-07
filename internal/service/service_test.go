package service

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/logger"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/storage"
)

func TestGenerateId(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	store, err := storage.New(storage.MapDriver, cfg, logger)
	require.NoError(t, err)

	repo := repository.New(store)
	event := audit.NewAuditEvent(logger)
	service := New(repo, logger, event)

	val := service.GenerateId()

	dec, err := base64.URLEncoding.DecodeString(val)

	require.NoError(t, err)

	assert.Len(t, dec, 6)
}

func ExampleService_GenerateId() {
	cfg, _ := config.New()

	logger, _ := logger.InitZap("debug")

	store, _ := storage.New(storage.MapDriver, cfg, logger)

	repo := repository.New(store)
	event := audit.NewAuditEvent(logger)
	service := New(repo, logger, event)

	val := service.GenerateId()

	fmt.Printf("Уникальный идентификатор: %v\n", val)
}

func BenchmarkGenerateId(b *testing.B) {
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

	repo := repository.New(store)
	event := audit.NewAuditEvent(logger)
	service := New(repo, logger, event)

	// Сбрасываем таймер
	b.ResetTimer()

	b.Run("GenerateShortURLID", func(b *testing.B) {
		for b.Loop() {
			service.GenerateId()
		}
	})
}
