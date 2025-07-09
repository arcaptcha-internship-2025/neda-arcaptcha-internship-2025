package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_USER_APARTMENT_TABLE = `CREATE TABLE IF NOT EXIST user_apartments(
		user_id SERIAL REFERENCE users(id) ON DELETE CASCADE,
		apartment_id SERIAL REFERENCE apartments(id) ON DELETE CASCADE,
		is_manager BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, apartment_id),
	);`
)

type UserApartmentRepository interface {
	CreateUserApartment(ctx context.Context, user_apartment models.User_apartment) error
	GetUserApartmentByID(userID, apartmentID int) (*models.User_apartment, error)
	UpdateUserApartment(ctx context.Context, user_apartment models.User_apartment) error
	DeleteUserApartment(userID, apartmentID int) error
}

type userApartmentRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserApartmentRepository(autoCreate bool, db *sqlx.DB) (UserApartmentRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_USER_APARTMENT_TABLE); err != nil {
			return nil, err
		}
	}
	return &userApartmentRepositoryImpl{db: db}, nil
}

func (r *userApartmentRepositoryImpl) CreateUserApartment(ctx context.Context, user_apartment models.User_apartment) error {
	query := `INSERT INTO user_apartments (user_id, apartment_id, is_manager) 
			  VALUES (:user_id, :apartment_id, :is_manager)`
	_, err := r.db.NamedExecContext(ctx, query, user_apartment)
	if err != nil {
		return err
	}
	return nil
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
	if err != nil {
		return err
	}
	return nil
}

func (r *userApartmentRepositoryImpl) DeleteUserApartment(userID, apartmentID int) error {
	query := `DELETE FROM user_apartments WHERE user_id = $1 AND apartment_id = $2`
	_, err := r.db.Exec(query, userID, apartmentID)
	if err != nil {
		return err
	}
	return nil
}
