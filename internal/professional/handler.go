package professional

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"barbercentral-core/internal/shared"
)

type ProfessionalHandler struct {
	service Service
}

func NewProfessionalHandler(service Service) *ProfessionalHandler {
	return &ProfessionalHandler{service: service}
}

func (h *ProfessionalHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	status := r.URL.Query().Get("status")

	list, err := h.service.List(r.Context(), clientID, status)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar profissionais", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *ProfessionalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	p, err := h.service.GetByID(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrProfessionalNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Profissional não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar profissional", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, p)
}

func (h *ProfessionalHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req CreateProfessionalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O nome do profissional é obrigatório", nil)
		return
	}

	p, err := h.service.Create(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao cadastrar profissional", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, p)
}

func (h *ProfessionalHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req UpdateProfessionalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O nome do profissional é obrigatório", nil)
		return
	}

	p, err := h.service.Update(r.Context(), clientID, id, req)
	if err != nil {
		if errors.Is(err, ErrProfessionalNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Profissional não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar profissional", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, p)
}

func (h *ProfessionalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	err := h.service.Delete(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrProfessionalNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Profissional não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao remover profissional", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *ProfessionalHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	list, err := h.service.GetSchedule(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar agenda", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *ProfessionalHandler) SaveSchedule(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req BulkUpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da grade semanal inválido", err)
		return
	}

	err := h.service.SaveSchedule(r.Context(), clientID, id, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao salvar grade semanal", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *ProfessionalHandler) GetLinkedServices(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	list, err := h.service.GetLinkedServices(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar serviços vinculados", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *ProfessionalHandler) LinkService(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req LinkServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Dados do vínculo inválidos", err)
		return
	}

	if req.ServiceID == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O ID do serviço é obrigatório", nil)
		return
	}

	err := h.service.LinkService(r.Context(), clientID, id, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao vincular serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *ProfessionalHandler) UnlinkService(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")
	serviceID := chi.URLParam(r, "service_id")

	err := h.service.UnlinkService(r.Context(), clientID, id, serviceID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao desvincular serviço", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *ProfessionalHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Falha ao processar arquivo multipart", err)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Arquivo 'photo' ausente", err)
		return
	}
	defer file.Close()

	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao preparar diretório de uploads", err)
		return
	}

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s-%s%s", id, uuid.New().String(), ext)
	filePath := filepath.Join("./uploads", filename)

	out, err := os.Create(filePath)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar arquivo no disco", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gravar arquivo no disco", err)
		return
	}

	photoURL := fmt.Sprintf("http://localhost:8080/uploads/%s", filename)

	err = h.service.UpdatePhoto(r.Context(), clientID, id, photoURL)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao associar foto ao profissional", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]string{"photo_url": photoURL})
}
