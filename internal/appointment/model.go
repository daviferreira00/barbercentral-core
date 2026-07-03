package appointment

import "time"

type BlockedSlot struct {
	ID             string    `json:"id" db:"id"`
	ClientID       string    `json:"client_id" db:"client_id"`
	ProfessionalID *string   `json:"professional_id,omitempty" db:"professional_id"`
	Date           string    `json:"date" db:"date"`             // YYYY-MM-DD
	StartTime      string    `json:"start_time" db:"start_time"` // HH:MM:SS
	EndTime        string    `json:"end_time" db:"end_time"`     // HH:MM:SS
	Reason         *string   `json:"reason,omitempty" db:"reason"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type Appointment struct {
	ID             string    `json:"id" db:"id"`
	ClientID       string    `json:"client_id" db:"client_id"`
	ProfessionalID string    `json:"professional_id" db:"professional_id"`
	CustomerID     *string   `json:"customer_id,omitempty" db:"customer_id"`
	Date           string    `json:"date" db:"date"`             // YYYY-MM-DD
	StartTime      string    `json:"start_time" db:"start_time"` // HH:MM:SS
	EndTime        string    `json:"end_time" db:"end_time"`     // HH:MM:SS
	Status         string    `json:"status" db:"status"`         // pending, confirmed, in_progress, completed, cancelled, no_show
	Notes          *string   `json:"notes,omitempty" db:"notes"`
	CancelToken    *string   `json:"cancel_token,omitempty" db:"cancel_token"`
	CustomerName   *string   `json:"customer_name,omitempty" db:"customer_name"`
	CustomerPhone  *string   `json:"customer_phone,omitempty" db:"customer_phone"`
	CustomerEmail  *string   `json:"customer_email,omitempty" db:"customer_email"`
	ReminderSent   int       `json:"reminder_sent" db:"reminder_sent"` // 0=false, 1=true
	Source         string    `json:"source" db:"source"`               // panel, online
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type AppointmentService struct {
	ID              string  `json:"id" db:"id"`
	AppointmentID   string  `json:"appointment_id" db:"appointment_id"`
	ServiceID       string  `json:"service_id" db:"service_id"`
	Price           float64 `json:"price" db:"price"`
	DurationMinutes int     `json:"duration_minutes" db:"duration_minutes"`
}

// Modelos enriquecidos para exibição na Agenda
type EnrichedAppointment struct {
	Appointment
	ProfessionalName string               `json:"professional_name"`
	Services         []EnrichedAppService `json:"services"`
	StartedAt        *time.Time           `json:"started_at,omitempty"`
}

type EnrichedAppService struct {
	ServiceID   string  `json:"service_id"`
	ServiceName string  `json:"service_name"`
	Price       float64 `json:"price"`
	Duration    int     `json:"duration_minutes"`
}

type CreateBlockedSlotRequest struct {
	ProfessionalID *string `json:"professional_id"`
	Date           string  `json:"date"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Reason         *string `json:"reason"`
}

type UpdateStatusRequest struct {
	Status string  `json:"status"`
	Notes  *string `json:"notes,omitempty"`
}

type CreatePublicAppointmentRequest struct {
	ProfessionalID string   `json:"professional_id"`
	ServiceIDs     []string `json:"service_ids"`
	Date           string   `json:"date"`
	StartTime      string   `json:"start_time"`
	CustomerName   string   `json:"customer_name"`
	CustomerPhone  string   `json:"customer_phone"`
	CustomerEmail  string   `json:"customer_email"`
	Notes          *string  `json:"notes,omitempty"`
}

type TimeSlot struct {
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
}

type AppointmentPayment struct {
	ID            string     `json:"id" db:"id"`
	AppointmentID string     `json:"appointment_id" db:"appointment_id"`
	ClientID      string     `json:"client_id" db:"client_id"`
	Amount        float64    `json:"amount" db:"amount"`
	Method        string     `json:"method" db:"method"`
	Status        string     `json:"status" db:"status"`
	PaidAt        *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	Notes         *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type AppointmentStatusLog struct {
	ID            string    `json:"id" db:"id"`
	AppointmentID string    `json:"appointment_id" db:"appointment_id"`
	FromStatus    *string   `json:"from_status,omitempty" db:"from_status"`
	ToStatus      string    `json:"to_status" db:"to_status"`
	ChangedBy     string    `json:"changed_by" db:"changed_by"`
	Notes         *string   `json:"notes,omitempty" db:"notes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type CreateAppointmentRequest struct {
	CustomerID     *string  `json:"customer_id,omitempty"`
	CustomerName   *string  `json:"customer_name,omitempty"`
	CustomerPhone  *string  `json:"customer_phone,omitempty"`
	CustomerEmail  *string  `json:"customer_email,omitempty"`
	ProfessionalID string   `json:"professional_id"`
	ServiceIDs     []string `json:"service_ids"`
	Date           string   `json:"date"`
	StartTime      string   `json:"start_time"`
	Notes          *string  `json:"notes,omitempty"`
}

type UpdateAppointmentRequest struct {
	ProfessionalID string `json:"professional_id"`
	Date           string `json:"date"`
	StartTime      string `json:"start_time"`
}


