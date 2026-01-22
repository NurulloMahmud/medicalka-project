package comment

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, c Comment) (*Comment, error)
	PostExists(ctx context.Context, postID uuid.UUID) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, c Comment) (*Comment, error) {
	query := `
	INSERT INTO comments (post_id, author_id, content)
	VALUES ($1, $2, $3)
	RETURNING id, created_at`

	err := r.db.QueryRowContext(
		ctx, query, c.PostID, c.AuthorID, c.Content,
	).Scan(&c.ID, &c.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *postgresRepo) PostExists(ctx context.Context, postID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *postgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*Comment, error) {
	query := `
	SELECT id, post_id, author_id, content, created_at
	FROM comments
	WHERE id = $1`

	var c Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.PostID,
		&c.AuthorID,
		&c.Content,
		&c.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}

func (r *postgresRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM comments WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}