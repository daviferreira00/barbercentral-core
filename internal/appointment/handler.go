package appointment

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
)

type AppointmentHandler struct {
	service           Service
	verificationCodes *verificationStore
}

func NewAppointmentHandler(service Service) *AppointmentHandler {
	return &AppointmentHandler{
		service: service,
		verificationCodes: &verificationStore{
			codes: make(map[string]verificationCode),
		},
	}
}

func (h *AppointmentHandler) ListBlockedSlots(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	profID := r.URL.Query().Get("professional_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	list, err := h.service.ListBlockedSlots(r.Context(), clientID, profID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar bloqueios", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *AppointmentHandler) CreateBlockedSlot(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)

	var req CreateBlockedSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo do bloqueio inválido", err)
		return
	}

	if req.Date == "" || req.StartTime == "" || req.EndTime == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Data, hora inicial e hora final são obrigatórias", nil)
		return
	}

	slot, err := h.service.CreateBlockedSlot(r.Context(), clientID, userID, req)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao registrar bloqueio de agenda", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, slot)
}

func (h *AppointmentHandler) DeleteBlockedSlot(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	err := h.service.DeleteBlockedSlot(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrBlockedSlotNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Bloqueio não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao remover bloqueio", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AppointmentHandler) List(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	profID := r.URL.Query().Get("professional_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	list, err := h.service.List(r.Context(), clientID, profID, startDate, endDate)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar agendamentos", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *AppointmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	app, err := h.service.GetByID(r.Context(), clientID, id)
	if err != nil {
		if errors.Is(err, ErrAppointmentNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Agendamento não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, app)
}

func (h *AppointmentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo do status inválido", err)
		return
	}

	if req.Status == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O campo status é obrigatório", nil)
		return
	}

	userID, _ := r.Context().Value("user_id").(string)
	notesVal := ""
	if req.Notes != nil {
		notesVal = *req.Notes
	}

	err := h.service.UpdateStatus(r.Context(), clientID, id, req.Status, userID, notesVal)
	if err != nil {
		if errors.Is(err, ErrAppointmentNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Agendamento não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar status do agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AppointmentHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	profID := r.URL.Query().Get("professional_id")
	date := r.URL.Query().Get("date")

	serviceIDs := r.URL.Query()["service_ids"]
	if len(serviceIDs) == 0 {
		sIDStr := r.URL.Query().Get("service_id")
		if sIDStr != "" {
			serviceIDs = strings.Split(sIDStr, ",")
		}
	}

	if date == "" || len(serviceIDs) == 0 {
		shared.RespondWithError(w, http.StatusBadRequest, "Data e IDs de serviços são obrigatórios", nil)
		return
	}

	slots, err := h.service.GetAvailability(r.Context(), slug, profID, serviceIDs, date)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao calcular disponibilidade", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, slots)
}

func (h *AppointmentHandler) CreatePublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req CreatePublicAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.ProfessionalID == "" || len(req.ServiceIDs) == 0 || req.Date == "" || req.StartTime == "" || req.CustomerName == "" || req.CustomerPhone == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Todos os campos obrigatórios do agendamento devem ser preenchidos", nil)
		return
	}

	app, err := h.service.CreatePublic(r.Context(), slug, req)
	if err != nil {
		if errors.Is(err, ErrSlotNotAvailable) {
			shared.RespondWithError(w, http.StatusConflict, "O horário selecionado já não está mais disponível", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao registrar agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, app)
}

func (h *AppointmentHandler) GetByCancelToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	app, err := h.service.GetByCancelToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrAppointmentNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Agendamento não encontrado", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao buscar agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, app)
}

func (h *AppointmentHandler) CancelByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := h.service.CancelByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrAppointmentNotFound) {
			shared.RespondWithError(w, http.StatusNotFound, "Agendamento não encontrado", err)
			return
		}
		if errors.Is(err, ErrCancelPolicy) {
			shared.RespondWithError(w, http.StatusForbidden, err.Error(), err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao cancelar agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AppointmentHandler) SendReminders(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	expectedToken := "service_token"

	if authHeader != "Bearer "+expectedToken {
		shared.RespondWithError(w, http.StatusUnauthorized, "Token de serviço inválido", nil)
		return
	}

	count, err := h.service.SendUpcomingReminders(r.Context())
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao disparar lembretes", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"sent": count})
}

func (h *AppointmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)

	var req CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.ProfessionalID == "" || len(req.ServiceIDs) == 0 || req.Date == "" || req.StartTime == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Campos obrigatórios ausentes", nil)
		return
	}

	app, err := h.service.Create(r.Context(), clientID, userID, req)
	if err != nil {
		if errors.Is(err, ErrSlotNotAvailable) {
			shared.RespondWithError(w, http.StatusConflict, "O horário selecionado já não está mais disponível ou colide com outro agendamento", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao criar agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusCreated, app)
}

func (h *AppointmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")

	var req UpdateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	app, err := h.service.Update(r.Context(), clientID, id, userID, req)
	if err != nil {
		if errors.Is(err, ErrSlotNotAvailable) {
			shared.RespondWithError(w, http.StatusConflict, "O horário selecionado já não está mais disponível ou colide com outro agendamento", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao atualizar agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, app)
}

func (h *AppointmentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	err := h.service.Cancel(r.Context(), clientID, id, userID, req.Reason)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao cancelar agendamento", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AppointmentHandler) GetStatusLogs(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	id := chi.URLParam(r, "id")

	logs, err := h.service.GetStatusLogs(r.Context(), clientID, id)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar logs de status", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, logs)
}

func (h *AppointmentHandler) GetAvailabilityInternal(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)
	profID := r.URL.Query().Get("professional_id")
	date := r.URL.Query().Get("date")

	serviceIDs := r.URL.Query()["service_ids"]
	if len(serviceIDs) == 0 {
		sIDStr := r.URL.Query().Get("service_id")
		if sIDStr != "" {
			serviceIDs = strings.Split(sIDStr, ",")
		}
	}

	if date == "" || len(serviceIDs) == 0 {
		shared.RespondWithError(w, http.StatusBadRequest, "Data e IDs de serviços são obrigatórios", nil)
		return
	}

	slots, err := h.service.GetAvailabilityInternal(r.Context(), clientID, profID, serviceIDs, date)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao calcular disponibilidade interna", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, slots)
}

// Suporte para código de verificação temporário por WhatsApp
type verificationCode struct {
	Code      string
	ExpiresAt time.Time
}

type verificationStore struct {
	mu    sync.RWMutex
	codes map[string]verificationCode
}

func (v *verificationStore) Set(phone, code string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.codes[phone] = verificationCode{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
}

func (v *verificationStore) Verify(phone, code string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	vc, exists := v.codes[phone]
	if !exists {
		return false
	}

	if time.Now().After(vc.ExpiresAt) {
		delete(v.codes, phone)
		return false
	}

	if vc.Code == code {
		delete(v.codes, phone)
		return true
	}

	return false
}

func cleanNumber(phone string) string {
	var clean []rune
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			clean = append(clean, r)
		}
	}
	s := string(clean)
	if len(s) == 10 || len(s) == 11 {
		s = "55" + s
	}
	return s
}

func (h *AppointmentHandler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	cleanPhone := cleanNumber(req.Phone)
	if cleanPhone == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Telefone inválido", nil)
		return
	}

	// Generate 6-digit code
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Send WhatsApp message
	msg := fmt.Sprintf("Seu código de confirmação para o agendamento no BarberCentral é: %s", code)
	err := h.service.SendWhatsAppNotification(r.Context(), slug, cleanPhone, msg)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Falha ao enviar código por WhatsApp: "+err.Error(), err)
		return
	}

	h.verificationCodes.Set(cleanPhone, code)

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AppointmentHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	cleanPhone := cleanNumber(req.Phone)
	if cleanPhone == "" || req.Code == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Telefone e código são obrigatórios", nil)
		return
	}

	valid := h.verificationCodes.Verify(cleanPhone, req.Code)
	if !valid {
		shared.RespondWithError(w, http.StatusUnauthorized, "Código inválido ou expirado", nil)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"verified": true})
}

