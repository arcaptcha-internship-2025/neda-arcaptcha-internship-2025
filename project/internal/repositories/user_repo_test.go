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
		Phone:    "1234567890",
		FullName: "Test User",
		UserType: models.Resident,
	}

	// successful creation
	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.Username, user.Password, user.Email, user.Phone, user.FullName, string(user.UserType)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		id, err := repo.CreateUser(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	})

	// duplicate username
	t.Run("duplicate username", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.Username, user.Password, user.Email, user.Phone, user.FullName, string(user.UserType)).
			WillReturnError(sql.ErrNoRows) // simulating unique constraint violation

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

	now := time.Now()

	// user found
	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "password", "email", "phone", "full_name", "user_type", "created_at", "updated_at"}).
			AddRow(1, "testuser", "hashedpassword", "test@example.com", "1234567890", "Test User", "resident", now, now)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id`).
			WithArgs(1).
			WillReturnRows(rows)

		user, err := repo.GetUserByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "hashedpassword", user.Password)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "1234567890", user.Phone)
		assert.Equal(t, "Test User", user.FullName)
		assert.Equal(t, models.Resident, user.UserType)
		assert.WithinDuration(t, now, user.CreatedAt, time.Second)
		assert.WithinDuration(t, now, user.UpdatedAt, time.Second)
	})

	// user not found
	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id`).
			WithArgs(2).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByID(2)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(false, sqlxDB)

	now := time.Now()

	// user found
	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "password", "email", "phone", "full_name", "user_type", "created_at", "updated_at"}).
			AddRow(1, "testuser", "hashedpassword", "test@example.com", "1234567890", "Test User", "resident", now, now)

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE username`).
			WithArgs("testuser").
			WillReturnRows(rows)

		user, err := repo.GetUserByUsername("testuser")
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, models.Resident, user.UserType)
	})

	// user not found
	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM users WHERE username`).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetUserByUsername("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(false, sqlxDB)

	user := models.User{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		Username: "updateduser",
		Password: "newhashed",
		Email:    "updated@example.com",
		Phone:    "9876543210",
		FullName: "Updated User",
		UserType: models.Manager,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET`).
			WithArgs(
				"updateduser",
				"newhashed",
				"updated@example.com",
				"9876543210",
				"Updated User",
				"manager",
				1,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateUser(context.Background(), user)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(`UPDATE users SET`).
			WithArgs(
				"updateduser",
				"newhashed",
				"updated@example.com",
				"9876543210",
				"Updated User",
				"manager",
				1,
			).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpdateUser(context.Background(), user)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(false, sqlxDB)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM users WHERE id`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteUser(1)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM users WHERE id`).
			WithArgs(2).
			WillReturnError(sql.ErrConnDone)

		err := repo.DeleteUser(2)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
