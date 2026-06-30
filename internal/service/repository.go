package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	ErrServiceNotFound  = errors.New("serviço não encontrado")
	ErrCategoryNotFound = errors.New("categoria não encontrada")
)

type ServiceRepository interface {
	List(ctx context.Context, clientID string, categoryID string) ([]Service, error)
	GetByID(ctx context.Context, clientID, id string) (*Service, error)
	Create(ctx context.Context, s *Service) error
	Update(ctx context.Context, s *Service) error
	Delete(ctx context.Context, clientID, id string) error
	ListCategories(ctx context.Context, clientID string) ([]ServiceCategory, error)
	CreateCategory(ctx context.Context, cat *ServiceCategory) error
}

type serviceRepository struct {
	db *sqlx.DB
}

func NewServiceRepository(db *sqlx.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) List(ctx context.Context, clientID string, categoryID string) ([]Service, error) {
	var list []Service
	var query string
	var args []interface{}

	if categoryID != "" {
		query = "SELECT * FROM service WHERE client_id = ? AND category_id = ? ORDER BY name ASC"
		args = []interface{}{clientID, categoryID}
	} else {
		query = "SELECT * FROM service WHERE client_id = ? ORDER BY name ASC"
		args = []interface{}{clientID}
	}

	err := r.db.SelectContext(ctx, &list, query, args...)
	return list, err
}

func (r *serviceRepository) GetByID(ctx context.Context, clientID, id string) (*Service, error) {
	var s Service
	query := "SELECT * FROM service WHERE client_id = ? AND id = ?"
	err := r.db.GetContext(ctx, &s, query, clientID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *serviceRepository) Create(ctx context.Context, s *Service) error {
	query := `INSERT INTO service (id, client_id, category_id, name, description, duration_minutes, price, photo_url, active, created_at)
	          VALUES (:id, :client_id, :category_id, :name, :description, :duration_minutes, :price, :photo_url, :active, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, s)
	return err
}

func (r *serviceRepository) Update(ctx context.Context, s *Service) error {
	query := `UPDATE service 
	          SET category_id = :category_id, name = :name, description = :description, duration_minutes = :duration_minutes, price = :price, active = :active 
	          WHERE client_id = :client_id AND id = :id`
	res, err := r.db.NamedExecContext(ctx, query, s)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrServiceNotFound
	}
	return nil
}

func (r *serviceRepository) Delete(ctx context.Context, clientID, id string) error {
	query := "DELETE FROM service WHERE client_id = ? AND id = ?"
	res, err := r.db.ExecContext(ctx, query, clientID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrServiceNotFound
	}
	return nil
}

func (r *serviceRepository) ListCategories(ctx context.Context, clientID string) ([]ServiceCategory, error) {
	var list []ServiceCategory
	query := "SELECT * FROM service_category WHERE client_id = ? ORDER BY name ASC"
	err := r.db.SelectContext(ctx, &list, query, clientID)
	return list, err
}

func (r *serviceRepository) CreateCategory(ctx context.Context, cat *ServiceCategory) error {
	query := `INSERT INTO service_category (id, client_id, name) VALUES (:id, :client_id, :name)`
	_, err := r.db.NamedExecContext(ctx, query, cat)
	return err
}
