package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"barbercentral-core/internal/planlimit"
)

type AdminService interface {
	ListClients(ctx context.Context) ([]Client, error)
	GetClientByID(ctx context.Context, id string) (*Client, error)
	CreateClient(ctx context.Context, req CreateClientRequest) (*Client, error)
	UpdateClient(ctx context.Context, id string, req UpdateClientRequest) (*Client, error)
	BlockClient(ctx context.Context, id string) error
	UnblockClient(ctx context.Context, id string) error
	BlockClientWithReason(ctx context.Context, id, reason, performedBy string) error
	UnblockClientWithReason(ctx context.Context, id, reason, performedBy string) error
	UpdateClientPlan(ctx context.Context, id, planID string) error
	ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error)
	CreateClientUser(ctx context.Context, clientID string, req CreateClientUserRequest) (*ClientUser, error)
	UpdateClientUser(ctx context.Context, userID string, req UpdateClientUserRequest) error
	DeleteClientUser(ctx context.Context, userID string) error

	ListPlans(ctx context.Context) ([]planlimit.Plan, error)
	CreatePlan(ctx context.Context, p *planlimit.Plan) (*planlimit.Plan, error)
	UpdatePlan(ctx context.Context, id string, p *planlimit.Plan) (*planlimit.Plan, error)
}

type adminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) ListClients(ctx context.Context) ([]Client, error) {
	return s.repo.ListClients(ctx)
}

func (s *adminService) GetClientByID(ctx context.Context, id string) (*Client, error) {
	return s.repo.GetClientByID(ctx, id)
}

func (s *adminService) CreateClient(ctx context.Context, req CreateClientRequest) (*Client, error) {
	c := &Client{
		ID:           uuid.New().String(),
		PlanID:       req.PlanID,
		Name:         req.Name,
		Slug:         req.Slug,
		CustomDomain: req.CustomDomain,
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	err := s.repo.CreateClient(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *adminService) UpdateClient(ctx context.Context, id string, req UpdateClientRequest) (*Client, error) {
	c, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.PlanID = req.PlanID
	c.Name = req.Name
	c.Slug = req.Slug
	c.CustomDomain = req.CustomDomain

	err = s.repo.UpdateClient(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *adminService) BlockClient(ctx context.Context, id string) error {
	return s.repo.UpdateClientStatus(ctx, id, "blocked")
}

func (s *adminService) UnblockClient(ctx context.Context, id string) error {
	return s.repo.UpdateClientStatus(ctx, id, "active")
}

func (s *adminService) BlockClientWithReason(ctx context.Context, id, reason, performedBy string) error {
	err := s.repo.UpdateClientStatus(ctx, id, "blocked")
	if err != nil {
		return err
	}
	logID := uuid.New().String()
	return s.repo.CreateBlockLog(ctx, logID, id, "block", reason, performedBy)
}

func (s *adminService) UnblockClientWithReason(ctx context.Context, id, reason, performedBy string) error {
	err := s.repo.UpdateClientStatus(ctx, id, "active")
	if err != nil {
		return err
	}
	logID := uuid.New().String()
	return s.repo.CreateBlockLog(ctx, logID, id, "unblock", reason, performedBy)
}

func (s *adminService) UpdateClientPlan(ctx context.Context, id, planID string) error {
	return s.repo.UpdateClientPlan(ctx, id, planID)
}

func (s *adminService) ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error) {
	return s.repo.ListClientUsers(ctx, clientID)
}

func (s *adminService) CreateClientUser(ctx context.Context, clientID string, req CreateClientUserRequest) (*ClientUser, error) {
	allowed, curr, max, err := planlimit.CheckUsersLimit(ctx, s.repo.GetDB(), clientID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("plan_limit_exceeded:max_users:%d:%d", curr, max)
	}

	u := &ClientUser{
		ID:        uuid.New().String(),
		ClientID:  clientID,
		Name:      req.Name,
		Email:     req.Email,
		Role:      req.Role,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	// Senha padrão local para testes/fallback
	pwd := req.Password
	if pwd == "" {
		pwd = "senha_padrao_local"
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	err = s.repo.CreateClientUser(ctx, u, string(hashBytes))
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *adminService) UpdateClientUser(ctx context.Context, userID string, req UpdateClientUserRequest) error {
	return s.repo.UpdateClientUser(ctx, userID, req)
}

func (s *adminService) DeleteClientUser(ctx context.Context, userID string) error {
	return s.repo.DeleteClientUser(ctx, userID)
}

func (s *adminService) ListPlans(ctx context.Context) ([]planlimit.Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *adminService) CreatePlan(ctx context.Context, p *planlimit.Plan) (*planlimit.Plan, error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	err := s.repo.CreatePlan(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *adminService) UpdatePlan(ctx context.Context, id string, p *planlimit.Plan) (*planlimit.Plan, error) {
	p.ID = id
	err := s.repo.UpdatePlan(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
