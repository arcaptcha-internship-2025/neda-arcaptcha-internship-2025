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
        units_num INTEGER NOT NULL,
        manager_id INTEGER REFERENCES users(id),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
)

type ApartmentRepository interface {
	CreateApartment(apartmentName, address string, unitsNum int, managerID int) (int, error)
	GetApartmentByID(id int) (*models.Apartment, error)
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

func (r *apartmentRepositoryImpl) CreateApartment(apartmentName, address string, unitsNum int, managerID int) (int, error) {
	var id int
	query := `INSERT INTO apartments (apartment_name, address, units_num, manager_id) 
              VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(query, apartmentName, address, unitsNum, managerID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *apartmentRepositoryImpl) GetApartmentByID(id int) (*models.Apartment, error) {
	var apartment models.Apartment
	query := `SELECT id, apartment_name, address, units_num, manager_id, created_at, updated_at 
              FROM apartments WHERE id = $1`
	err := r.db.Get(&apartment, query, id)
	if err != nil {
		return nil, err
	}
	return &apartment, nil
}
