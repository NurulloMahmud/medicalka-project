package like

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository interface {
	Create(ctx context.Context, l Like) (*Like, error)
	Delete(ctx context.Context, userID, postID uuid.UUID) error
	Exists(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	GetPostAuthorID(ctx context.Context, postID uuid.UUID) (*uuid.UUID, error)
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, l Like) (*Like, error) {
	query := `
	INSERT INTO likes (user_id, post_id)
	VALUES ($1, $2)
	RETURNING created_at`

	err := r.db.QueryRowContext(ctx, query, l.UserID, l.PostID).Scan(&l.CreatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errAlreadyLiked
		}
		return nil, err
	}

	return &l, nil
}

func (r *postgresRepo) Delete(ctx context.Context, userID, postID uuid.UUID) error {
	query := `DELETE FROM likes WHERE user_id = $1 AND post_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, postID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errLikeNotFound
	}

	return nil
}

func (r *postgresRepo) Exists(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND post_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, postID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *postgresRepo) GetPostAuthorID(ctx context.Context, postID uuid.UUID) (*uuid.UUID, error) {
	query := `SELECT author_id FROM posts WHERE id = $1`

	var authorID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &authorID, nil
}
