package comment

import (
	"errors"
	"strings"
)

var (
	errContentLength   = errors.New("comment content must be between 1 and 2000 characters")
	errPostNotFound    = errors.New("post not found")
	errNotVerified     = errors.New("email verification required to create comments")
	errCommentNotFound = errors.New("comment not found")
	errNotAuthor       = errors.New("you can only delete your own comments")
)

type createCommentRequest struct {
	Content string `json:"content"`
}

func (r *createCommentRequest) Validate() error {
	contentLen := len(strings.TrimSpace(r.Content))
	if contentLen < 1 || contentLen > 2000 {
		return errContentLength
	}

	return nil
}
