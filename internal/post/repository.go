package post

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, p Post) (*Post, error)
	GetAll(ctx context.Context, req getPostsRequest) ([]Post, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*postDetailResponse, error)
	GetByIDSimple(ctx context.Context, id uuid.UUID) (*Post, error)
	Update(ctx context.Context, p *Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetFeed(ctx context.Context, req getFeedRequest) ([]feedUserResponse, int, error)
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, p Post) (*Post, error) {
	query := `
	INSERT INTO posts (author_id, title, content)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query, p.AuthorID, p.Title, p.Content,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *postgresRepo) GetAll(ctx context.Context, req getPostsRequest) ([]Post, int, error) {
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, author_id, title, content, created_at, updated_at
	FROM posts
	WHERE ($1 = '' OR title ILIKE '%%' || $1 || '%%' OR content ILIKE '%%' || $1 || '%%')
	AND ($2::timestamptz IS NULL OR created_at >= $2)
	AND ($3::timestamptz IS NULL OR created_at <= $3)
	ORDER BY %s
	LIMIT $4 OFFSET $5`, req.GetSort())

	rows, err := r.db.QueryContext(
		ctx, query,
		req.Search,
		req.DateFrom,
		req.DateTo,
		req.Limit(),
		req.Offset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []Post
	var total int

	for rows.Next() {
		var p Post
		err := rows.Scan(
			&total,
			&p.ID,
			&p.AuthorID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*postDetailResponse, error) {
	query := `
	SELECT 
		p.id, 
		p.author_id, 
		p.title, 
		p.content, 
		p.created_at, 
		p.updated_at,
		COALESCE(
			(
				SELECT json_agg(
					jsonb_build_object(
						'id', c.id,
						'author_id', c.author_id,
						'content', c.content,
						'created_at', c.created_at
					)
					ORDER BY c.created_at
				)
				FROM comments c
				WHERE c.post_id = p.id
			),
			'[]'
		) AS comments
	FROM posts p
	WHERE p.id = $1`

	var post postDetailResponse
	var commentsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		&commentsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	json.Unmarshal(commentsJSON, &post.Comments)

	return &post, nil
}

func (r *postgresRepo) GetByIDSimple(ctx context.Context, id uuid.UUID) (*Post, error) {
	query := `
	SELECT id, author_id, title, content, created_at, updated_at
	FROM posts
	WHERE id = $1`

	var post Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

func (r *postgresRepo) Update(ctx context.Context, p *Post) error {
	query := `
	UPDATE posts
	SET title = $1, content = $2, updated_at = NOW()
	WHERE id = $3
	RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query, p.Title, p.Content, p.ID).Scan(&p.UpdatedAt)
}

func (r *postgresRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresRepo) GetFeed(ctx context.Context, req getFeedRequest) ([]feedUserResponse, int, error) {
	countQuery := `SELECT COUNT(DISTINCT u.id) FROM users u INNER JOIN posts p ON p.author_id = u.id`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
	SELECT 
		u.id,
		u.username,
		COALESCE(
			(
				SELECT json_agg(
					jsonb_build_object(
						'id', p.id,
						'title', p.title,
						'content', p.content,
						'likes', COALESCE(
							(
								SELECT json_agg(l.user_id)
								FROM likes l
								WHERE l.post_id = p.id
							),
							'[]'
						)
					)
					ORDER BY p.created_at DESC
				)
				FROM posts p
				WHERE p.author_id = u.id
			),
			'[]'
		) AS posts
	FROM users u
	WHERE EXISTS (SELECT 1 FROM posts p WHERE p.author_id = u.id)
	ORDER BY u.username
	LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, req.Limit(), req.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var feed []feedUserResponse

	for rows.Next() {
		var userID uuid.UUID
		var f feedUserResponse
		var postsJSON []byte

		err := rows.Scan(&userID, &f.Username, &postsJSON)
		if err != nil {
			return nil, 0, err
		}

		json.Unmarshal(postsJSON, &f.Posts)
		feed = append(feed, f)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return feed, total, nil
}
