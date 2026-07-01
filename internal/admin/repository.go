package admin

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"barbercentral-core/internal/planlimit"
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
	UpdateClientUser(ctx context.Context, userID string, req UpdateClientUserRequest) error
	DeleteClientUser(ctx context.Context, userID string) error
	UpdateClientPlan(ctx context.Context, id, planID string) error
	CreateBlockLog(ctx context.Context, id, clientID, action, reason, performedBy string) error
	ListPlans(ctx context.Context) ([]planlimit.Plan, error)
	CreatePlan(ctx context.Context, p *planlimit.Plan) error
	UpdatePlan(ctx context.Context, p *planlimit.Plan) error
	GetDB() *sqlx.DB
}

type adminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetDB() *sqlx.DB {
	return r.db
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

func (r *adminRepository) UpdateClientUser(ctx context.Context, userID string, req UpdateClientUserRequest) error {
	query := `UPDATE client_user SET name = ?, email = ?, role = ?, status = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, req.Name, req.Email, req.Role, req.Status, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("usuário não encontrado")
	}
	return nil
}

func (r *adminRepository) DeleteClientUser(ctx context.Context, userID string) error {
	query := "DELETE FROM client_user WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("usuário não encontrado")
	}
	return nil
}

func (r *adminRepository) UpdateClientPlan(ctx context.Context, id, planID string) error {
	query := "UPDATE client SET plan_id = ? WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, planID, id)
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

func (r *adminRepository) CreateBlockLog(ctx context.Context, id, clientID, action, reason, performedBy string) error {
	query := `INSERT INTO client_block_log (id, client_id, action, reason, performed_by, created_at)
	          VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, id, clientID, action, reason, performedBy)
	return err
}

func (r *adminRepository) ListPlans(ctx context.Context) ([]planlimit.Plan, error) {
	var list []planlimit.Plan
	query := "SELECT * FROM plan ORDER BY price ASC"
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *adminRepository) CreatePlan(ctx context.Context, p *planlimit.Plan) error {
	query := `INSERT INTO plan (id, name, max_professionals, max_customers, max_users, has_loyalty, has_stock, has_reports, has_online_booking, is_public, price, created_at)
	          VALUES (:id, :name, :max_professionals, :max_customers, :max_users, :has_loyalty, :has_stock, :has_reports, :has_online_booking, :is_public, :price, NOW())`
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
}

func (r *adminRepository) UpdatePlan(ctx context.Context, p *planlimit.Plan) error {
	query := `UPDATE plan 
	          SET name = :name, max_professionals = :max_professionals, max_customers = :max_customers, max_users = :max_users, 
	              has_loyalty = :has_loyalty, has_stock = :has_stock, has_reports = :has_reports, 
	              has_online_booking = :has_online_booking, is_public = :is_public, price = :price 
	          WHERE id = :id`
	res, err := r.db.NamedExecContext(ctx, query, p)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("plano não encontrado")
	}
	return nil
}
