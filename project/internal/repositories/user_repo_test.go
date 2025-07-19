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
		Username: "testuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
		Phone:    "+1234567890",
		FullName: "Test User",
		UserType: models.Resident,
	}

	//successful creation
	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.Username, user.Password, user.Email, user.Phone, user.FullName, user.UserType).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		id, err := repo.CreateUser(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	})

	//duplicate username
	t.Run("duplicate username", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.Username, user.Password, user.Email, user.Phone, user.FullName, user.UserType).
			WillReturnError(sql.ErrNoRows) //simulating unique constraint violation

		_, err := repo.CreateUser(context.Background(), user)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(false, sqlxDB)

	//user found
	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "phone", "full_name", "user_type", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "+1234567890", "Test User", "resident", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT \* FROM users WHERE id = \$1`).
			WithArgs(1).
			WillReturnRows(rows)

		user, err := repo.GetUserByID(1)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
	})

	//user not found
	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM users WHERE id = \$1`).
			WithArgs(2).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetUserByID(2)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
