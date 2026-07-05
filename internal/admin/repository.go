package admin

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"barbercentral-core/internal/planlimit"
)

var (
	ErrClientNotFound         = errors.New("cliente não encontrado")
	ErrClientUserLinkNotFound = errors.New("vínculo de usuário não encontrado")
	ErrLinkAlreadyExists      = errors.New("este usuário já possui vínculo com esta barbearia")
)

type AdminRepository interface {
	ListClients(ctx context.Context) ([]Client, error)
	GetClientByID(ctx context.Context, id string) (*Client, error)
	CreateClient(ctx context.Context, c *Client) error
	UpdateClient(ctx context.Context, c *Client) error
	UpdateClientStatus(ctx context.Context, id, status string) error

	ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error)
	FindUserAccountByEmail(ctx context.Context, email string) (*UserAccountRef, error)
	EmailExistsAsAdmin(ctx context.Context, email string) (bool, error)
	GetUserIDByLinkID(ctx context.Context, linkID string) (string, error)
	CreateUserWithLink(ctx context.Context, clientID string, req CreateClientUserRequest, passwordHash string) (*ClientUser, error)
	LinkExistingUser(ctx context.Context, clientID, userID, role string) (*ClientUser, error)
	UpdateClientUserLink(ctx context.Context, linkID string, req UpdateClientUserRequest) error
	DeleteClientUserLink(ctx context.Context, linkID string) error

	UpdateClientPlan(ctx context.Context, id, planID string) error
	CreateBlockLog(ctx context.Context, id, clientID, action, reason, performedBy string) error
	ListPlans(ctx context.Context) ([]planlimit.Plan, error)
	CreatePlan(ctx context.Context, p *planlimit.Plan) error
	UpdatePlan(ctx context.Context, p *planlimit.Plan) error
	GetDB() *sqlx.DB

	ListAllUsers(ctx context.Context) ([]AdminUserResponse, error)
	CheckEmailExists(ctx context.Context, email string, excludeID string) (bool, error)
	CreateAdminUser(ctx context.Context, id, name, email, passwordHash string) error
	UpdateAdminUser(ctx context.Context, id, name, email, passwordHash string) error
	DeleteAdminUser(ctx context.Context, id string) error
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
	query := `INSERT INTO client (id, plan_id, name, slug, custom_domain, status, created_at)
	          VALUES (:id, :plan_id, :name, :slug, :custom_domain, :status, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, c)
	return err
}

func (r *adminRepository) UpdateClient(ctx context.Context, c *Client) error {
	query := `UPDATE client SET plan_id = :plan_id, name = :name, slug = :slug, custom_domain = :custom_domain WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, c)
	return err
}

func (r *adminRepository) UpdateClientStatus(ctx context.Context, id, status string) error {
	query := "UPDATE client SET status = ? WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *adminRepository) ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error) {
	var list []ClientUser
	query := `
		SELECT cul.id, cul.user_id, cul.client_id, ua.name, ua.email, cul.role, cul.status, cul.created_at
		FROM client_user_link cul
		JOIN user_account ua ON ua.id = cul.user_id
		WHERE cul.client_id = ?
		ORDER BY ua.name ASC
	`
	err := r.db.SelectContext(ctx, &list, query, clientID)
	return list, err
}

func (r *adminRepository) FindUserAccountByEmail(ctx context.Context, email string) (*UserAccountRef, error) {
	var u UserAccountRef
	query := `SELECT id, name, email, status FROM user_account WHERE email = ? LIMIT 1`
	err := r.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *adminRepository) EmailExistsAsAdmin(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM platform_admin WHERE email = ?`, email)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *adminRepository) GetUserIDByLinkID(ctx context.Context, linkID string) (string, error) {
	var userID string
	err := r.db.GetContext(ctx, &userID, `SELECT user_id FROM client_user_link WHERE id = ?`, linkID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrClientUserLinkNotFound
		}
		return "", err
	}
	return userID, nil
}

func (r *adminRepository) CreateUserWithLink(ctx context.Context, clientID string, req CreateClientUserRequest, passwordHash string) (*ClientUser, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID := uuid.New().String()
	linkID := uuid.New().String()
	now := time.Now()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_account (id, name, email, password_hash, status, created_at) VALUES (?, ?, ?, ?, 'active', ?)`,
		userID, req.Name, req.Email, passwordHash, now)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO client_user_link (id, user_id, client_id, role, status, created_at) VALUES (?, ?, ?, ?, 'active', ?)`,
		linkID, userID, clientID, req.Role, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ClientUser{
		ID: linkID, UserID: userID, ClientID: clientID,
		Name: req.Name, Email: req.Email, Role: req.Role,
		Status: "active", CreatedAt: now,
	}, nil
}

func (r *adminRepository) LinkExistingUser(ctx context.Context, clientID, userID, role string) (*ClientUser, error) {
	var existingCount int
	err := r.db.GetContext(ctx, &existingCount,
		`SELECT COUNT(*) FROM client_user_link WHERE user_id = ? AND client_id = ?`, userID, clientID)
	if err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, ErrLinkAlreadyExists
	}

	linkID := uuid.New().String()
	now := time.Now()
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO client_user_link (id, user_id, client_id, role, status, created_at) VALUES (?, ?, ?, ?, 'active', ?)`,
		linkID, userID, clientID, role, now)
	if err != nil {
		return nil, err
	}

	var acc UserAccountRef
	_ = r.db.GetContext(ctx, &acc, `SELECT id, name, email, status FROM user_account WHERE id = ?`, userID)

	return &ClientUser{
		ID: linkID, UserID: userID, ClientID: clientID,
		Name: acc.Name, Email: acc.Email, Role: role,
		Status: "active", CreatedAt: now,
	}, nil
}

func (r *adminRepository) UpdateClientUserLink(ctx context.Context, linkID string, req UpdateClientUserRequest) error {
	userID, err := r.GetUserIDByLinkID(ctx, linkID)
	if err != nil {
		return err
	}

	if req.PasswordHash != "" {
		_, err = r.db.ExecContext(ctx,
			`UPDATE user_account SET name = ?, email = ?, password_hash = ? WHERE id = ?`,
			req.Name, req.Email, req.PasswordHash, userID)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE user_account SET name = ?, email = ? WHERE id = ?`,
			req.Name, req.Email, userID)
	}
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE client_user_link SET role = ?, status = ? WHERE id = ?`,
		req.Role, req.Status, linkID)
	return err
}

func (r *adminRepository) DeleteClientUserLink(ctx context.Context, linkID string) error {
	query := "DELETE FROM client_user_link WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrClientUserLinkNotFound
	}
	return nil
}

func (r *adminRepository) UpdateClientPlan(ctx context.Context, id, planID string) error {
	query := "UPDATE client SET plan_id = ? WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, planID, id)
	return err
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
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
}

func (r *adminRepository) ListAllUsers(ctx context.Context) ([]AdminUserResponse, error) {
	var list []AdminUserResponse
	query := `
		SELECT
			cul.id AS id,
			cul.user_id AS user_id,
			'client' as type,
			ua.name,
			ua.email,
			cul.role,
			cul.status,
			cul.client_id,
			c.name as client_name,
			cul.created_at
		FROM client_user_link cul
		JOIN user_account ua ON ua.id = cul.user_id
		LEFT JOIN client c ON c.id = cul.client_id
		UNION ALL
		SELECT
			pa.id AS id,
			pa.id AS user_id,
			'admin' as type,
			pa.name,
			pa.email,
			'admin' as role,
			'active' as status,
			'' as client_id,
			'' as client_name,
			pa.created_at
		FROM platform_admin pa
		ORDER BY name ASC
	`
	err := r.db.SelectContext(ctx, &list, query)
	return list, err
}

func (r *adminRepository) CheckEmailExists(ctx context.Context, email string, excludeID string) (bool, error) {
	var count int
	// Check in user_account
	query := `SELECT COUNT(*) FROM user_account WHERE email = ? AND id != ?`
	err := r.db.GetContext(ctx, &count, query, email, excludeID)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Check in platform_admin
	query = `SELECT COUNT(*) FROM platform_admin WHERE email = ? AND id != ?`
	err = r.db.GetContext(ctx, &count, query, email, excludeID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *adminRepository) CreateAdminUser(ctx context.Context, id, name, email, passwordHash string) error {
	query := `INSERT INTO platform_admin (id, name, email, password_hash, created_at) VALUES (?, ?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, id, name, email, passwordHash)
	return err
}

func (r *adminRepository) UpdateAdminUser(ctx context.Context, id, name, email, passwordHash string) error {
	var query string
	var args []interface{}

	if passwordHash != "" {
		query = `UPDATE platform_admin SET name = ?, email = ?, password_hash = ? WHERE id = ?`
		args = []interface{}{name, email, passwordHash, id}
	} else {
		query = `UPDATE platform_admin SET name = ?, email = ? WHERE id = ?`
		args = []interface{}{name, email, id}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *adminRepository) DeleteAdminUser(ctx context.Context, id string) error {
	query := `DELETE FROM platform_admin WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("admin não encontrado")
	}
	return nil
}
