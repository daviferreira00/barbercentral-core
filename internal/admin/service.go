package admin

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminService interface {
	ListClients(ctx context.Context) ([]Client, error)
	GetClientByID(ctx context.Context, id string) (*Client, error)
	CreateClient(ctx context.Context, req CreateClientRequest) (*Client, error)
	UpdateClient(ctx context.Context, id string, req UpdateClientRequest) (*Client, error)
	BlockClient(ctx context.Context, id string) error
	UnblockClient(ctx context.Context, id string) error
	ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error)
	CreateClientUser(ctx context.Context, clientID string, req CreateClientUserRequest) (*ClientUser, error)
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
		ID:        uuid.New().String(),
		PlanID:    req.PlanID,
		Name:      req.Name,
		Slug:      req.Slug,
		Status:    "active",
		CreatedAt: time.Now(),
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

func (s *adminService) ListClientUsers(ctx context.Context, clientID string) ([]ClientUser, error) {
	return s.repo.ListClientUsers(ctx, clientID)
}

func (s *adminService) CreateClientUser(ctx context.Context, clientID string, req CreateClientUserRequest) (*ClientUser, error) {
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
