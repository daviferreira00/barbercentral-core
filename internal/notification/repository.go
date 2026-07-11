package notification

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotificationNotFound = errors.New("regra de notificação não encontrada")
)

type Repository interface {
	List(ctx context.Context, clientID string) ([]NotificationConfig, error)
	GetByID(ctx context.Context, clientID, id string) (*NotificationConfig, error)
	Create(ctx context.Context, config *NotificationConfig) error
	Update(ctx context.Context, config *NotificationConfig) error
	Delete(ctx context.Context, clientID, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context, clientID string) ([]NotificationConfig, error) {
	var list []NotificationConfig
	query := `
		SELECT nc.*, wi.instance_name AS channel_name
		FROM notification_config nc
		LEFT JOIN whatsapp_instance wi ON nc.channel_id = wi.id
		WHERE nc.client_id = ?
		ORDER BY nc.created_at DESC
	`
	err := r.db.SelectContext(ctx, &list, query, clientID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetByID(ctx context.Context, clientID, id string) (*NotificationConfig, error) {
	var config NotificationConfig
	query := `
		SELECT nc.*, wi.instance_name AS channel_name
		FROM notification_config nc
		LEFT JOIN whatsapp_instance wi ON nc.channel_id = wi.id
		WHERE nc.client_id = ? AND nc.id = ?
	`
	err := r.db.GetContext(ctx, &config, query, clientID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}
		return nil, err
	}
	return &config, nil
}

func (r *repository) Create(ctx context.Context, config *NotificationConfig) error {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	query := `
		INSERT INTO notification_config (id, client_id, name, trigger_type, trigger_value, trigger_unit, message_template, channel_id, active)
		VALUES (:id, :client_id, :name, :trigger_type, :trigger_value, :trigger_unit, :message_template, :channel_id, :active)
	`
	_, err := r.db.NamedExecContext(ctx, query, config)
	return err
}

func (r *repository) Update(ctx context.Context, config *NotificationConfig) error {
	query := `
		UPDATE notification_config
		SET name = :name,
		    trigger_type = :trigger_type,
		    trigger_value = :trigger_value,
		    trigger_unit = :trigger_unit,
		    message_template = :message_template,
		    channel_id = :channel_id,
		    active = :active
		WHERE id = :id AND client_id = :client_id
	`
	_, err := r.db.NamedExecContext(ctx, query, config)
	return err
}

func (r *repository) Delete(ctx context.Context, clientID, id string) error {
	query := `DELETE FROM notification_config WHERE client_id = ? AND id = ?`
	res, err := r.db.ExecContext(ctx, query, clientID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotificationNotFound
	}
	return nil
}
