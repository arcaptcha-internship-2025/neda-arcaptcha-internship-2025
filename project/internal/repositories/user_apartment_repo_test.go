package repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserApartmentRepository_IsUserManagerOfApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userID := 1
	apartmentID := 2

	t.Run("is manager", func(t *testing.T) {
		mock.ExpectQuery(`SELECT is_manager FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"is_manager"}).AddRow(true))

		isManager, err := repo.IsUserManagerOfApartment(context.Background(), userID, apartmentID)
		assert.NoError(t, err)
		assert.True(t, isManager)
	})

	t.Run("not manager", func(t *testing.T) {
		mock.ExpectQuery(`SELECT is_manager FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"is_manager"}).AddRow(false))

		isManager, err := repo.IsUserManagerOfApartment(context.Background(), userID, apartmentID)
		assert.NoError(t, err)
		assert.False(t, isManager)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT is_manager FROM user_apartments`).
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.IsUserManagerOfApartment(context.Background(), userID, apartmentID)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserApartmentRepository_IsUserInApartment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserApartmentRepository(false, sqlxDB)

	userID := 1
	apartmentID := 2

	t.Run("user in apartment", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.IsUserInApartment(context.Background(), userID, apartmentID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("user not in apartment", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID, apartmentID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		exists, err := repo.IsUserInApartment(context.Background(), userID, apartmentID)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(userID, apartmentID).
			WillReturnError(sql.ErrConnDone)

		_, err := repo.IsUserInApartment(context.Background(), userID, apartmentID)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
