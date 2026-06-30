package finance

import (
	"time"
)

type CashRegister struct {
	ID             string     `db:"id" json:"id"`
	ClientID       string     `db:"client_id" json:"client_id"`
	OpenedBy       string     `db:"opened_by" json:"opened_by"`
	OpenedAt       time.Time  `db:"opened_at" json:"opened_at"`
	ClosedBy       *string    `db:"closed_by" json:"closed_by,omitempty"`
	ClosedAt       *time.Time `db:"closed_at" json:"closed_at,omitempty"`
	OpeningBalance float64    `db:"opening_balance" json:"opening_balance"`
	ClosingBalance *float64   `db:"closing_balance" json:"closing_balance,omitempty"`
	Notes          *string    `db:"notes" json:"notes,omitempty"`
	Status         string     `db:"status" json:"status"` // open, closed
}

type CashTransaction struct {
	ID                   string     `db:"id" json:"id"`
	RegisterID           string     `db:"register_id" json:"register_id"`
	ClientID             string     `db:"client_id" json:"client_id"`
	AppointmentPaymentID *string    `db:"appointment_payment_id" json:"appointment_payment_id,omitempty"`
	Type                 string     `db:"type" json:"type"` // income, expense
	Amount               float64    `db:"amount" json:"amount"`
	Method               string     `db:"method" json:"method"` // cash, pix, card_debit, card_credit, other
	Description          string     `db:"description" json:"description"`
	Category             *string    `db:"category" json:"category,omitempty"`
	CreatedBy            string     `db:"created_by" json:"created_by"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
}

type OpenRegisterRequest struct {
	OpeningBalance float64 `json:"opening_balance"`
	Notes          *string `json:"notes,omitempty"`
}

type CloseRegisterRequest struct {
	ClosingBalance float64 `json:"closing_balance"`
	Notes          *string `json:"notes,omitempty"`
}

type CreateTransactionRequest struct {
	Type        string  `json:"type"` // income, expense
	Amount      float64 `json:"amount"`
	Method      string  `json:"method"` // cash, pix, card_debit, card_credit, other
	Description string  `json:"description"`
	Category    *string `json:"category,omitempty"`
}

type CashRegisterSummary struct {
	Register       *CashRegister      `json:"register"`
	TotalIncome    float64            `json:"total_income"`
	TotalExpense   float64            `json:"total_expense"`
	ExpectedBal    float64            `json:"expected_balance"`
	MethodTotals   map[string]float64 `json:"method_totals"` // totals per pix, cash, card_debit, card_credit, other
}
