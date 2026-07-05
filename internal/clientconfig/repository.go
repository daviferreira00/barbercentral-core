package clientconfig

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrConfigNotFound = errors.New("configurações da barbearia não encontradas")
	ErrClientBlocked  = errors.New("esta barbearia está temporariamente indisponível")
)

type ConfigRepository interface {
	GetByClientID(ctx context.Context, clientID string) (*ClientConfig, error)
	Update(ctx context.Context, config *ClientConfig) error
	GetBySlug(ctx context.Context, slug string) (*PublicClientData, error)
	UpdateLogo(ctx context.Context, clientID, logoURL string) error
	UpdateLogoCentral(ctx context.Context, clientID, logoURL string) error
	GetClientName(ctx context.Context, clientID string) (string, error)
	GetClientSlug(ctx context.Context, clientID string) (string, error)
}

type configRepository struct {
	db *sqlx.DB
}

func NewConfigRepository(db *sqlx.DB) ConfigRepository {
	return &configRepository{db: db}
}

func (r *configRepository) GetByClientID(ctx context.Context, clientID string) (*ClientConfig, error) {
	var cfg ClientConfig
	query := "SELECT * FROM client_config WHERE client_id = ?"
	err := r.db.GetContext(ctx, &cfg, query, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *configRepository) Update(ctx context.Context, cfg *ClientConfig) error {
	query := `UPDATE client_config SET
		logo_url = :logo_url,
		logo_central = :logo_central,
		color_primary = :color_primary,
		color_secondary = :color_secondary,
		color_button = :color_button,
		background_type = :background_type,
		font_family = :font_family,
		address = :address,
		neighborhood = :neighborhood,
		city = :city,
		state = :state,
		phone = :phone,
		whatsapp = :whatsapp,
		instagram = :instagram,
		timezone = :timezone,
		cancellation_policy_hours = :cancellation_policy_hours,
		booking_requires_login = :booking_requires_login,
		min_advance_hours = :min_advance_hours,
		max_advance_days = :max_advance_days,
		interval_between_minutes = :interval_between_minutes,
		kds_pin = :kds_pin
		WHERE client_id = :client_id`
	res, err := r.db.NamedExecContext(ctx, query, cfg)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// Se não existe, cria
		queryInsert := `INSERT INTO client_config (
			client_id, logo_url, logo_central, color_primary, color_secondary, color_button, background_type, font_family, address, neighborhood, city, state,
			phone, whatsapp, instagram, timezone, cancellation_policy_hours, booking_requires_login,
			min_advance_hours, max_advance_days, interval_between_minutes, active, kds_pin
		) VALUES (
			:client_id, :logo_url, :logo_central, :color_primary, :color_secondary, :color_button, :background_type, :font_family, :address, :neighborhood, :city, :state,
			:phone, :whatsapp, :instagram, :timezone, :cancellation_policy_hours, :booking_requires_login,
			:min_advance_hours, :max_advance_days, :interval_between_minutes, 1, :kds_pin
		)`
		_, err = r.db.NamedExecContext(ctx, queryInsert, cfg)
		return err
	}
	return nil
}

func (r *configRepository) GetBySlug(ctx context.Context, slug string) (*PublicClientData, error) {
	// Verifica primeiro se a barbearia está ativa
	var c struct {
		ID           string  `db:"id"`
		Name         string  `db:"name"`
		Slug         string  `db:"slug"`
		Status       string  `db:"status"`
		CustomDomain *string `db:"custom_domain"`
	}
	queryClient := "SELECT id, name, slug, status, custom_domain FROM client WHERE slug = ? OR custom_domain = ?"
	err := r.db.GetContext(ctx, &c, queryClient, slug, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}

	if c.Status != "active" {
		return nil, ErrClientBlocked
	}

	var cfg ClientConfig
	queryCfg := "SELECT * FROM client_config WHERE client_id = ?"
	err = r.db.GetContext(ctx, &cfg, queryCfg, c.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Retorna padrão
			cfg = ClientConfig{
				ClientID:                 c.ID,
				ColorPrimary:             "#1a1a1a",
				ColorSecondary:           "#c9a84c",
				FontFamily:               "Inter",
				Timezone:                 "America/Sao_Paulo",
				CancellationPolicyHours: 2,
				MinAdvanceHours:         1,
				MaxAdvanceDays:          30,
				Active:                   1,
			}
		} else {
			return nil, err
		}
	}

	if cfg.Active != 1 {
		return nil, ErrClientBlocked
	}

	return &PublicClientData{
		ClientConfig: cfg,
		ClientName:   c.Name,
		ClientSlug:   c.Slug,
		CustomDomain: c.CustomDomain,
	}, nil
}

func (r *configRepository) UpdateLogo(ctx context.Context, clientID, logoURL string) error {
	query := "UPDATE client_config SET logo_url = ? WHERE client_id = ?"
	res, err := r.db.ExecContext(ctx, query, logoURL, clientID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// Tenta criar com logo
		cfg := &ClientConfig{
			ClientID:                 clientID,
			LogoURL:                  &logoURL,
			ColorPrimary:             "#1a1a1a",
			ColorSecondary:           "#c9a84c",
			FontFamily:               "Inter",
			Timezone:                 "America/Sao_Paulo",
			CancellationPolicyHours: 2,
			MinAdvanceHours:         1,
			MaxAdvanceDays:          30,
			Active:                   1,
		}
		queryInsert := `INSERT INTO client_config (
			client_id, logo_url, color_primary, color_secondary, font_family, timezone, 
			cancellation_policy_hours, booking_requires_login, min_advance_hours, max_advance_days, 
			interval_between_minutes, active
		) VALUES (
			:client_id, :logo_url, :color_primary, :color_secondary, :font_family, :timezone, 
			:cancellation_policy_hours, :booking_requires_login, :min_advance_hours, :max_advance_days, 
			:interval_between_minutes, 1
		)`
		_, err = r.db.NamedExecContext(ctx, queryInsert, cfg)
		return err
	}
	return nil
}

func (r *configRepository) UpdateLogoCentral(ctx context.Context, clientID, logoURL string) error {
	query := "UPDATE client_config SET logo_central = ? WHERE client_id = ?"
	res, err := r.db.ExecContext(ctx, query, logoURL, clientID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// Tenta criar com logo central
		cfg := &ClientConfig{
			ClientID:                 clientID,
			LogoCentral:              &logoURL,
			ColorPrimary:             "#1a1a1a",
			ColorSecondary:           "#c9a84c",
			FontFamily:               "Inter",
			Timezone:                 "America/Sao_Paulo",
			CancellationPolicyHours: 2,
			MinAdvanceHours:         1,
			MaxAdvanceDays:          30,
			Active:                   1,
		}
		queryInsert := `INSERT INTO client_config (
			client_id, logo_central, color_primary, color_secondary, font_family, timezone, 
			cancellation_policy_hours, booking_requires_login, min_advance_hours, max_advance_days, 
			interval_between_minutes, active
		) VALUES (
			:client_id, :logo_central, :color_primary, :color_secondary, :font_family, :timezone, 
			:cancellation_policy_hours, :booking_requires_login, :min_advance_hours, :max_advance_days, 
			:interval_between_minutes, 1
		)`
		_, err = r.db.NamedExecContext(ctx, queryInsert, cfg)
		return err
	}
	return nil
}

func (r *configRepository) GetClientName(ctx context.Context, clientID string) (string, error) {
	var name string
	query := "SELECT name FROM client WHERE id = ?"
	err := r.db.GetContext(ctx, &name, query, clientID)
	return name, err
}

func (r *configRepository) GetClientSlug(ctx context.Context, clientID string) (string, error) {
	var slug string
	query := "SELECT slug FROM client WHERE id = ?"
	err := r.db.GetContext(ctx, &slug, query, clientID)
	return slug, err
}
