package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/middleware"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
)

func main() {
	app := newApp()

	if err := app.Run(); err != nil {
		log.Fatal("Cannot run app", err)
	}

	app.logger.Debug("Config: ", app.cfg)

	repo := repository.New(app.store)
	service := service.New(repo, app.logger)
	handler := handler.New(app.cfg, app.logger, service)
	middlewares := middleware.New(app.logger)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.WithLogging)
		r.Use(middlewares.Gzip)

		r.Post("/", http.HandlerFunc(handler.GenerateId))
		r.Get("/{id}", http.HandlerFunc(handler.GetUrl))
		r.Post("/api/shorten", http.HandlerFunc(handler.GetShortenUrl))
		r.Get("/ping", http.HandlerFunc(handler.Ping))
		r.Post("/api/shorten/batch", http.HandlerFunc(handler.GetShortenUrls))
		r.Get("/api/user/urls", http.HandlerFunc(handler.Urls))
		r.Delete("/api/user/urls", http.HandlerFunc(handler.DeleteUrls))
	})

	srv := &http.Server{
		Addr:         app.cfg.ServerAddress,
		Handler:      r,
		ReadTimeout:  app.cfg.ReadTimeout,
		WriteTimeout: app.cfg.WriteTimeout,
		IdleTimeout:  app.cfg.IdleTimeout,
	}

	log.Printf("Listen on: %s", app.cfg.ServerAddress)

	if err := srv.ListenAndServe(); err != nil {
		app.logger.Fatalf("Server error: %s", err)
	}

	app.logger.Infof("Starting server on %s", app.cfg.ServerAddress)
}
