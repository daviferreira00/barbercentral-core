package admin

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrClientNotFound = errors.New("cliente não encontrado")
)

type AdminRepository interface {
	ListClients(ctx context.Context) ([]Client, error)
	GetClientByID(ctx context.Context, id string) (*Client, error)
	CreateClient(ctx context.Context, c *Client) error
	UpdateClient(ctx context.Context, c *Client) error
	UpdateClientStatus(ctx context.Context, id, status string) error
	ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error)
	CreateClientUser(ctx context.Context, u *ClientUser, passwordHash string) error
}

type adminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) ListClients(ctx context.Context) ([]Client, error) {
	var list []Client
	query := "SELECT * FROM client ORDER BY name ASC"
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *adminRepository) GetClientByID(ctx context.Context, id string) (*Client, error) {
	var c Client
	query := "SELECT * FROM client WHERE id = ?"
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *adminRepository) CreateClient(ctx context.Context, c *Client) error {
	query := `INSERT INTO client (id, plan_id, name, slug, status, created_at)
	          VALUES (:id, :plan_id, :name, :slug, :status, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, c)
	return err
}

func (r *adminRepository) UpdateClient(ctx context.Context, c *Client) error {
	query := `UPDATE client SET plan_id = :plan_id, name = :name, slug = :slug WHERE id = :id`
	res, err := r.db.NamedExecContext(ctx, query, c)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrClientNotFound
	}
	return nil
}

func (r *adminRepository) UpdateClientStatus(ctx context.Context, id, status string) error {
	query := "UPDATE client SET status = ? WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrClientNotFound
	}
	return nil
}

func (r *adminRepository) ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error) {
	var list []ClientUser
	query := "SELECT id, client_id, name, email, role, status, created_at FROM client_user WHERE client_id = ? ORDER BY name ASC"
	err := r.db.SelectContext(ctx, &list, query, clientID)
	return list, err
}

func (r *adminRepository) CreateClientUser(ctx context.Context, u *ClientUser, passwordHash string) error {
	query := `INSERT INTO client_user (id, client_id, name, email, password_hash, role, status, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.ClientID, u.Name, u.Email, passwordHash, u.Role, u.Status, u.CreatedAt)
	return err
}
