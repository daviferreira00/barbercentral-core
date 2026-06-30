package professional

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	List(ctx context.Context, clientID string, status string) ([]Professional, error)
	GetByID(ctx context.Context, clientID, id string) (*Professional, error)
	Create(ctx context.Context, clientID string, req CreateProfessionalRequest) (*Professional, error)
	Update(ctx context.Context, clientID, id string, req UpdateProfessionalRequest) (*Professional, error)
	Delete(ctx context.Context, clientID, id string) error
	GetSchedule(ctx context.Context, clientID, professionalID string) ([]ProfessionalSchedule, error)
	SaveSchedule(ctx context.Context, clientID, professionalID string, req BulkUpdateScheduleRequest) error
	GetLinkedServices(ctx context.Context, clientID, professionalID string) ([]ProfessionalServiceLink, error)
	LinkService(ctx context.Context, clientID, professionalID string, req LinkServiceRequest) error
	UnlinkService(ctx context.Context, clientID, professionalID, serviceID string) error
	UpdatePhoto(ctx context.Context, clientID, id, photoURL string) error
}

type service struct {
	repo ProfessionalRepository
}

func NewService(repo ProfessionalRepository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, clientID string, status string) ([]Professional, error) {
	return s.repo.List(ctx, clientID, status)
}

func (s *service) GetByID(ctx context.Context, clientID, id string) (*Professional, error) {
	return s.repo.GetByID(ctx, clientID, id)
}

func (s *service) Create(ctx context.Context, clientID string, req CreateProfessionalRequest) (*Professional, error) {
	pID := uuid.New().String()
	p := &Professional{
		ID:        pID,
		ClientID:  clientID,
		Name:      req.Name,
		Bio:       req.Bio,
		Status:    req.Status,
		CreatedAt: time.Now(),
	}

	if p.Status == "" {
		p.Status = "active"
	}

	err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	// Grade padrão de Seg a Sáb (weekday 1 a 6) das 09h às 19h
	var defaultSchedules []ProfessionalSchedule
	for day := 1; day <= 6; day++ {
		defaultSchedules = append(defaultSchedules, ProfessionalSchedule{
			ID:             uuid.New().String(),
			ProfessionalID: pID,
			ClientID:       clientID,
			Weekday:        day,
			StartTime:      "09:00:00",
			EndTime:        "19:00:00",
			Enabled:        1,
		})
	}
	_ = s.repo.SaveSchedule(ctx, defaultSchedules)

	return p, nil
}

func (s *service) Update(ctx context.Context, clientID, id string, req UpdateProfessionalRequest) (*Professional, error) {
	p, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	p.Name = req.Name
	p.Bio = req.Bio
	p.Status = req.Status

	err = s.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *service) Delete(ctx context.Context, clientID, id string) error {
	return s.repo.Delete(ctx, clientID, id)
}

func (s *service) GetSchedule(ctx context.Context, clientID, professionalID string) ([]ProfessionalSchedule, error) {
	return s.repo.GetSchedule(ctx, clientID, professionalID)
}

func (s *service) SaveSchedule(ctx context.Context, clientID, professionalID string, req BulkUpdateScheduleRequest) error {
	for i := range req.Schedules {
		if req.Schedules[i].ID == "" {
			req.Schedules[i].ID = uuid.New().String()
		}
		req.Schedules[i].ProfessionalID = professionalID
		req.Schedules[i].ClientID = clientID
	}
	return s.repo.SaveSchedule(ctx, req.Schedules)
}

func (s *service) GetLinkedServices(ctx context.Context, clientID, professionalID string) ([]ProfessionalServiceLink, error) {
	return s.repo.GetLinkedServices(ctx, clientID, professionalID)
}

func (s *service) LinkService(ctx context.Context, clientID, professionalID string, req LinkServiceRequest) error {
	link := &ProfessionalServiceLink{
		ProfessionalID: professionalID,
		ServiceID:      req.ServiceID,
		ClientID:       clientID,
		CustomPrice:    req.CustomPrice,
		CustomDuration: req.CustomDuration,
	}
	return s.repo.LinkService(ctx, link)
}

func (s *service) UnlinkService(ctx context.Context, clientID, professionalID, serviceID string) error {
	return s.repo.UnlinkService(ctx, clientID, professionalID, serviceID)
}

func (s *service) UpdatePhoto(ctx context.Context, clientID, id, photoURL string) error {
	return s.repo.UpdatePhoto(ctx, clientID, id, photoURL)
}
