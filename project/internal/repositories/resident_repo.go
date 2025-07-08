package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

// user_unit
const (
	CREATE_RESIDENTS_TABLE = `CREATE TABLE IF NOT EXISTS residents(
		id SERIAL PRIMARY KEY,
		user_id SERIAL REFERENCES users(id) ON DELETE CASCADE,
		unit_id SERIAL REFERENCES units(id) ON DELETE CASCADE,
		move_in_date DATE NOT NULL,
		move_out_date DATE,
		is_primary_resident BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		);`
)

type ResidentRepository interface {
	CreateResident(userID, unitID int, moveInDate string, moveOutDate *string, isPrimaryResident bool) (int, error)
	GetResidentByID(id int) (*models.Resident, error)
	GetResidentsByUnitID(unitID int) ([]models.Resident, error)
}

type residentRepositoryImpl struct {
	db *sqlx.DB
}

func NewResidentRepository(autoCreate bool, db *sqlx.DB) (ResidentRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_RESIDENTS_TABLE); err != nil {
			return nil, err
		}
	}
	return &residentRepositoryImpl{db: db}, nil
}

func (r *residentRepositoryImpl) CreateResident(userID, unitID int, moveInDate string, moveOutDate *string, isPrimaryResident bool) (int, error) {
	var id int
	query := `INSERT INTO residents (user_id, unit_id, move_in_date, move_out_date, is_primary_resident) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, userID, unitID, moveInDate, moveOutDate, isPrimaryResident).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (r *residentRepositoryImpl) GetResidentByID(id int) (*models.Resident, error) {
	var resident models.Resident
	query := `SELECT id, user_id, unit_id, move_in_date, move_out_date, is_primary_resident, created_at, updated_at 
			  FROM residents WHERE id = $1`
	err := r.db.Get(&resident, query, id)
	if err != nil {
		return nil, err
	}
	return &resident, nil
}

func (r *residentRepositoryImpl) GetResidentsByUnitID(unitID int) ([]models.Resident, error) {
	var residents []models.Resident
	query := `SELECT id, user_id, unit_id, move_in_date, move_out_date, is_primary_resident, created_at, updated_at 
			  FROM residents WHERE unit_id = $1`
	err := r.db.Select(&residents, query, unitID)
	if err != nil {
		return nil, err
	}
	return residents, nil
}
