package chat

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	ListChats(ctx context.Context, clientID string) ([]Chat, error)
	GetChatByID(ctx context.Context, clientID, chatID string) (*Chat, error)
	GetOrCreateChat(ctx context.Context, clientID, contactNumber, contactName string) (*Chat, error)
	ResetUnreadCount(ctx context.Context, clientID, chatID string) error
	SaveMessage(ctx context.Context, msg *Message) error
	ListMessages(ctx context.Context, chatID string) ([]Message, error)
	HasMessage(ctx context.Context, messageID string) (bool, error)
	UpdateChatLastMessage(ctx context.Context, chatID string, lastMessage string, incrementUnread bool) error

	// Appointment helpers for buttons response automation
	GetLatestPendingAppointment(ctx context.Context, clientID, cleanPhone string) (string, error)
	UpdateAppointmentStatus(ctx context.Context, clientID, appointmentID, status string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListChats(ctx context.Context, clientID string) ([]Chat, error) {
	list := []Chat{}
	query := `
		SELECT * FROM whatsapp_chat
		WHERE client_id = ?
		ORDER BY updated_at DESC
	`
	err := r.db.SelectContext(ctx, &list, query, clientID)
	return list, err
}

func (r *repository) GetChatByID(ctx context.Context, clientID, chatID string) (*Chat, error) {
	var c Chat
	query := `SELECT * FROM whatsapp_chat WHERE client_id = ? AND id = ?`
	err := r.db.GetContext(ctx, &c, query, clientID, chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("chat não encontrado")
		}
		return nil, err
	}
	return &c, nil
}

func (r *repository) GetOrCreateChat(ctx context.Context, clientID, contactNumber, contactName string) (*Chat, error) {
	var c Chat
	querySelect := `SELECT * FROM whatsapp_chat WHERE client_id = ? AND contact_number = ?`
	err := r.db.GetContext(ctx, &c, querySelect, clientID, contactNumber)
	if err == nil {
		return &c, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Create
	newChat := Chat{
		ID:            uuid.New().String(),
		ClientID:      clientID,
		ContactNumber: contactNumber,
		ContactName:   &contactName,
		UnreadCount:   0,
		UpdatedAt:     time.Now(),
		CreatedAt:     time.Now(),
	}

	queryInsert := `
		INSERT INTO whatsapp_chat (id, client_id, contact_number, contact_name, last_message, unread_count, updated_at, created_at)
		VALUES (:id, :client_id, :contact_number, :contact_name, :last_message, :unread_count, :updated_at, :created_at)
	`
	_, err = r.db.NamedExecContext(ctx, queryInsert, &newChat)
	if err != nil {
		return nil, err
	}

	return &newChat, nil
}

func (r *repository) ResetUnreadCount(ctx context.Context, clientID, chatID string) error {
	query := `UPDATE whatsapp_chat SET unread_count = 0 WHERE client_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, clientID, chatID)
	return err
}

func (r *repository) SaveMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	query := `
		INSERT INTO whatsapp_message (id, chat_id, message_id, direction, content, created_at)
		VALUES (:id, :chat_id, :message_id, :direction, :content, :created_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, msg)
	return err
}

func (r *repository) ListMessages(ctx context.Context, chatID string) ([]Message, error) {
	list := []Message{}
	query := `
		SELECT * FROM whatsapp_message
		WHERE chat_id = ?
		ORDER BY created_at ASC
	`
	err := r.db.SelectContext(ctx, &list, query, chatID)
	return list, err
}

func (r *repository) UpdateChatLastMessage(ctx context.Context, chatID string, lastMessage string, incrementUnread bool) error {
	var query string
	if incrementUnread {
		query = `
			UPDATE whatsapp_chat
			SET last_message = ?, unread_count = unread_count + 1, updated_at = NOW()
			WHERE id = ?
		`
	} else {
		query = `
			UPDATE whatsapp_chat
			SET last_message = ?, updated_at = NOW()
			WHERE id = ?
		`
	}
	_, err := r.db.ExecContext(ctx, query, lastMessage, chatID)
	return err
}

func (r *repository) GetLatestPendingAppointment(ctx context.Context, clientID, cleanPhone string) (string, error) {
	var appID string
	query := `
		SELECT id FROM appointment
		WHERE client_id = ?
		  AND (customer_phone = ? OR customer_phone LIKE ?)
		  AND status = 'pending'
		ORDER BY date DESC, start_time DESC
		LIMIT 1
	`
	phoneLike := "%" + cleanPhone
	err := r.db.GetContext(ctx, &appID, query, clientID, cleanPhone, phoneLike)
	if err != nil {
		return "", err
	}
	return appID, nil
}

func (r *repository) UpdateAppointmentStatus(ctx context.Context, clientID, appointmentID, status string) error {
	query := `UPDATE appointment SET status = ? WHERE client_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, status, clientID, appointmentID)
	return err
}

func (r *repository) HasMessage(ctx context.Context, messageID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM whatsapp_message WHERE message_id = ?)`
	err := r.db.GetContext(ctx, &exists, query, messageID)
	return exists, err
}
