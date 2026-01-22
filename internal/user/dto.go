package user

import (
	"errors"
	"regexp"
	"strings"
)

var (
	errEmailFormat     = errors.New("invalid email format")
	errUsernameLength  = errors.New("username must contain between 3 and 32 characters")
	errUsernameFormat  = errors.New("username can only contain latin letters, digits and underscore")
	errFullNameLength  = errors.New("full name must contain between 2 and 100 characters")
	errFullNameFormat  = errors.New("full name can only contain letters, spaces and hyphens")
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

	err = validateUsername(r.Username)
	if err != nil {
		return err
	}

	err = validateFullName(r.FullName)
	if err != nil {
		return err
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
		err := validateUsername(*r.Username)
		if err != nil {
			return err
		}
	}

	if r.FullName != nil {
		err := validateFullName(*r.FullName)
		if err != nil {
			return err
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

func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 32 {
		return errUsernameLength
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return errUsernameFormat
	}

	return nil
}

func validateFullName(fullName string) error {
	fullName = strings.TrimSpace(fullName)
	if len(fullName) < 2 || len(fullName) > 100 {
		return errFullNameLength
	}

	fullNameRegex := regexp.MustCompile(`^[\p{L}\s\-]+$`)
	if !fullNameRegex.MatchString(fullName) {
		return errFullNameFormat
	}

	return nil
}
