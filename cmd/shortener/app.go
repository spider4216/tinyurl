package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"dario.cat/mergo"
	"go.uber.org/zap"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/config/db"
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
	var cfg config.Config

	cfgEnv, err := config.New()
	if err != nil {
		return err
	}

	flags := NewFlags()

	if errInit := flags.Init(); errInit != nil {
		return errInit
	}

	cfgFlags := makeFlagConfig(flags)

	cfgPath := cfgEnv.CfgFile

	if cfgPath == "" {
		cfgPath = flags.CfgFile
	}

	// Получаем файл конфигурации
	cfgFile, err := app.makeFileConfig(cfgPath)
	if err != nil {
		return err
	}

	if err := mergo.Merge(&cfg, cfgEnv); err != nil {
		return err
	}

	if err := mergo.Merge(&cfg, cfgFlags); err != nil {
		return err
	}

	if err := mergo.Merge(&cfg, cfgFile); err != nil {
		return err
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
	// Для boolean оставляем кастом логику
	cfg.Https = cfgFile.Https

	// Поскольку в конфигурации переменка bool, а у нее значение false по умолчанию
	// Нужно понять была ли передана env
	if value, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
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

	app.cfg = &cfg

	return nil
}

func makeFlagConfig(flag Flags) *config.Config {
	return &config.Config{
		ServerAddress: flag.ServerAddress,
		BaseUrl:       flag.BaseUrl,
		LogLvl:        flag.LogLvl,
		FileStorePath: flag.FileStorePath,
		DbConfig: db.DbConfig{
			DbDsn: flag.DbCon,
		},
		AuditFile:   flag.AuditFile,
		AuditURL:    flag.AuditURL,
		Https:       flag.HttpsSet,
		CfgFile:     flag.CfgFile,
		TrustSubnet: flag.TrustSubnet,
	}
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
