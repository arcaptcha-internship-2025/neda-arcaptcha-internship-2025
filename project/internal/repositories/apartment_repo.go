package repositories

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

const (
	CREATE_APARTMENTS_TABLE = `CREATE TABLE IF NOT EXIST apartments(
		id SERIAL PRIMARY KEY,
        apartment_name VARCHAR(100) NOT NULL,
        address TEXT NOT NULL,
        units_count INTEGER NOT NULL,
        manager_id INTEGER REFERENCES users(id),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
)

type ApartmentRepository interface {
	CreateApartment(ctx context.Context, apartment models.Apartment) (int, error)
	GetApartmentByID(id int) (*models.Apartment, error)
	UpdateApartment(ctx context.Context, apartment models.Apartment) error
	DeleteApartment(id int) error
	GetAllApartments(ctx context.Context) ([]models.Apartment, error)
}

type apartmentRepositoryImpl struct {
	db *sqlx.DB
}

func NewApartmentRepository(autoCreate bool, db *sqlx.DB) ApartmentRepository {
	if autoCreate {
		if _, err := db.Exec(CREATE_APARTMENTS_TABLE); err != nil {
			log.Fatalf("failed to create apartments table: %v", err)
		}
	}
	return &apartmentRepositoryImpl{db: db}
}

func (r *apartmentRepositoryImpl) CreateApartment(ctx context.Context, apartment models.Apartment) (int, error) {
	query := `INSERT INTO apartments (apartment_name, address, units_count, manager_id) 
	          VALUES (:apartment_name, :address, :units_count, :manager_id) 
	          RETURNING id`
	var id int
	if err := r.db.QueryRowxContext(ctx, query, apartment).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *apartmentRepositoryImpl) GetApartmentByID(id int) (*models.Apartment, error) {
	var apartment models.Apartment
	query := `SELECT id, apartment_name, address, units_count, manager_id, created_at, updated_at 
              FROM apartments WHERE id = $1`
	err := r.db.Get(&apartment, query, id)
	if err != nil {
		return nil, err
	}
	return &apartment, nil
}

func (r *apartmentRepositoryImpl) UpdateApartment(ctx context.Context, apartment models.Apartment) error {
	query := `UPDATE apartments SET apartment_name = :apartment_name, address = :address, 
	          units_count = :units_count, manager_id = :manager_id, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, apartment)
	return err
}

func (r *apartmentRepositoryImpl) DeleteApartment(id int) error {
	query := `DELETE FROM apartments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *apartmentRepositoryImpl) GetAllApartments(ctx context.Context) ([]models.Apartment, error) {
	var apartments []models.Apartment
	query := `SELECT id, apartment_name, address, units_count, manager_id, created_at, updated_at 
			  FROM apartments`
	err := r.db.SelectContext(ctx, &apartments, query)
	if err != nil {
		return nil, err
	}
	return apartments, nil
}
