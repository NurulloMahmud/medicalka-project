package server

import (
	"database/sql"
	"log"
	"os"

	"github.com/NurulloMahmud/medicalka-project/config"
	"github.com/NurulloMahmud/medicalka-project/internal/comment"
	"github.com/NurulloMahmud/medicalka-project/internal/like"
	"github.com/NurulloMahmud/medicalka-project/internal/middleware"
	"github.com/NurulloMahmud/medicalka-project/internal/platform/database"
	"github.com/NurulloMahmud/medicalka-project/internal/post"
	"github.com/NurulloMahmud/medicalka-project/internal/tasks"
	"github.com/NurulloMahmud/medicalka-project/internal/user"
	"github.com/NurulloMahmud/medicalka-project/migrations"
)

type Application struct {
	Logger         *log.Logger
	DB             *sql.DB
	Cfg            config.Config
	UserHandler    user.UserHandler
	PostHandler    post.PostHandler
	CommentHandler comment.CommentHandler
	LikeHandler    like.LikeHandler
	Middleware     middleware.Middleware
}

func NewApplication(cfg config.Config) (*Application, error) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	pgDB, err := database.New(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	err = database.Migrate(pgDB, migrations.FS, ".")
	if err != nil {
		return nil, err
	}

	// email sender
	emailSender := tasks.NewEmailSender(cfg.SMTP, logger)

	// repositories
	userPostgresRepo := user.NewPostgresRepository(pgDB)
	postRepo := post.NewPostgresRepository(pgDB)
	commentPostgresRepo := comment.NewPostgresRepository(pgDB)
	likePostgresRepo := like.NewPostgresRepository(pgDB)

	// services
	userService := user.NewService(userPostgresRepo, emailSender)
	postService := post.NewService(postRepo)
	commentService := comment.NewService(commentPostgresRepo)
	likeService := like.NewService(likePostgresRepo)

	// handlers
	userHandler := user.NewHandler(userService, logger, cfg)
	postHandler := post.NewHandler(postService, logger)
	commentHandler := comment.NewHandler(commentService, logger)
	likeHandler := like.NewHandler(likeService, logger)

	// middlewares
	middleware := middleware.NewMiddleware(logger, userPostgresRepo, cfg)

	app := &Application{
		Logger:         logger,
		DB:             pgDB,
		Cfg:            cfg,
		UserHandler:    *userHandler,
		Middleware:     *middleware,
		LikeHandler:    *likeHandler,
		PostHandler:    *postHandler,
		CommentHandler: *commentHandler,
	}

	return app, nil
}
