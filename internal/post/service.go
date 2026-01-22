package post

import (
	"context"
	"errors"

	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/google/uuid"
)

var (
	errPostNotFound = errors.New("post not found")
	errNotAuthor    = errors.New("you can only edit your own posts")
)

type PostService struct {
	repo Repository
}

func NewService(repo Repository) PostService {
	return PostService{repo: repo}
}

func (s *PostService) create(ctx context.Context, req createPostRequest, authorID uuid.UUID) (*Post, error) {
	post := Post{
		AuthorID: authorID,
		Title:    req.Title,
		Content:  req.Content,
	}

	return s.repo.Create(ctx, post)
}

func (s *PostService) getAll(ctx context.Context, req getPostsRequest) ([]Post, utils.Metadata, error) {
	posts, total, err := s.repo.GetAll(ctx, req)
	if err != nil {
		return nil, utils.Metadata{}, err
	}

	metadata := utils.CalculateMetadata(total, req.Page, req.PageSize)
	return posts, metadata, nil
}

func (s *PostService) getByID(ctx context.Context, id uuid.UUID) (*postDetailResponse, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, errPostNotFound
	}

	return post, nil
}

func (s *PostService) update(ctx context.Context, postID, userID uuid.UUID, req updatePostRequest) (*Post, error) {
	post, err := s.repo.GetByIDSimple(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, errPostNotFound
	}

	if post.AuthorID != userID {
		return nil, errNotAuthor
	}

	if req.Title != nil {
		post.Title = *req.Title
	}

	if req.Content != nil {
		post.Content = *req.Content
	}

	err = s.repo.Update(ctx, post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) delete(ctx context.Context, postID, userID uuid.UUID) error {
	post, err := s.repo.GetByIDSimple(ctx, postID)
	if err != nil {
		return err
	}

	if post == nil {
		return errPostNotFound
	}

	if post.AuthorID != userID {
		return errNotAuthor
	}

	return s.repo.Delete(ctx, postID)
}

func (s *PostService) getFeed(ctx context.Context, req getFeedRequest) ([]feedUserResponse, utils.Metadata, error) {
	feed, total, err := s.repo.GetFeed(ctx, req)
	if err != nil {
		return nil, utils.Metadata{}, err
	}

	metadata := utils.CalculateMetadata(total, req.Page, req.PageSize)
	return feed, metadata, nil
}
