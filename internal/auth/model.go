package auth

import "time"

// Entidades do banco de dados

type ClientUser struct {
	ID           string    `db:"id"`
	ClientID     string    `db:"client_id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"` // owner, manager, professional, receptionist
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
}

type PlatformAdmin struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type AuthToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	UserType  string    `db:"user_type"` // client_user, platform_admin
	Type      string    `db:"type"`      // magic_link, password_reset
	Token     string    `db:"token"`
	Used      int       `db:"used"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// Representação unificada do usuário logado na sessão

type Usuario struct {
	ID       string `json:"id"`
	ClientID string `json:"client_id,omitempty"`
	Nome     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"` // admin, owner, manager, professional, receptionist
}

// Structs de requisição e resposta para os endpoints de API

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  *Usuario `json:"user"`
}

type MagicLinkRequest struct {
	Email string `json:"email"`
}

type MagicLinkVerifyRequest struct {
	Token string `json:"token"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetConfirmRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type ProfileResponse struct {
	User *Usuario `json:"user"`
}
