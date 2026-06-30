package finance

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetByID(ctx context.Context, clientID, id string) (*CashRegister, error)
	GetCurrent(ctx context.Context, clientID string) (*CashRegister, error)
	List(ctx context.Context, clientID string, page, pageSize int) ([]CashRegister, int, error)
	Open(ctx context.Context, register *CashRegister) error
	Close(ctx context.Context, clientID, id, closedBy string, closingBalance float64, notes *string) error

	ListTransactions(ctx context.Context, clientID, registerID string) ([]CashTransaction, error)
	CreateTransaction(ctx context.Context, tx *CashTransaction) error
	DeleteTransaction(ctx context.Context, clientID, registerID, txID string) error
	GetTransactionByID(ctx context.Context, clientID, registerID, txID string) (*CashTransaction, error)

	// Métodos de Pagamento de Agendamento
	CreateAppointmentPayment(ctx context.Context, paymentID, appID, clientID string, amount float64, method, status string, notes *string) error
	GetAppointmentPaymentByAppID(ctx context.Context, clientID, appID string) (*AppointmentPaymentDB, error)
	UpdateAppointmentPayment(ctx context.Context, clientID, appID string, amount float64, method, status string, notes *string) error
}

type AppointmentPaymentDB struct {
	ID            string  `db:"id"`
	AppointmentID string  `db:"appointment_id"`
	ClientID      string  `db:"client_id"`
	Amount        float64 `db:"amount"`
	Method        string  `db:"method"`
	Status        string  `db:"status"`
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, clientID, id string) (*CashRegister, error) {
	var reg CashRegister
	query := "SELECT * FROM cash_register WHERE client_id = ? AND id = ?"
	err := r.db.GetContext(ctx, &reg, query, clientID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("caixa não encontrado")
		}
		return nil, err
	}
	return &reg, nil
}

func (r *repository) GetCurrent(ctx context.Context, clientID string) (*CashRegister, error) {
	var reg CashRegister
	query := "SELECT * FROM cash_register WHERE client_id = ? AND status = 'open' LIMIT 1"
	err := r.db.GetContext(ctx, &reg, query, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Sem caixa aberto
		}
		return nil, err
	}
	return &reg, nil
}

func (r *repository) List(ctx context.Context, clientID string, page, pageSize int) ([]CashRegister, int, error) {
	var list []CashRegister
	offset := (page - 1) * pageSize

	queryList := "SELECT * FROM cash_register WHERE client_id = ? ORDER BY opened_at DESC LIMIT ? OFFSET ?"
	err := r.db.SelectContext(ctx, &list, queryList, clientID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	queryCount := "SELECT COUNT(*) FROM cash_register WHERE client_id = ?"
	err = r.db.GetContext(ctx, &total, queryCount, clientID)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) Open(ctx context.Context, register *CashRegister) error {
	query := `INSERT INTO cash_register (id, client_id, opened_by, opened_at, opening_balance, status) 
	          VALUES (:id, :client_id, :opened_by, :opened_at, :opening_balance, :status)`
	_, err := r.db.NamedExecContext(ctx, query, register)
	return err
}

func (r *repository) Close(ctx context.Context, clientID, id, closedBy string, closingBalance float64, notes *string) error {
	query := `UPDATE cash_register 
	          SET closed_by = ?, closed_at = NOW(), closing_balance = ?, notes = ?, status = 'closed' 
	          WHERE client_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, closedBy, closingBalance, notes, clientID, id)
	return err
}

func (r *repository) ListTransactions(ctx context.Context, clientID, registerID string) ([]CashTransaction, error) {
	var list []CashTransaction
	query := "SELECT * FROM cash_transaction WHERE client_id = ? AND register_id = ? ORDER BY created_at DESC"
	err := r.db.SelectContext(ctx, &list, query, clientID, registerID)
	return list, err
}

func (r *repository) CreateTransaction(ctx context.Context, tx *CashTransaction) error {
	query := `INSERT INTO cash_transaction (id, register_id, client_id, appointment_payment_id, type, amount, method, description, category, created_by, created_at) 
	          VALUES (:id, :register_id, :client_id, :appointment_payment_id, :type, :amount, :method, :description, :category, :created_by, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, tx)
	return err
}

func (r *repository) DeleteTransaction(ctx context.Context, clientID, registerID, txID string) error {
	query := "DELETE FROM cash_transaction WHERE client_id = ? AND register_id = ? AND id = ?"
	_, err := r.db.ExecContext(ctx, query, clientID, registerID, txID)
	return err
}

func (r *repository) GetTransactionByID(ctx context.Context, clientID, registerID, txID string) (*CashTransaction, error) {
	var tx CashTransaction
	query := "SELECT * FROM cash_transaction WHERE client_id = ? AND register_id = ? AND id = ?"
	err := r.db.GetContext(ctx, &tx, query, clientID, registerID, txID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("transação não encontrada")
		}
		return nil, err
	}
	return &tx, nil
}

func (r *repository) CreateAppointmentPayment(ctx context.Context, paymentID, appID, clientID string, amount float64, method, status string, notes *string) error {
	// Se já existe, atualiza, senão insere (fluxo robusto)
	query := `INSERT INTO appointment_payment (id, appointment_id, client_id, amount, method, status, paid_at, notes) 
	          VALUES (?, ?, ?, ?, ?, ?, CASE WHEN ? = 'paid' THEN NOW() ELSE NULL END, ?) 
	          ON DUPLICATE KEY UPDATE amount = VALUES(amount), method = VALUES(method), status = VALUES(status), paid_at = VALUES(paid_at), notes = VALUES(notes)`
	_, err := r.db.ExecContext(ctx, query, paymentID, appID, clientID, amount, method, status, status, notes)
	return err
}

func (r *repository) GetAppointmentPaymentByAppID(ctx context.Context, clientID, appID string) (*AppointmentPaymentDB, error) {
	var pay AppointmentPaymentDB
	query := "SELECT id, appointment_id, client_id, amount, method, status FROM appointment_payment WHERE client_id = ? AND appointment_id = ?"
	err := r.db.GetContext(ctx, &pay, query, clientID, appID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pay, nil
}

func (r *repository) UpdateAppointmentPayment(ctx context.Context, clientID, appID string, amount float64, method, status string, notes *string) error {
	query := `UPDATE appointment_payment 
	          SET amount = ?, method = ?, status = ?, paid_at = CASE WHEN ? = 'paid' THEN NOW() ELSE NULL END, notes = ? 
	          WHERE client_id = ? AND appointment_id = ?`
	_, err := r.db.ExecContext(ctx, query, amount, method, status, status, notes, clientID, appID)
	return err
}
