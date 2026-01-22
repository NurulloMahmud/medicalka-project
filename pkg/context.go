package pkg

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const userContextKey = contextKey("user")

type User struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	FullName   string    `json:"full_name"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

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
