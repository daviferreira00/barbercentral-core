package finance

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"barbercentral-core/internal/loyalty"
)

type Service interface {
	GetByID(ctx context.Context, clientID, id string) (*CashRegister, error)
	GetCurrent(ctx context.Context, clientID string) (*CashRegister, error)
	List(ctx context.Context, clientID string, page, pageSize int) ([]CashRegister, int, error)
	Open(ctx context.Context, clientID, userID string, req OpenRegisterRequest) (*CashRegister, error)
	Close(ctx context.Context, clientID, id, userID string, req CloseRegisterRequest) error
	GetSummary(ctx context.Context, clientID, id string) (*CashRegisterSummary, error)
	ExportCSV(ctx context.Context, clientID, id string) ([]byte, error)

	ListTransactions(ctx context.Context, clientID, registerID string) ([]CashTransaction, error)
	CreateTransaction(ctx context.Context, clientID, userID, registerID string, req CreateTransactionRequest) (*CashTransaction, error)
	DeleteTransaction(ctx context.Context, clientID, registerID, txID string) error

	// Vínculo com agendamento
	RegisterPayment(ctx context.Context, clientID, userID, appID string, amount float64, method string, notes *string) error
	UpdatePayment(ctx context.Context, clientID, userID, appID string, amount float64, method string, notes *string) error
}

type service struct {
	repo           Repository
	loyaltyService loyalty.Service
}

func NewService(repo Repository, loyaltyService loyalty.Service) Service {
	return &service{
		repo:           repo,
		loyaltyService: loyaltyService,
	}
}

func (s *service) GetByID(ctx context.Context, clientID, id string) (*CashRegister, error) {
	return s.repo.GetByID(ctx, clientID, id)
}

func (s *service) GetCurrent(ctx context.Context, clientID string) (*CashRegister, error) {
	return s.repo.GetCurrent(ctx, clientID)
}

func (s *service) List(ctx context.Context, clientID string, page, pageSize int) ([]CashRegister, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return s.repo.List(ctx, clientID, page, pageSize)
}

func (s *service) Open(ctx context.Context, clientID, userID string, req OpenRegisterRequest) (*CashRegister, error) {
	// Verifica se já existe caixa aberto
	current, err := s.repo.GetCurrent(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if current != nil {
		return nil, errors.New("já existe um caixa aberto para esta barbearia")
	}

	reg := &CashRegister{
		ID:             uuid.New().String(),
		ClientID:       clientID,
		OpenedBy:       userID,
		OpenedAt:       time.Now(),
		OpeningBalance: req.OpeningBalance,
		Status:         "open",
	}

	err = s.repo.Open(ctx, reg)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func (s *service) Close(ctx context.Context, clientID, id, userID string, req CloseRegisterRequest) error {
	reg, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return err
	}
	if reg.Status == "closed" {
		return errors.New("este caixa já está fechado")
	}

	return s.repo.Close(ctx, clientID, id, userID, req.ClosingBalance, req.Notes)
}

func (s *service) GetSummary(ctx context.Context, clientID, id string) (*CashRegisterSummary, error) {
	reg, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	txs, err := s.repo.ListTransactions(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	var totalIncome float64
	var totalExpense float64
	methodTotals := map[string]float64{
		"cash":        0.0,
		"pix":         0.0,
		"card_debit":  0.0,
		"card_credit": 0.0,
		"other":       0.0,
	}

	for _, tx := range txs {
		if tx.Type == "income" {
			totalIncome += tx.Amount
			methodTotals[tx.Method] += tx.Amount
		} else {
			totalExpense += tx.Amount
			// Despesas diminuem do método que foi pago (ex: se saiu dinheiro do caixa físico)
			methodTotals[tx.Method] -= tx.Amount
		}
	}

	// O saldo esperado em dinheiro físico é: saldo inicial + entradas em cash - saídas em cash
	expectedCashBal := reg.OpeningBalance + methodTotals["cash"]

	summary := &CashRegisterSummary{
		Register:     reg,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		ExpectedBal:  expectedCashBal,
		MethodTotals: methodTotals,
	}

	return summary, nil
}

func (s *service) ListTransactions(ctx context.Context, clientID, registerID string) ([]CashTransaction, error) {
	return s.repo.ListTransactions(ctx, clientID, registerID)
}

func (s *service) CreateTransaction(ctx context.Context, clientID, userID, registerID string, req CreateTransactionRequest) (*CashTransaction, error) {
	reg, err := s.repo.GetByID(ctx, clientID, registerID)
	if err != nil {
		return nil, err
	}
	if reg.Status == "closed" {
		return nil, errors.New("não é possível lançar transações em um caixa fechado")
	}

	tx := &CashTransaction{
		ID:          uuid.New().String(),
		RegisterID:  registerID,
		ClientID:    clientID,
		Type:        req.Type,
		Amount:      req.Amount,
		Method:      req.Method,
		Description: req.Description,
		Category:    req.Category,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	err = s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *service) DeleteTransaction(ctx context.Context, clientID, registerID, txID string) error {
	reg, err := s.repo.GetByID(ctx, clientID, registerID)
	if err != nil {
		return err
	}
	if reg.Status == "closed" {
		return errors.New("não é possível estornar lançamentos em um caixa fechado")
	}

	return s.repo.DeleteTransaction(ctx, clientID, registerID, txID)
}

func (s *service) RegisterPayment(ctx context.Context, clientID, userID, appID string, amount float64, method string, notes *string) error {
	// Verifica se há caixa aberto
	current, err := s.repo.GetCurrent(ctx, clientID)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("não há nenhum caixa aberto. Abra o caixa primeiro para registrar pagamentos")
	}

	// Cria o ID do pagamento
	payID := uuid.New().String()

	err = s.repo.CreateAppointmentPayment(ctx, payID, appID, clientID, amount, method, "paid", notes)
	if err != nil {
		return err
	}

	// Lança a transação no caixa aberto
	tx := &CashTransaction{
		ID:                   uuid.New().String(),
		RegisterID:           current.ID,
		ClientID:             clientID,
		AppointmentPaymentID: &payID,
		Type:                 "income",
		Amount:               amount,
		Method:               method,
		Description:          fmt.Sprintf("Recebimento do Agendamento %s", appID[:8]),
		CreatedBy:            userID,
		CreatedAt:            time.Now(),
	}

	err = s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return err
	}

	// Trigger fidelidade automática se houver cliente e agendamento concluído
	custID, status, errDetails := s.repo.GetAppointmentDetails(ctx, clientID, appID)
	if errDetails == nil && custID != nil && *custID != "" && status == "completed" && s.loyaltyService != nil {
		_ = s.loyaltyService.TriggerAutomaticEarn(ctx, clientID, *custID, appID, amount)
	}

	return nil
}

func (s *service) UpdatePayment(ctx context.Context, clientID, userID, appID string, amount float64, method string, notes *string) error {
	pay, err := s.repo.GetAppointmentPaymentByAppID(ctx, clientID, appID)
	if err != nil {
		return err
	}
	if pay == nil {
		// Se não existe, podemos registrá-lo
		return s.RegisterPayment(ctx, clientID, userID, appID, amount, method, notes)
	}

	if pay.Status == "paid" {
		// Verifica se o caixa onde o pagamento foi lançado ainda está aberto
		current, err := s.repo.GetCurrent(ctx, clientID)
		if err != nil {
			return err
		}
		if current == nil {
			return errors.New("caixa aberto não encontrado para edição do pagamento")
		}
	}

	return s.repo.UpdateAppointmentPayment(ctx, clientID, appID, amount, method, "paid", notes)
}

func (s *service) ExportCSV(ctx context.Context, clientID, id string) ([]byte, error) {
	txs, err := s.repo.ListTransactions(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	// Escreve o UTF-8 BOM para abrir corretamente no Excel brasileiro
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	w.Comma = ';'

	// Cabeçalho
	_ = w.Write([]string{"Data", "Tipo", "Valor", "Metodo", "Descricao", "Categoria", "Registrado Por"})

	for _, tx := range txs {
		createdAtStr := tx.CreatedAt.Format("02/01/2006 15:04")
		valStr := fmt.Sprintf("%.2f", tx.Amount)

		methodLabel := tx.Method
		switch tx.Method {
		case "cash":
			methodLabel = "Dinheiro"
		case "pix":
			methodLabel = "PIX"
		case "card_debit":
			methodLabel = "Cartão de Débito"
		case "card_credit":
			methodLabel = "Cartão de Crédito"
		case "other":
			methodLabel = "Outros"
		}

		typeLabel := "Entrada"
		if tx.Type == "expense" {
			typeLabel = "Saída"
		}

		cat := ""
		if tx.Category != nil {
			cat = *tx.Category
		}

		_ = w.Write([]string{
			createdAtStr,
			typeLabel,
			valStr,
			methodLabel,
			tx.Description,
			cat,
			tx.CreatedBy,
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}
