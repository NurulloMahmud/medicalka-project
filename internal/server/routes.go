package server

import "github.com/go-chi/chi/v5"

func (app *Application) Routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(app.Middleware.RateLimit)

	// user management endpoints (public)
	r.Post("/api/auth/register", app.UserHandler.HandleRegister)
	r.Post("/api/auth/login", app.UserHandler.HandleLogin)

	// unprotected post endpoints
	r.Get("/api/posts", app.PostHandler.HandleGetAll)
	r.Get("/api/posts/{id}", app.PostHandler.HandleGetByID)
	r.Get("/api/feed", app.PostHandler.HandleGetFeed)

	r.Group(func(r chi.Router) {
		r.Use(app.Middleware.Authenticate)

		// user management
		r.Get("/api/auth/me", app.UserHandler.HandleMe)
		r.Patch("/api/auth/me", app.UserHandler.HandleUpdate)
		r.Get("/api/verify-email", app.UserHandler.HandleEmailVerification)

		// posts
		r.Post("/api/posts", app.PostHandler.HandleCreate)
		r.Patch("/api/posts/{id}", app.PostHandler.HandleUpdate)
		r.Delete("/api/posts/{id}", app.PostHandler.HandleDelete)

		// posts
		r.Post("/api/posts/{id}/comments", app.CommentHandler.HandleCreate)
		r.Delete("/api/posts/{postID}/comments/{commentID}", app.CommentHandler.HandleDelete)

		// likes
		r.Post("/api/posts/{id}/like", app.LikeHandler.HandleCreate)
		r.Delete("/api/posts/{id}/like", app.LikeHandler.HandleDelete)
	})

	return r
}
