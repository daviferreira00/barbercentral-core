package customer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
	"barbercentral-core/internal/planlimit"
)

type CustomerHandler struct {
	service Service
}

func NewCustomerHandler(service Service) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" || req.Phone == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome e telefone são obrigatórios", nil)
		return
	}

	c, err := h.service.Create(r.Context(), clientID, req)
	if err != nil {
		var limit string
		var curr, max int
		n, scanErr := fmt.Sscanf(err.Error(), "plan_limit_exceeded:%s:%d:%d", &limit, &curr, &max)
		if scanErr == nil && n == 3 {
			planlimit.RespondWithLimitExceeded(w, limit, curr, max)
			return
		}
		if errors.Is(err, ErrCPFInvalid) {
			shared.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		if errors.Is(err, ErrPhoneDuplicate) {
			shared.RespondWithError(w, http.StatusConflict, err.Error(), err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao cadastrar cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, c)
}

func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" || req.Phone == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome e telefone são obrigatórios", nil)
		return
	}

	c, err := h.service.Update(r.Context(), clientID, id, req)
	if err != nil {
		if errors.Is(err, ErrCustomerNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, err.Error(), err)
			return
		}
		if errors.Is(err, ErrCPFInvalid) {
			shared.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		if errors.Is(err, ErrPhoneDuplicate) {
			shared.RespondWithError(w, http.StatusConflict, err.Error(), err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar dados do cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, c)
}

func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	err := h.service.Delete(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao deletar cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	stats, err := h.service.GetByID(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusNotFound, "Cliente não encontrado", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, stats)
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	query := r.URL.Query().Get("query")
	birthMonthStr := r.URL.Query().Get("birth_month")
	birthMonth := 0
	if birthMonthStr != "" {
		birthMonth, _ = strconv.Atoi(birthMonthStr)
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	list, total, err := h.service.List(r.Context(), clientID, query, birthMonth, page, pageSize)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar clientes", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":  list,
		"total": total,
	})
}

func (h *CustomerHandler) Search(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	query := r.URL.Query().Get("q")

	list, err := h.service.Search(r.Context(), clientID, query)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao pesquisar clientes", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *CustomerHandler) GetAppointmentsHistory(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	list, err := h.service.GetAppointmentsHistory(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar histórico de atendimentos", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *CustomerHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	csvBytes, err := h.service.ExportCSV(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gerar exportação dos clientes", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=clientes.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvBytes)
}
