package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/logger"
	"github.com/spider4216/tinyurl/internal/middleware"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
	"github.com/spider4216/tinyurl/internal/storage"
	"github.com/spider4216/tinyurl/migrations"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New()

	if err != nil {
		fmt.Println(err)
		return
	}

	flags := NewFlags()

	if err := flags.Init(); err != nil {
		fmt.Println(err)
		return
	}

	if cfg.BaseUrl == "" {
		cfg.BaseUrl = flags.BaseUrl
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = flags.ServerAddress
	}

	if cfg.LogLvl == "" {
		cfg.LogLvl = flags.LogLvl
	}

	if cfg.FileStorePath == "" {
		cfg.FileStorePath = flags.FileStorePath
	}

	if cfg.DbDsn == "" {
		cfg.DbDsn = flags.DbCon
	}

	// Если DSN установлен, значит дайвер pgx
	if cfg.DbDsn != "" {
		cfg.StoreDriver = storage.PostgresDriver
	} else if cfg.FileStorePath != "" {
		// Если DSN не укакзан, следующий приоритет - это файл
		cfg.StoreDriver = storage.FileDriver
	} else {
		// По умолчанию - мап драйвер (память)
		cfg.StoreDriver = storage.MapDriver
	}

	logger, err := logger.InitZap(cfg.LogLvl)

	if err != nil {
		fmt.Println(err)
		return
	}

	logger.Debug("Config: ", cfg)

	if err != nil {
		logger.Fatal("Error while creating db", zap.Error(err))
	}

	store, err := storage.New(cfg.StoreDriver, cfg)

	if err != nil {
		logger.Fatal("Error while creating store driver", zap.Error(err))
	}

	// Если драйвер postgres, то придется запускать миграции
	// из приложения по условиям задания
	// Подробюнее: migrations.embed.go
	if store.StoreName() == storage.PostgresDriver {
		logger.Debug("Up migrations")
		st := store.(*storage.PgxStorage)
		if err := migrations.Run(st.Con); err != nil {
			logger.Fatal("Migration up error", zap.Error(err))
		}
	}

	repo := repository.New(store)
	service := service.New(repo)
	handler := handler.New(cfg, logger, service)
	middlewares := middleware.New(logger)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.WithLogging)
		r.Use(middlewares.Gzip)

		r.Post("/", http.HandlerFunc(handler.GenerateId))
		r.Get("/{id}", http.HandlerFunc(handler.GetUrl))
		r.Post("/api/shorten", http.HandlerFunc(handler.GetShortenUrl))
		r.Get("/ping", http.HandlerFunc(handler.Ping))
	})

	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("Listen on: %s", cfg.ServerAddress)

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("Server error: %s", err)
	}

	logger.Infof("Starting server on %s", cfg.ServerAddress)
}
