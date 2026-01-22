package server

import "github.com/go-chi/chi/v5"

func (app *Application) Routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(app.Middleware.RateLimit)

	// user management endpoints (public)
	r.Post("/api/auth/register", app.UserHandler.HandleRegister)
	r.Post("/api/auth/login", app.UserHandler.HandleLogin)

	r.Group(func(r chi.Router) {
		r.Use(app.Middleware.Authenticate)
		
		// user management
		r.Get("/api/auth/me", app.UserHandler.HandleMe)
		r.Patch("/api/auth/me", app.UserHandler.HandleUpdate)
		r.Get("/api/verify-email", app.UserHandler.HandleEmailVerification)
	})

	return r
}
