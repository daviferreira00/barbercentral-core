package clientconfig

import "context"

type ConfigService interface {
	GetByClientID(ctx context.Context, clientID string) (*ClientConfig, error)
	Update(ctx context.Context, clientID string, req UpdateConfigRequest) (*ClientConfig, error)
	GetBySlug(ctx context.Context, slug string) (*PublicClientData, error)
	UpdateLogo(ctx context.Context, clientID, logoURL string) error
	UpdateLogoCentral(ctx context.Context, clientID, logoURL string) error
}

type configService struct {
	repo ConfigRepository
}

func NewConfigService(repo ConfigRepository) ConfigService {
	return &configService{repo: repo}
}

func (s *configService) GetByClientID(ctx context.Context, clientID string) (*ClientConfig, error) {
	return s.repo.GetByClientID(ctx, clientID)
}

func (s *configService) Update(ctx context.Context, clientID string, req UpdateConfigRequest) (*ClientConfig, error) {
	cfg := &ClientConfig{
		ClientID:                 clientID,
		LogoURL:                  req.LogoURL,
		LogoCentral:              req.LogoCentral,
		ColorPrimary:             req.ColorPrimary,
		ColorSecondary:           req.ColorSecondary,
		ColorButton:              req.ColorButton,
		BackgroundType:           req.BackgroundType,
		FontFamily:               req.FontFamily,
		Address:                  req.Address,
		Neighborhood:             req.Neighborhood,
		City:                     req.City,
		State:                    req.State,
		Phone:                    req.Phone,
		WhatsApp:                 req.WhatsApp,
		Instagram:                req.Instagram,
		Timezone:                 req.Timezone,
		CancellationPolicyHours: req.CancellationPolicyHours,
		BookingRequiresLogin:    req.BookingRequiresLogin,
		MinAdvanceHours:         req.MinAdvanceHours,
		MaxAdvanceDays:          req.MaxAdvanceDays,
		IntervalBetweenMinutes:  req.IntervalBetweenMinutes,
		Active:                   1,
	}

	if cfg.ColorPrimary == "" {
		cfg.ColorPrimary = "#1a1a1a"
	}
	if cfg.ColorSecondary == "" {
		cfg.ColorSecondary = "#c9a84c"
	}
	if cfg.FontFamily == "" {
		cfg.FontFamily = "Inter"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "America/Sao_Paulo"
	}

	err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (s *configService) GetBySlug(ctx context.Context, slug string) (*PublicClientData, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *configService) UpdateLogo(ctx context.Context, clientID, logoURL string) error {
	return s.repo.UpdateLogo(ctx, clientID, logoURL)
}

func (s *configService) UpdateLogoCentral(ctx context.Context, clientID, logoURL string) error {
	return s.repo.UpdateLogoCentral(ctx, clientID, logoURL)
}
