package planlimit

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type Plan struct {
	ID               string  `db:"id" json:"id"`
	Name             string  `db:"name" json:"name"`
	MaxProfessionals int     `db:"max_professionals" json:"max_professionals"`
	MaxCustomers     int     `db:"max_customers" json:"max_customers"`
	MaxUsers         int     `db:"max_users" json:"max_users"`
	HasLoyalty       int     `db:"has_loyalty" json:"has_loyalty"`
	HasStock         int     `db:"has_stock" json:"has_stock"`
	HasReports       int     `db:"has_reports" json:"has_reports"`
	HasOnlineBooking int     `db:"has_online_booking" json:"has_online_booking"`
	IsPublic         int     `db:"is_public" json:"is_public"`
	BillingType      string  `db:"billing_type" json:"billing_type"`
	Price            float64 `db:"price" json:"price"`
	FeaturesJSON     *string `db:"features_json" json:"features_json,omitempty"`
	CreatedAt        string  `db:"created_at" json:"created_at,omitempty"`
}

type UsageResponse struct {
	Plan          *Plan `json:"plan"`
	Professionals int   `json:"professionals"`
	Customers     int   `json:"customers"`
	Users         int   `json:"users"`
}

type ClientBlockLog struct {
	ID          string `db:"id" json:"id"`
	ClientID    string `db:"client_id" json:"client_id"`
	Action      string `db:"action" json:"action"`
	Reason      string `db:"reason" json:"reason"`
	PerformedBy string `db:"performed_by" json:"performed_by"`
	CreatedAt   string `db:"created_at" json:"created_at"`
}

func GetClientPlan(ctx context.Context, db *sqlx.DB, clientID string) (*Plan, error) {
	var plan Plan
	query := `SELECT p.* FROM plan p JOIN client c ON c.plan_id = p.id WHERE c.id = ? LIMIT 1`
	err := db.GetContext(ctx, &plan, query, clientID)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func CheckProfessionalsLimit(ctx context.Context, db *sqlx.DB, clientID string) (bool, int, int, error) {
	p, err := GetClientPlan(ctx, db, clientID)
	if err != nil {
		return false, 0, 0, err
	}
	if p.MaxProfessionals == -1 {
		return true, 0, 0, nil
	}

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM professional WHERE client_id = ?", clientID)
	if err != nil {
		return false, 0, 0, err
	}

	if count >= p.MaxProfessionals {
		return false, count, p.MaxProfessionals, nil
	}
	return true, count, p.MaxProfessionals, nil
}

func CheckCustomersLimit(ctx context.Context, db *sqlx.DB, clientID string) (bool, int, int, error) {
	p, err := GetClientPlan(ctx, db, clientID)
	if err != nil {
		return false, 0, 0, err
	}
	if p.MaxCustomers == -1 {
		return true, 0, 0, nil
	}

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM customer WHERE client_id = ?", clientID)
	if err != nil {
		return false, 0, 0, err
	}

	if count >= p.MaxCustomers {
		return false, count, p.MaxCustomers, nil
	}
	return true, count, p.MaxCustomers, nil
}

func CheckUsersLimit(ctx context.Context, db *sqlx.DB, clientID string) (bool, int, int, error) {
	p, err := GetClientPlan(ctx, db, clientID)
	if err != nil {
		return false, 0, 0, err
	}
	if p.MaxUsers == -1 {
		return true, 0, 0, nil
	}

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM client_user_link WHERE client_id = ? AND status = 'active'", clientID)
	if err != nil {
		return false, 0, 0, err
	}

	if count >= p.MaxUsers {
		return false, count, p.MaxUsers, nil
	}
	return true, count, p.MaxUsers, nil
}

func RespondWithLimitExceeded(w http.ResponseWriter, limitName string, current, max int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "plan_limit_exceeded",
		"limit":   limitName,
		"current": current,
		"max":     max,
	})
}

// MIDDLEWARES DE RECURSOS E PLANOS

func RequireFeatureLoyalty(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID, _ := r.Context().Value("client_id").(string)
			if clientID == "" {
				next.ServeHTTP(w, r)
				return
			}
			p, err := GetClientPlan(r.Context(), db, clientID)
			if err != nil || p.HasLoyalty == 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"plan_feature_not_included","feature":"has_loyalty"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireFeatureStock(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID, _ := r.Context().Value("client_id").(string)
			if clientID == "" {
				next.ServeHTTP(w, r)
				return
			}
			p, err := GetClientPlan(r.Context(), db, clientID)
			if err != nil || p.HasStock == 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"plan_feature_not_included","feature":"has_stock"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireFeatureReports(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID, _ := r.Context().Value("client_id").(string)
			if clientID == "" {
				next.ServeHTTP(w, r)
				return
			}
			p, err := GetClientPlan(r.Context(), db, clientID)
			if err != nil || p.HasReports == 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"plan_feature_not_included","feature":"has_reports"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireClientSelected substitui a antiga RequireActiveClient: além de checar
// se a barbearia está bloqueada, bloqueia ativamente quem não é admin e ainda
// não escolheu uma barbearia ativa (client_id vazio) — caso de usuários com
// 2+ vínculos que ainda não passaram pelo seletor.
func RequireClientSelected(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value("user_role").(string)
			clientID, _ := r.Context().Value("client_id").(string)

			if clientID == "" {
				if role == "admin" {
					next.ServeHTTP(w, r)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"client_not_selected","message":"Selecione uma barbearia para continuar."}`))
				return
			}

			var status string
			query := "SELECT status FROM client WHERE id = ? LIMIT 1"
			err := db.GetContext(r.Context(), &status, query, clientID)
			if err == nil && status == "blocked" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"client_blocked","message":"Sua barbearia está suspensa ou bloqueada pelo administrador."}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// HANDLER PARA INDICADORES DE USO DO CLIENTE

type LimitHandler struct {
	db *sqlx.DB
}

func NewLimitHandler(db *sqlx.DB) *LimitHandler {
	return &LimitHandler{db: db}
}

func (h *LimitHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	clientID, _ := r.Context().Value("client_id").(string)

	p, err := GetClientPlan(r.Context(), h.db, clientID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"erro ao carregar plano"}`))
		return
	}

	var countProfs int
	_ = h.db.GetContext(r.Context(), &countProfs, "SELECT COUNT(*) FROM professional WHERE client_id = ?", clientID)

	var countCustomers int
	_ = h.db.GetContext(r.Context(), &countCustomers, "SELECT COUNT(*) FROM customer WHERE client_id = ?", clientID)

	var countUsers int
	_ = h.db.GetContext(r.Context(), &countUsers, "SELECT COUNT(*) FROM client_user_link WHERE client_id = ? AND status = 'active'", clientID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(UsageResponse{
		Plan:          p,
		Professionals: countProfs,
		Customers:     countCustomers,
		Users:         countUsers,
	})
}
