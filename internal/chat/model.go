package chat

import "time"

type MessageDirection string
const (
	DirectionInbound  MessageDirection = "inbound"
	DirectionOutbound MessageDirection = "outbound"
)

type Chat struct {
	ID            string    `json:"id" db:"id"`
	ClientID      string    `json:"client_id" db:"client_id"`
	ContactNumber string    `json:"contact_number" db:"contact_number"`
	ContactName   *string   `json:"contact_name,omitempty" db:"contact_name"`
	LastMessage   *string   `json:"last_message,omitempty" db:"last_message"`
	UnreadCount   int       `json:"unread_count" db:"unread_count"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type Message struct {
	ID        string           `json:"id" db:"id"`
	ChatID    string           `json:"chat_id" db:"chat_id"`
	MessageID string           `json:"message_id" db:"message_id"`
	Direction MessageDirection `json:"direction" db:"direction"`
	Content   string           `json:"content" db:"content"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

type SendMessageRequest struct {
	ContactNumber string `json:"contact_number"`
	Content       string `json:"content"`
}

type WebhookPayload struct {
	Event    string `json:"event"`
	Instance string `json:"instance"`
	Data     struct {
		Key struct {
			RemoteJid string `json:"remoteJid"`
			FromMe    bool   `json:"fromMe"`
			ID        string `json:"id"`
		} `json:"key"`
		Message struct {
			Conversation        string `json:"conversation"`
			ExtendedTextMessage struct {
				Text string `json:"text"`
			} `json:"extendedTextMessage"`
			ButtonsResponseMessage struct {
				SelectedButtonId string `json:"selectedButtonId"`
				DisplayText      string `json:"displayText"`
			} `json:"buttonsResponseMessage"`
		} `json:"message"`
		PushName string `json:"pushName"`
	} `json:"data"`
}
