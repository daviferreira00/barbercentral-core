package auth

import "time"

// Entidades do banco de dados

// UserAccount é a identidade global do usuário (login), independente de barbearia.
// Antes chamada client_user (que tinha client_id/role fixos, 1:1 com uma barbearia).
type UserAccount struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Status       string    `db:"status"` // active, inactive, pending — status GLOBAL da conta
	CreatedAt    time.Time `db:"created_at"`
	PhotoURL     *string   `db:"photo_url" json:"photo_url,omitempty"`
}

// ClientUserLink é o vínculo entre um usuário e uma barbearia específica.
// Um usuário pode ter múltiplos vínculos ativos (uma barbearia cada), cada um
// com seu próprio role e status.
type ClientUserLink struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	ClientID  string    `db:"client_id"`
	Role      string    `db:"role"`   // owner, manager, professional, receptionist
	Status    string    `db:"status"` // status DESTE vínculo específico
	CreatedAt time.Time `db:"created_at"`
}

// ClientMembership enriquece o vínculo com dados da barbearia, usada para
// alimentar o seletor de barbearias (GET /auth/my-clients).
type ClientMembership struct {
	LinkID     string `db:"link_id" json:"link_id"`
	ClientID   string `db:"client_id" json:"client_id"`
	ClientName string `db:"client_name" json:"client_name"`
	ClientSlug string `db:"client_slug" json:"client_slug"`
	Role       string `db:"role" json:"role"`
	Status     string `db:"status" json:"status"`
}

type PlatformAdmin struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	PhotoURL     *string   `db:"photo_url" json:"photo_url,omitempty"`
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
	ID                   string     `json:"id"`
	ClientID             string     `json:"client_id,omitempty"`
	Nome                 string     `json:"name"`
	Email                string     `json:"email"`
	Role                 string     `json:"role"` // admin, owner, manager, professional, receptionist, ou "" (aguardando seleção)
	Impersonating        bool       `json:"impersonating,omitempty"`
	NeedsClientSelection bool       `json:"needs_client_selection,omitempty"`
	PhotoURL             *string    `json:"photo_url,omitempty"`
	CompletedOnboarding  bool       `json:"completed_onboarding"`
	SubscriptionStatus   string     `json:"subscription_status,omitempty"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty"`
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

type RegisterTrialRequest struct {
	BarberName string `json:"barber_name"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
}

type WizardOnboardingRequest struct {
	BarberName  string  `json:"barber_name"`
	Phone       string  `json:"phone"`
	ServiceName string  `json:"service_name,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Duration    int     `json:"duration,omitempty"`
	OwnerName   string  `json:"owner_name,omitempty"`
}

type MarkTutorialSeenRequest struct {
	TutorialID string `json:"tutorial_id"`
}

type OnboardingStatusResponse struct {
	CompletedOnboarding bool       `json:"completed_onboarding"`
	SeenTutorials       []string   `json:"seen_tutorials"`
	SubscriptionStatus  string     `json:"subscription_status"`
	TrialEndsAt         *time.Time `json:"trial_ends_at,omitempty"`
	DaysRemaining       int        `json:"days_remaining"`
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

type SwitchClientRequest struct {
	ClientID string `json:"client_id"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdateProfileRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password,omitempty"`
	PhotoBase64 string `json:"photo_base64,omitempty"`
}


