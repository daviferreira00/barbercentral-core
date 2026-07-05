package clientconfig

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

	"barbercentral-core/internal/professional"
	"barbercentral-core/internal/service"
	"barbercentral-core/internal/shared"
)

type ConfigHandler struct {
	service     ConfigService
	profService professional.Service
	svcService  service.ServiceService
}

func NewConfigHandler(service ConfigService, profService professional.Service, svcService service.ServiceService) *ConfigHandler {
	return &ConfigHandler{
		service:     service,
		profService: profService,
		svcService:  svcService,
	}
}

func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	cfg, err := h.service.GetByClientID(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			// Cria configuração padrão se não existir
			cfg = &ClientConfig{
				ClientID:                 clientID,
				ColorPrimary:             "#1a1a1a",
				ColorSecondary:           "#c9a84c",
				FontFamily:               "Inter",
				Timezone:                 "America/Sao_Paulo",
				CancellationPolicyHours: 2,
				MinAdvanceHours:         1,
				MaxAdvanceDays:          30,
				Active:                   1,
			}
			err = h.service.UpdateLogo(r.Context(), clientID, "") // Só inicializa se der erro
		} else {
			shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar configurações", err)
			return
		}
	}

	shared.RespondWithJSON(w, http.StatusOK, cfg)
}

// GetBranding devolve config + nome/slug da barbearia ATIVA da sessão (JWT),
// mesmo formato de GetPublicData mas sem depender de slug na URL — usado pelo
// layout autenticado para pintar branding sem subdomínio.
func (h *ConfigHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	if clientID == "" {
		shared.RespondWithError(w, http.StatusNotFound, "Nenhuma barbearia selecionada", nil)
		return
	}

	data, err := h.service.GetBrandingByClientID(r.Context(), clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar identidade visual da barbearia", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, data)
}

func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	cfg, err := h.service.Update(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar configurações", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, cfg)
}

func (h *ConfigHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	err := r.ParseMultipartForm(5 * 1024 * 1024) // 5MB limit
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Arquivo inválido ou muito grande", err)
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "O campo 'logo' é obrigatório", err)
		return
	}
	defer file.Close()

	_ = os.MkdirAll(shared.GetUploadsDir(), 0755)

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("logo_%s_%d%s", clientID, time.Now().Unix(), ext)
	outPath := filepath.Join(shared.GetUploadsDir(), filename)

	out, err := os.Create(outPath)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao salvar arquivo localmente", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gravar conteúdo da logo", err)
		return
	}

	logoURL := fmt.Sprintf("/uploads/%s", filename)
	err = h.service.UpdateLogo(r.Context(), clientID, logoURL)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar logo no banco de dados", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]string{"logo_url": logoURL})
}

func (h *ConfigHandler) UploadLogoCentral(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	err := r.ParseMultipartForm(5 * 1024 * 1024) // 5MB limit
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Arquivo inválido ou muito grande", err)
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "O campo 'logo' é obrigatório", err)
		return
	}
	defer file.Close()

	_ = os.MkdirAll(shared.GetUploadsDir(), 0755)

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("logo_central_%s_%d%s", clientID, time.Now().Unix(), ext)
	outPath := filepath.Join(shared.GetUploadsDir(), filename)

	out, err := os.Create(outPath)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao salvar arquivo localmente", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gravar conteúdo da logo", err)
		return
	}

	logoURL := fmt.Sprintf("/uploads/%s", filename)
	err = h.service.UpdateLogoCentral(r.Context(), clientID, logoURL)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar logo no banco de dados", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]string{"logo_url": logoURL})
}

func (h *ConfigHandler) GetPublicData(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Barbearia não encontrada", err)
			return
		}
		if errors.Is(err, ErrClientBlocked) {
			shared.RespondWithError(w, http.StatusForbidden, "Esta barbearia está temporariamente indisponível", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar dados públicos da barbearia", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, data)
}

func (h *ConfigHandler) GetPublicServices(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Barbearia não encontrada", err)
			return
		}
		if errors.Is(err, ErrClientBlocked) {
			shared.RespondWithError(w, http.StatusForbidden, "Esta barbearia está temporariamente indisponível", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar dados da barbearia", err)
		return
	}

	// Lista os serviços ativos desta barbearia (clientID)
	services, err := h.svcService.List(r.Context(), data.ClientID, "")
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar serviços", err)
		return
	}

	// Retorna apenas ativos
	var activeServices []service.Service
	for _, s := range services {
		if s.Active == 1 {
			activeServices = append(activeServices, s)
		}
	}

	shared.RespondWithJSON(w, http.StatusOK, activeServices)
}

func (h *ConfigHandler) GetPublicProfessionals(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	data, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Barbearia não encontrada", err)
			return
		}
		if errors.Is(err, ErrClientBlocked) {
			shared.RespondWithError(w, http.StatusForbidden, "Esta barbearia está temporariamente indisponível", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar dados da barbearia", err)
		return
	}

	// Lista os profissionais ativos desta barbearia
	professionals, err := h.profService.List(r.Context(), data.ClientID, "active")
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar profissionais", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, professionals)
}

// Admin handlers - operate on any client by ID from URL

func (h *ConfigHandler) GetConfigByClientID(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")

	cfg, err := h.service.GetByClientID(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			cfg = &ClientConfig{
				ClientID:                clientID,
				ColorPrimary:            "#1a1a1a",
				ColorSecondary:          "#c9a84c",
				FontFamily:              "Inter",
				Timezone:                "America/Sao_Paulo",
				CancellationPolicyHours: 2,
				MinAdvanceHours:         1,
				MaxAdvanceDays:          30,
				Active:                  1,
			}
			shared.RespondWithJSON(w, http.StatusOK, cfg)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar configurações", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, cfg)
}

func (h *ConfigHandler) UpdateConfigByClientID(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	cfg, err := h.service.Update(r.Context(), clientID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar configurações", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, cfg)
}
