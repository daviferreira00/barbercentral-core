package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound  = errors.New("usuário não encontrado")
	ErrTokenNotFound = errors.New("token não encontrado")
)

type AuthRepository interface {
	GetAdminByEmail(ctx context.Context, email string) (*PlatformAdmin, error)
	GetUserAccountByEmail(ctx context.Context, email string) (*UserAccount, error)
	GetAdminByID(ctx context.Context, id string) (*PlatformAdmin, error)
	GetUserAccountByID(ctx context.Context, id string) (*UserAccount, error)
	UpdateAdminProfile(ctx context.Context, id, name, email, passwordHash string, photoURL *string) error
	UpdateUserProfile(ctx context.Context, id, name, email, passwordHash string, photoURL *string) error

	// Vínculos usuário↔barbearia
	ListActiveMemberships(ctx context.Context, userID string) ([]ClientUserLink, error)
	GetMembership(ctx context.Context, userID, clientID string) (*ClientUserLink, error)
	ListMembershipsWithClient(ctx context.Context, userID string) ([]ClientMembership, error)

	// Password Resets & Magic Links Tokens
	CreateToken(ctx context.Context, token *AuthToken) error
	GetToken(ctx context.Context, tokenStr string) (*AuthToken, error)
	MarkTokenUsed(ctx context.Context, tokenStr string) error

	// Password Updates
	UpdateAdminPassword(ctx context.Context, id string, passwordHash string) error
	UpdateUserPassword(ctx context.Context, id string, passwordHash string) error
	GetClientStatus(ctx context.Context, clientID string) (string, error)

	// Trial & Onboarding
	GetDefaultTrialPlanID(ctx context.Context) (string, error)
	RegisterTrialClient(ctx context.Context, clientID, planID, name, slug, phone string, trialEndsAt time.Time) error
	CreateUserAccount(ctx context.Context, userID, name, email, passwordHash string) error
	CreateClientUserLink(ctx context.Context, linkID, userID, clientID, role, status string) error
	CreateDefaultClientConfig(ctx context.Context, clientID, phone string) error
	GetOnboardingData(ctx context.Context, userID, clientID string) (completedOnboarding bool, seenTutorialsJSON string, subscriptionStatus string, trialEndsAt *time.Time, err error)
	CompleteOnboarding(ctx context.Context, userID, clientID string) error
	UpdateSeenTutorials(ctx context.Context, userID, clientID string, jsonStr string) error
}

type authRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetAdminByEmail(ctx context.Context, email string) (*PlatformAdmin, error) {
	var admin PlatformAdmin
	query := `SELECT id, name, email, password_hash, created_at, photo_url FROM platform_admin WHERE email = ? LIMIT 1`
	err := r.db.GetContext(ctx, &admin, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *authRepository) GetUserAccountByEmail(ctx context.Context, email string) (*UserAccount, error) {
	var user UserAccount
	query := `SELECT id, name, email, password_hash, status, created_at, photo_url FROM user_account WHERE email = ? LIMIT 1`
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
	query := `SELECT id, name, email, password_hash, created_at, photo_url FROM platform_admin WHERE id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &admin, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *authRepository) GetUserAccountByID(ctx context.Context, id string) (*UserAccount, error) {
	var user UserAccount
	query := `SELECT id, name, email, password_hash, status, created_at, photo_url FROM user_account WHERE id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) ListActiveMemberships(ctx context.Context, userID string) ([]ClientUserLink, error) {
	var links []ClientUserLink
	query := `SELECT id, user_id, client_id, role, status, created_at FROM client_user_link WHERE user_id = ? AND status = 'active'`
	err := r.db.SelectContext(ctx, &links, query, userID)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (r *authRepository) GetMembership(ctx context.Context, userID, clientID string) (*ClientUserLink, error) {
	var link ClientUserLink
	query := `SELECT id, user_id, client_id, role, status, created_at FROM client_user_link WHERE user_id = ? AND client_id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &link, query, userID, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &link, nil
}

func (r *authRepository) ListMembershipsWithClient(ctx context.Context, userID string) ([]ClientMembership, error) {
	var list []ClientMembership
	query := `
		SELECT cul.id AS link_id, cul.client_id, c.name AS client_name, c.slug AS client_slug, cul.role, cul.status
		FROM client_user_link cul
		JOIN client c ON c.id = cul.client_id
		WHERE cul.user_id = ? AND cul.status = 'active'
		ORDER BY c.name ASC
	`
	err := r.db.SelectContext(ctx, &list, query, userID)
	if err != nil {
		return nil, err
	}
	return list, nil
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
	query := `UPDATE user_account SET password_hash = ? WHERE id = ?`
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

func (r *authRepository) UpdateAdminProfile(ctx context.Context, id, name, email, passwordHash string, photoURL *string) error {
	var err error
	if passwordHash != "" {
		if photoURL != nil {
			query := `UPDATE platform_admin SET name = ?, email = ?, password_hash = ?, photo_url = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, passwordHash, photoURL, id)
		} else {
			query := `UPDATE platform_admin SET name = ?, email = ?, password_hash = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, passwordHash, id)
		}
	} else {
		if photoURL != nil {
			query := `UPDATE platform_admin SET name = ?, email = ?, photo_url = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, photoURL, id)
		} else {
			query := `UPDATE platform_admin SET name = ?, email = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, id)
		}
	}
	return err
}

func (r *authRepository) UpdateUserProfile(ctx context.Context, id, name, email, passwordHash string, photoURL *string) error {
	var err error
	if passwordHash != "" {
		if photoURL != nil {
			query := `UPDATE user_account SET name = ?, email = ?, password_hash = ?, photo_url = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, passwordHash, photoURL, id)
		} else {
			query := `UPDATE user_account SET name = ?, email = ?, password_hash = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, passwordHash, id)
		}
	} else {
		if photoURL != nil {
			query := `UPDATE user_account SET name = ?, email = ?, photo_url = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, photoURL, id)
		} else {
			query := `UPDATE user_account SET name = ?, email = ? WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, name, email, id)
		}
	}
	return err
}

func (r *authRepository) GetDefaultTrialPlanID(ctx context.Context) (string, error) {
	var planID string
	query := `SELECT id FROM plan WHERE id = 'plan-pro' OR name LIKE '%Pro%' ORDER BY (CASE WHEN id = 'plan-pro' THEN 1 WHEN name LIKE '%Pro%' THEN 2 ELSE 3 END) LIMIT 1`
	err := r.db.GetContext(ctx, &planID, query)
	if err != nil || planID == "" {
		fallbackQuery := `SELECT id FROM plan LIMIT 1`
		_ = r.db.GetContext(ctx, &planID, fallbackQuery)
	}
	if planID == "" {
		planID = "plan-pro"
	}
	return planID, nil
}

func (r *authRepository) RegisterTrialClient(ctx context.Context, clientID, planID, name, slug, phone string, trialEndsAt time.Time) error {
	query := `
		INSERT INTO client (id, plan_id, name, slug, status, trial_ends_at, subscription_status, phone, created_at)
		VALUES (?, ?, ?, ?, 'active', ?, 'trial', ?, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, clientID, planID, name, slug, trialEndsAt, phone)
	if err != nil {
		// Fallback caso a tabela client ainda não possua trial_ends_at ou subscription_status ou phone
		fallbackQuery := `
			INSERT INTO client (id, plan_id, name, slug, status, created_at)
			VALUES (?, ?, ?, ?, 'active', NOW())
		`
		_, err = r.db.ExecContext(ctx, fallbackQuery, clientID, planID, name, slug)
	}
	return err
}

func (r *authRepository) CreateUserAccount(ctx context.Context, userID, name, email, passwordHash string) error {
	query := `
		INSERT INTO user_account (id, name, email, password_hash, status, created_at)
		VALUES (?, ?, ?, ?, 'active', NOW())
	`
	_, err := r.db.ExecContext(ctx, query, userID, name, email, passwordHash)
	return err
}

func (r *authRepository) CreateClientUserLink(ctx context.Context, linkID, userID, clientID, role, status string) error {
	query := `
		INSERT INTO client_user_link (id, user_id, client_id, role, status, completed_onboarding, created_at)
		VALUES (?, ?, ?, ?, ?, 0, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, linkID, userID, clientID, role, status)
	if err != nil {
		fallbackQuery := `
			INSERT INTO client_user_link (id, user_id, client_id, role, status, created_at)
			VALUES (?, ?, ?, ?, ?, NOW())
		`
		_, err = r.db.ExecContext(ctx, fallbackQuery, linkID, userID, clientID, role, status)
	}
	return err
}

func (r *authRepository) CreateDefaultClientConfig(ctx context.Context, clientID, phone string) error {
	query := `
		INSERT INTO client_config (client_id, color_primary, color_secondary, timezone, cancellation_policy_hours, booking_requires_login, min_advance_hours, max_advance_days, interval_between_minutes, active, phone, whatsapp)
		VALUES (?, '#0F172A', '#D97706', 'America/Sao_Paulo', 2, 0, 1, 30, 30, 1, ?, ?)
		ON DUPLICATE KEY UPDATE phone = VALUES(phone), whatsapp = VALUES(whatsapp)
	`
	_, err := r.db.ExecContext(ctx, query, clientID, phone, phone)
	return err
}

func (r *authRepository) GetOnboardingData(ctx context.Context, userID, clientID string) (completedOnboarding bool, seenTutorialsJSON string, subscriptionStatus string, trialEndsAt *time.Time, err error) {
	var row struct {
		CompletedOnboarding int        `db:"completed_onboarding"`
		SeenTutorialsJSON   *string    `db:"seen_tutorials_json"`
		SubscriptionStatus  string     `db:"subscription_status"`
		TrialEndsAt         *time.Time `db:"trial_ends_at"`
	}
	query := `
		SELECT cul.completed_onboarding, cul.seen_tutorials_json, c.subscription_status, c.trial_ends_at
		FROM client_user_link cul
		JOIN client c ON c.id = cul.client_id
		WHERE cul.user_id = ? AND cul.client_id = ?
		LIMIT 1
	`
	err = r.db.GetContext(ctx, &row, query, userID, clientID)
	if err != nil {
		return false, "", "active", nil, err
	}
	seen := ""
	if row.SeenTutorialsJSON != nil {
		seen = *row.SeenTutorialsJSON
	}
	return row.CompletedOnboarding == 1, seen, row.SubscriptionStatus, row.TrialEndsAt, nil
}

func (r *authRepository) CompleteOnboarding(ctx context.Context, userID, clientID string) error {
	query := `UPDATE client_user_link SET completed_onboarding = 1 WHERE user_id = ? AND client_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID, clientID)
	return err
}

func (r *authRepository) UpdateSeenTutorials(ctx context.Context, userID, clientID string, jsonStr string) error {
	query := `UPDATE client_user_link SET seen_tutorials_json = ? WHERE user_id = ? AND client_id = ?`
	_, err := r.db.ExecContext(ctx, query, jsonStr, userID, clientID)
	return err
}

