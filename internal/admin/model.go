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

type ClientUser struct {
	ID        string    `json:"id" db:"id"`
	ClientID  string    `json:"client_id" db:"client_id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`     // owner, manager, professional, receptionist
	Status    string    `json:"status" db:"status"` // active, inactive
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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

type UpdateClientUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	ClientID     string `json:"client_id"`
	PasswordHash string `json:"password_hash,omitempty"`
}

type AdminUserResponse struct {
	ID         string    `json:"id" db:"id"`
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

