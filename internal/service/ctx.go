package service

import "context"

type (
	UserIdKey      string
	IsSignValidKey string
)

const (
	userKey      UserIdKey      = "user_id"
	validSignKey IsSignValidKey = "is_sign_valid"
)

func (s Service) SetUserIdToCtx(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, userKey, userId)
}

func (s Service) SetIsSignValidToCtx(ctx context.Context, isValid bool) context.Context {
	return context.WithValue(ctx, validSignKey, isValid)
}

func (s Service) GetUserIdFromCtx(ctx context.Context) string {
	userId, ok := ctx.Value(userKey).(string)

	if !ok {
		return ""
	}

	return userId
}

func (s Service) IsSignValidFromCtx(ctx context.Context) bool {
	isValid, ok := ctx.Value(validSignKey).(bool)

	if !ok {
		return false
	}

	return isValid
}
