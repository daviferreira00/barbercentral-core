package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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
		if errors.Is(err, ErrNoActiveMembership) {
			shared.RespondWithError(w, http.StatusForbidden, "Sua conta não está vinculada a nenhuma barbearia ativa no momento. Fale com o administrador da plataforma.", nil)
			return
		}
		if errors.Is(err, ErrClientBlocked) {
			shared.RespondWithError(w, http.StatusForbidden, "Sua barbearia está suspensa ou bloqueada pelo administrador.", nil)
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
	userID, ok1 := ctx.Value("user_id").(string)
	userRole, _ := ctx.Value("user_role").(string)
	clientID, _ := ctx.Value("client_id").(string)
	originalAdminID, _ := ctx.Value("original_admin_id").(string)
	if !ok1 || userID == "" {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	profile, err := h.service.GetProfile(ctx, userID, userRole, clientID, originalAdminID)
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

	adminID, ok := r.Context().Value("user_id").(string)
	if !ok || adminID == "" {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	tokenStr, err := h.service.Impersonate(r.Context(), adminID, clientID)
	if err != nil {
		if errors.Is(err, ErrClientBlocked) {
			shared.RespondWithError(w, http.StatusForbidden, "Esta barbearia está suspensa ou bloqueada", nil)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao realizar impersonação", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, TokenResponse{Token: tokenStr})
}

func (h *AuthHandler) SwitchClient(w http.ResponseWriter, r *http.Request) {
	var req SwitchClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}
	if req.ClientID == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "O client_id é obrigatório", nil)
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	tokenStr, err := h.service.SwitchClient(r.Context(), userID, req.ClientID)
	if err != nil {
		if errors.Is(err, ErrNoActiveMembership) {
			shared.RespondWithError(w, http.StatusForbidden, "Você não tem vínculo ativo com esta barbearia", nil)
			return
		}
		if errors.Is(err, ErrClientBlocked) {
			shared.RespondWithError(w, http.StatusForbidden, "Esta barbearia está suspensa ou bloqueada", nil)
			return
		}
		if errors.Is(err, ErrInactiveUser) {
			shared.RespondWithError(w, http.StatusForbidden, "Sua conta está inativa", nil)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao trocar de barbearia", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, TokenResponse{Token: tokenStr})
}

func (h *AuthHandler) MyClients(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	list, err := h.service.MyClients(r.Context(), userID)
	if err != nil {
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao listar barbearias vinculadas", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, list)
}

func (h *AuthHandler) ReturnToAdmin(w http.ResponseWriter, r *http.Request) {
	originalAdminID, _ := r.Context().Value("original_admin_id").(string)

	tokenStr, err := h.service.ReturnToAdmin(r.Context(), originalAdminID)
	if err != nil {
		if errors.Is(err, ErrNotImpersonating) {
			shared.RespondWithError(w, http.StatusBadRequest, "Esta sessão não está impersonando nenhum admin", nil)
			return
		}
		shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao voltar para o painel admin", err)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, TokenResponse{Token: tokenStr})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	userRole, _ := ctx.Value("user_role").(string)
	if !ok || userID == "" {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Senha atual e nova senha são obrigatórias", nil)
		return
	}

	if len(req.NewPassword) < 8 {
		shared.RespondWithError(w, http.StatusBadRequest, "A nova senha deve ter no mínimo 8 caracteres", nil)
		return
	}

	err := h.service.ChangePassword(ctx, userID, userRole, req.CurrentPassword, req.NewPassword)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	userRole, _ := ctx.Value("user_role").(string)
	if !ok || userID == "" {
		shared.RespondWithError(w, http.StatusUnauthorized, "Não autenticado", nil)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, "Corpo da requisição inválido", err)
		return
	}

	if req.Name == "" || req.Email == "" {
		shared.RespondWithError(w, http.StatusBadRequest, "Nome e e-mail são obrigatórios", nil)
		return
	}

	if req.Password != "" && len(req.Password) < 8 {
		shared.RespondWithError(w, http.StatusBadRequest, "A nova senha deve ter no mínimo 8 caracteres", nil)
		return
	}

	var photoURL *string
	if req.PhotoBase64 != "" {
		if strings.HasPrefix(req.PhotoBase64, "data:") {
			parts := strings.SplitN(req.PhotoBase64, ",", 2)
			if len(parts) == 2 {
				headerPart := parts[0]
				dataPart := parts[1]

				ext := ".png"
				if strings.Contains(headerPart, "image/jpeg") || strings.Contains(headerPart, "image/jpg") {
					ext = ".jpg"
				} else if strings.Contains(headerPart, "image/webp") {
					ext = ".webp"
				}

				decoded, err := base64.StdEncoding.DecodeString(dataPart)
				if err != nil {
					shared.RespondWithError(w, http.StatusBadRequest, "Foto em base64 inválida", err)
					return
				}

				err = os.MkdirAll(shared.GetUploadsDir(), os.ModePerm)
				if err != nil {
					shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao preparar diretório de uploads", err)
					return
				}

				filename := fmt.Sprintf("profile-%s-%s%s", userID, uuid.New().String(), ext)
				filePath := filepath.Join(shared.GetUploadsDir(), filename)

				err = os.WriteFile(filePath, decoded, 0644)
				if err != nil {
					shared.RespondWithError(w, http.StatusInternalServerError, "Erro ao gravar foto no disco", err)
					return
				}

				pathStr := fmt.Sprintf("/uploads/%s", filename)
				photoURL = &pathStr
			}
		} else if strings.HasPrefix(req.PhotoBase64, "/uploads/") {
			pathStr := req.PhotoBase64
			photoURL = &pathStr
		}
	}

	tokenStr, err := h.service.UpdateProfile(ctx, userID, userRole, req.Name, req.Email, req.Password, photoURL)
	if err != nil {
		shared.RespondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	shared.RespondWithJSON(w, http.StatusOK, TokenResponse{Token: tokenStr})
}


