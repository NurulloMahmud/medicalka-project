package user

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/NurulloMahmud/medicalka-project/config"
	"github.com/NurulloMahmud/medicalka-project/internal/auth"
	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/google/uuid"
)

type UserHandler struct {
	service UserService
	logger  *log.Logger
	cfg     config.Config
}

func NewHandler(s UserService, log *log.Logger, cfg config.Config) *UserHandler {
	return &UserHandler{
		service: s,
		logger:  log,
		cfg:     cfg,
	}
}

func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var data registerUserRequest

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	err = data.validate()
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	user, err := h.service.register(r.Context(), data)
	if err != nil {
		if err == errUsernameEmailTaken {
			utils.BadRequest(w, r, err, h.logger)
			return
		}

		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": user})
}

func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var data loginRequest
	var username, email string

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	err = data.Validate()
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	if data.Username != nil {
		username = *data.Username
	}
	if data.Email != nil {
		email = *data.Email
	}

	user, err := h.service.get(r.Context(), uuid.Nil, username, email)
	if err != nil {
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	if user == nil {
		utils.BadRequest(w, r, errors.New("Invalid credentials"), h.logger)
		return
	}

	match, err := user.Password.Matches(data.Password)
	if err != nil {
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	if !match {
		utils.BadRequest(w, r, errors.New("Invalid credentials"), h.logger)
		return
	}

	tokenClaims := auth.TokenClaims{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}

	token, err := auth.GenerateAccessToken(tokenClaims, h.cfg.JWTSecret)
	if err != nil {
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"access_token": token})
}

func (h *UserHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "User not logged in")
		return
	}

	user, err := h.service.get(r.Context(), ctxUser.ID, ctxUser.Username, ctxUser.Email)
	if err != nil {
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	if user == nil {
		utils.Unauthorized(w, r, "User not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": user})
}

func (h *UserHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "You should be logged in to change data")
		return
	}

	var data updateUserRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	err = data.Validate()
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	updatedUser, err := h.service.update(r.Context(), data)

	if err != nil {
		switch err {
		case errUserNotFound:
			utils.BadRequest(w, r, err, h.logger)
		case errEmailTaken:
			utils.BadRequest(w, r, err, h.logger)
		case errUsernameTaken:
			utils.BadRequest(w, r, err, h.logger)
		default:
			utils.InternalServerError(w, r, err, h.logger)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": updatedUser})
}

func (h *UserHandler) HandleEmailVerification(w http.ResponseWriter, r *http.Request) {
	token := utils.ReadString(r, "token", "")
	if token == "" {
		utils.BadRequest(w, r, errors.New("token is required"), h.logger)
		return
	}

	err := h.service.verifyUser(r.Context(), token)
	if err != nil {
		if err == errUserNotFound {
			utils.BadRequest(w, r, err, h.logger)
			return
		}
		utils.InternalServerError(w, r, err, h.logger)
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Email verified successfully"})
}
