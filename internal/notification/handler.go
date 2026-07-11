package notification

import (
	"encoding/json"
	"errors"
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
	list, err := h.service.List(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar configurações de notificações", err)
		return
	}
	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	config, err := h.service.GetByID(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Configuração de notificação não encontrada", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar configuração de notificação", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, config)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	config, err := h.service.Create(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, config)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req UpdateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	config, err := h.service.Update(r.Context(), clientID, id, req)
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Configuração de notificação não encontrada", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, config)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	err := h.service.Delete(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrNotificationNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Configuração de notificação não encontrada", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao excluir configuração de notificação", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
