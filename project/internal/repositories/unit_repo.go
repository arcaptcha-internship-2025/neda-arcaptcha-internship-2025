package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/nda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_UNITS_TABLE = `CREATE TABLE IF NOT EXISTS units(
		id SERIAL PRIMARY KEY,
		unit_number VARCHAR(20) NOT NULL,
		apartment_id INTEGER REFERENCES apartments(id) ON DELETE CASCADE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE (unit_number, apartment_id)
	);`
)

type UnitRepository interface {
	CreateUnit(unitNumber string, apartmentID int) (int, error)
	GetUnitByID(id int) (*models.Unit, error)
}

type unitRepositoryImpl struct {
	db *sqlx.DB
}

func NewUnitRepository(autoCreate bool, db *sqlx.DB) (UnitRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_UNITS_TABLE); err != nil {
			return nil, err
		}
	}
	return &unitRepositoryImpl{db: db}, nil
}

func (r *unitRepositoryImpl) CreateUnit(unitNumber string, apartmentID int) (int, error) {
	var id int
	query := `INSERT INTO units (unit_number, apartment_id) 
			  VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRow(query, unitNumber, apartmentID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *unitRepositoryImpl) GetUnitByID(id int) (*models.Unit, error) {
	var unit models.Unit
	query := `SELECT id, unit_number, apartment_id, created_at, updated_at 
			  FROM units WHERE id = $1`
	err := r.db.Get(&unit, query, id)
	if err != nil {
		return nil, err
	}
	return &unit, nil
}
