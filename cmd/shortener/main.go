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
)

func main() {
	cfg := config.New()
	flags := NewFlags()
	flags.Init()

	if cfg.BaseUrl == "" {
		cfg.BaseUrl = flags.BaseUrl
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = flags.ServerAddress
	}

	if cfg.LogLvl == "" {
		cfg.LogLvl = flags.LogLvl
	}

	logger, err := logger.InitZap(cfg.LogLvl)

	if err != nil {
		fmt.Println(err)
		return
	}

	store := map[string]string{}
	repo := repository.New(store)
	service := service.New(repo)
	handler := handler.New(cfg, logger, service)
	middlewares := middleware.New(logger)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.WithLogging)

		r.Post("/", http.HandlerFunc(handler.GenerateId))
		r.Get("/{id}", http.HandlerFunc(handler.GetUrl))
		r.Post("/api/shorten", http.HandlerFunc(handler.GetShortenUrl))
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
