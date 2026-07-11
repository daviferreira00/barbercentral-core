package whatsapp

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	ListInstances(ctx context.Context) ([]WhatsAppInstance, error)
	GetInstanceByName(ctx context.Context, name string) (*WhatsAppInstance, error)
	CreateInstance(ctx context.Context, inst *WhatsAppInstance) error
	UpdateInstanceLink(ctx context.Context, name string, clientID, professionalID *string) error
	DeleteInstance(ctx context.Context, name string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListInstances(ctx context.Context) ([]WhatsAppInstance, error) {
	var list []WhatsAppInstance
	query := `
		SELECT wi.*, c.name AS client_name, c.slug AS client_slug, p.name AS professional_name
		FROM whatsapp_instance wi
		LEFT JOIN client c ON wi.client_id = c.id
		LEFT JOIN professional p ON wi.professional_id = p.id
		ORDER BY wi.instance_name ASC
	`
	err := r.db.SelectContext(ctx, &list, query)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetInstanceByName(ctx context.Context, name string) (*WhatsAppInstance, error) {
	var inst WhatsAppInstance
	query := `
		SELECT wi.*, c.name AS client_name, c.slug AS client_slug, p.name AS professional_name
		FROM whatsapp_instance wi
		LEFT JOIN client c ON wi.client_id = c.id
		LEFT JOIN professional p ON wi.professional_id = p.id
		WHERE wi.instance_name = ?
	`
	err := r.db.GetContext(ctx, &inst, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inst, nil
}

func (r *repository) CreateInstance(ctx context.Context, inst *WhatsAppInstance) error {
	if inst.ID == "" {
		inst.ID = uuid.New().String()
	}
	query := `
		INSERT INTO whatsapp_instance (id, instance_name, client_id, professional_id)
		VALUES (:id, :instance_name, :client_id, :professional_id)
	`
	_, err := r.db.NamedExecContext(ctx, query, inst)
	return err
}

func (r *repository) UpdateInstanceLink(ctx context.Context, name string, clientID, professionalID *string) error {
	query := `
		UPDATE whatsapp_instance
		SET client_id = ?, professional_id = ?
		WHERE instance_name = ?
	`
	_, err := r.db.ExecContext(ctx, query, clientID, professionalID, name)
	return err
}

func (r *repository) DeleteInstance(ctx context.Context, name string) error {
	query := `DELETE FROM whatsapp_instance WHERE instance_name = ?`
	_, err := r.db.ExecContext(ctx, query, name)
	return err
}
