package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/calyx/smart-inventory/internal/domain/repository"
	"github.com/calyx/smart-inventory/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	uc *usecase.InventoryUseCase
}

func NewInventoryHandler(uc *usecase.InventoryUseCase) *InventoryHandler {
	return &InventoryHandler{uc: uc}
}

func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input usecase.CreateItemInput
	if err := DecodeJSON(r, &input); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.uc.CreateItem(r.Context(), input)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, item)
}

func (h *InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }

	filter := repository.ItemFilter{
		Name:  r.URL.Query().Get("name"),
		SKU:   r.URL.Query().Get("sku"),
		Page:  page,
		Limit: limit,
	}

	if custID := r.URL.Query().Get("customer_id"); custID != "" {
		if id, err := uuid.Parse(custID); err == nil {
			filter.CustomerID = &id
		}
	}

	items, total, err := h.uc.List(r.Context(), filter)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONList(w, items, page, limit, total)
}

func (h *InventoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	item, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "item not found")
		return
	}

	JSON(w, http.StatusOK, item)
}

func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input usecase.UpdateItemInput
	if err := DecodeJSON(r, &input); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.uc.UpdateItem(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			JSONError(w, http.StatusNotFound, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, item)
}

func (h *InventoryHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input usecase.AdjustStockInput
	if err := DecodeJSON(r, &input); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.uc.AdjustStock(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			JSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrStockCannotBeNeg) {
			JSONError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrInvalidInput) {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, item)
}
