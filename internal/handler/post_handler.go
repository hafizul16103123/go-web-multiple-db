package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"webapp/internal/domain"
	"webapp/internal/repository"
	"webapp/internal/service"
	"webapp/pkg/response"
)

// PostHandler holds the HTTP handlers for the posts resource.
type PostHandler struct {
	svc service.PostService
}

func NewPostHandler(svc service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// RegisterRoutes attaches all post routes to the given mux.
// Go 1.22+ supports "METHOD /path" patterns natively.
func (h *PostHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/posts", h.getAll)
	mux.HandleFunc("POST /api/v1/posts", h.create)
	mux.HandleFunc("GET /api/v1/posts/{id}", h.getByID)
	mux.HandleFunc("PUT /api/v1/posts/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/posts/{id}", h.delete)
}

// getAll handles GET /api/v1/posts
func (h *PostHandler) getAll(w http.ResponseWriter, r *http.Request) {
	posts, err := h.svc.GetAll()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch posts")
		return
	}
	response.JSON(w, http.StatusOK, posts)
}

// create handles POST /api/v1/posts
func (h *PostHandler) create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.AuthorID == 0 {
		response.Error(w, http.StatusUnprocessableEntity, "title and author_id are required")
		return
	}

	post, err := h.svc.Create(&req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create post")
		return
	}
	response.JSON(w, http.StatusCreated, post)
}

// getByID handles GET /api/v1/posts/{id}
func (h *PostHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}

	post, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "post not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to fetch post")
		return
	}
	response.JSON(w, http.StatusOK, post)
}

// update handles PUT /api/v1/posts/{id}
func (h *PostHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var req domain.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := h.svc.Update(id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "post not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update post")
		return
	}
	response.JSON(w, http.StatusOK, post)
}

// delete handles DELETE /api/v1/posts/{id}
func (h *PostHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid post id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "post not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete post")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "post deleted successfully"})
}

// pathID extracts and parses the {id} path parameter.
func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
