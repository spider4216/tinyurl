package main

import (
	"fmt"

	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/logger"
	"github.com/spider4216/tinyurl/internal/storage"
	"github.com/spider4216/tinyurl/migrations"
	"go.uber.org/zap"
)

type app struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	store  storage.Storage
}

func newApp() app {
	return app{}
}

func (app *app) Run() error {
	if err := app.initConfig(); err != nil {
		return err
	}

	if err := app.initLogger(); err != nil {
		return err
	}

	if err := app.initStore(); err != nil {
		return err
	}

	if err := app.initMigrations(); err != nil {
		return err
	}

	return nil
}

// Если драйвер postgres, то придется запускать миграции
// из приложения по условиям задания
// Подробюнее: migrations.embed.go
func (app *app) initMigrations() error {
	if app.store.StoreName() == storage.PostgresDriver {
		app.logger.Debug("Up migrations")
		st, ok := app.store.(*storage.PgxStorage)

		if !ok {
			return fmt.Errorf("cannot cast to pgx store type in init migration")
		}

		if err := migrations.Run(st.Con); err != nil {
			return err
		}
	}

	return nil
}

func (app *app) initStore() error {
	store, err := storage.New(app.cfg.StoreDriver, app.cfg, app.logger)
	if err != nil {
		return err
	}

	app.store = store

	return nil
}

func (app *app) initLogger() error {
	logger, err := logger.InitZap(app.cfg.LogLvl)
	if err != nil {
		return err
	}

	app.logger = logger

	return nil
}

func (app *app) initConfig() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	flags := NewFlags()

	if err := flags.Init(); err != nil {
		return err
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

	app.cfg = cfg

	return nil
}
