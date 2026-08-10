package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/spider4216/tinyurl/internal/config"
	mygrpc "github.com/spider4216/tinyurl/internal/grpc"
	"github.com/spider4216/tinyurl/internal/handler"
	"github.com/spider4216/tinyurl/internal/middleware"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
	pb "github.com/spider4216/tinyurl/proto"

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

	pretyCfg, err := json.MarshalIndent(app.cfg, "", " ")
	if err != nil {
		app.logger.Fatalf("Error while run app: %s", err)
	}

	app.logger.Debug("Config: ", string(pretyCfg))

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

		r.Group(func(r chi.Router) {
			r.Use(middlewares.WithSubnet)

			r.Get("/api/internal/stats", http.HandlerFunc(handler.Stat))
		})
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

	grpcSrv := grpc.NewServer()

	// Если GRPC в конфигурации указан, то запускаем сервер
	if app.cfg.GRPCHost != "" {
		reflection.Register(grpcSrv)
		myGrpcSrv := mygrpc.New(app.cfg, service, app.logger)
		go runGRPC(app.cfg.GRPCHost, myGrpcSrv, grpcSrv, app.logger)
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

		if app.cfg.GRPCHost != "" {
			grpcSrv.GracefulStop()
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

func runGRPC(host string, srv *mygrpc.ShortenerServer, s *grpc.Server, logger *zap.SugaredLogger) error {
	listen, err := net.Listen("tcp", host)

	if err != nil {
		return err
	}

	pb.RegisterShortenerServiceServer(s, srv)

	logger.Infof("Listen GRPC on port %s", host)

	return s.Serve(listen)
}
