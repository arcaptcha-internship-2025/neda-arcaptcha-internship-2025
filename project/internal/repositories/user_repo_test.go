package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(false, sqlxDB)

	user := models.User{
		Username:     "testuser",
		Password:     "hashedpassword",
		Email:        "test@example.com",
		Phone:        "1234567890",
		FullName:     "Test User",
		UserType:     models.Resident,
		TelegramUser: "testtelegram",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(
				user.Username,
				user.Password,
				user.Email,
				user.Phone,
				user.FullName,
				string(user.UserType),
				user.TelegramUser,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		id, err := repo.CreateUser(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	})

	t.Run("duplicate username", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(
				user.Username,
				user.Password,
				user.Email,
				user.Phone,
				user.FullName,
				string(user.UserType),
				user.TelegramUser,
			).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.CreateUser(context.Background(), user)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByTelegramUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(false, sqlxDB)

	now := time.Now()
	telegramUser := "testtelegram"

	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "username", "password", "email", "phone", "full_name",
			"user_type", "telegram_user", "telegram_chat_id", "created_at", "updated_at",
		}).
			AddRow(1, "testuser", "hashedpassword", "test@example.com",
				"1234567890", "Test User", "resident", telegramUser, 12345, now, now)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE telegram_user`).
			WithArgs(telegramUser).
			WillReturnRows(rows)

		user, err := repo.GetUserByTelegramUser(telegramUser)
		assert.NoError(t, err)
		assert.Equal(t, telegramUser, user.TelegramUser)
		assert.Equal(t, int64(12345), user.TelegramChatID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM users WHERE telegram_user`).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetUserByTelegramUser("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateTelegramChatID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(false, sqlxDB)

	telegramUsername := "testuser"
	chatID := int64(12345)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET telegram_chat_id`).
			WithArgs(chatID, telegramUsername).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateTelegramChatID(context.Background(), telegramUsername, chatID)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET telegram_chat_id`).
			WithArgs(chatID, telegramUsername).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpdateTelegramChatID(context.Background(), telegramUsername, chatID)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
