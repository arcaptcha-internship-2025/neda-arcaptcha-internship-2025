package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
)

const (
	CREATE_USERS_TABLE = `CREATE TABLE IF NOT EXISTS users(
        id SERIAL PRIMARY KEY,
        username VARCHAR(100) UNIQUE NOT NULL,
        password VARCHAR(2000) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        phone VARCHAR(20) UNIQUE NOT NULL,
        full_name VARCHAR(100) NOT NULL,
        user_type VARCHAR(20) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(id int) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByPhone(phone string) (*models.User, error)
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

func (r *userRepositoryImpl) CreateUser(ctx context.Context, user models.User) (int, error) {
	query := `INSERT INTO users (username, password_hash, email, phone, full_name, user_type) 
	          VALUES (:username, :password_hash, :email, :phone, :full_name, :user_type) 
	          RETURNING id`
	var id int
	if err := r.db.QueryRowxContext(ctx, query, user).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *userRepositoryImpl) GetUserByID(id int) (*models.User, error) {
	query := `SELECT * FROM users WHERE id = $1`
	var user models.User
	if err := r.db.Get(&user, query, id); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateUser(ctx context.Context, user models.User) error {
	query := `UPDATE users SET username = :username, password_hash = :password_hash, 
	          email = :email, phone = :phone, full_name = :full_name, user_type = :user_type, 
	          updated_at = CURRENT_TIMESTAMP WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepositoryImpl) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *userRepositoryImpl) GetAllUsers(ctx context.Context) ([]models.User, error) {
	query := `SELECT * FROM users`
	var users []models.User
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryImpl) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT * FROM users WHERE username = $1`
	var user models.User
	if err := r.db.Get(&user, query, username); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT * FROM users WHERE email = $1`
	var user models.User
	if err := r.db.Get(&user, query, email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetUserByPhone(phone string) (*models.User, error) {
	query := `SELECT * FROM users WHERE phone = $1`
	var user models.User
	if err := r.db.Get(&user, query, phone); err != nil {
		return nil, err
	}
	return &user, nil
}
