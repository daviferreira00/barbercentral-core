package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service interface {
	List(ctx context.Context, clientID string) ([]NotificationConfig, error)
	GetByID(ctx context.Context, clientID, id string) (*NotificationConfig, error)
	Create(ctx context.Context, clientID string, req CreateNotificationRequest) (*NotificationConfig, error)
	Update(ctx context.Context, clientID, id string, req UpdateNotificationRequest) (*NotificationConfig, error)
	Delete(ctx context.Context, clientID, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, clientID string) ([]NotificationConfig, error) {
	return s.repo.List(ctx, clientID)
}

func (s *service) GetByID(ctx context.Context, clientID, id string) (*NotificationConfig, error) {
	return s.repo.GetByID(ctx, clientID, id)
}

func (s *service) Create(ctx context.Context, clientID string, req CreateNotificationRequest) (*NotificationConfig, error) {
	if req.Name == "" || req.TriggerType == "" || req.MessageTemplate == "" {
		return nil, fmt.Errorf("campos obrigatórios ausentes")
	}

	config := &NotificationConfig{
		ID:              uuid.New().String(),
		ClientID:        clientID,
		Name:            req.Name,
		TriggerType:     req.TriggerType,
		TriggerValue:    req.TriggerValue,
		TriggerUnit:     req.TriggerUnit,
		MessageTemplate: req.MessageTemplate,
		ChannelID:       req.ChannelID,
		Active:          req.Active,
	}

	err := s.repo.Create(ctx, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *service) Update(ctx context.Context, clientID, id string, req UpdateNotificationRequest) (*NotificationConfig, error) {
	config, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	config.Name = req.Name
	config.TriggerType = req.TriggerType
	config.TriggerValue = req.TriggerValue
	config.TriggerUnit = req.TriggerUnit
	config.MessageTemplate = req.MessageTemplate
	config.ChannelID = req.ChannelID
	config.Active = req.Active

	err = s.repo.Update(ctx, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *service) Delete(ctx context.Context, clientID, id string) error {
	return s.repo.Delete(ctx, clientID, id)
}
