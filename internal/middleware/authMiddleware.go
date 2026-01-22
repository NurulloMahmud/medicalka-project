package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/NurulloMahmud/medicalka-project/internal/auth"
	"github.com/NurulloMahmud/medicalka-project/utils"
)

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			r = utils.SetUser(r, utils.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			utils.Unauthorized(w, r, "invalid authorization header")
			return
		}

		token := headerParts[1]
		claims, err := auth.VerifyToken(string(token), m.cfg.JWTSecret)
		if err != nil {
			m.logger.Printf(token)
			m.logger.Printf("error -> ", err.Error())
			utils.Unauthorized(w, r, "invalid token")
			return
		}

		user, err := m.userRepo.Get(r.Context(), claims.ID, claims.Username, claims.Email)
		if err != nil {
			utils.InternalServerError(w, r, err, m.logger)
			return
		}

		if user == nil {
			utils.BadRequest(w, r, errors.New("user not found"), m.logger)
			return
		}

		contextUser := utils.User{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Username: user.Username,
		}

		r = utils.SetUser(r, &contextUser)
		next.ServeHTTP(w, r)
	})
}
