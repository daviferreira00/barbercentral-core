package stock

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetByID(ctx context.Context, clientID, id string) (*Product, error)
	List(ctx context.Context, clientID string, page, pageSize int, queryStr, filter string) ([]Product, int, error)
	Create(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, clientID, id string) error
	GetLowStock(ctx context.Context, clientID string) ([]Product, error)

	// Movimentações
	ListMovementsByProduct(ctx context.Context, clientID, productID string) ([]StockMovement, error)
	ListAllMovements(ctx context.Context, clientID, typeFilter string) ([]StockMovement, error)
	CreateMovement(ctx context.Context, tx *sqlx.Tx, move *StockMovement) error
	UpdateQuantityInStock(ctx context.Context, tx *sqlx.Tx, clientID, productID string, newQty float64) error

	// Vínculos de serviços
	ListServiceProducts(ctx context.Context, clientID, serviceID string) ([]EnrichedServiceProduct, error)
	LinkServiceProduct(ctx context.Context, link *ServiceProduct) error
	UpdateServiceProductQty(ctx context.Context, clientID, serviceID, productID string, qty float64) error
	UnlinkServiceProduct(ctx context.Context, clientID, serviceID, productID string) error

	// Transação auxiliar
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

func (r *repository) GetByID(ctx context.Context, clientID, id string) (*Product, error) {
	var prod Product
	query := "SELECT * FROM product WHERE client_id = ? AND id = ?"
	err := r.db.GetContext(ctx, &prod, query, clientID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("produto não encontrado")
		}
		return nil, err
	}
	return &prod, nil
}

func (r *repository) List(ctx context.Context, clientID string, page, pageSize int, queryStr, filter string) ([]Product, int, error) {
	var list []Product
	offset := (page - 1) * pageSize

	sqlQuery := "SELECT * FROM product WHERE client_id = ?"
	var args []interface{}
	args = append(args, clientID)

	if queryStr != "" {
		sqlQuery += " AND (name LIKE ? OR sku LIKE ?)"
		likeTerm := "%" + queryStr + "%"
		args = append(args, likeTerm, likeTerm)
	}

	if filter == "active" {
		sqlQuery += " AND active = 1"
	} else if filter == "inactive" {
		sqlQuery += " AND active = 0"
	} else if filter == "low_stock" {
		sqlQuery += " AND active = 1 AND quantity_in_stock <= low_stock_alert"
	}

	// Ordena
	sqlQuery += " ORDER BY name ASC"

	// Count
	var total int
	countQuery := "SELECT COUNT(*) FROM (" + sqlQuery + ") AS sub"
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Limit / Offset
	sqlQuery += " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	err = r.db.SelectContext(ctx, &list, sqlQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) Create(ctx context.Context, product *Product) error {
	query := `INSERT INTO product (id, client_id, name, description, sku, price, cost_price, quantity_in_stock, low_stock_alert, unit, active) 
	          VALUES (:id, :client_id, :name, :description, :sku, :price, :cost_price, :quantity_in_stock, :low_stock_alert, :unit, :active)`
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

func (r *repository) Update(ctx context.Context, product *Product) error {
	query := `UPDATE product 
	          SET name = :name, description = :description, sku = :sku, price = :price, cost_price = :cost_price, low_stock_alert = :low_stock_alert, unit = :unit, active = :active 
	          WHERE client_id = :client_id AND id = :id`
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

func (r *repository) Delete(ctx context.Context, clientID, id string) error {
	query := "UPDATE product SET active = 0 WHERE client_id = ? AND id = ?"
	_, err := r.db.ExecContext(ctx, query, clientID, id)
	return err
}

func (r *repository) GetLowStock(ctx context.Context, clientID string) ([]Product, error) {
	var list []Product
	query := "SELECT * FROM product WHERE client_id = ? AND active = 1 AND quantity_in_stock <= low_stock_alert"
	err := r.db.SelectContext(ctx, &list, query, clientID)
	return list, err
}

func (r *repository) ListMovementsByProduct(ctx context.Context, clientID, productID string) ([]StockMovement, error) {
	var list []StockMovement
	query := "SELECT * FROM stock_movement WHERE client_id = ? AND product_id = ? ORDER BY created_at DESC"
	err := r.db.SelectContext(ctx, &list, query, clientID, productID)
	return list, err
}

func (r *repository) ListAllMovements(ctx context.Context, clientID, typeFilter string) ([]StockMovement, error) {
	var list []StockMovement
	query := "SELECT * FROM stock_movement WHERE client_id = ?"
	var args []interface{}
	args = append(args, clientID)

	if typeFilter != "" {
		query += " AND type = ?"
		args = append(args, typeFilter)
	}

	query += " ORDER BY created_at DESC"
	err := r.db.SelectContext(ctx, &list, query, args...)
	return list, err
}

func (r *repository) CreateMovement(ctx context.Context, tx *sqlx.Tx, move *StockMovement) error {
	query := `INSERT INTO stock_movement (id, product_id, client_id, type, quantity, reason, appointment_id, created_by, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, move.ID, move.ProductID, move.ClientID, move.Type, move.Quantity, move.Reason, move.AppointmentID, move.CreatedBy, move.CreatedAt)
	} else {
		_, err = r.db.ExecContext(ctx, query, move.ID, move.ProductID, move.ClientID, move.Type, move.Quantity, move.Reason, move.AppointmentID, move.CreatedBy, move.CreatedAt)
	}
	return err
}

func (r *repository) UpdateQuantityInStock(ctx context.Context, tx *sqlx.Tx, clientID, productID string, newQty float64) error {
	query := "UPDATE product SET quantity_in_stock = ? WHERE client_id = ? AND id = ?"
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, newQty, clientID, productID)
	} else {
		_, err = r.db.ExecContext(ctx, query, newQty, clientID, productID)
	}
	return err
}

func (r *repository) ListServiceProducts(ctx context.Context, clientID, serviceID string) ([]EnrichedServiceProduct, error) {
	var list []EnrichedServiceProduct
	query := `SELECT sp.service_id, sp.product_id, p.name AS product_name, sp.quantity, p.unit 
	          FROM service_product sp 
	          JOIN product p ON sp.product_id = p.id 
	          WHERE sp.client_id = ? AND sp.service_id = ?`
	err := r.db.SelectContext(ctx, &list, query, clientID, serviceID)
	return list, err
}

func (r *repository) LinkServiceProduct(ctx context.Context, link *ServiceProduct) error {
	query := `INSERT INTO service_product (service_id, product_id, client_id, quantity) 
	          VALUES (?, ?, ?, ?) 
	          ON DUPLICATE KEY UPDATE quantity = VALUES(quantity)`
	_, err := r.db.ExecContext(ctx, query, link.ServiceID, link.ProductID, link.ClientID, link.Quantity)
	return err
}

func (r *repository) UpdateServiceProductQty(ctx context.Context, clientID, serviceID, productID string, qty float64) error {
	query := "UPDATE service_product SET quantity = ? WHERE client_id = ? AND service_id = ? AND product_id = ?"
	_, err := r.db.ExecContext(ctx, query, qty, clientID, serviceID, productID)
	return err
}

func (r *repository) UnlinkServiceProduct(ctx context.Context, clientID, serviceID, productID string) error {
	query := "DELETE FROM service_product WHERE client_id = ? AND service_id = ? AND product_id = ?"
	_, err := r.db.ExecContext(ctx, query, clientID, serviceID, productID)
	return err
}
