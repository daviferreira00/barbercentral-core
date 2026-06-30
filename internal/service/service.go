package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ServiceService interface {
	List(ctx context.Context, clientID string, categoryID string) ([]Service, error)
	GetByID(ctx context.Context, clientID, id string) (*Service, error)
	Create(ctx context.Context, clientID string, req CreateServiceRequest) (*Service, error)
	Update(ctx context.Context, clientID, id string, req CreateServiceRequest) (*Service, error)
	Delete(ctx context.Context, clientID, id string) error
	ListCategories(ctx context.Context, clientID string) ([]ServiceCategory, error)
	CreateCategory(ctx context.Context, clientID string, req CreateCategoryRequest) (*ServiceCategory, error)
}

type serviceService struct {
	repo ServiceRepository
}

func NewServiceService(repo ServiceRepository) ServiceService {
	return &serviceService{repo: repo}
}

func (s *serviceService) List(ctx context.Context, clientID string, categoryID string) ([]Service, error) {
	return s.repo.List(ctx, clientID, categoryID)
}

func (s *serviceService) GetByID(ctx context.Context, clientID, id string) (*Service, error) {
	return s.repo.GetByID(ctx, clientID, id)
}

func (s *serviceService) Create(ctx context.Context, clientID string, req CreateServiceRequest) (*Service, error) {
	svc := &Service{
		ID:              uuid.New().String(),
		ClientID:        clientID,
		CategoryID:      req.CategoryID,
		Name:            req.Name,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
		Active:          req.Active,
		CreatedAt:       time.Now(),
	}

	if svc.DurationMinutes <= 0 {
		svc.DurationMinutes = 30
	}

	err := s.repo.Create(ctx, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *serviceService) Update(ctx context.Context, clientID, id string, req CreateServiceRequest) (*Service, error) {
	svc, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	svc.CategoryID = req.CategoryID
	svc.Name = req.Name
	svc.Description = req.Description
	svc.DurationMinutes = req.DurationMinutes
	svc.Price = req.Price
	svc.Active = req.Active

	err = s.repo.Update(ctx, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *serviceService) Delete(ctx context.Context, clientID, id string) error {
	return s.repo.Delete(ctx, clientID, id)
}

func (s *serviceService) ListCategories(ctx context.Context, clientID string) ([]ServiceCategory, error) {
	return s.repo.ListCategories(ctx, clientID)
}

func (s *serviceService) CreateCategory(ctx context.Context, clientID string, req CreateCategoryRequest) (*ServiceCategory, error) {
	cat := &ServiceCategory{
		ID:       uuid.New().String(),
		ClientID: clientID,
		Name:     req.Name,
	}

	err := s.repo.CreateCategory(ctx, cat)
	if err != nil {
		return nil, err
	}

	return cat, nil
}
