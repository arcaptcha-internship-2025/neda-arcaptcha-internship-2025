package repositories

import (
	"context"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
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
        telegram_user VARCHAR(100) UNIQUE,
        telegram_chat_id BIGINT,
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
	GetUserByTelegramUser(telegramUser string) (*models.User, error)
	UpdateTelegramChatID(ctx context.Context, telegramUsername string, chatID int64) error
}

type userRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserRepository(autoCreate bool, db *sqlx.DB) UserRepository {
	if autoCreate {
		if _, err := db.Exec(CREATE_USERS_TABLE); err != nil {
			log.Fatalf("failed to create users table: %v", err)
		}
	}
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) CreateUser(ctx context.Context, user models.User) (int, error) {
	query := `INSERT INTO users (username, password, email, phone, full_name, user_type, telegram_user) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) 
	          RETURNING id`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Password,
		user.Email,
		user.Phone,
		user.FullName,
		user.UserType,
		user.TelegramUser).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *userRepositoryImpl) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, username, password, email, phone, full_name, user_type, 
	          telegram_user, telegram_chat_id, created_at, updated_at 
	          FROM users WHERE id = $1`
	var user models.User
	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateUser(ctx context.Context, user models.User) error {
	query := `UPDATE users SET 
		username = :username, 
		password = :password, 
		email = :email, 
		phone = :phone, 
		full_name = :full_name, 
		user_type = :user_type, 
		telegram_user = :telegram_user,
		telegram_chat_id = :telegram_chat_id,
		updated_at = CURRENT_TIMESTAMP 
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepositoryImpl) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *userRepositoryImpl) GetAllUsers(ctx context.Context) ([]models.User, error) {
	query := `SELECT id, username, password, email, phone, full_name, user_type, 
	          telegram_user, telegram_chat_id, created_at, updated_at 
	          FROM users`
	var users []models.User
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryImpl) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password, email, phone, full_name, user_type, 
	          telegram_user, telegram_chat_id, created_at, updated_at 
	          FROM users WHERE username = $1`
	var user models.User
	if err := r.db.Get(&user, query, username); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, password, email, phone, full_name, user_type, 
	          telegram_user, telegram_chat_id, created_at, updated_at 
	          FROM users WHERE email = $1`
	var user models.User
	if err := r.db.Get(&user, query, email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetUserByPhone(phone string) (*models.User, error) {
	query := `SELECT id, username, password, email, phone, full_name, user_type, 
	          telegram_user, telegram_chat_id, created_at, updated_at 
	          FROM users WHERE phone = $1`
	var user models.User
	if err := r.db.Get(&user, query, phone); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetUserByTelegramUser(telegramUser string) (*models.User, error) {
	query := `SELECT id, username, password, email, phone, full_name, user_type, 
	          telegram_user, telegram_chat_id, created_at, updated_at 
	          FROM users WHERE telegram_user = $1`
	var user models.User
	if err := r.db.Get(&user, query, telegramUser); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateTelegramChatID(ctx context.Context, telegramUsername string, chatID int64) error {
	query := `UPDATE users SET telegram_chat_id = $1, updated_at = CURRENT_TIMESTAMP WHERE telegram_user = $2`
	_, err := r.db.ExecContext(ctx, query, chatID, telegramUsername)
	return err
}
