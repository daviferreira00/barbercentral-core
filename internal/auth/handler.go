package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"barbercentral-core/internal/shared"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/magic-link", h.RequestMagicLink)
		r.Post("/magic-link/verify", h.VerifyMagicLink)
		r.Post("/password-reset", h.ForgotPassword)
		r.Post("/password-reset/confirm", h.ResetPassword)
		r.Post("/logout", h.Logout)
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "E-mail e senha são obrigatórios", nil)
		return
	}

	resp, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			shared.RespondWithError(w, http.StatusUnauthorized, "E-mail ou senha incorretos", nil)
			return
		}
		if errors.Is(err, ErrInactiveUser) {
			shared.RespondWithError(w, http.StatusForbidden, "Esta conta está inativa", nil)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro interno ao processar login", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var req MagicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Email == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O e-mail é obrigatório", nil)
		return
	}

	host := r.Header.Get("X-Frontend-Host")
	if host == "" {
		host = r.Header.Get("Origin")
	}
	if host == "" {
		host = "http://localhost:3000"
	}

	_, err := h.service.RequestMagicLink(r.Context(), req.Email, host)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Prevenir enumeração de emails
			shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
			return
		}
		if errors.Is(err, ErrInactiveUser) {
			shared.RespondWithError(w, http.StatusForbidden, "Esta conta está inativa", nil)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao processar magic link", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	var req MagicLinkVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Token == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O token é obrigatório", nil)
		return
	}

	resp, err := h.service.VerifyMagicLink(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenUsed) {
			shared.RespondWithError(w, http.StatusBadRequest, "Link de acesso inválido ou já utilizado.", err)
			return
		}
		if errors.Is(err, ErrTokenExpired) {
			shared.RespondWithError(w, http.StatusBadRequest, "Link de acesso expirado. Solicite um novo.", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao verificar magic link", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Email == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O e-mail é obrigatório", nil)
		return
	}

	host := r.Header.Get("X-Frontend-Host")
	if host == "" {
		host = r.Header.Get("Origin")
	}
	if host == "" {
		host = "http://localhost:3000"
	}

	_, err := h.service.ForgotPassword(r.Context(), req.Email, host)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Prevenir enumeração de emails
			shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao processar recuperação de senha", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Token == "" || req.Password == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Token e nova senha são obrigatórios", nil)
		return
	}

	if len(req.Password) < 8 {
		shared.RespondWithError(w, http.StatusBadRequest, "A senha deve ter no mínimo 8 caracteres", nil)
		return
	}

	_, err := h.service.ResetPassword(r.Context(), req.Token, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenUsed) {
			shared.RespondWithError(w, http.StatusBadRequest, "Link de redefinição inválido ou já utilizado.", err)
			return
		}
		if errors.Is(err, ErrTokenExpired) {
			shared.RespondWithError(w, http.StatusBadRequest, "Link de redefinição expirado. Solicite um novo.", err)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao redefinir senha", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDVal := ctx.Value("user_id")
	userRoleVal := ctx.Value("user_role")
	if userIDVal == nil || userRoleVal == nil {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	userID, ok1 := userIDVal.(string)
	userRole, ok2 := userRoleVal.(string)
	if !ok1 || !ok2 {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	profile, err := h.service.GetProfile(ctx, userID, userRole)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao carregar perfil", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, profile)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) Impersonate(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "client_id")
	if clientID == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O client_id é obrigatório", nil)
		return
	}

	adminIDVal := r.Context().Value("user_id")
	if adminIDVal == nil {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}
	adminID, ok := adminIDVal.(string)
	if !ok {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	tokenStr, err := h.service.Impersonate(r.Context(), adminID, clientID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao realizar impersonação", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]string{"token": tokenStr})
}
