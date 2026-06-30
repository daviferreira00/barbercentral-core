package professional

import "time"

type Professional struct {
	ID        string    `json:"id" db:"id"`
	ClientID  string    `json:"client_id" db:"client_id"`
	UserID    *string   `json:"user_id,omitempty" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Bio       *string   `json:"bio,omitempty" db:"bio"`
	PhotoURL  *string   `json:"photo_url,omitempty" db:"photo_url"`
	Status    string    `json:"status" db:"status"` // active, inactive
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ProfessionalSchedule struct {
	ID             string `json:"id" db:"id"`
	ProfessionalID string `json:"professional_id" db:"professional_id"`
	ClientID       string `json:"client_id" db:"client_id"`
	Weekday        int    `json:"weekday" db:"weekday"` // 0=Dom, 1=Seg, ..., 6=Sab
	StartTime      string `json:"start_time" db:"start_time"`
	EndTime        string `json:"end_time" db:"end_time"`
	Enabled        int    `json:"enabled" db:"enabled"` // 0=false, 1=true
}

type ProfessionalServiceLink struct {
	ProfessionalID string   `json:"professional_id" db:"professional_id"`
	ServiceID      string   `json:"service_id" db:"service_id"`
	ClientID       string   `json:"client_id" db:"client_id"`
	CustomPrice    *float64 `json:"custom_price,omitempty" db:"custom_price"`
	CustomDuration *int     `json:"custom_duration,omitempty" db:"custom_duration"`
}

type CreateProfessionalRequest struct {
	Name   string  `json:"name"`
	Bio    *string `json:"bio"`
	Status string  `json:"status"`
}

type UpdateProfessionalRequest struct {
	Name   string  `json:"name"`
	Bio    *string `json:"bio"`
	Status string  `json:"status"`
}

type BulkUpdateScheduleRequest struct {
	Schedules []ProfessionalSchedule `json:"schedules"`
}

type LinkServiceRequest struct {
	ServiceID      string   `json:"service_id"`
	CustomPrice    *float64 `json:"custom_price"`
	CustomDuration *int     `json:"custom_duration"`
}
