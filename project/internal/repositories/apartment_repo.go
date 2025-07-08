package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
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
	CreateApartment(apartmentName, address string, unitsCount int, managerID int) (int, error)
	GetApartmentByID(id int) (*models.Apartment, error)
	UpdateApartment(id int, apartmentName, address string, unitsCount int, managerID int) error
	DeleteApartment(id int) error
}

type apartmentRepositoryImpl struct {
	db *sqlx.DB
}

func NewApartmentRepository(autoCreate bool, db *sqlx.DB) (ApartmentRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_APARTMENTS_TABLE); err != nil {
			return nil, err
		}
	}
	return &apartmentRepositoryImpl{db: db}, nil
}

func (r *apartmentRepositoryImpl) CreateApartment(apartmentName, address string, unitsCount int, managerID int) (int, error) {
	var id int
	query := `INSERT INTO apartments (apartment_name, address, units_count, manager_id) 
              VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(query, apartmentName, address, unitsCount, managerID).Scan(&id)
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

func (r *apartmentRepositoryImpl) UpdateApartment(id int, apartmentName, address string, unitsCount int, managerID int) error {
	query := `UPDATE apartments SET apartment_name = $1, address = $2, units_count = $3, manager_id = $4, 
			  updated_at = CURRENT_TIMESTAMP WHERE id = $5`
	_, err := r.db.Exec(query, apartmentName, address, unitsCount, managerID, id)
	return err
}

func (r *apartmentRepositoryImpl) DeleteApartment(id int) error {
	query := `DELETE FROM apartments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
