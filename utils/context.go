package utils

import (
	"context"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func SetUser(r *http.Request, user *User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func GetUser(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
