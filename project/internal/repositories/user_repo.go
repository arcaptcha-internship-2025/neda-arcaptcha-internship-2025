package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_USERS_TABLE = `CREATE TABLE IF NOT EXISTS users(
        id SERIAL PRIMARY KEY,
        username VARCHAR(100) UNIQUE NOT NULL,
        password_hash VARCHAR(2000) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        phone VARCHAR(20) UNIQUE NOT NULL,
        full_name VARCHAR(100) NOT NULL,
        user_type VARCHAR(20) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
)

type UserRepository interface {
	CreateUser(username, passwordHash, email, phone, fullName, userType string) (int, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUser(id int, username, passwordHash, email, phone, fullName, userType string) error
	DeleteUser(id int) error
}

type userRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserRepository(autoCreate bool, db *sqlx.DB) (UserRepository, error) {
	if autoCreate {
		if _, err := db.Exec(CREATE_USERS_TABLE); err != nil {
			return nil, err
		}
	}
	return &userRepositoryImpl{db: db}, nil
}

func (r *userRepositoryImpl) CreateUser(username, passwordHash, email, phone, fullName, userType string) (int, error) {
	var id int
	query := `INSERT INTO users (username, password_hash, email, phone, full_name, user_type) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := r.db.QueryRow(query, username, passwordHash, email, phone, fullName, userType).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *userRepositoryImpl) GetUserByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password_hash, email, phone, full_name, user_type, created_at, updated_at 
			  FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateUser(id int, username, passwordHash, email, phone, fullName, userType string) error {
	query := `UPDATE users SET username = $1, password_hash = $2, email = $3, phone = $4, 
			  full_name = $5, user_type = $6, updated_at = CURRENT_TIMESTAMP WHERE id = $7`
	_, err := r.db.Exec(query, username, passwordHash, email, phone, fullName, userType, id)
	return err
}

func (r *userRepositoryImpl) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
