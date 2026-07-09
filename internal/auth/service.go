package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"barbercentral-core/internal/email"
	"barbercentral-core/internal/middleware"
)

var (
	ErrInvalidCredentials  = errors.New("credenciais inválidas")
	ErrInvalidToken        = errors.New("token inválido")
	ErrTokenExpired        = errors.New("token expirado")
	ErrTokenUsed           = errors.New("token já utilizado")
	ErrInactiveUser        = errors.New("usuário inativo")
	ErrNoActiveMembership  = errors.New("nenhum vínculo ativo com barbearia")
	ErrClientBlocked       = errors.New("barbearia suspensa ou bloqueada")
	ErrNotImpersonating    = errors.New("esta sessão não está impersonando nenhum admin")
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	ValidateToken(tokenStr string) (*middleware.Claims, error)
	GetProfile(ctx context.Context, userID, role, clientID, originalAdminID string) (*ProfileResponse, error)
	ForgotPassword(ctx context.Context, email, host string) (string, error)
	ValidateResetToken(ctx context.Context, token string) (bool, *AuthToken, error)
	ResetPassword(ctx context.Context, token, newPassword string) (bool, error)
	RequestMagicLink(ctx context.Context, email, host string) (string, error)
	VerifyMagicLink(ctx context.Context, token string) (*LoginResponse, error)
	Impersonate(ctx context.Context, adminID, targetClientID string) (string, error)
	SwitchClient(ctx context.Context, userID, targetClientID string) (string, error)
	MyClients(ctx context.Context, userID string) ([]ClientMembership, error)
	ReturnToAdmin(ctx context.Context, originalAdminID string) (string, error)
	ChangePassword(ctx context.Context, userID, userRole, currentPassword, newPassword string) error
	UpdateProfile(ctx context.Context, userID, userRole, name, email, newPassword string, photoURL *string) (string, error)
}

type authService struct {
	repo        AuthRepository
	jwtSecret   string
	emailClient *email.Client
}

func NewAuthService(repo AuthRepository, jwtSecret string, emailClient *email.Client) AuthService {
	return &authService{
		repo:        repo,
		jwtSecret:   jwtSecret,
		emailClient: emailClient,
	}
}

// resolveClientSelection decide o client_id/role iniciais do JWT com base nos
// vínculos ativos do usuário. clientID/role vazios significam "2+ vínculos,
// precisa escolher pelo seletor".
func (s *authService) resolveClientSelection(ctx context.Context, userID string) (string, string, error) {
	memberships, err := s.repo.ListActiveMemberships(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if len(memberships) == 0 {
		return "", "", ErrNoActiveMembership
	}
	if len(memberships) == 1 {
		m := memberships[0]
		status, err := s.repo.GetClientStatus(ctx, m.ClientID)
		if err == nil && status == "blocked" {
			return "", "", ErrClientBlocked
		}
		return m.ClientID, m.Role, nil
	}
	// 2+ vínculos ativos: frontend deve forçar o seletor de barbearia
	return "", "", nil
}

func (s *authService) Login(ctx context.Context, emailVal, password string) (*LoginResponse, error) {
	var user *Usuario
	var passwordHash string
	isAdmin := false

	admin, err := s.repo.GetAdminByEmail(ctx, emailVal)
	if err == nil && admin != nil {
		isAdmin = true
		passwordHash = admin.PasswordHash
		user = &Usuario{
			ID:       admin.ID,
			Nome:     admin.Name,
			Email:    admin.Email,
			Role:     "admin",
			PhotoURL: admin.PhotoURL,
		}
	} else {
		acc, err := s.repo.GetUserAccountByEmail(ctx, emailVal)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
		if acc.Status != "active" {
			return nil, ErrInactiveUser
		}
		passwordHash = acc.PasswordHash
		user = &Usuario{
			ID:       acc.ID,
			Nome:     acc.Name,
			Email:    acc.Email,
			PhotoURL: acc.PhotoURL,
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !isAdmin {
		clientID, role, err := s.resolveClientSelection(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		user.ClientID = clientID
		user.Role = role
	}

	tokenStr, err := s.generateJWT(user, "")
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: tokenStr,
		User:  user,
	}, nil
}

func (s *authService) ValidateToken(tokenStr string) (*middleware.Claims, error) {
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Role != "admin" && claims.ClientID != "" {
		status, err := s.repo.GetClientStatus(context.Background(), claims.ClientID)
		if err == nil && status == "blocked" {
			return nil, errors.New("client_blocked")
		}
	}

	return claims, nil
}

// GetProfile monta o perfil da sessão atual a partir do CONTEXTO da requisição
// (client_id/role já resolvidos pelo JWT), não re-buscando vínculo no banco —
// isso é o que permite representar corretamente um admin impersonando (que não
// tem linha própria em client_user_link).
func (s *authService) GetProfile(ctx context.Context, userID, role, clientID, originalAdminID string) (*ProfileResponse, error) {
	if role == "admin" {
		admin, err := s.repo.GetAdminByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &ProfileResponse{User: &Usuario{
			ID:       admin.ID,
			Nome:     admin.Name,
			Email:    admin.Email,
			Role:     "admin",
			PhotoURL: admin.PhotoURL,
		}}, nil
	}

	if originalAdminID != "" {
		// Admin impersonando: user_id do token é o ID real do admin (nunca muda,
		// mesmo em impersonate encadeado), sem linha própria em client_user_link.
		admin, err := s.repo.GetAdminByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &ProfileResponse{User: &Usuario{
			ID:            admin.ID,
			ClientID:      clientID,
			Nome:          admin.Name,
			Email:         admin.Email,
			Role:          role,
			Impersonating: true,
			PhotoURL:      admin.PhotoURL,
		}}, nil
	}

	acc, err := s.repo.GetUserAccountByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{User: &Usuario{
		ID:                   acc.ID,
		ClientID:             clientID,
		Nome:                 acc.Name,
		Email:                acc.Email,
		Role:                 role,
		NeedsClientSelection: clientID == "" && role == "",
		PhotoURL:             acc.PhotoURL,
	}}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, emailVal, host string) (string, error) {
	var userID string
	var userType string
	var userName string

	admin, err := s.repo.GetAdminByEmail(ctx, emailVal)
	if err == nil && admin != nil {
		userID = admin.ID
		userType = "platform_admin"
		userName = admin.Name
	} else {
		acc, err := s.repo.GetUserAccountByEmail(ctx, emailVal)
		if err != nil {
			return "", ErrUserNotFound
		}
		userID = acc.ID
		userType = "client_user"
		userName = acc.Name
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(1 * time.Hour)
	tokenRecord := &AuthToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userType,
		Type:      "password_reset",
		Token:     tokenStr,
		Used:      0,
		ExpiresAt: expiresAt,
	}

	err = s.repo.CreateToken(ctx, tokenRecord)
	if err != nil {
		return "", err
	}

	resetLink := fmt.Sprintf("%s/criar-senha?token=%s", host, tokenStr)
	err = s.emailClient.SendResetPasswordEmail(emailVal, userName, resetLink)
	if err != nil {
		log.Error().Err(err).Str("email", emailVal).Msg("falha ao enviar email de recuperacao de senha")
	}

	log.Info().
		Str("email", emailVal).
		Str("token", tokenStr).
		Msgf("[DEV] Link de redefinição de senha: %s", resetLink)

	return tokenStr, nil
}

func (s *authService) ValidateResetToken(ctx context.Context, tokenStr string) (bool, *AuthToken, error) {
	token, err := s.repo.GetToken(ctx, tokenStr)
	if err != nil {
		return false, nil, err
	}
	if token.Type != "password_reset" {
		return false, nil, ErrInvalidToken
	}
	if token.Used == 1 {
		return false, nil, ErrTokenUsed
	}
	if time.Now().After(token.ExpiresAt) {
		return false, nil, ErrTokenExpired
	}
	return true, token, nil
}

func (s *authService) ResetPassword(ctx context.Context, tokenStr, newPassword string) (bool, error) {
	valid, token, err := s.ValidateResetToken(ctx, tokenStr)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	if token.UserType == "platform_admin" {
		err = s.repo.UpdateAdminPassword(ctx, token.UserID, string(hash))
	} else {
		err = s.repo.UpdateUserPassword(ctx, token.UserID, string(hash))
	}
	if err != nil {
		return false, err
	}

	err = s.repo.MarkTokenUsed(ctx, tokenStr)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *authService) RequestMagicLink(ctx context.Context, emailVal, host string) (string, error) {
	var userID string
	var userType string
	var userName string

	admin, err := s.repo.GetAdminByEmail(ctx, emailVal)
	if err == nil && admin != nil {
		userID = admin.ID
		userType = "platform_admin"
		userName = admin.Name
	} else {
		acc, err := s.repo.GetUserAccountByEmail(ctx, emailVal)
		if err != nil {
			return "", ErrUserNotFound
		}
		if acc.Status != "active" {
			return "", ErrInactiveUser
		}
		userID = acc.ID
		userType = "client_user"
		userName = acc.Name
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(15 * time.Minute)
	tokenRecord := &AuthToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserType:  userType,
		Type:      "magic_link",
		Token:     tokenStr,
		Used:      0,
		ExpiresAt: expiresAt,
	}

	err = s.repo.CreateToken(ctx, tokenRecord)
	if err != nil {
		return "", err
	}

	verifyLink := fmt.Sprintf("%s/magic-link?token=%s", host, tokenStr)
	err = s.emailClient.SendMagicLinkEmail(emailVal, userName, verifyLink)
	if err != nil {
		log.Error().Err(err).Str("email", emailVal).Msg("falha ao enviar email de login rapido (magic link)")
	}

	log.Info().
		Str("email", emailVal).
		Str("token", tokenStr).
		Msgf("[DEV] Link de login rápido (Magic Link): %s", verifyLink)

	return tokenStr, nil
}

func (s *authService) VerifyMagicLink(ctx context.Context, tokenStr string) (*LoginResponse, error) {
	token, err := s.repo.GetToken(ctx, tokenStr)
	if err != nil {
		return nil, err
	}
	if token.Type != "magic_link" {
		return nil, ErrInvalidToken
	}
	if token.Used == 1 {
		return nil, ErrTokenUsed
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	err = s.repo.MarkTokenUsed(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	var user *Usuario
	if token.UserType == "platform_admin" {
		admin, err := s.repo.GetAdminByID(ctx, token.UserID)
		if err != nil {
			return nil, err
		}
		user = &Usuario{
			ID:       admin.ID,
			Nome:     admin.Name,
			Email:    admin.Email,
			Role:     "admin",
			PhotoURL: admin.PhotoURL,
		}
	} else {
		acc, err := s.repo.GetUserAccountByID(ctx, token.UserID)
		if err != nil {
			return nil, err
		}
		if acc.Status != "active" {
			return nil, ErrInactiveUser
		}

		clientID, role, err := s.resolveClientSelection(ctx, acc.ID)
		if err != nil {
			return nil, err
		}

		user = &Usuario{
			ID:       acc.ID,
			ClientID: clientID,
			Nome:     acc.Name,
			Email:    acc.Email,
			Role:     role,
			PhotoURL: acc.PhotoURL,
		}
	}

	tokenJWT, err := s.generateJWT(user, "")
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: tokenJWT,
		User:  user,
	}, nil
}

// generateJWT emite um token novo. originalAdminID, quando preenchido, marca
// a sessão como "impersonando" (permite depois voltar para o painel admin).
func (s *authService) generateJWT(user *Usuario, originalAdminID string) (string, error) {
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &middleware.Claims{
		UserID:          user.ID,
		Email:           user.Email,
		Role:            user.Role,
		ClientID:        user.ClientID,
		OriginalAdminID: originalAdminID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// Impersonate reemite um JWT trocando o client_id/role para o admin "entrar"
// numa barbearia. adminID é sempre o ID real do admin (extraído do token atual
// pelo handler) — em impersonate encadeado (admin já impersonando pula para
// outra barbearia), o handler já resolve isso corretamente porque o user_id do
// token impersonado permanece o ID real do admin.
func (s *authService) Impersonate(ctx context.Context, adminID, targetClientID string) (string, error) {
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return "", err
	}
	if admin == nil {
		return "", errors.New("admin não encontrado")
	}

	status, err := s.repo.GetClientStatus(ctx, targetClientID)
	if err == nil && status == "blocked" {
		return "", ErrClientBlocked
	}

	user := &Usuario{
		ID:       admin.ID,
		ClientID: targetClientID,
		Nome:     admin.Name,
		Email:    admin.Email,
		Role:     "owner",
		PhotoURL: admin.PhotoURL,
	}

	return s.generateJWT(user, adminID)
}

// SwitchClient troca a barbearia ativa de um usuário comum (não-admin) que
// tenha 2+ vínculos, validando que o vínculo exista e esteja ativo.
func (s *authService) SwitchClient(ctx context.Context, userID, targetClientID string) (string, error) {
	link, err := s.repo.GetMembership(ctx, userID, targetClientID)
	if err != nil || link.Status != "active" {
		return "", ErrNoActiveMembership
	}

	status, err := s.repo.GetClientStatus(ctx, targetClientID)
	if err == nil && status == "blocked" {
		return "", ErrClientBlocked
	}

	acc, err := s.repo.GetUserAccountByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if acc.Status != "active" {
		return "", ErrInactiveUser
	}

	user := &Usuario{
		ID:       acc.ID,
		ClientID: targetClientID,
		Nome:     acc.Name,
		Email:    acc.Email,
		Role:     link.Role,
	}

	return s.generateJWT(user, "")
}

func (s *authService) MyClients(ctx context.Context, userID string) ([]ClientMembership, error) {
	return s.repo.ListMembershipsWithClient(ctx, userID)
}

// ReturnToAdmin restaura uma sessão de admin limpa a partir do OriginalAdminID
// preservado no token impersonado.
func (s *authService) ReturnToAdmin(ctx context.Context, originalAdminID string) (string, error) {
	if originalAdminID == "" {
		return "", ErrNotImpersonating
	}

	admin, err := s.repo.GetAdminByID(ctx, originalAdminID)
	if err != nil {
		return "", err
	}

	user := &Usuario{
		ID:       admin.ID,
		Nome:     admin.Name,
		Email:    admin.Email,
		Role:     "admin",
		PhotoURL: admin.PhotoURL,
	}

	return s.generateJWT(user, "")
}

func (s *authService) ChangePassword(ctx context.Context, userID, userRole, currentPassword, newPassword string) error {
	var passwordHash string
	var err error

	if userRole == "admin" {
		admin, err := s.repo.GetAdminByID(ctx, userID)
		if err != nil {
			return errors.New("usuário administrador não encontrado")
		}
		passwordHash = admin.PasswordHash
	} else {
		acc, err := s.repo.GetUserAccountByID(ctx, userID)
		if err != nil {
			return errors.New("usuário não encontrado")
		}
		passwordHash = acc.PasswordHash
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(currentPassword)); err != nil {
		return errors.New("senha atual incorreta")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("erro ao gerar hash da nova senha")
	}

	if userRole == "admin" {
		err = s.repo.UpdateAdminPassword(ctx, userID, string(newHash))
	} else {
		err = s.repo.UpdateUserPassword(ctx, userID, string(newHash))
	}

	if err != nil {
		return errors.New("erro ao atualizar a senha no banco de dados")
	}

	return nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID, userRole, name, email, newPassword string, photoURL *string) (string, error) {
	var passwordHash string
	if newPassword != "" {
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return "", errors.New("erro ao gerar hash da nova senha")
		}
		passwordHash = string(hashBytes)
	}

	var err error
	if userRole == "admin" {
		err = s.repo.UpdateAdminProfile(ctx, userID, name, email, passwordHash, photoURL)
	} else {
		err = s.repo.UpdateUserProfile(ctx, userID, name, email, passwordHash, photoURL)
	}

	if err != nil {
		return "", err
	}

	var user *Usuario
	if userRole == "admin" {
		admin, err := s.repo.GetAdminByID(ctx, userID)
		if err != nil {
			return "", err
		}
		user = &Usuario{
			ID:       admin.ID,
			Nome:     admin.Name,
			Email:    admin.Email,
			Role:     "admin",
			PhotoURL: admin.PhotoURL,
		}
	} else {
		acc, err := s.repo.GetUserAccountByID(ctx, userID)
		if err != nil {
			return "", err
		}

		clientID, role, err := s.resolveClientSelection(ctx, acc.ID)
		if err != nil {
			role = userRole
		}

		user = &Usuario{
			ID:       acc.ID,
			ClientID: clientID,
			Nome:     acc.Name,
			Email:    acc.Email,
			Role:     role,
			PhotoURL: acc.PhotoURL,
		}
	}

	return s.generateJWT(user, "")
}


