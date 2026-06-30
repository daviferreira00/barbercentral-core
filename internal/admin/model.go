package admin

import "time"

type Client struct {
	ID        string    `json:"id" db:"id"`
	PlanID    string    `json:"plan_id" db:"plan_id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	Status    string    `json:"status" db:"status"` // active, blocked
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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
	PlanID string `json:"plan_id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
}

type UpdateClientRequest struct {
	PlanID string `json:"plan_id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
}

type CreateClientUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"password"` // Default local password for dev fallback
}
