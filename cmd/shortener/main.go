package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/middleware"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"

	_ "net/http/pprof"
)

const (
	serverTimeout time.Duration = 5 * time.Second
)

func main() {
	app := newApp()

	if err := app.Run(); err != nil {
		log.Fatal("Cannot run app", err)
	}

	app.logger.Debug("Config: ", app.cfg)

	repo := repository.New(app.store)
	service := service.New(repo, app.logger, app.audit)
	handler := handler.New(app.cfg, app.logger, service, app.reqPools)
	middlewares := middleware.New(app.logger, app.cfg, service)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middlewares.WithLogging)
		r.Use(middlewares.Gzip)
		r.Use(middlewares.WithAuth)

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

	fmt.Printf("\nBuild version: %s\n", app.buildVersion)
	fmt.Printf("Build date: %s\n", app.buildDate)
	fmt.Printf("Build commit: %s\n\n", app.buildCommit)

	app.logger.Infof("Listen profile on: %s", app.cfg.ProfileHost)

	srvProfile := &http.Server{
		Addr: app.cfg.ProfileHost,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	go func() {
		defer wg.Done()

		app.logger.Debug("Graceful shutdown mode on")
		<-ctx.Done()

		// Тут тоже останавляваем перехват сигналов
		stop()

		app.logger.Debug("Shutdown all servers...")

		ctxShutdown, cancel := context.WithTimeout(context.Background(), serverTimeout)
		defer cancel()

		if err := srvProfile.Shutdown(ctxShutdown); err != nil {
			app.logger.Warnf("Cannot shutdown profile server: %s", err)
		}

		if err := srv.Shutdown(ctxShutdown); err != nil {
			app.logger.Warnf("Cannot shutdown main server: %s", err)
		}
	}()

	// Run profile server
	go func() {
		if err := srvProfile.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Fatalf("Profile server error: %s", err)
		}
	}()

	app.logger.Infof("Listen on: %s", app.cfg.ServerAddress)

	if err := runServer(srv, app.cfg, app.logger); err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.logger.Fatalf("Server error: %s", err)
	}

	wg.Wait()
}

func runServer(srv *http.Server, cfg *config.Config, logger *zap.SugaredLogger) error {
	if cfg.Https {
		logger.Info("Run HTTPS mode")
		return srv.ListenAndServeTLS(cfg.CrtPath, cfg.PKPath)
	}

	logger.Info("Run HTTP mode")
	return srv.ListenAndServe()
}
