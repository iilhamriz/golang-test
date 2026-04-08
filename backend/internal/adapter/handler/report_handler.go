package handler

import (
	"net/http"
	"strconv"

	"github.com/calyx/smart-inventory/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReportHandler struct {
	uc *usecase.ReportUseCase
}

func NewReportHandler(uc *usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

func (h *ReportHandler) ListDoneTransactions(w http.ResponseWriter, r *http.Request) {
	txnType := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }

	txns, total, err := h.uc.ListDoneTransactions(r.Context(), txnType, page, limit)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONList(w, txns, page, limit, total)
}

func (h *ReportHandler) GetTransactionDetail(w http.ResponseWriter, r *http.Request) {
	txnType := chi.URLParam(r, "type")
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	detail, err := h.uc.GetTransactionDetail(r.Context(), txnType, id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "transaction not found")
		return
	}

	JSON(w, http.StatusOK, detail)
}
