package stock

import (
	"time"
)

type Product struct {
	ID              string    `db:"id" json:"id"`
	ClientID        string    `db:"client_id" json:"client_id"`
	Name            string    `db:"name" json:"name"`
	Description     *string   `db:"description" json:"description,omitempty"`
	SKU             *string   `db:"sku" json:"sku,omitempty"`
	Price           float64   `db:"price" json:"price"`
	CostPrice       float64   `db:"cost_price" json:"cost_price"`
	QuantityInStock float64   `db:"quantity_in_stock" json:"quantity_in_stock"`
	LowStockAlert   float64   `db:"low_stock_alert" json:"low_stock_alert"`
	Unit            string    `db:"unit" json:"unit"` // un, ml, g, etc.
	Active          int       `db:"active" json:"active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type StockMovement struct {
	ID            string    `db:"id" json:"id"`
	ProductID     string    `db:"product_id" json:"product_id"`
	ClientID      string    `db:"client_id" json:"client_id"`
	Type          string    `db:"type" json:"type"` // in, out, adjustment
	Quantity      float64   `db:"quantity" json:"quantity"`
	Reason        *string   `db:"reason" json:"reason,omitempty"`
	AppointmentID *string   `db:"appointment_id" json:"appointment_id,omitempty"`
	CreatedBy     string    `db:"created_by" json:"created_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type ServiceProduct struct {
	ServiceID string  `db:"service_id" json:"service_id"`
	ProductID string  `db:"product_id" json:"product_id"`
	ClientID  string  `db:"client_id" json:"client_id"`
	Quantity  float64 `db:"quantity" json:"quantity"`
}

type EnrichedServiceProduct struct {
	ServiceID   string  `db:"service_id" json:"service_id"`
	ProductID   string  `db:"product_id" json:"product_id"`
	ProductName string  `db:"product_name" json:"product_name"`
	Quantity    float64 `db:"quantity" json:"quantity"`
	Unit        string  `db:"unit" json:"unit"`
}

type CreateProductRequest struct {
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	SKU             *string `json:"sku,omitempty"`
	Price           float64 `json:"price"`
	CostPrice       float64 `json:"cost_price"`
	QuantityInStock float64 `json:"quantity_in_stock"`
	LowStockAlert   float64 `json:"low_stock_alert"`
	Unit            string  `json:"unit"`
}

type UpdateProductRequest struct {
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	SKU           *string `json:"sku,omitempty"`
	Price         float64 `json:"price"`
	CostPrice     float64 `json:"cost_price"`
	LowStockAlert float64 `json:"low_stock_alert"`
	Unit          string  `json:"unit"`
	Active        int     `json:"active"`
}

type CreateMovementRequest struct {
	Type     string  `json:"type"` // in, out, adjustment
	Quantity float64 `json:"quantity"`
	Reason   *string `json:"reason,omitempty"`
}

type StockReportItem struct {
	ProductID       string  `db:"product_id" json:"product_id"`
	ProductName     string  `db:"product_name" json:"product_name"`
	SKU             *string `db:"sku" json:"sku,omitempty"`
	QuantityInStock float64 `db:"quantity_in_stock" json:"quantity_in_stock"`
	LowStockAlert   float64 `db:"low_stock_alert" json:"low_stock_alert"`
	Unit            string  `db:"unit" json:"unit"`
	IsLowStock      bool    `json:"is_low_stock"`
}
