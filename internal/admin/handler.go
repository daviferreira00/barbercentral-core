package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
)

type AdminHandler struct {
	service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListClients(r.Context())
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar clientes", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *AdminHandler) GetClientByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	c, err := h.service.GetClientByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Cliente não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, c)
}

func (h *AdminHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" || req.Slug == "" || req.PlanID == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome, slug e plano são obrigatórios", nil)
		return
	}

	c, err := h.service.CreateClient(r.Context(), req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, c)
}

func (h *AdminHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" || req.Slug == "" || req.PlanID == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome, slug e plano são obrigatórios", nil)
		return
	}

	c, err := h.service.UpdateClient(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Cliente não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, c)
}

func (h *AdminHandler) BlockClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.BlockClient(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Cliente não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao bloquear cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AdminHandler) UnblockClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.UnblockClient(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Cliente não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao desbloquear cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AdminHandler) ListClientUsers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	list, err := h.service.ListClientUsers(r.Context(), id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar usuários do cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *AdminHandler) CreateClientUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req CreateClientUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo do usuário inválido", err)
		return
	}

	if req.Name == "" || req.Email == "" || req.Role == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome, e-mail e função (role) são obrigatórios", nil)
		return
	}

	u, err := h.service.CreateClientUser(r.Context(), id, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar usuário para barbearia", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, u)
}
