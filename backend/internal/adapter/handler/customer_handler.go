package handler

import (
	"net/http"
	"strconv"

	"github.com/calyx/smart-inventory/internal/domain/entity"
	"github.com/calyx/smart-inventory/internal/domain/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	repo repository.CustomerRepository
}

func NewCustomerHandler(repo repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c entity.Customer
	if err := DecodeJSON(r, &c); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if c.Name == "" {
		JSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := h.repo.Create(r.Context(), &c); err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, c)
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }

	customers, total, err := h.repo.List(r.Context(), page, limit)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONList(w, customers, page, limit, total)
}

func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	c, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "customer not found")
		return
	}

	JSON(w, http.StatusOK, c)
}
