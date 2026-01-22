package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, u User, token string) (*User, error)
	UserExists(ctx context.Context, username, email string) (bool, error)
	Get(ctx context.Context, id uuid.UUID, username, email string) (*User, error)
	VerifyToken(ctx context.Context, token string) (*uuid.UUID, error)
	Update(ctx context.Context, user *User, token string) error
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, u User, token string) (*User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO users (email, username, full_name, password_hash)
	VALUES ($1, $2, $3, $4)
	RETURNING id, is_verified, created_at, updated_at`

	err = tx.QueryRowContext(
		ctx, query, u.Email, u.Username, u.FullName, u.Password.hash,
	).Scan(
		&u.ID, &u.IsVerified, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	tokenQuery := `
	INSERT INTO email_verification_tokens(user_id, token, expires_at)
	VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, tokenQuery, u.ID, token, time.Now().Add(24*time.Hour))

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *postgresRepo) UserExists(ctx context.Context, username, email string) (bool, error) {
	var exists bool
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM users
		WHERE username = $1 OR email = $2
	)`

	err := r.db.QueryRowContext(ctx, query, username, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *postgresRepo) Get(ctx context.Context, id uuid.UUID, username, email string) (*User, error) {
	var user User
	query := `
	SELECT id, email, username, full_name, password_hash, is_verified, created_at, updated_at
	FROM users
	WHERE id = $1 OR username = $2 OR email = $3`

	err := r.db.QueryRowContext(
		ctx, query,
		id, username, email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FullName,
		&user.Password.hash,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresRepo) Update(ctx context.Context, user *User, token string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	UPDATE users
	SET 
		username = $1, 
		email = $2, 
		full_name = $3, 
		password_hash = $4,
		is_verified = $5,
		updated_at = NOW()
	WHERE id = $6`

	_, err = tx.ExecContext(
		ctx, query,
		user.Username,
		user.Email,
		user.FullName,
		user.Password.hash,
		user.IsVerified,
	)

	if err != nil {
		return err
	}

	// agar token bosh bolmasa demak email update qilingan -> verify qilish kk
	if token != "" {
		tokenQuery := `
		INSERT INTO email_verification_tokens(user_id, token, expires_at)
		VALUES ($1, $2, $3)`

		_, err = tx.ExecContext(ctx, tokenQuery, user.ID, token, time.Now().Add(24*time.Hour))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresRepo) VerifyToken(ctx context.Context, token string) (*uuid.UUID, error) {
	var userID uuid.UUID
	query := `
	SELECT user_id
	FROM email_verification_tokens
	WHERE token = $1 AND expires_at > NOW()`

	err := r.db.QueryRowContext(ctx, query, token).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &userID, nil
}
