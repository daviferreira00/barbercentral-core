package stock

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

	query := r.URL.Query().Get("query")
	filter := r.URL.Query().Get("filter") // all, active, low_stock, inactive
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	list, total, err := h.service.List(r.Context(), clientID, page, pageSize, query, filter)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar produtos", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":  list,
		"total": total,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	prod, err := h.service.GetByID(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, prod)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)

	var req CreateProductRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	prod, err := h.service.Create(r.Context(), clientID, userID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, prod)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req UpdateProductRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	prod, err := h.service.Update(r.Context(), clientID, id, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, prod)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	err := h.service.Delete(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	list, err := h.service.GetLowStock(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar produtos com baixo estoque", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) ListMovementsByProduct(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	list, err := h.service.ListMovementsByProduct(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar movimentações do produto", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateMovement(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")

	var req CreateMovementRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	move, err := h.service.CreateMovement(r.Context(), clientID, userID, id, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, move)
}

func (h *Handler) ListAllMovements(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	typeFilter := r.URL.Query().Get("type")

	list, err := h.service.ListAllMovements(r.Context(), clientID, typeFilter)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar movimentações", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) StockReport(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	report, err := h.service.StockReport(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gerar relatório de estoque", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, report)
}

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	csvData, err := h.service.ExportCSV(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao exportar estoque para CSV", err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=estoque.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

// Vínculos de serviços e consumo de produtos

func (h *Handler) ListServiceProducts(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	list, err := h.service.ListServiceProducts(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar insumos do serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *Handler) LinkServiceProduct(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req ServiceProduct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	err = h.service.LinkServiceProduct(r.Context(), clientID, id, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao vincular insumo ao serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type UpdateServiceProductQtyRequest struct {
	Quantity float64 `json:"quantity"`
}

func (h *Handler) UpdateServiceProductQty(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")
	productID := chi.URLParam(r, "product_id")

	var req UpdateServiceProductQtyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	err = h.service.UpdateServiceProductQty(r.Context(), clientID, id, productID, req.Quantity)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao alterar quantidade do insumo", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) UnlinkServiceProduct(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")
	productID := chi.URLParam(r, "product_id")

	err := h.service.UnlinkServiceProduct(r.Context(), clientID, id, productID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao desvincular insumo", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
