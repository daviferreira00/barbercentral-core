package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
	"barbercentral-core/internal/planlimit"
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
	performedBy, _ := r.Context().Value("user_id").(string)
	if performedBy == "" {
		performedBy = "admin-session"
	}

	var req BlockRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Se falhar, reason fica vazio

	err := h.service.BlockClientWithReason(r.Context(), id, req.Reason, performedBy)
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
	performedBy, _ := r.Context().Value("user_id").(string)
	if performedBy == "" {
		performedBy = "admin-session"
	}

	var req BlockRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	err := h.service.UnblockClientWithReason(r.Context(), id, req.Reason, performedBy)
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
		var limit string
		var curr, max int
		n, scanErr := fmt.Sscanf(err.Error(), "plan_limit_exceeded:%s:%d:%d", &limit, &curr, &max)
		if scanErr == nil && n == 3 {
			planlimit.RespondWithLimitExceeded(w, limit, curr, max)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar usuário para barbearia", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, u)
}

func (h *AdminHandler) UpdateClientUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	var req UpdateClientUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" || req.Email == "" || req.Role == "" || req.Status == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome, e-mail, função e status são obrigatórios", nil)
		return
	}

	err := h.service.UpdateClientUser(r.Context(), userID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar usuário", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AdminHandler) DeleteClientUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	err := h.service.DeleteClientUser(r.Context(), userID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao excluir usuário", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AdminHandler) UpdateClientPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo inválido", err)
		return
	}

	if req.PlanID == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "plan_id é obrigatório", nil)
		return
	}

	err := h.service.UpdateClientPlan(r.Context(), id, req.PlanID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao alterar plano do cliente", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AdminHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListPlans(r.Context())
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar planos", err)
		return
	}
	shared.RespondWithJSON(w, http.StatusOK, plans)
}

func (h *AdminHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var p planlimit.Plan
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo inválido", err)
		return
	}

	res, err := h.service.CreatePlan(r.Context(), &p)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar plano", err)
		return
	}
	shared.RespondWithJSON(w, http.StatusCreated, res)
}

func (h *AdminHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var p planlimit.Plan
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo inválido", err)
		return
	}

	res, err := h.service.UpdatePlan(r.Context(), id, &p)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar plano", err)
		return
	}
	shared.RespondWithJSON(w, http.StatusOK, res)
}

func (h *AdminHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 * 1024 * 1024) // 5MB limit
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Arquivo inválido ou muito grande", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "O campo 'file' é obrigatório", err)
		return
	}
	defer file.Close()

	_ = os.MkdirAll("./uploads", 0755)

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("upload_%d%s", time.Now().UnixNano(), ext)
	outPath := filepath.Join("./uploads", filename)

	out, err := os.Create(outPath)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao salvar arquivo localmente", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gravar conteúdo do arquivo", err)
		return
	}

	fileURL := fmt.Sprintf("/uploads/%s", filename)
	shared.RespondWithJSON(w, http.StatusOK, map[string]string{"url": fileURL})
}
