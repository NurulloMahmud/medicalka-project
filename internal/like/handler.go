package like

import (
	"errors"
	"log"
	"net/http"

	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LikeHandler struct {
	service LikeService
	logger  *log.Logger
}

func NewHandler(s LikeService, log *log.Logger) *LikeHandler {
	return &LikeHandler{
		service: s,
		logger:  log,
	}
}

func (h *LikeHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "authentication required")
		return
	}

	postIDParam := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid post id"), h.logger)
		return
	}

	like, err := h.service.create(r.Context(), postID, ctxUser.ID)
	if err != nil {
		switch err {
		case errPostNotFound:
			utils.NotFound(w, r, h.logger)
		case errCannotLikeOwn:
			utils.Forbidden(w, r, err.Error())
		case errAlreadyLiked:
			utils.BadRequest(w, r, err, h.logger)
		default:
			utils.InternalServerError(w, r, err, h.logger)
		}
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": like})
}

func (h *LikeHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "authentication required")
		return
	}

	postIDParam := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid post id"), h.logger)
		return
	}

	err = h.service.delete(r.Context(), postID, ctxUser.ID)
	if err != nil {
		switch err {
		case errPostNotFound, errLikeNotFound:
			utils.NotFound(w, r, h.logger)
		default:
			utils.InternalServerError(w, r, err, h.logger)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "like removed successfully"})
}
