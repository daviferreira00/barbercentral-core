package loyalty

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetProgram(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	p, err := h.service.GetProgram(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar programa de fidelidade", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, p)
}

func (h *Handler) SaveProgram(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req SaveProgramRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	p, err := h.service.SaveProgram(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, p)
}

func (h *Handler) DeleteProgram(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	err := h.service.DeleteProgram(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) GetCardByCustomerID(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	customerID := chi.URLParam(r, "id")

	card, txs, err := h.service.GetCardByCustomerID(r.Context(), clientID, customerID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar cartão de fidelidade", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"card":         card,
		"transactions": txs,
	})
}

func (h *Handler) RedeemReward(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	customerID := chi.URLParam(r, "id")

	var req RedeemRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	tx, err := h.service.RedeemReward(r.Context(), clientID, userID, customerID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, tx)
}
