package logger

import (
	"go.uber.org/zap"
)

func InitZap(lvl string) (*zap.SugaredLogger, error) {

	level, err := zap.ParseAtomicLevel(lvl)

	if err != nil {
		return nil, err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = level

	logger, err := cfg.Build()

	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
