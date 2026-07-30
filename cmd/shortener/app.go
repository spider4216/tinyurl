package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/logger"
	"github.com/spider4216/tinyurl/internal/models"
	"github.com/spider4216/tinyurl/internal/pool"
	"github.com/spider4216/tinyurl/internal/storage"
	"github.com/spider4216/tinyurl/migrations"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

type app struct {
	buildVersion string
	buildDate    string
	buildCommit  string
	cfg          *config.Config
	logger       *zap.SugaredLogger
	store        storage.Storage
	audit        audit.Publisher
	reqPools     pool.ReqPools
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

	if err := app.initAudit(); err != nil {
		return err
	}

	app.initReqPools()
	app.initAppMeta()

	return nil
}

func (app *app) initAppMeta() {
	app.buildVersion = buildVersion
	app.buildDate = buildDate
	app.buildCommit = buildCommit
}

func (app *app) initReqPools() {
	shReq := pool.New(func() *models.ShortenReq {
		return &models.ShortenReq{}
	})

	app.reqPools = pool.ReqPools{
		ShortenReq: shReq,
	}
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

	if errInit := flags.Init(); errInit != nil {
		return errInit
	}

	if cfg.CfgFile == "" {
		cfg.CfgFile = flags.CfgFile
	}

	// Получаем файл конфигурации
	fcfg, err := app.makeFileConfig(cfg.CfgFile)
	if err != nil {
		return err
	}

	if cfg.BaseUrl == "" {
		cfg.BaseUrl = flags.BaseUrl
	}

	if cfg.BaseUrl == "" {
		cfg.BaseUrl = fcfg.BaseUrl
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = flags.ServerAddress
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = fcfg.ServerAddress
	}

	if cfg.LogLvl == "" {
		cfg.LogLvl = flags.LogLvl
	}

	if cfg.LogLvl == "" {
		cfg.LogLvl = fcfg.LogLvl
	}

	if cfg.FileStorePath == "" {
		cfg.FileStorePath = flags.FileStorePath
	}

	if cfg.FileStorePath == "" {
		cfg.FileStorePath = fcfg.FileStorePath
	}

	if cfg.DbDsn == "" {
		cfg.DbDsn = flags.DbCon
	}

	if cfg.DbDsn == "" {
		cfg.DbDsn = fcfg.DbDsn
	}

	if cfg.AuditFile == "" {
		cfg.AuditFile = flags.AuditFile
	}

	if cfg.AuditFile == "" {
		cfg.AuditFile = fcfg.AuditFile
	}

	if cfg.AuditURL == "" {
		cfg.AuditURL = flags.AuditURL
	}

	if cfg.AuditURL == "" {
		cfg.AuditURL = fcfg.AuditURL
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

	// Значение из файла - наименьший приоритет
	cfg.Https = fcfg.Https

	// Поскольку в конфигурации переменка bool, а у нее значение false по умолчанию
	// Нужно понять была ли передана env
	if value, exists := os.LookupEnv("ENABLE_HTTPS"); exists {
		https, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		cfg.Https = https
	}

	// Если флаг был передан, то учитываем его
	if flags.HttpsSet {
		cfg.Https = flags.Https
	}

	app.cfg = cfg

	return nil
}

func (app *app) makeFileConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &config.Config{}, nil
		}

		return nil, err
	}

	var cfg config.Config

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (app *app) initAudit() error {
	event := audit.NewAuditEvent(app.logger)

	if app.cfg.AuditFile != "" {
		ob, err := audit.NewAuditFileObserver(app.cfg.AuditFile, app.logger)
		if err != nil {
			return err
		}

		event.Register(ob)
	}

	if app.cfg.AuditURL != "" {
		u, err := url.Parse(app.cfg.AuditURL)
		if err != nil {
			return err
		}

		host := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

		ob := audit.NewAuditServerObserver(host, u.Path, app.logger)

		event.Register(ob)
	}

	app.audit = event

	return nil
}
