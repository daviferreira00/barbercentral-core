package customer

import (
	"time"
)

type Customer struct {
	ID        string    `db:"id" json:"id"`
	ClientID  string    `db:"client_id" json:"client_id"`
	Name      string    `db:"name" json:"name"`
	Phone     string    `db:"phone" json:"phone"`
	Email     *string   `db:"email" json:"email"`
	CPF       *string   `db:"cpf" json:"cpf"`
	BirthDate *string   `db:"birth_date" json:"birth_date"`
	Notes     *string   `db:"notes" json:"notes"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CustomerStats struct {
	Customer
	TotalVisits int     `db:"total_visits" json:"total_visits"`
	TotalSpent  float64 `db:"total_spent" json:"total_spent"`
	LastVisit   *string `db:"last_visit" json:"last_visit"`
	FirstVisit  *string `db:"first_visit" json:"first_visit"`
}

type CreateCustomerRequest struct {
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Email     *string `json:"email"`
	CPF       *string `json:"cpf"`
	BirthDate *string `json:"birth_date"`
	Notes     *string `json:"notes"`
}

type UpdateCustomerRequest struct {
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Email     *string `json:"email"`
	CPF       *string `json:"cpf"`
	BirthDate *string `json:"birth_date"`
	Notes     *string `json:"notes"`
}

type EnrichedAppointmentHistory struct {
	ID               string   `db:"id" json:"id"`
	Date             string   `db:"date" json:"date"`
	StartTime        string   `db:"start_time" json:"start_time"`
	EndTime          string   `db:"end_time" json:"end_time"`
	Status           string   `db:"status" json:"status"`
	ProfessionalName string   `db:"professional_name" json:"professional_name"`
	ServicesList     string   `db:"services_list" json:"services_list"`
	TotalPrice       float64  `db:"total_price" json:"total_price"`
	PaymentStatus    *string  `db:"payment_status" json:"payment_status"`
	PaymentMethod    *string  `db:"payment_method" json:"payment_method"`
}
