package service

import "time"

type ServiceCategory struct {
	ID       string `json:"id" db:"id"`
	ClientID string `json:"client_id" db:"client_id"`
	Name     string `json:"name" db:"name"`
}

type Service struct {
	ID              string    `json:"id" db:"id"`
	ClientID        string    `json:"client_id" db:"client_id"`
	CategoryID      *string   `json:"category_id,omitempty" db:"category_id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description,omitempty" db:"description"`
	DurationMinutes int       `json:"duration_minutes" db:"duration_minutes"`
	Price           float64   `json:"price" db:"price"`
	PhotoURL        *string   `json:"photo_url,omitempty" db:"photo_url"`
	Active          int       `json:"active" db:"active"` // 1=true, 0=false
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type CreateServiceRequest struct {
	CategoryID      *string  `json:"category_id"`
	Name            string   `json:"name"`
	Description     *string  `json:"description"`
	DurationMinutes int      `json:"duration_minutes"`
	Price           float64  `json:"price"`
	Active          int      `json:"active"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}
