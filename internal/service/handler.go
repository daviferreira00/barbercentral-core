package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
)

type ServiceHandler struct {
	service ServiceService
}

func NewServiceHandler(service ServiceService) *ServiceHandler {
	return &ServiceHandler{service: service}
}

func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	categoryID := r.URL.Query().Get("category_id")

	list, err := h.service.List(r.Context(), clientID, categoryID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar serviços", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *ServiceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	s, err := h.service.GetByID(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrServiceNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Serviço não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, s)
}

func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O nome do serviço é obrigatório", nil)
		return
	}

	s, err := h.service.Create(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, s)
}

func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O nome do serviço é obrigatório", nil)
		return
	}

	s, err := h.service.Update(r.Context(), clientID, id, req)
	if err != nil {
		if errors.Is(err, ErrServiceNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Serviço não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, s)
}

func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	err := h.service.Delete(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrServiceNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Serviço não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao deletar serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *ServiceHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	list, err := h.service.ListCategories(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar categorias", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *ServiceHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da categoria inválido", err)
		return
	}

	if req.Name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O nome da categoria é obrigatório", nil)
		return
	}

	cat, err := h.service.CreateCategory(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar categoria", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, cat)
}
