package stock

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service interface {
	GetByID(ctx context.Context, clientID, id string) (*Product, error)
	List(ctx context.Context, clientID string, page, pageSize int, query, filter string) ([]Product, int, error)
	Create(ctx context.Context, clientID, userID string, req CreateProductRequest) (*Product, error)
	Update(ctx context.Context, clientID, id string, req UpdateProductRequest) (*Product, error)
	Delete(ctx context.Context, clientID, id string) error
	GetLowStock(ctx context.Context, clientID string) ([]Product, error)

	// Movimentações
	ListMovementsByProduct(ctx context.Context, clientID, productID string) ([]StockMovement, error)
	ListAllMovements(ctx context.Context, clientID, typeFilter string) ([]StockMovement, error)
	CreateMovement(ctx context.Context, clientID, userID, productID string, req CreateMovementRequest) (*StockMovement, error)
	StockReport(ctx context.Context, clientID string) ([]StockReportItem, error)
	ExportCSV(ctx context.Context, clientID string) ([]byte, error)

	// Vínculos de serviços
	ListServiceProducts(ctx context.Context, clientID, serviceID string) ([]EnrichedServiceProduct, error)
	LinkServiceProduct(ctx context.Context, clientID string, serviceID string, req ServiceProduct) error
	UpdateServiceProductQty(ctx context.Context, clientID, serviceID, productID string, qty float64) error
	UnlinkServiceProduct(ctx context.Context, clientID, serviceID, productID string) error

	// Baixa automática de agendamentos (consumido internamente)
	TriggerAutomaticDrop(ctx context.Context, clientID, appointmentID string, serviceIDs []string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, clientID, id string) (*Product, error) {
	return s.repo.GetByID(ctx, clientID, id)
}

func (s *service) List(ctx context.Context, clientID string, page, pageSize int, query, filter string) ([]Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return s.repo.List(ctx, clientID, page, pageSize, query, filter)
}

func (s *service) Create(ctx context.Context, clientID, userID string, req CreateProductRequest) (*Product, error) {
	p := &Product{
		ID:              uuid.New().String(),
		ClientID:        clientID,
		Name:            req.Name,
		Description:     req.Description,
		SKU:             req.SKU,
		Price:           req.Price,
		CostPrice:       req.CostPrice,
		QuantityInStock: req.QuantityInStock,
		LowStockAlert:   req.LowStockAlert,
		Unit:            req.Unit,
		Active:          1,
		CreatedAt:       time.Now(),
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	err = s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	// Se tem estoque inicial, gera movimentação de entrada
	if req.QuantityInStock > 0 {
		m := &StockMovement{
			ID:        uuid.New().String(),
			ProductID: p.ID,
			ClientID:  clientID,
			Type:      "in",
			Quantity:  req.QuantityInStock,
			Reason:    strPtr("Estoque inicial cadastrado"),
			CreatedBy: userID,
			CreatedAt: time.Now(),
		}
		err = s.repo.CreateMovement(ctx, tx, m)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *service) Update(ctx context.Context, clientID, id string, req UpdateProductRequest) (*Product, error) {
	p, err := s.repo.GetByID(ctx, clientID, id)
	if err != nil {
		return nil, err
	}

	p.Name = req.Name
	p.Description = req.Description
	p.SKU = req.SKU
	p.Price = req.Price
	p.CostPrice = req.CostPrice
	p.LowStockAlert = req.LowStockAlert
	p.Unit = req.Unit
	p.Active = req.Active

	err = s.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) Delete(ctx context.Context, clientID, id string) error {
	return s.repo.Delete(ctx, clientID, id)
}

func (s *service) GetLowStock(ctx context.Context, clientID string) ([]Product, error) {
	return s.repo.GetLowStock(ctx, clientID)
}

func (s *service) ListMovementsByProduct(ctx context.Context, clientID, productID string) ([]StockMovement, error) {
	return s.repo.ListMovementsByProduct(ctx, clientID, productID)
}

func (s *service) ListAllMovements(ctx context.Context, clientID, typeFilter string) ([]StockMovement, error) {
	return s.repo.ListAllMovements(ctx, clientID, typeFilter)
}

func (s *service) CreateMovement(ctx context.Context, clientID, userID, productID string, req CreateMovementRequest) (*StockMovement, error) {
	p, err := s.repo.GetByID(ctx, clientID, productID)
	if err != nil {
		return nil, err
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var newQty float64
	var actualQty float64
	var moveType string
	var reasonStr *string = req.Reason

	switch req.Type {
	case "in":
		newQty = p.QuantityInStock + req.Quantity
		actualQty = req.Quantity
		moveType = "in"
	case "out":
		if p.QuantityInStock < req.Quantity {
			return nil, errors.New("estoque insuficiente para efetuar a saída")
		}
		newQty = p.QuantityInStock - req.Quantity
		actualQty = req.Quantity
		moveType = "out"
	case "adjustment":
		diff := req.Quantity - p.QuantityInStock
		if diff == 0 {
			return nil, errors.New("a nova quantidade informada é idêntica ao estoque atual")
		}
		newQty = req.Quantity
		actualQty = math.Abs(diff)
		if diff > 0 {
			moveType = "in"
			reasonStr = strPtr(fmt.Sprintf("Ajuste de inventário (Acréscimo). Motivo: %s", optStr(req.Reason)))
		} else {
			moveType = "out"
			reasonStr = strPtr(fmt.Sprintf("Ajuste de inventário (Dedução). Motivo: %s", optStr(req.Reason)))
		}
	default:
		return nil, errors.New("tipo de movimentação inválido")
	}

	m := &StockMovement{
		ID:        uuid.New().String(),
		ProductID: productID,
		ClientID:  clientID,
		Type:      moveType,
		Quantity:  actualQty,
		Reason:    reasonStr,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}

	err = s.repo.CreateMovement(ctx, tx, m)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateQuantityInStock(ctx, tx, clientID, productID, newQty)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *service) StockReport(ctx context.Context, clientID string) ([]StockReportItem, error) {
	prods, _, err := s.repo.List(ctx, clientID, 1, 1000, "", "")
	if err != nil {
		return nil, err
	}

	var report []StockReportItem
	for _, p := range prods {
		report = append(report, StockReportItem{
			ProductID:       p.ID,
			ProductName:     p.Name,
			SKU:             p.SKU,
			QuantityInStock: p.QuantityInStock,
			LowStockAlert:   p.LowStockAlert,
			Unit:            p.Unit,
			IsLowStock:      p.QuantityInStock <= p.LowStockAlert,
		})
	}
	return report, nil
}

func (s *service) ExportCSV(ctx context.Context, clientID string) ([]byte, error) {
	prods, _, err := s.repo.List(ctx, clientID, 1, 1000, "", "")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	_ = w.Write([]string{"Produto", "SKU", "Estoque Atual", "Alerta Minimo", "Unidade", "Preco Custo", "Preco Venda", "Status"})

	for _, p := range prods {
		skuStr := ""
		if p.SKU != nil {
			skuStr = *p.SKU
		}
		statusStr := "Ativo"
		if p.Active == 0 {
			statusStr = "Inativo"
		}
		if p.QuantityInStock <= p.LowStockAlert {
			statusStr += " (Estoque Baixo)"
		}

		_ = w.Write([]string{
			p.Name,
			skuStr,
			fmt.Sprintf("%.3f", p.QuantityInStock),
			fmt.Sprintf("%.3f", p.LowStockAlert),
			p.Unit,
			fmt.Sprintf("%.2f", p.CostPrice),
			fmt.Sprintf("%.2f", p.Price),
			statusStr,
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (s *service) ListServiceProducts(ctx context.Context, clientID, serviceID string) ([]EnrichedServiceProduct, error) {
	return s.repo.ListServiceProducts(ctx, clientID, serviceID)
}

func (s *service) LinkServiceProduct(ctx context.Context, clientID string, serviceID string, req ServiceProduct) error {
	req.ClientID = clientID
	req.ServiceID = serviceID
	return s.repo.LinkServiceProduct(ctx, &req)
}

func (s *service) UpdateServiceProductQty(ctx context.Context, clientID, serviceID, productID string, qty float64) error {
	return s.repo.UpdateServiceProductQty(ctx, clientID, serviceID, productID, qty)
}

func (s *service) UnlinkServiceProduct(ctx context.Context, clientID, serviceID, productID string) error {
	return s.repo.UnlinkServiceProduct(ctx, clientID, serviceID, productID)
}

func (s *service) TriggerAutomaticDrop(ctx context.Context, clientID, appointmentID string, serviceIDs []string) error {
	// 1. Para cada serviço, carrega os produtos consumidos
	productConsumptions := make(map[string]float64)

	for _, svcID := range serviceIDs {
		links, err := s.repo.ListServiceProducts(ctx, clientID, svcID)
		if err != nil {
			log.Error().Err(err).Msg("Erro ao listar insumos do serviço para baixa automática")
			continue
		}
		for _, link := range links {
			productConsumptions[link.ProductID] += link.Quantity
		}
	}

	if len(productConsumptions) == 0 {
		return nil // Sem produtos associados a nenhum dos serviços concluídos
	}

	// 2. Para cada produto, efetua a baixa transacional
	for prodID, qty := range productConsumptions {
		p, err := s.repo.GetByID(ctx, clientID, prodID)
		if err != nil {
			log.Error().Err(err).Str("product_id", prodID).Msg("Produto não encontrado para baixa automática")
			continue
		}

		tx, err := s.repo.BeginTx(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Erro ao iniciar transação de baixa automática")
			continue
		}

		newQty := p.QuantityInStock - qty
		if newQty < 0 {
			newQty = 0 // Impede estoque negativo no automático
		}

		m := &StockMovement{
			ID:            uuid.New().String(),
			ProductID:     prodID,
			ClientID:      clientID,
			Type:          "out",
			Quantity:      qty,
			Reason:        strPtr(fmt.Sprintf("Baixa automática de atendimento finalizado (%s)", appointmentID[:8])),
			AppointmentID: &appointmentID,
			CreatedBy:     "system",
			CreatedAt:     time.Now(),
		}

		err = s.repo.CreateMovement(ctx, tx, m)
		if err != nil {
			_ = tx.Rollback()
			log.Error().Err(err).Msg("Erro ao registrar movimentação de baixa automática")
			continue
		}

		err = s.repo.UpdateQuantityInStock(ctx, tx, clientID, prodID, newQty)
		if err != nil {
			_ = tx.Rollback()
			log.Error().Err(err).Msg("Erro ao atualizar quantidade após baixa automática")
			continue
		}

		err = tx.Commit()
		if err != nil {
			log.Error().Err(err).Msg("Erro ao commitar baixa automática de estoque")
			continue
		}

		if newQty <= p.LowStockAlert {
			log.Warn().Str("product", p.Name).Float64("quantity", newQty).Msg("ALERTA: Estoque do produto atingiu nível crítico!")
		}
	}

	return nil
}

// Helpers
func strPtr(s string) *string {
	return &s
}

func optStr(s *string) string {
	if s != nil {
		return *s
	}
	return "Sem observação"
}
