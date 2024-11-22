package context

import (
	"context"

	"tgpt/internal/models"
)

type ContextKeyUserID struct{}

func CtxWithUserID(ctx context.Context, userID models.UserID) context.Context {
	return context.WithValue(ctx, ContextKeyUserID{}, userID)
}

func UserIDFromCtx(ctx context.Context) (models.UserID, bool) {
	v, ok := ctx.Value(ContextKeyUserID{}).(models.UserID)
	if !ok {
		return models.UserID{}, false
	}
	return v, true
}
