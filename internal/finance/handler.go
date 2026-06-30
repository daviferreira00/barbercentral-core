package finance

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	list, total, err := h.service.List(r.Context(), clientID, page, pageSize)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar históricos de caixa", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":  list,
		"total": total,
	})
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	current, err := h.service.GetCurrent(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar caixa atual", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, current)
}

func (h *Handler) Open(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)

	var req OpenRegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	reg, err := h.service.Open(r.Context(), clientID, userID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, reg)
}

func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")

	var req CloseRegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	err = h.service.Close(r.Context(), clientID, id, userID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	summary, err := h.service.GetSummary(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar resumo do caixa", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, summary)
}

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	csvData, err := h.service.ExportCSV(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao exportar caixa para CSV", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=caixa.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	list, err := h.service.ListTransactions(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar transações", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")

	var req CreateTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	tx, err := h.service.CreateTransaction(r.Context(), clientID, userID, id, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, tx)
}

func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")
	txID := chi.URLParam(r, "tx_id")

	err := h.service.DeleteTransaction(r.Context(), clientID, id, txID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// Registro e atualização de pagamentos em agendamentos

type PaymentRequest struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"` // cash, pix, card_debit, card_credit, other
	Notes  *string `json:"notes,omitempty"`
}

func (h *Handler) RegisterPayment(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")

	var req PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	err = h.service.RegisterPayment(r.Context(), clientID, userID, id, req.Amount, req.Method, req.Notes)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")

	var req PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	err = h.service.UpdatePayment(r.Context(), clientID, userID, id, req.Amount, req.Method, req.Notes)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
