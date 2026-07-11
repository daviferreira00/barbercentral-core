package chat

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	list, err := h.service.ListChats(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar conversas", err)
		return
	}
	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	chatID := chi.URLParam(r, "id")

	list, err := h.service.ListMessages(r.Context(), clientID, chatID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar histórico de mensagens", err)
		return
	}
	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	msg, err := h.service.SendMessage(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, msg)
}

func (h *Handler) ProcessWebhook(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		// Retorna 200 OK para a Evolution API em formatos de payload diferentes que ignoramos
		shared.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"processed": false, "reason": "payload format ignored"})
		return
	}

	err := h.service.ProcessWebhook(r.Context(), payload)
	if err != nil {
		// Retorna 200 OK para evitar retentativas infinitas da Evolution para webhooks normais
		shared.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"processed": false, "error": err.Error()})
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"processed": true})
}
