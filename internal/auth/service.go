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
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrInvalidToken       = errors.New("token inválido")
	ErrTokenExpired       = errors.New("token expirado")
	ErrTokenUsed          = errors.New("token já utilizado")
	ErrInactiveUser       = errors.New("usuário inativo")
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*LoginResponse, error)
	ValidateToken(tokenStr string) (*middleware.Claims, error)
	GetProfile(ctx context.Context, userID, role string) (*ProfileResponse, error)
	ForgotPassword(ctx context.Context, email, host string) (string, error)
	ValidateResetToken(ctx context.Context, token string) (bool, *AuthToken, error)
	ResetPassword(ctx context.Context, token, newPassword string) (bool, error)
	RequestMagicLink(ctx context.Context, email, host string) (string, error)
	VerifyMagicLink(ctx context.Context, token string) (*LoginResponse, error)
	Impersonate(ctx context.Context, adminID, targetClientID string) (string, error)
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

func (s *authService) Login(ctx context.Context, emailVal, password string) (*LoginResponse, error) {
	var user *Usuario
	var passwordHash string

	// Tenta admin primeiro
	admin, err := s.repo.GetAdminByEmail(ctx, emailVal)
	if err == nil && admin != nil {
		passwordHash = admin.PasswordHash
		user = &Usuario{
			ID:    admin.ID,
			Nome:  admin.Name,
			Email: admin.Email,
			Role:  "admin",
		}
	} else {
		// Tenta client_user
		cUser, err := s.repo.GetUserByEmail(ctx, emailVal)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
		if cUser.Status != "active" {
			return nil, ErrInactiveUser
		}
		// Verificar se a barbearia está bloqueada
		status, err := s.repo.GetClientStatus(ctx, cUser.ClientID)
		if err == nil && status == "blocked" {
			return nil, errors.New("sua barbearia está suspensa ou bloqueada pelo administrador")
		}

		passwordHash = cUser.PasswordHash
		user = &Usuario{
			ID:       cUser.ID,
			ClientID: cUser.ClientID,
			Nome:     cUser.Name,
			Email:    cUser.Email,
			Role:     cUser.Role,
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	tokenStr, err := s.generateJWT(user)
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

func (s *authService) GetProfile(ctx context.Context, userID, role string) (*ProfileResponse, error) {
	var user *Usuario
	if role == "admin" {
		admin, err := s.repo.GetAdminByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		user = &Usuario{
			ID:    admin.ID,
			Nome:  admin.Name,
			Email: admin.Email,
			Role:  "admin",
		}
	} else {
		cUser, err := s.repo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		user = &Usuario{
			ID:       cUser.ID,
			ClientID: cUser.ClientID,
			Nome:     cUser.Name,
			Email:    cUser.Email,
			Role:     cUser.Role,
		}
	}

	return &ProfileResponse{User: user}, nil
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
		cUser, err := s.repo.GetUserByEmail(ctx, emailVal)
		if err != nil {
			return "", ErrUserNotFound
		}
		userID = cUser.ID
		userType = "client_user"
		userName = cUser.Name
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
		cUser, err := s.repo.GetUserByEmail(ctx, emailVal)
		if err != nil {
			return "", ErrUserNotFound
		}
		if cUser.Status != "active" {
			return "", ErrInactiveUser
		}
		userID = cUser.ID
		userType = "client_user"
		userName = cUser.Name
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
			ID:    admin.ID,
			Nome:  admin.Name,
			Email: admin.Email,
			Role:  "admin",
		}
	} else {
		cUser, err := s.repo.GetUserByID(ctx, token.UserID)
		if err != nil {
			return nil, err
		}
		if cUser.Status != "active" {
			return nil, ErrInactiveUser
		}

		// Verificar se a barbearia está bloqueada
		status, err := s.repo.GetClientStatus(ctx, cUser.ClientID)
		if err == nil && status == "blocked" {
			return nil, errors.New("sua barbearia está suspensa ou bloqueada pelo administrador")
		}

		user = &Usuario{
			ID:       cUser.ID,
			ClientID: cUser.ClientID,
			Nome:     cUser.Name,
			Email:    cUser.Email,
			Role:     cUser.Role,
		}
	}

	tokenJWT, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: tokenJWT,
		User:  user,
	}, nil
}

func (s *authService) generateJWT(user *Usuario) (string, error) {
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &middleware.Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		ClientID: user.ClientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) Impersonate(ctx context.Context, adminID, targetClientID string) (string, error) {
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return "", err
	}
	if admin == nil {
		return "", errors.New("admin não encontrado")
	}

	user := &Usuario{
		ID:       admin.ID,
		ClientID: targetClientID,
		Nome:     admin.Name + " (Impersonated)",
		Email:    admin.Email,
		Role:     "owner",
	}

	return s.generateJWT(user)
}
