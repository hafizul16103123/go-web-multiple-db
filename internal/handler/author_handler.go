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

// AuthorHandler holds the HTTP handlers for the authors resource.
type AuthorHandler struct {
	svc service.AuthorService
}

func NewAuthorHandler(svc service.AuthorService) *AuthorHandler {
	return &AuthorHandler{svc: svc}
}

// RegisterRoutes attaches all author routes to the given mux.
func (h *AuthorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/authors", h.getAll)
	mux.HandleFunc("POST /api/v1/authors", h.create)
	mux.HandleFunc("GET /api/v1/authors/{id}", h.getByID)
	mux.HandleFunc("PUT /api/v1/authors/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/authors/{id}", h.delete)
}

func (h *AuthorHandler) getAll(w http.ResponseWriter, r *http.Request) {
	authors, err := h.svc.GetAll()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch authors")
		return
	}
	response.JSON(w, http.StatusOK, authors)
}

func (h *AuthorHandler) create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		response.Error(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	author, err := h.svc.Create(&req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create author")
		return
	}
	response.JSON(w, http.StatusCreated, author)
}

func (h *AuthorHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := authorPathID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid author id")
		return
	}

	author, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "author not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to fetch author")
		return
	}
	response.JSON(w, http.StatusOK, author)
}

func (h *AuthorHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := authorPathID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid author id")
		return
	}

	var req domain.UpdateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	author, err := h.svc.Update(id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "author not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update author")
		return
	}
	response.JSON(w, http.StatusOK, author)
}

func (h *AuthorHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := authorPathID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid author id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "author not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete author")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "author deleted successfully"})
}

func authorPathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
