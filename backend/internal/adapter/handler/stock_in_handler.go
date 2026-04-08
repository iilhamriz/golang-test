package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/calyx/smart-inventory/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StockInHandler struct {
	uc *usecase.StockInUseCase
}

func NewStockInHandler(uc *usecase.StockInUseCase) *StockInHandler {
	return &StockInHandler{uc: uc}
}

func (h *StockInHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input usecase.CreateStockInInput
	if err := DecodeJSON(r, &input); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txn, err := h.uc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, txn)
}

func (h *StockInHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }

	txns, total, err := h.uc.List(r.Context(), status, page, limit)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONList(w, txns, page, limit, total)
}

func (h *StockInHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	txn, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "transaction not found")
		return
	}

	JSON(w, http.StatusOK, txn)
}

func (h *StockInHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := DecodeJSON(r, &input); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txn, err := h.uc.UpdateStatus(r.Context(), id, input.Status)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidTransition) {
			JSONError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			JSONError(w, http.StatusNotFound, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, txn)
}

func (h *StockInHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	txn, err := h.uc.Cancel(r.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrCannotCancelDone) {
			JSONError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrNotFound) {
			JSONError(w, http.StatusNotFound, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, txn)
}
