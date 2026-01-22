package user

import (
	"errors"
	"regexp"
	"strings"
)

var (
	errEmailFormat     = errors.New("invalid email format")
	errUsernameLength  = errors.New("Username must contain between 3 and 32 characters")
	errFullNameLength  = errors.New("Full name must contain between 3 and 32 characters")
	errPasswordLength  = errors.New("password length must be between 4 and 32")
	errNoUsernameEmail = errors.New("username or email is mandatory to login")
)

type registerUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

func (r *registerUserRequest) validate() error {
	err := validateEmailFormat(strings.TrimSpace(r.Email))
	if err != nil {
		return errEmailFormat
	}

	if len(strings.TrimSpace(r.Username)) < 3 || len(strings.TrimSpace(r.Username)) > 32 {
		return errUsernameLength
	}

	if len(strings.TrimSpace(r.FullName)) < 3 || len(strings.TrimSpace(r.FullName)) > 32 {
		return errFullNameLength
	}

	passwordLen := len(strings.TrimSpace(r.Password))
	if passwordLen > 32 || passwordLen < 4 {
		return errPasswordLength
	}

	return nil
}

func validateEmailFormat(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errEmailFormat
	}
	return nil
}

type loginRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password string  `json:"password"`
}

func (r *loginRequest) Validate() error {
	if r.Username == nil && r.Email == nil {
		return errNoUsernameEmail
	}

	if strings.TrimSpace(r.Password) == "" {
		return errPasswordLength
	}

	return nil
}

type updateUserRequest struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	FullName *string `json:"full_name"`
	Password *string `json:"password"`
}

func (r *updateUserRequest) Validate() error {
	if r.Email != nil {
		err := validateEmailFormat(strings.TrimSpace(*r.Email))
		if err != nil {
			return errEmailFormat
		}
	}

	if r.Username != nil {
		if len(strings.TrimSpace(*r.Username)) < 3 || len(strings.TrimSpace(*r.Username)) > 32 {
			return errUsernameLength
		}
	}

	if r.FullName != nil {
		if len(strings.TrimSpace(*r.FullName)) < 3 || len(strings.TrimSpace(*r.FullName)) > 32 {
			return errFullNameLength
		}
	}

	if r.Password != nil {
		passwordLen := len(strings.TrimSpace(*r.Password))
		if passwordLen > 32 || passwordLen < 4 {
			return errPasswordLength
		}
	}

	return nil
}
