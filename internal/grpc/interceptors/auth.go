package interceptors

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authKey string = "authorization"
)

func (i *Interceptor) WithAuth(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if i.cfg.SignKey == "" {
		msg := "cannot find sign key. Set and try again."

		i.logger.Error(msg)
		return nil, status.Error(codes.Unauthenticated, msg)
	}

	var userId string
	var sign string

	// Извлекаем meta-данные из контекста
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		msg := "something went wrong while getting md"

		i.logger.Error(msg)
		return nil, status.Error(codes.Unauthenticated, msg)
	}

	values := md.Get(authKey)

	// Если токен не передан
	if len(values) == 0 {
		// Генерируем User ID
		userId = uuid.NewString()

		// Подписываем новый идентификатор
		sign, err := i.service.SignVal(userId, i.cfg.SignKey)

		if err != nil {
			msg := "cannot sign cookie"
			i.logger.Error(msg, zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, msg)
		}

		i.logger.Debug("Set user_id and sign validate result to ctx ", userId, true)

		// Устанавливаем user ID в контекст
		ctx = i.service.SetUserIdToCtx(ctx, userId)
		ctx = i.service.SetIsSignValidToCtx(ctx, true)

		// Устанавливаем ключ в metadata
		err = grpc.SetHeader(ctx, metadata.Pairs(
			authKey, userId+"."+sign,
		))

		if err != nil {
			msg := "cannot set sign to metadata header"
			i.logger.Error(msg, zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, msg)
		}

		return handler(ctx, req)

	}

	// Токен был передан
	i.logger.Debug("Token found")

	sig := values[0]

	// Получаем отдельно ID и отдельно подпись
	parts := strings.Split(sig, ".")

	if len(parts) < 2 {
		msg := "bad sign"
		i.logger.Error(msg)
		return nil, status.Error(codes.Unauthenticated, msg)
	}

	userId = parts[0]
	sign = parts[1]

	i.logger.Debug("Validate sign...")

	isValid := false

	// Валидируем токен
	if err := i.service.ValidateSign(userId, i.cfg.SignKey, sign); err == nil {
		isValid = true
	}

	i.logger.Debug("Sign validation result ", isValid)

	// Устанавливаем user ID в контекст
	ctx = i.service.SetUserIdToCtx(ctx, userId)
	// Поскольку по требованию инкрементов, некоторые эндпоинты должны
	// пропускать невалидность, устанавливаем флаг валидности куки
	// в контекст и делегируем поведение невалидности обработчикам
	ctx = i.service.SetIsSignValidToCtx(ctx, isValid)

	return handler(ctx, req)
}
