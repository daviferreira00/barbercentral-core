package professional

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrProfessionalNotFound = errors.New("profissional não encontrado")
)

type ProfessionalRepository interface {
	List(ctx context.Context, clientID string, status string) ([]Professional, error)
	GetByID(ctx context.Context, clientID, id string) (*Professional, error)
	Create(ctx context.Context, p *Professional) error
	Update(ctx context.Context, p *Professional) error
	Delete(ctx context.Context, clientID, id string) error
	GetSchedule(ctx context.Context, clientID, professionalID string) ([]ProfessionalSchedule, error)
	SaveSchedule(ctx context.Context, schedules []ProfessionalSchedule) error
	GetLinkedServices(ctx context.Context, clientID, professionalID string) ([]ProfessionalServiceLink, error)
	LinkService(ctx context.Context, link *ProfessionalServiceLink) error
	UnlinkService(ctx context.Context, clientID, professionalID, serviceID string) error
	UpdatePhoto(ctx context.Context, clientID, id, photoURL string) error
}

type professionalRepository struct {
	db *sqlx.DB
}

func NewProfessionalRepository(db *sqlx.DB) ProfessionalRepository {
	return &professionalRepository{db: db}
}

func (r *professionalRepository) List(ctx context.Context, clientID string, status string) ([]Professional, error) {
	var list []Professional
	var query string
	var args []interface{}

	if status != "" {
		query = "SELECT * FROM professional WHERE client_id = ? AND status = ? ORDER BY name ASC"
		args = []interface{}{clientID, status}
	} else {
		query = "SELECT * FROM professional WHERE client_id = ? ORDER BY name ASC"
		args = []interface{}{clientID}
	}

	err := r.db.SelectContext(ctx, &list, query, args...)
	return list, err
}

func (r *professionalRepository) GetByID(ctx context.Context, clientID, id string) (*Professional, error) {
	var p Professional
	query := "SELECT * FROM professional WHERE client_id = ? AND id = ?"
	err := r.db.GetContext(ctx, &p, query, clientID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProfessionalNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *professionalRepository) Create(ctx context.Context, p *Professional) error {
	query := `INSERT INTO professional (id, client_id, user_id, name, bio, photo_url, status, created_at)
	          VALUES (:id, :client_id, :user_id, :name, :bio, :photo_url, :status, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
}

func (r *professionalRepository) Update(ctx context.Context, p *Professional) error {
	query := `UPDATE professional 
	          SET name = :name, bio = :bio, status = :status 
	          WHERE client_id = :client_id AND id = :id`
	res, err := r.db.NamedExecContext(ctx, query, p)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProfessionalNotFound
	}
	return nil
}

func (r *professionalRepository) Delete(ctx context.Context, clientID, id string) error {
	query := "DELETE FROM professional WHERE client_id = ? AND id = ?"
	res, err := r.db.ExecContext(ctx, query, clientID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProfessionalNotFound
	}
	return nil
}

func (r *professionalRepository) GetSchedule(ctx context.Context, clientID, professionalID string) ([]ProfessionalSchedule, error) {
	var list []ProfessionalSchedule
	query := "SELECT * FROM professional_schedule WHERE client_id = ? AND professional_id = ? ORDER BY weekday ASC"
	err := r.db.SelectContext(ctx, &list, query, clientID, professionalID)
	return list, err
}

func (r *professionalRepository) SaveSchedule(ctx context.Context, schedules []ProfessionalSchedule) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryDelete := "DELETE FROM professional_schedule WHERE client_id = ? AND professional_id = ?"
	queryInsert := `INSERT INTO professional_schedule (id, professional_id, client_id, weekday, start_time, end_time, enabled)
	                VALUES (:id, :professional_id, :client_id, :weekday, :start_time, :end_time, :enabled)`

	if len(schedules) > 0 {
		_, err = tx.ExecContext(ctx, queryDelete, schedules[0].ClientID, schedules[0].ProfessionalID)
		if err != nil {
			return err
		}

		for _, s := range schedules {
			_, err = tx.NamedExecContext(ctx, queryInsert, &s)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *professionalRepository) GetLinkedServices(ctx context.Context, clientID, professionalID string) ([]ProfessionalServiceLink, error) {
	var list []ProfessionalServiceLink
	query := "SELECT * FROM professional_service WHERE client_id = ? AND professional_id = ?"
	err := r.db.SelectContext(ctx, &list, query, clientID, professionalID)
	return list, err
}

func (r *professionalRepository) LinkService(ctx context.Context, link *ProfessionalServiceLink) error {
	query := `INSERT INTO professional_service (professional_id, service_id, client_id, custom_price, custom_duration)
	          VALUES (:professional_id, :service_id, :client_id, :custom_price, :custom_duration)
	          ON DUPLICATE KEY UPDATE custom_price = VALUES(custom_price), custom_duration = VALUES(custom_duration)`
	_, err := r.db.NamedExecContext(ctx, query, link)
	return err
}

func (r *professionalRepository) UnlinkService(ctx context.Context, clientID, professionalID, serviceID string) error {
	query := "DELETE FROM professional_service WHERE client_id = ? AND professional_id = ? AND service_id = ?"
	_, err := r.db.ExecContext(ctx, query, clientID, professionalID, serviceID)
	return err
}

func (r *professionalRepository) UpdatePhoto(ctx context.Context, clientID, id, photoURL string) error {
	query := "UPDATE professional SET photo_url = ? WHERE client_id = ? AND id = ?"
	res, err := r.db.ExecContext(ctx, query, photoURL, clientID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProfessionalNotFound
	}
	return nil
}
