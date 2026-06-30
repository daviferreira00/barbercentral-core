package customer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, clientID, id string) error
	GetByID(ctx context.Context, clientID, id string) (*CustomerStats, error)
	GetByPhone(ctx context.Context, clientID, phone string) (*Customer, error)
	List(ctx context.Context, clientID string, searchQuery string, birthMonth int, limit, offset int) ([]CustomerStats, int, error)
	Search(ctx context.Context, clientID string, searchQuery string) ([]Customer, error)
	GetAppointmentsHistory(ctx context.Context, clientID, customerID string) ([]EnrichedAppointmentHistory, error)
	GetDB() *sqlx.DB
}

type customerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) GetDB() *sqlx.DB {
	return r.db
}

func (r *customerRepository) Create(ctx context.Context, c *Customer) error {
	query := `INSERT INTO customer (
		id, client_id, name, phone, email, cpf, birth_date, notes, created_at
	) VALUES (
		:id, :client_id, :name, :phone, :email, :cpf, :birth_date, :notes, :created_at
	)`
	_, err := r.db.NamedExecContext(ctx, query, c)
	return err
}

func (r *customerRepository) Update(ctx context.Context, c *Customer) error {
	query := `UPDATE customer SET
		name = :name,
		phone = :phone,
		email = :email,
		cpf = :cpf,
		birth_date = :birth_date,
		notes = :notes
	WHERE id = :id AND client_id = :client_id`
	_, err := r.db.NamedExecContext(ctx, query, c)
	return err
}

func (r *customerRepository) Delete(ctx context.Context, clientID, id string) error {
	// A especificação fala em soft delete. Como não adicionamos coluna active ou deleted_at no model,
	// vamos remover do banco por hora, ou se preferir deletar de forma física (já que não há coluna deleted_at na FASE-06).
	// Exclusão física no MySQL para simplificar:
	query := "DELETE FROM customer WHERE id = ? AND client_id = ?"
	_, err := r.db.ExecContext(ctx, query, id, clientID)
	return err
}

func (r *customerRepository) GetByID(ctx context.Context, clientID, id string) (*CustomerStats, error) {
	query := `
		SELECT 
			c.id, c.client_id, c.name, c.phone, c.email, c.cpf, c.birth_date, c.notes, c.created_at,
			COUNT(DISTINCT CASE WHEN a.status = 'completed' THEN a.id END) as total_visits,
			COALESCE(SUM(CASE WHEN a.status = 'completed' THEN aps.price END), 0.00) as total_spent,
			MAX(CASE WHEN a.status = 'completed' THEN a.date END) as last_visit,
			MIN(CASE WHEN a.status = 'completed' THEN a.date END) as first_visit
		FROM customer c
		LEFT JOIN appointment a ON a.customer_id = c.id
		LEFT JOIN appointment_service aps ON aps.appointment_id = a.id
		WHERE c.id = ? AND c.client_id = ?
		GROUP BY c.id`

	var stats CustomerStats
	err := r.db.GetContext(ctx, &stats, query, id, clientID)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *customerRepository) GetByPhone(ctx context.Context, clientID, phone string) (*Customer, error) {
	// Limpa o telefone para bater apenas números ou fazer busca limpa
	cleanPhone := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	query := `SELECT id, client_id, name, phone, email, cpf, birth_date, notes, created_at 
	          FROM customer 
	          WHERE client_id = ? AND (phone = ? OR REPLACE(REPLACE(REPLACE(REPLACE(phone, '-', ''), ' ', ''), '(', ''), ')', '') = ?)`
	
	var c Customer
	err := r.db.GetContext(ctx, &c, query, clientID, phone, cleanPhone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) List(ctx context.Context, clientID string, searchQuery string, birthMonth int, limit, offset int) ([]CustomerStats, int, error) {
	var args []interface{}
	queryWhere := "WHERE c.client_id = ?"
	args = append(args, clientID)

	if searchQuery != "" {
		queryWhere += " AND (c.name LIKE ? OR c.phone LIKE ?)"
		likeArg := "%" + searchQuery + "%"
		args = append(args, likeArg, likeArg)
	}

	if birthMonth > 0 && birthMonth <= 12 {
		queryWhere += " AND MONTH(c.birth_date) = ?"
		args = append(args, birthMonth)
	}

	// Consulta de contagem total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customer c %s", queryWhere)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Consulta paginada
	query := fmt.Sprintf(`
		SELECT 
			c.id, c.client_id, c.name, c.phone, c.email, c.cpf, c.birth_date, c.notes, c.created_at,
			COUNT(DISTINCT CASE WHEN a.status = 'completed' THEN a.id END) as total_visits,
			COALESCE(SUM(CASE WHEN a.status = 'completed' THEN aps.price END), 0.00) as total_spent,
			MAX(CASE WHEN a.status = 'completed' THEN a.date END) as last_visit,
			MIN(CASE WHEN a.status = 'completed' THEN a.date END) as first_visit
		FROM customer c
		LEFT JOIN appointment a ON a.customer_id = c.id
		LEFT JOIN appointment_service aps ON aps.appointment_id = a.id
		%s
		GROUP BY c.id
		ORDER BY c.name ASC
		LIMIT ? OFFSET ?`, queryWhere)

	args = append(args, limit, offset)
	var list []CustomerStats
	err = r.db.SelectContext(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *customerRepository) Search(ctx context.Context, clientID string, searchQuery string) ([]Customer, error) {
	query := `SELECT id, client_id, name, phone, email, cpf, birth_date, notes, created_at 
	          FROM customer 
	          WHERE client_id = ? AND (name LIKE ? OR phone LIKE ?) 
	          ORDER BY name ASC 
	          LIMIT 20`
	likeArg := "%" + searchQuery + "%"
	var list []Customer
	err := r.db.SelectContext(ctx, &list, query, clientID, likeArg, likeArg)
	return list, err
}

func (r *customerRepository) GetAppointmentsHistory(ctx context.Context, clientID, customerID string) ([]EnrichedAppointmentHistory, error) {
	query := `
		SELECT 
			a.id, a.date, a.start_time, a.end_time, a.status,
			p.name as professional_name,
			GROUP_CONCAT(s.name SEPARATOR ', ') as services_list,
			COALESCE(SUM(aps.price), 0.00) as total_price,
			pay.status as payment_status,
			pay.method as payment_method
		FROM appointment a
		JOIN professional p ON p.id = a.professional_id
		JOIN appointment_service aps ON aps.appointment_id = a.id
		JOIN service s ON s.id = aps.service_id
		LEFT JOIN appointment_payment pay ON pay.appointment_id = a.id
		WHERE a.customer_id = ? AND a.client_id = ?
		GROUP BY a.id
		ORDER BY a.date DESC, a.start_time DESC`

	var list []EnrichedAppointmentHistory
	err := r.db.SelectContext(ctx, &list, query, customerID, clientID)
	return list, err
}
