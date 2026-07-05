package admin

import "time"

type Client struct {
	ID           string    `json:"id" db:"id"`
	PlanID       string    `json:"plan_id" db:"plan_id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	CustomDomain *string   `json:"custom_domain" db:"custom_domain"`
	Status       string    `json:"status" db:"status"` // active, blocked
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ClientUser representa 1 VÍNCULO (client_user_link), não mais 1 usuário 1:1.
// Um mesmo e-mail (user_id) pode aparecer em várias linhas, uma por barbearia.
type ClientUser struct {
	ID        string    `json:"id" db:"id"` // = link_id (client_user_link.id)
	UserID    string    `json:"user_id" db:"user_id"`
	ClientID  string    `json:"client_id" db:"client_id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`     // owner, manager, professional, receptionist
	Status    string    `json:"status" db:"status"` // active, inactive
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserAccountRef é uma referência mínima à identidade global do usuário,
// usada para decidir se um e-mail já existe na plataforma (achar ou criar).
type UserAccountRef struct {
	ID     string `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Status string `db:"status"`
}

type CreateClientRequest struct {
	PlanID       string  `json:"plan_id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	CustomDomain *string `json:"custom_domain"`
}

type UpdateClientRequest struct {
	PlanID       string  `json:"plan_id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	CustomDomain *string `json:"custom_domain"`
}

type CreateClientUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"password"` // Default local password for dev fallback
}

type BlockRequest struct {
	Reason string `json:"reason"`
}

type UpdatePlanRequest struct {
	PlanID string `json:"plan_id"`
}

// UpdateClientUserRequest atualiza um vínculo específico (identificado por
// link_id). Name/Email/PasswordHash afetam a identidade GLOBAL do usuário
// (user_account); Role/Status afetam só este vínculo (client_user_link).
type UpdateClientUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	PasswordHash string `json:"password_hash,omitempty"`
}

type AdminUserResponse struct {
	ID         string    `json:"id" db:"id"` // link_id (client) ou o próprio id (admin)
	UserID     string    `json:"user_id,omitempty" db:"user_id"`
	Type       string    `json:"type" db:"type"` // "admin" or "client"
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Role       string    `json:"role,omitempty" db:"role"`
	Status     string    `json:"status,omitempty" db:"status"`
	ClientID   string    `json:"client_id,omitempty" db:"client_id"`
	ClientName string    `json:"client_name,omitempty" db:"client_name"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CreateUserRequest struct {
	Type     string `json:"type"` // "admin" or "client"
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
	ClientID string `json:"client_id,omitempty"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"` // optional password
	Role     string `json:"role,omitempty"`
	Status   string `json:"status,omitempty"`
	ClientID string `json:"client_id,omitempty"`
}

