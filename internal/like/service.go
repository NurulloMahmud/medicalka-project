package like

import (
	"context"

	"github.com/google/uuid"
)

type LikeService struct {
	repo Repository
}

func NewService(repo Repository) LikeService {
	return LikeService{repo: repo}
}

func (s *LikeService) create(ctx context.Context, postID, userID uuid.UUID) (*Like, error) {
	authorID, err := s.repo.GetPostAuthorID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if authorID == nil {
		return nil, errPostNotFound
	}

	if *authorID == userID {
		return nil, errCannotLikeOwn
	}

	like := Like{
		UserID: userID,
		PostID: postID,
	}

	return s.repo.Create(ctx, like)
}

func (s *LikeService) delete(ctx context.Context, postID, userID uuid.UUID) error {
	authorID, err := s.repo.GetPostAuthorID(ctx, postID)
	if err != nil {
		return err
	}

	if authorID == nil {
		return errPostNotFound
	}

	return s.repo.Delete(ctx, userID, postID)
}
