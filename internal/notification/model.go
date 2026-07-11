package notification

import "time"

type TriggerType string
const (
	TriggerBookingConfirmation TriggerType = "booking_confirmation"
	TriggerBookingReminder     TriggerType = "booking_reminder"
	TriggerCustomerRetention   TriggerType = "customer_retention"
)

type TriggerUnit string
const (
	UnitHours TriggerUnit = "hours"
	UnitDays  TriggerUnit = "days"
)

type NotificationConfig struct {
	ID              string      `json:"id" db:"id"`
	ClientID        string      `json:"client_id" db:"client_id"`
	Name            string      `json:"name" db:"name"`
	TriggerType     TriggerType `json:"trigger_type" db:"trigger_type"`
	TriggerValue    int         `json:"trigger_value" db:"trigger_value"`
	TriggerUnit     TriggerUnit `json:"trigger_unit" db:"trigger_unit"`
	MessageTemplate string      `json:"message_template" db:"message_template"`
	ChannelID       *string     `json:"channel_id,omitempty" db:"channel_id"`
	Active          bool        `json:"active" db:"active"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`

	// Join field
	ChannelName *string `json:"channel_name,omitempty" db:"channel_name"`
}

type CreateNotificationRequest struct {
	Name            string      `json:"name"`
	TriggerType     TriggerType `json:"trigger_type"`
	TriggerValue    int         `json:"trigger_value"`
	TriggerUnit     TriggerUnit `json:"trigger_unit"`
	MessageTemplate string      `json:"message_template"`
	ChannelID       *string     `json:"channel_id,omitempty"`
	Active          bool        `json:"active"`
}

type UpdateNotificationRequest struct {
	Name            string      `json:"name"`
	TriggerType     TriggerType `json:"trigger_type"`
	TriggerValue    int         `json:"trigger_value"`
	TriggerUnit     TriggerUnit `json:"trigger_unit"`
	MessageTemplate string      `json:"message_template"`
	ChannelID       *string     `json:"channel_id,omitempty"`
	Active          bool        `json:"active"`
}
