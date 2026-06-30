package appointment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrAppointmentNotFound = errors.New("agendamento não encontrado")
	ErrBlockedSlotNotFound = errors.New("bloqueio de agenda não encontrado")
)

type AppointmentRepository interface {
	ListBlockedSlots(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]BlockedSlot, error)
	CreateBlockedSlot(ctx context.Context, slot *BlockedSlot) error
	DeleteBlockedSlot(ctx context.Context, clientID, id string) error
	List(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]EnrichedAppointment, error)
	GetByID(ctx context.Context, clientID, id string) (*EnrichedAppointment, error)
	UpdateStatus(ctx context.Context, clientID, id, status string) error
	Create(ctx context.Context, app *Appointment, services []AppointmentService) error
	GetByCancelToken(ctx context.Context, token string) (*EnrichedAppointment, error)
	GetUpcomingAppointmentsForReminder(ctx context.Context, date string) ([]EnrichedAppointment, error)
	MarkReminderSent(ctx context.Context, id string) error

	// Novos métodos para logs, edição e remoção
	CreateStatusLog(ctx context.Context, log *AppointmentStatusLog) error
	GetStatusLogs(ctx context.Context, clientID, appointmentID string) ([]AppointmentStatusLog, error)
	Update(ctx context.Context, app *Appointment, services []AppointmentService) error
	Delete(ctx context.Context, clientID, id string) error
}

type appointmentRepository struct {
	db *sqlx.DB
}

func NewAppointmentRepository(db *sqlx.DB) AppointmentRepository {
	return &appointmentRepository{db: db}
}

func (r *appointmentRepository) ListBlockedSlots(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]BlockedSlot, error) {
	var list []BlockedSlot
	query := "SELECT * FROM blocked_slot WHERE client_id = ?"
	args := []interface{}{clientID}

	if professionalID != "" {
		query += " AND (professional_id = ? OR professional_id IS NULL)"
		args = append(args, professionalID)
	}

	if startDate != "" && endDate != "" {
		query += " AND date BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " ORDER BY date ASC, start_time ASC"

	err := r.db.SelectContext(ctx, &list, query, args...)
	return list, err
}

func (r *appointmentRepository) CreateBlockedSlot(ctx context.Context, slot *BlockedSlot) error {
	query := `INSERT INTO blocked_slot (id, client_id, professional_id, date, start_time, end_time, reason, created_by, created_at)
	          VALUES (:id, :client_id, :professional_id, :date, :start_time, :end_time, :reason, :created_by, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, slot)
	return err
}

func (r *appointmentRepository) DeleteBlockedSlot(ctx context.Context, clientID, id string) error {
	query := "DELETE FROM blocked_slot WHERE client_id = ? AND id = ?"
	res, err := r.db.ExecContext(ctx, query, clientID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBlockedSlotNotFound
	}
	return nil
}

func (r *appointmentRepository) List(ctx context.Context, clientID, professionalID, startDate, endDate string) ([]EnrichedAppointment, error) {
	var appointments []Appointment
	query := "SELECT * FROM appointment WHERE client_id = ?"
	args := []interface{}{clientID}

	if professionalID != "" {
		query += " AND professional_id = ?"
		args = append(args, professionalID)
	}

	if startDate != "" && endDate != "" {
		query += " AND date BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " ORDER BY date ASC, start_time ASC"

	err := r.db.SelectContext(ctx, &appointments, query, args...)
	if err != nil {
		return nil, err
	}

	var enrichedList []EnrichedAppointment
	for _, app := range appointments {
		// Busca nome do profissional
		var profName string
		_ = r.db.GetContext(ctx, &profName, "SELECT name FROM professional WHERE id = ?", app.ProfessionalID)

		// Busca serviços
		var services []EnrichedAppService
		querySrv := `SELECT s.id as service_id, s.name as service_name, aps.price, aps.duration_minutes
		             FROM appointment_service aps
		             JOIN service s ON aps.service_id = s.id
		             WHERE aps.appointment_id = ?`
		_ = r.db.SelectContext(ctx, &services, querySrv, app.ID)

		enrichedList = append(enrichedList, EnrichedAppointment{
			Appointment:      app,
			ProfessionalName: profName,
			Services:         services,
		})
	}

	return enrichedList, nil
}

func (r *appointmentRepository) GetByID(ctx context.Context, clientID, id string) (*EnrichedAppointment, error) {
	var app Appointment
	query := "SELECT * FROM appointment WHERE client_id = ? AND id = ?"
	err := r.db.GetContext(ctx, &app, query, clientID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	var profName string
	_ = r.db.GetContext(ctx, &profName, "SELECT name FROM professional WHERE id = ?", app.ProfessionalID)

	var services []EnrichedAppService
	querySrv := `SELECT s.id as service_id, s.name as service_name, aps.price, aps.duration_minutes
	             FROM appointment_service aps
	             JOIN service s ON aps.service_id = s.id
	             WHERE aps.appointment_id = ?`
	_ = r.db.SelectContext(ctx, &services, querySrv, app.ID)

	return &EnrichedAppointment{
		Appointment:      app,
		ProfessionalName: profName,
		Services:         services,
	}, nil
}

func (r *appointmentRepository) UpdateStatus(ctx context.Context, clientID, id, status string) error {
	query := "UPDATE appointment SET status = ? WHERE client_id = ? AND id = ?"
	res, err := r.db.ExecContext(ctx, query, status, clientID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAppointmentNotFound
	}
	return nil
}

func (r *appointmentRepository) Create(ctx context.Context, app *Appointment, services []AppointmentService) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryApp := `INSERT INTO appointment (id, client_id, professional_id, customer_id, date, start_time, end_time, status, notes, cancel_token, customer_name, customer_phone, customer_email, reminder_sent, source, created_at)
	             VALUES (:id, :client_id, :professional_id, :customer_id, :date, :start_time, :end_time, :status, :notes, :cancel_token, :customer_name, :customer_phone, :customer_email, :reminder_sent, :source, :created_at)`
	_, err = tx.NamedExecContext(ctx, queryApp, app)
	if err != nil {
		return err
	}

	querySrv := `INSERT INTO appointment_service (id, appointment_id, service_id, price, duration_minutes)
	             VALUES (:id, :appointment_id, :service_id, :price, :duration_minutes)`
	for _, s := range services {
		_, err = tx.NamedExecContext(ctx, querySrv, &s)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *appointmentRepository) GetByCancelToken(ctx context.Context, token string) (*EnrichedAppointment, error) {
	var app Appointment
	query := "SELECT * FROM appointment WHERE cancel_token = ?"
	err := r.db.GetContext(ctx, &app, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}

	var profName string
	_ = r.db.GetContext(ctx, &profName, "SELECT name FROM professional WHERE id = ?", app.ProfessionalID)

	var services []EnrichedAppService
	querySrv := `SELECT s.id as service_id, s.name as service_name, aps.price, aps.duration_minutes
	             FROM appointment_service aps
	             JOIN service s ON aps.service_id = s.id
	             WHERE aps.appointment_id = ?`
	_ = r.db.SelectContext(ctx, &services, querySrv, app.ID)

	return &EnrichedAppointment{
		Appointment:      app,
		ProfessionalName: profName,
		Services:         services,
	}, nil
}

func (r *appointmentRepository) GetUpcomingAppointmentsForReminder(ctx context.Context, date string) ([]EnrichedAppointment, error) {
	var appointments []Appointment
	query := "SELECT * FROM appointment WHERE date = ? AND status IN ('pending', 'confirmed') AND reminder_sent = 0"
	err := r.db.SelectContext(ctx, &appointments, query, date)
	if err != nil {
		return nil, err
	}

	var enrichedList []EnrichedAppointment
	for _, app := range appointments {
		var profName string
		_ = r.db.GetContext(ctx, &profName, "SELECT name FROM professional WHERE id = ?", app.ProfessionalID)

		var services []EnrichedAppService
		querySrv := `SELECT s.id as service_id, s.name as service_name, aps.price, aps.duration_minutes
		             FROM appointment_service aps
		             JOIN service s ON aps.service_id = s.id
		             WHERE aps.appointment_id = ?`
		_ = r.db.SelectContext(ctx, &services, querySrv, app.ID)

		enrichedList = append(enrichedList, EnrichedAppointment{
			Appointment:      app,
			ProfessionalName: profName,
			Services:         services,
		})
	}

	return enrichedList, nil
}

func (r *appointmentRepository) MarkReminderSent(ctx context.Context, id string) error {
	query := "UPDATE appointment SET reminder_sent = 1 WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *appointmentRepository) CreateStatusLog(ctx context.Context, log *AppointmentStatusLog) error {
	query := `INSERT INTO appointment_status_log (id, appointment_id, from_status, to_status, changed_by, notes, created_at)
	          VALUES (:id, :appointment_id, :from_status, :to_status, :changed_by, :notes, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, log)
	return err
}

func (r *appointmentRepository) GetStatusLogs(ctx context.Context, clientID, appointmentID string) ([]AppointmentStatusLog, error) {
	// Garante que o agendamento pertence ao client_id
	var check int
	err := r.db.GetContext(ctx, &check, "SELECT COUNT(*) FROM appointment WHERE id = ? AND client_id = ?", appointmentID, clientID)
	if err != nil {
		return nil, err
	}
	if check == 0 {
		return []AppointmentStatusLog{}, nil
	}

	var logs []AppointmentStatusLog
	query := "SELECT * FROM appointment_status_log WHERE appointment_id = ? ORDER BY created_at ASC"
	err = r.db.SelectContext(ctx, &logs, query, appointmentID)
	return logs, err
}

func (r *appointmentRepository) Update(ctx context.Context, app *Appointment, services []AppointmentService) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryApp := `UPDATE appointment SET
		professional_id = :professional_id,
		customer_id = :customer_id,
		date = :date,
		start_time = :start_time,
		end_time = :end_time,
		status = :status,
		notes = :notes,
		customer_name = :customer_name,
		customer_phone = :customer_phone,
		customer_email = :customer_email
	WHERE id = :id AND client_id = :client_id`
	_, err = tx.NamedExecContext(ctx, queryApp, app)
	if err != nil {
		return err
	}

	// Deleta serviços antigos
	_, err = tx.ExecContext(ctx, "DELETE FROM appointment_service WHERE appointment_id = ?", app.ID)
	if err != nil {
		return err
	}

	// Insere novos
	querySrv := `INSERT INTO appointment_service (id, appointment_id, service_id, price, duration_minutes)
	             VALUES (:id, :appointment_id, :service_id, :price, :duration_minutes)`
	for _, s := range services {
		_, err = tx.NamedExecContext(ctx, querySrv, &s)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *appointmentRepository) Delete(ctx context.Context, clientID, id string) error {
	query := "DELETE FROM appointment WHERE id = ? AND client_id = ?"
	_, err := r.db.ExecContext(ctx, query, id, clientID)
	return err
}

