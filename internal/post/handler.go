package post

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PostHandler struct {
	service PostService
	logger  *log.Logger
}

func NewHandler(s PostService, log *log.Logger) *PostHandler {
	return &PostHandler{
		service: s,
		logger:  log,
	}
}

func (h *PostHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "authentication required")
		return
	}

	if !ctxUser.IsVerified {
		utils.Forbidden(w, r, "email verification required to create posts")
		return
	}

	var data createPostRequest
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

	post, err := h.service.create(r.Context(), data, ctxUser.ID)
	if err != nil {
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": post})
}

func (h *PostHandler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	var req getPostsRequest

	req.Search = utils.ReadString(r, "search", "")
	req.Page = utils.ReadInt(r, "page", 1)
	req.PageSize = utils.ReadInt(r, "page_size", 10)
	req.Sort = utils.ReadString(r, "sort", "-created_at")
	req.SortSafeList = []string{"created_at", "title"}

	dateFromStr := utils.ReadString(r, "date_from", "")
	if dateFromStr != "" {
		parsed, err := time.Parse(time.RFC3339, dateFromStr)
		if err != nil {
			utils.BadRequest(w, r, errInvalidDateFormat, h.logger)
			return
		}
		req.DateFrom = &parsed
	}

	dateToStr := utils.ReadString(r, "date_to", "")
	if dateToStr != "" {
		parsed, err := time.Parse(time.RFC3339, dateToStr)
		if err != nil {
			utils.BadRequest(w, r, errInvalidDateFormat, h.logger)
			return
		}
		req.DateTo = &parsed
	}

	err := req.Validate()
	if err != nil {
		utils.BadRequest(w, r, err, h.logger)
		return
	}

	posts, metadata, err := h.service.getAll(r.Context(), req)
	if err != nil {
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"data":     posts,
		"metadata": metadata,
	})
}

func (h *PostHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid post id"), h.logger)
		return
	}

	post, err := h.service.getByID(r.Context(), id)
	if err != nil {
		if err == errPostNotFound {
			utils.NotFound(w, r, h.logger)
			return
		}
		utils.InternalServerError(w, r, err, h.logger)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": post})
}

func (h *PostHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "authentication required")
		return
	}

	idParam := chi.URLParam(r, "id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid post id"), h.logger)
		return
	}

	var req updatePostRequest
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

	post, err := h.service.update(r.Context(), postID, ctxUser.ID, req)
	if err != nil {
		switch err {
		case errPostNotFound:
			utils.NotFound(w, r, h.logger)
		case errNotAuthor:
			utils.Forbidden(w, r, err.Error())
		default:
			utils.InternalServerError(w, r, err, h.logger)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": post})
}

func (h *PostHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	ctxUser := utils.GetUser(r.Context())
	if ctxUser.IsAnonymous() {
		utils.Unauthorized(w, r, "authentication required")
		return
	}

	idParam := chi.URLParam(r, "id")
	postID, err := uuid.Parse(idParam)
	if err != nil {
		utils.BadRequest(w, r, errors.New("invalid post id"), h.logger)
		return
	}

	err = h.service.delete(r.Context(), postID, ctxUser.ID)
	if err != nil {
		switch err {
		case errPostNotFound:
			utils.NotFound(w, r, h.logger)
		case errNotAuthor:
			utils.Forbidden(w, r, err.Error())
		default:
			utils.InternalServerError(w, r, err, h.logger)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "post deleted successfully"})
}
