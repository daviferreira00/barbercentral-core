package loyalty

import (
	"time"
)

type LoyaltyProgram struct {
	ID                string     `db:"id" json:"id"`
	ClientID          string     `db:"client_id" json:"client_id"`
	Name              string     `db:"name" json:"name"`
	Type              string     `db:"type" json:"type"` // stamps, points
	StampsToReward    *int       `db:"stamps_to_reward" json:"stamps_to_reward,omitempty"`
	PointsPerReal     *float64   `db:"points_per_real" json:"points_per_real,omitempty"`
	RewardDescription string     `db:"reward_description" json:"reward_description"`
	Active            int        `db:"active" json:"active"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
}

type LoyaltyCard struct {
	ID            string    `db:"id" json:"id"`
	CustomerID    string    `db:"customer_id" json:"customer_id"`
	ClientID      string    `db:"client_id" json:"client_id"`
	ProgramID     string    `db:"program_id" json:"program_id"`
	StampsCount   int       `db:"stamps_count" json:"stamps_count"`
	PointsBalance float64   `db:"points_balance" json:"points_balance"`
	Status        string    `db:"status" json:"status"` // active, inactive
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type LoyaltyTransaction struct {
	ID            string     `db:"id" json:"id"`
	CardID        string     `db:"card_id" json:"card_id"`
	ClientID      string     `db:"client_id" json:"client_id"`
	AppointmentID *string    `db:"appointment_id" json:"appointment_id,omitempty"`
	Type          string     `db:"type" json:"type"` // earn, redeem
	StampsValue   *int       `db:"stamps_value" json:"stamps_value,omitempty"`
	PointsValue   *float64   `db:"points_value" json:"points_value,omitempty"`
	Description   string     `db:"description" json:"description"`
	CreatedBy     string     `db:"created_by" json:"created_by"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
}

type SaveProgramRequest struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"` // stamps, points
	StampsToReward    *int     `json:"stamps_to_reward,omitempty"`
	PointsPerReal     *float64 `json:"points_per_real,omitempty"`
	RewardDescription string   `json:"reward_description"`
	Active            int      `json:"active"`
}

type RedeemRequest struct {
	Notes *string `json:"notes,omitempty"`
}
