package middleware

import "go.uber.org/zap"

type Middleware struct {
	logger *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger) Middleware {
	return Middleware{
		logger: logger,
	}
}
