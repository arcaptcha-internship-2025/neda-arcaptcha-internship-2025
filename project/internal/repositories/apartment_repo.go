package repositories

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

const (
	CREATE_APARTMENTS_TABLE = `CREATE TABLE IF NOT EXISTS apartments(
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
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		apartment.ApartmentName,
		apartment.Address,
		apartment.UnitsCount,
		apartment.ManagerID).Scan(&id)
	if err != nil {
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
	query := `UPDATE apartments SET apartment_name = $1, address = $2,
		units_count = $3, manager_id = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query,
		apartment.ApartmentName,
		apartment.Address,
		apartment.UnitsCount,
		apartment.ManagerID,
		apartment.ID)
	return err
}

func (r *apartmentRepositoryImpl) DeleteApartment(id int) error {
	query := `DELETE FROM apartments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
