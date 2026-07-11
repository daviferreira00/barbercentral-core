package whatsapp

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
	list, err := h.service.ListInstances(r.Context())
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar instâncias", err)
		return
	}
	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.InstanceName == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome da instância é obrigatório", nil)
		return
	}

	inst, err := h.service.CreateInstance(r.Context(), req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, inst)
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome da instância é obrigatório", nil)
		return
	}

	res, err := h.service.ConnectInstance(r.Context(), name)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao obter QR Code", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, res)
}

func (h *Handler) State(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome da instância é obrigatório", nil)
		return
	}

	res, err := h.service.StateInstance(r.Context(), name)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar status da conexão", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome da instância é obrigatório", nil)
		return
	}

	res, err := h.service.LogoutInstance(r.Context(), name)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao desconectar instância", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, res)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome da instância é obrigatório", nil)
		return
	}

	res, err := h.service.DeleteInstance(r.Context(), name)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao excluir instância", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, res)
}

func (h *Handler) Link(w http.ResponseWriter, r *http.Request) {
	var req LinkInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.InstanceName == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome da instância é obrigatório", nil)
		return
	}

	err := h.service.LinkInstance(r.Context(), req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao vincular instância", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
