package comment

import (
	"context"

	"github.com/google/uuid"
)

type CommentService struct {
	repo Repository
}

func NewService(repo Repository) CommentService {
	return CommentService{repo: repo}
}

func (s *CommentService) create(ctx context.Context, postID, authorID uuid.UUID, req createCommentRequest) (*Comment, error) {
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errPostNotFound
	}

	comment := Comment{
		PostID:   postID,
		AuthorID: authorID,
		Content:  req.Content,
	}

	return s.repo.Create(ctx, comment)
}

func (s *CommentService) delete(ctx context.Context, postID, commentID, userID uuid.UUID) error {
	exists, err := s.repo.PostExists(ctx, postID)
	if err != nil {
		return err
	}

	if !exists {
		return errPostNotFound
	}

	comment, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	if comment == nil {
		return errCommentNotFound
	}

	if comment.PostID != postID {
		return errCommentNotFound
	}

	if comment.AuthorID != userID {
		return errNotAuthor
	}

	return s.repo.Delete(ctx, commentID)
}
