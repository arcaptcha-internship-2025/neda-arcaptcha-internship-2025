package repositories

import (
	"context"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

const (
	CREATE_USER_APARTMENT_TABLE = `CREATE TABLE IF NOT EXISTS user_apartments(
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		apartment_id INTEGER REFERENCES apartments(id) ON DELETE CASCADE,
		is_manager BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, apartment_id)
	);`
)

type UserApartmentRepository interface {
	CreateUserApartment(ctx context.Context, user_apartment models.User_apartment) error
	GetResidentsInApartment(apartmentID int) ([]models.User, error)
	GetUserApartmentByID(userID, apartmentID int) (*models.User_apartment, error)
	UpdateUserApartment(ctx context.Context, user_apartment models.User_apartment) error
	DeleteUserApartment(userID, apartmentID int) error
	GetAllApartmentsForAResident(residentID int) ([]models.Apartment, error)
	IsUserManagerOfApartment(ctx context.Context, userID, apartmentID int) (bool, error)
	IsUserInApartment(ctx context.Context, userID, apartmentID int) (bool, error)
}

type userApartmentRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserApartmentRepository(autoCreate bool, db *sqlx.DB) UserApartmentRepository {
	if autoCreate {
		if _, err := db.Exec(CREATE_USER_APARTMENT_TABLE); err != nil {
			log.Fatalf("failed to create user_apartments table: %v", err)
		}
	}
	return &userApartmentRepositoryImpl{db: db}
}

func (r *userApartmentRepositoryImpl) CreateUserApartment(ctx context.Context, user_apartment models.User_apartment) error {
	query := `INSERT INTO user_apartments (user_id, apartment_id, is_manager) 
			  VALUES (:user_id, :apartment_id, :is_manager)`
	_, err := r.db.NamedExecContext(ctx, query, user_apartment)
	return err
}

func (r *userApartmentRepositoryImpl) GetUserApartmentByID(userID, apartmentID int) (*models.User_apartment, error) {
	var userApartment models.User_apartment
	query := `SELECT user_id, apartment_id, is_manager, created_at, updated_at 
			  FROM user_apartments WHERE user_id = $1 AND apartment_id = $2`
	err := r.db.Get(&userApartment, query, userID, apartmentID)
	if err != nil {
		return nil, err
	}
	return &userApartment, nil
}

func (r *userApartmentRepositoryImpl) UpdateUserApartment(ctx context.Context, user_apartment models.User_apartment) error {
	query := `UPDATE user_apartments 
			  SET is_manager = :is_manager, updated_at = CURRENT_TIMESTAMP 
			  WHERE user_id = :user_id AND apartment_id = :apartment_id`
	_, err := r.db.NamedExecContext(ctx, query, user_apartment)
	return err
}

func (r *userApartmentRepositoryImpl) DeleteUserApartment(userID, apartmentID int) error {
	query := `DELETE FROM user_apartments WHERE user_id = $1 AND apartment_id = $2`
	_, err := r.db.Exec(query, userID, apartmentID)
	return err
}

func (r *userApartmentRepositoryImpl) GetResidentsInApartment(apartmentID int) ([]models.User, error) {
	var residents []models.User
	query := `SELECT u.id, u.username, u.email, u.phone, u.full_name, u.user_type, u.created_at, u.updated_at
          FROM users u
          JOIN user_apartments ua ON u.id = ua.user_id
          WHERE ua.apartment_id = $1`
	err := r.db.Select(&residents, query, apartmentID)
	if err != nil {
		return nil, err
	}
	return residents, nil
}

func (r *userApartmentRepositoryImpl) GetAllApartmentsForAResident(residentID int) ([]models.Apartment, error) {
	var apartments []models.Apartment
	query := `SELECT a.id, a.apartment_name, a.address, a.units_count, a.manager_id, a.created_at, a.updated_at
			  FROM apartments a
			  JOIN user_apartments ua ON a.id = ua.apartment_id
			  WHERE ua.user_id = $1`
	err := r.db.Select(&apartments, query, residentID)
	if err != nil {
		return nil, err
	}
	return apartments, nil
}

func (r *userApartmentRepositoryImpl) IsUserManagerOfApartment(ctx context.Context, userID, apartmentID int) (bool, error) {
	var isManager bool
	query := `SELECT is_manager FROM user_apartments 
			  WHERE user_id = $1 AND apartment_id = $2`
	err := r.db.GetContext(ctx, &isManager, query, userID, apartmentID)
	if err != nil || !isManager {
		if !isManager {
			return false, errors.New("not manager")
		}
		return false, err
	}
	return true, nil
}

func (r *userApartmentRepositoryImpl) IsUserInApartment(ctx context.Context, userID, apartmentID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM user_apartments 
		WHERE user_id = $1 AND apartment_id = $2
	)`
	err := r.db.GetContext(ctx, &exists, query, userID, apartmentID)
	if err != nil || !exists {
		if !exists {
			return false, errors.New("not in apartment")
		}
		return false, err
	}
	return true, nil
}
