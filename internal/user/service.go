package user

import (
	"context"
	"errors"
	"time"

	"github.com/NurulloMahmud/medicalka-project/internal/tasks"
	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/google/uuid"
)

var (
	errUsernameEmailTaken = errors.New("This username and/or email already taken. Try login")
	errUserNotFound       = errors.New("User not found")
	errEmailTaken         = errors.New("this email already taken")
	errUsernameTaken      = errors.New("this username already taken")
)

type UserService struct {
	repo        Repository
	emailSender *tasks.EmailSender
}

func NewService(repo Repository, emailSender *tasks.EmailSender) UserService {
	return UserService{
		repo:        repo,
		emailSender: emailSender,
	}
}

func (s *UserService) register(ctx context.Context, req registerUserRequest) (*User, error) {
	exists, err := s.repo.UserExists(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errUsernameEmailTaken
	}

	user := User{
		Email:    req.Email,
		Username: req.Username,
		FullName: req.FullName,
	}

	err = user.Password.Set(req.Password)
	if err != nil {
		return nil, err
	}

	verificationToken, err := utils.GenerateVerificationToken()
	if err != nil {
		return nil, err
	}

	newUser, err := s.repo.Create(ctx, user, verificationToken)
	if err != nil {
		return nil, err
	}

	s.emailSender.SendVerificationEmail(newUser.Email, verificationToken)
	return newUser, nil
}

func (s *UserService) get(ctx context.Context, id uuid.UUID, username, email string) (*User, error) {
	return s.repo.Get(ctx, id, username, email)
}

func (s *UserService) verifyUser(ctx context.Context, token string) error {
	userID, err := s.repo.verifyToken(ctx, token)
	if err != nil {
		return err
	}

	user, err := s.repo.Get(ctx, *userID, "", "")
	if err != nil {
		return err
	}

	if user == nil {
		return errUserNotFound
	}

	user.IsVerified = true
	user.UpdatedAt = time.Now()

	return s.repo.update(ctx, user, "")
}

func (s *UserService) update(ctx context.Context, data updateUserRequest) (*User, error) {
	ctxUser := utils.GetUser(ctx)

	user, err := s.repo.Get(ctx, ctxUser.ID, "", "")
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errUserNotFound
	}

	var token string

	if data.Email != nil {
		anotherUser, _ := s.repo.Get(ctx, uuid.Nil, "", *data.Email)
		if anotherUser != nil && anotherUser.ID != user.ID {
			return nil, errEmailTaken
		}

		user.Email = *data.Email
		user.IsVerified = false
		token, err = utils.GenerateVerificationToken()
		if err != nil {
			return nil, err
		}
	}

	if data.Username != nil {
		anotherUser, _ := s.repo.Get(ctx, uuid.Nil, *data.Username, "")
		if anotherUser != nil && anotherUser.ID != user.ID {
			return nil, errUsernameTaken
		}
		user.Username = *data.Username
	}

	if data.FullName != nil {
		user.FullName = *data.FullName
	}

	if data.Password != nil {
		// odatda old_password va new_password validationlar qilgan bolar edim
		err = user.Password.Set(*data.Password)
		if err != nil {
			return nil, err
		}
	}

	user.UpdatedAt = time.Now()
	err = s.repo.update(ctx, user, token)
	if err != nil {
		return nil, err
	}

	if token != "" {
		s.emailSender.SendVerificationEmail(user.Email, token)
	}
	return user, nil
}
