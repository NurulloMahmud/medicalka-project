package comment

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CommentHandler struct {
	service CommentService
	logger  *log.Logger
}

func NewHandler(s CommentService, log *log.Logger) *CommentHandler {
	return &CommentHandler{
		service: s,
		logger:  log,
	}
}

func (h *CommentHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "authentication required")
		return
	}

	if !ctxUser.IsVerified {
		utils.Forbidden(w, r, errNotVerified.Error())
		return
	}

	postIDParam := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid post id"), h.logger)
		return
	}

	var req createCommentRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	err = req.Validate()
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	comment, err := h.service.create(r.Context(), postID, ctxUser.ID, req)
	if err != nil {
		if err == errPostNotFound {
			utils.NotFound(w, r, h.logger)
			return
		}
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": comment})
}

func (h *CommentHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "authentication required")
		return
	}

	postIDParam := chi.URLParam(r, "postID")
	postID, err := uuid.Parse(postIDParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid post id"), h.logger)
		return
	}

	commentIDParam := chi.URLParam(r, "commentID")
	commentID, err := uuid.Parse(commentIDParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid comment id"), h.logger)
		return
	}

	err = h.service.delete(r.Context(), postID, commentID, ctxUser.ID)
	if err != nil {
		switch err {
		case errPostNotFound, errCommentNotFound:
			utils.NotFound(w, r, h.logger)
		case errNotAuthor:
			utils.Forbidden(w, r, err.Error())
		default:
			utils.InternalServerError(w, r, err, h.logger)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "comment deleted successfully"})
}
