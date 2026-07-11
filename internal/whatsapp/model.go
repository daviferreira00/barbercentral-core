package whatsapp

import "time"

type WhatsAppInstance struct {
	ID             string     `json:"id" db:"id"`
	InstanceName   string     `json:"instance_name" db:"instance_name"`
	ClientID       *string    `json:"client_id,omitempty" db:"client_id"`
	ProfessionalID *string    `json:"professional_id,omitempty" db:"professional_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	// Join fields
	ClientName       *string `json:"client_name,omitempty" db:"client_name"`
	ClientSlug       *string `json:"client_slug,omitempty" db:"client_slug"`
	ProfessionalName *string `json:"professional_name,omitempty" db:"professional_name"`

	// Evolution status fields (populated on fetch)
	ConnectionStatus string `json:"connection_status,omitempty"`
	OwnerJid         string `json:"owner_jid,omitempty"`
	ProfilePicUrl    string `json:"profile_pic_url,omitempty"`
	Number           string `json:"number,omitempty"`
}

type CreateInstanceRequest struct {
	InstanceName   string  `json:"instance_name"`
	ClientID       *string `json:"client_id,omitempty"`
	ProfessionalID *string `json:"professional_id,omitempty"`
}

type LinkInstanceRequest struct {
	InstanceName   string  `json:"instance_name"`
	ClientID       *string `json:"client_id,omitempty"`
	ProfessionalID *string `json:"professional_id,omitempty"`
}

type EvoInstance struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ConnectionStatus string `json:"connectionStatus"`
	OwnerJid         string `json:"ownerJid"`
	ProfilePicUrl    string `json:"profilePicUrl"`
	Number           string `json:"number"`
}
