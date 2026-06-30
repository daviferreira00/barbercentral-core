package loyalty

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service interface {
	GetProgram(ctx context.Context, clientID string) (*LoyaltyProgram, error)
	SaveProgram(ctx context.Context, clientID string, req SaveProgramRequest) (*LoyaltyProgram, error)
	DeleteProgram(ctx context.Context, clientID string) error

	GetCardByCustomerID(ctx context.Context, clientID, customerID string) (*LoyaltyCard, []LoyaltyTransaction, error)
	RedeemReward(ctx context.Context, clientID, userID, customerID string, req RedeemRequest) (*LoyaltyTransaction, error)

	// Lógica interna chamada pelo agendamento
	TriggerAutomaticEarn(ctx context.Context, clientID, customerID, appointmentID string, amount float64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetProgram(ctx context.Context, clientID string) (*LoyaltyProgram, error) {
	return s.repo.GetProgram(ctx, clientID)
}

func (s *service) SaveProgram(ctx context.Context, clientID string, req SaveProgramRequest) (*LoyaltyProgram, error) {
	if req.Type != "stamps" && req.Type != "points" {
		return nil, errors.New("tipo de programa inválido. Escolha 'stamps' ou 'points'")
	}

	p, err := s.repo.GetProgram(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if p == nil {
		p = &LoyaltyProgram{
			ID:       uuid.New().String(),
			ClientID: clientID,
			Active:   req.Active,
		}
	}

	p.Name = req.Name
	p.Type = req.Type
	p.RewardDescription = req.RewardDescription
	p.Active = req.Active
	p.StampsToReward = req.StampsToReward
	p.PointsPerReal = req.PointsPerReal

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
		err = s.repo.CreateProgram(ctx, p)
	} else {
		err = s.repo.UpdateProgram(ctx, p)
	}

	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) DeleteProgram(ctx context.Context, clientID string) error {
	return s.repo.DeleteProgram(ctx, clientID)
}

func (s *service) GetCardByCustomerID(ctx context.Context, clientID, customerID string) (*LoyaltyCard, []LoyaltyTransaction, error) {
	card, err := s.repo.GetCardByCustomerID(ctx, clientID, customerID)
	if err != nil {
		return nil, nil, err
	}
	if card == nil {
		return nil, nil, nil // Sem fidelidade ativa para este cliente
	}

	txs, err := s.repo.ListTransactionsByCard(ctx, clientID, card.ID)
	if err != nil {
		return card, nil, err
	}
	return card, txs, nil
}

func (s *service) RedeemReward(ctx context.Context, clientID, userID, customerID string, req RedeemRequest) (*LoyaltyTransaction, error) {
	p, err := s.repo.GetProgram(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Active == 0 {
		return nil, errors.New("não há programa de fidelidade ativo para esta barbearia")
	}

	card, err := s.repo.GetCardByCustomerID(ctx, clientID, customerID)
	if err != nil {
		return nil, err
	}
	if card == nil {
		return nil, errors.New("o cliente não possui um cartão de fidelidade ativo")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var stampsDeduction *int
	var pointsDeduction *float64
	var newStamps int = card.StampsCount
	var newPoints float64 = card.PointsBalance

	desc := fmt.Sprintf("Resgate manual de recompensa: %s", p.RewardDescription)
	if req.Notes != nil && *req.Notes != "" {
		desc += fmt.Sprintf(" (%s)", *req.Notes)
	}

	if p.Type == "stamps" {
		target := 10
		if p.StampsToReward != nil {
			target = *p.StampsToReward
		}
		if card.StampsCount < target {
			return nil, fmt.Errorf("carimbos insuficientes para resgate (necessário: %d, atual: %d)", target, card.StampsCount)
		}
		newStamps = card.StampsCount - target
		stampsDeduction = &target
	} else {
		// points
		target := 100.0
		if p.StampsToReward != nil {
			target = float64(*p.StampsToReward)
		}
		if card.PointsBalance < target {
			return nil, fmt.Errorf("pontos insuficientes para resgate (necessário: %.2f, atual: %.2f)", target, card.PointsBalance)
		}
		newPoints = card.PointsBalance - target
		pointsDeduction = &target
	}

	// Cria transação
	ltx := &LoyaltyTransaction{
		ID:            uuid.New().String(),
		CardID:        card.ID,
		ClientID:      clientID,
		Type:          "redeem",
		StampsValue:   stampsDeduction,
		PointsValue:   pointsDeduction,
		Description:   desc,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
	}

	err = s.repo.CreateTransaction(ctx, tx, ltx)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateCardBalance(ctx, tx, clientID, card.ID, newStamps, newPoints)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return ltx, nil
}

func (s *service) TriggerAutomaticEarn(ctx context.Context, clientID, customerID, appointmentID string, amount float64) error {
	p, err := s.repo.GetProgram(ctx, clientID)
	if err != nil {
		return err
	}
	if p == nil || p.Active == 0 {
		return nil // Sem programa de fidelidade ativo
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	card, err := s.repo.GetCardByCustomerID(ctx, clientID, customerID)
	if err != nil {
		return err
	}

	if card == nil {
		card = &LoyaltyCard{
			ID:            uuid.New().String(),
			CustomerID:    customerID,
			ClientID:      clientID,
			ProgramID:     p.ID,
			StampsCount:   0,
			PointsBalance: 0.0,
			Status:        "active",
			CreatedAt:     time.Now(),
		}
		err = s.repo.CreateCard(ctx, tx, card)
		if err != nil {
			return err
		}
	}

	var newStamps int = card.StampsCount
	var newPoints float64 = card.PointsBalance
	var stampsEarned *int
	var pointsEarned *float64
	var desc string

	if p.Type == "stamps" {
		val := 1
		newStamps = card.StampsCount + val
		stampsEarned = &val
		desc = fmt.Sprintf("Acúmulo automático por agendamento concluído (%s)", appointmentID[:8])
	} else {
		factor := 1.0
		if p.PointsPerReal != nil {
			factor = *p.PointsPerReal
		}
		val := amount * factor
		newPoints = card.PointsBalance + val
		pointsEarned = &val
		desc = fmt.Sprintf("Acúmulo automático por agendamento concluído (%s) - R$ %.2f gastos", appointmentID[:8], amount)
	}

	ltx := &LoyaltyTransaction{
		ID:            uuid.New().String(),
		CardID:        card.ID,
		ClientID:      clientID,
		AppointmentID: &appointmentID,
		Type:          "earn",
		StampsValue:   stampsEarned,
		PointsValue:   pointsEarned,
		Description:   desc,
		CreatedBy:     "system",
		CreatedAt:     time.Now(),
	}

	err = s.repo.CreateTransaction(ctx, tx, ltx)
	if err != nil {
		return err
	}

	err = s.repo.UpdateCardBalance(ctx, tx, clientID, card.ID, newStamps, newPoints)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Info().Str("client_id", clientID).Str("customer_id", customerID).Msg("Fidelidade acumulada automaticamente com sucesso!")
	return nil
}
