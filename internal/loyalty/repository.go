package loyalty

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetProgram(ctx context.Context, clientID string) (*LoyaltyProgram, error)
	CreateProgram(ctx context.Context, program *LoyaltyProgram) error
	UpdateProgram(ctx context.Context, program *LoyaltyProgram) error
	DeleteProgram(ctx context.Context, clientID string) error

	GetCardByCustomerID(ctx context.Context, clientID, customerID string) (*LoyaltyCard, error)
	GetCardByID(ctx context.Context, clientID, cardID string) (*LoyaltyCard, error)
	CreateCard(ctx context.Context, tx *sqlx.Tx, card *LoyaltyCard) error
	UpdateCardBalance(ctx context.Context, tx *sqlx.Tx, clientID, cardID string, stamps int, points float64) error

	ListTransactionsByCard(ctx context.Context, clientID, cardID string) ([]LoyaltyTransaction, error)
	CreateTransaction(ctx context.Context, tx *sqlx.Tx, transaction *LoyaltyTransaction) error

	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *repository) GetProgram(ctx context.Context, clientID string) (*LoyaltyProgram, error) {
	var program LoyaltyProgram
	query := "SELECT * FROM loyalty_program WHERE client_id = ? LIMIT 1"
	err := r.db.GetContext(ctx, &program, query, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Sem programa configurado
		}
		return nil, err
	}
	return &program, nil
}

func (r *repository) CreateProgram(ctx context.Context, program *LoyaltyProgram) error {
	query := `INSERT INTO loyalty_program (id, client_id, name, type, stamps_to_reward, points_per_real, reward_description, active) 
	          VALUES (:id, :client_id, :name, :type, :stamps_to_reward, :points_per_real, :reward_description, :active)`
	_, err := r.db.NamedExecContext(ctx, query, program)
	return err
}

func (r *repository) UpdateProgram(ctx context.Context, program *LoyaltyProgram) error {
	query := `UPDATE loyalty_program 
	          SET name = :name, type = :type, stamps_to_reward = :stamps_to_reward, points_per_real = :points_per_real, reward_description = :reward_description, active = :active 
	          WHERE client_id = :client_id AND id = :id`
	_, err := r.db.NamedExecContext(ctx, query, program)
	return err
}

func (r *repository) DeleteProgram(ctx context.Context, clientID string) error {
	query := "UPDATE loyalty_program SET active = 0 WHERE client_id = ?"
	_, err := r.db.ExecContext(ctx, query, clientID)
	return err
}

func (r *repository) GetCardByCustomerID(ctx context.Context, clientID, customerID string) (*LoyaltyCard, error) {
	var card LoyaltyCard
	query := "SELECT * FROM loyalty_card WHERE client_id = ? AND customer_id = ? LIMIT 1"
	err := r.db.GetContext(ctx, &card, query, clientID, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Sem cartão cadastrado
		}
		return nil, err
	}
	return &card, nil
}

func (r *repository) GetCardByID(ctx context.Context, clientID, cardID string) (*LoyaltyCard, error) {
	var card LoyaltyCard
	query := "SELECT * FROM loyalty_card WHERE client_id = ? AND id = ? LIMIT 1"
	err := r.db.GetContext(ctx, &card, query, clientID, cardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("cartão de fidelidade não encontrado")
		}
		return nil, err
	}
	return &card, nil
}

func (r *repository) CreateCard(ctx context.Context, tx *sqlx.Tx, card *LoyaltyCard) error {
	query := `INSERT INTO loyalty_card (id, customer_id, client_id, program_id, stamps_count, points_balance, status) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, card.ID, card.CustomerID, card.ClientID, card.ProgramID, card.StampsCount, card.PointsBalance, card.Status)
	} else {
		_, err = r.db.ExecContext(ctx, query, card.ID, card.CustomerID, card.ClientID, card.ProgramID, card.StampsCount, card.PointsBalance, card.Status)
	}
	return err
}

func (r *repository) UpdateCardBalance(ctx context.Context, tx *sqlx.Tx, clientID, cardID string, stamps int, points float64) error {
	query := "UPDATE loyalty_card SET stamps_count = ?, points_balance = ? WHERE client_id = ? AND id = ?"
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, stamps, points, clientID, cardID)
	} else {
		_, err = r.db.ExecContext(ctx, query, stamps, points, clientID, cardID)
	}
	return err
}

func (r *repository) ListTransactionsByCard(ctx context.Context, clientID, cardID string) ([]LoyaltyTransaction, error) {
	var list []LoyaltyTransaction
	query := "SELECT * FROM loyalty_transaction WHERE client_id = ? AND card_id = ? ORDER BY created_at DESC"
	err := r.db.SelectContext(ctx, &list, query, clientID, cardID)
	return list, err
}

func (r *repository) CreateTransaction(ctx context.Context, tx *sqlx.Tx, transaction *LoyaltyTransaction) error {
	query := `INSERT INTO loyalty_transaction (id, card_id, client_id, appointment_id, type, stamps_value, points_value, description, created_by, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, transaction.ID, transaction.CardID, transaction.ClientID, transaction.AppointmentID, transaction.Type, transaction.StampsValue, transaction.PointsValue, transaction.Description, transaction.CreatedBy, transaction.CreatedAt)
	} else {
		_, err = r.db.ExecContext(ctx, query, transaction.ID, transaction.CardID, transaction.ClientID, transaction.AppointmentID, transaction.Type, transaction.StampsValue, transaction.PointsValue, transaction.Description, transaction.CreatedBy, transaction.CreatedAt)
	}
	return err
}
