package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound  = errors.New("usuário não encontrado")
	ErrTokenNotFound = errors.New("token não encontrado")
)

type AuthRepository interface {
	GetAdminByEmail(ctx context.Context, email string) (*PlatformAdmin, error)
	GetUserByEmail(ctx context.Context, email string) (*ClientUser, error)
	GetAdminByID(ctx context.Context, id string) (*PlatformAdmin, error)
	GetUserByID(ctx context.Context, id string) (*ClientUser, error)

	// Password Resets & Magic Links Tokens
	CreateToken(ctx context.Context, token *AuthToken) error
	GetToken(ctx context.Context, tokenStr string) (*AuthToken, error)
	MarkTokenUsed(ctx context.Context, tokenStr string) error

	// Password Updates
	UpdateAdminPassword(ctx context.Context, id string, passwordHash string) error
	UpdateUserPassword(ctx context.Context, id string, passwordHash string) error
	GetClientStatus(ctx context.Context, clientID string) (string, error)
}

type authRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetAdminByEmail(ctx context.Context, email string) (*PlatformAdmin, error) {
	var admin PlatformAdmin
	query := `SELECT id, name, email, password_hash, created_at FROM platform_admin WHERE email = ? LIMIT 1`
	err := r.db.GetContext(ctx, &admin, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*ClientUser, error) {
	var user ClientUser
	query := `SELECT id, client_id, name, email, password_hash, role, status, created_at FROM client_user WHERE email = ? LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetAdminByID(ctx context.Context, id string) (*PlatformAdmin, error) {
	var admin PlatformAdmin
	query := `SELECT id, name, email, password_hash, created_at FROM platform_admin WHERE id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &admin, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id string) (*ClientUser, error) {
	var user ClientUser
	query := `SELECT id, client_id, name, email, password_hash, role, status, created_at FROM client_user WHERE id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) CreateToken(ctx context.Context, token *AuthToken) error {
	query := `INSERT INTO auth_token (id, user_id, user_type, type, token, used, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.UserType, token.Type, token.Token, token.Used, token.ExpiresAt)
	return err
}

func (r *authRepository) GetToken(ctx context.Context, tokenStr string) (*AuthToken, error) {
	var token AuthToken
	query := `SELECT id, user_id, user_type, type, token, used, expires_at, created_at FROM auth_token WHERE token = ? LIMIT 1`
	err := r.db.GetContext(ctx, &token, query, tokenStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

func (r *authRepository) MarkTokenUsed(ctx context.Context, tokenStr string) error {
	query := `UPDATE auth_token SET used = 1 WHERE token = ?`
	_, err := r.db.ExecContext(ctx, query, tokenStr)
	return err
}

func (r *authRepository) UpdateAdminPassword(ctx context.Context, id string, passwordHash string) error {
	query := `UPDATE platform_admin SET password_hash = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	return err
}

func (r *authRepository) UpdateUserPassword(ctx context.Context, id string, passwordHash string) error {
	query := `UPDATE client_user SET password_hash = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	return err
}

func (r *authRepository) GetClientStatus(ctx context.Context, clientID string) (string, error) {
	var status string
	query := `SELECT status FROM client WHERE id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &status, query, clientID)
	if err != nil {
		return "", err
	}
	return status, nil
}
