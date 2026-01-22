package post

import (
	"errors"
	"strings"
	"time"

	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/google/uuid"
)

var (
	errTitleLength       = errors.New("title must be between 5 and 255 characters")
	errContentLength     = errors.New("content must not exceed 10000 characters")
	errInvalidDateFormat = errors.New("invalid date format, use RFC3339")
)

type createPostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (r *createPostRequest) Validate() error {
	titleLen := len(strings.TrimSpace(r.Title))
	if titleLen < 5 || titleLen > 255 {
		return errTitleLength
	}

	contentLen := len(strings.TrimSpace(r.Content))
	if contentLen == 0 || contentLen > 10000 {
		return errContentLength
	}

	return nil
}

type getPostsRequest struct {
	DateFrom *time.Time
	DateTo   *time.Time
	utils.Filter
}

func (r *getPostsRequest) Validate() error {
	return r.Filter.Validate()
}

type postResponse struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type postDetailResponse struct {
	ID        uuid.UUID         `json:"id"`
	AuthorID  uuid.UUID         `json:"author_id"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	Comments  []commentResponse `json:"comments"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type commentResponse struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type updatePostRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

func (r *updatePostRequest) Validate() error {
	if r.Title != nil {
		titleLen := len(strings.TrimSpace(*r.Title))
		if titleLen < 5 || titleLen > 255 {
			return errTitleLength
		}
	}

	if r.Content != nil {
		contentLen := len(strings.TrimSpace(*r.Content))
		if contentLen == 0 || contentLen > 10000 {
			return errContentLength
		}
	}

	return nil
}

type getFeedRequest struct {
	utils.Filter
}

func (r *getFeedRequest) Validate() error {
	return r.Filter.Validate()
}

type feedPostResponse struct {
	ID      uuid.UUID   `json:"id"`
	Title   string      `json:"title"`
	Content string      `json:"content"`
	Likes   []uuid.UUID `json:"likes"`
}

type feedUserResponse struct {
	Username string             `json:"username"`
	Posts    []feedPostResponse `json:"posts"`
}
