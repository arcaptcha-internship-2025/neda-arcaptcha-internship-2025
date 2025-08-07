package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvitationLinkRepository_CreateInvitation(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	repo := &invitationLinkRepository{
		redisClient: db,
		expiration:  24 * time.Hour,
	}
	repo = NewInvitationLinkRepository(db, "test-salt").(*invitationLinkRepository)

	userID := 1
	apartmentID := 2
	managerID := 3
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedKey := "invitation:1:2:3"
		mock.ExpectSet(expectedKey, "1", 24*time.Hour).SetVal("OK")

		code, err := repo.CreateInvitation(ctx, userID, apartmentID, managerID)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.GreaterOrEqual(t, len(code), 8)
	})

	t.Run("redis error", func(t *testing.T) {
		expectedKey := "invitation:1:2:3"
		mock.ExpectSet(expectedKey, "1", 24*time.Hour).SetErr(redis.Nil)

		code, err := repo.CreateInvitation(ctx, userID, apartmentID, managerID)
		assert.Error(t, err)
		assert.Empty(t, code)
		assert.Contains(t, err.Error(), "failed to save invitation")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInvitationLinkRepository_ValidateAndConsumeInvitation(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	repo := NewInvitationLinkRepository(db, "test-salt").(*invitationLinkRepository)

	userID := 1
	apartmentID := 2
	managerID := 3
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedKey := "invitation:1:2:3"
		mock.ExpectSet(expectedKey, "1", 24*time.Hour).SetVal("OK")

		code, err := repo.CreateInvitation(ctx, userID, apartmentID, managerID)
		require.NoError(t, err)

		//testing validation and consumption
		mock.ExpectDel(expectedKey).SetVal(1)

		resultApartmentID, err := repo.ValidateAndConsumeInvitation(ctx, code)
		assert.NoError(t, err)
		assert.Equal(t, apartmentID, resultApartmentID)
	})

	t.Run("invalid code", func(t *testing.T) {
		invalidCode := "invalid-code"

		resultApartmentID, err := repo.ValidateAndConsumeInvitation(ctx, invalidCode)
		assert.Error(t, err)
		assert.Equal(t, 0, resultApartmentID)
		assert.Contains(t, err.Error(), "invalid or tampered code")
	})

	t.Run("invitation not found", func(t *testing.T) {
		//a valid code first
		expectedKey := "invitation:1:2:3"
		mock.ExpectSet(expectedKey, "1", 24*time.Hour).SetVal("OK")

		code, err := repo.CreateInvitation(ctx, userID, apartmentID, managerID)
		require.NoError(t, err)

		//mocking deletion returning 0 (key not found)
		mock.ExpectDel(expectedKey).SetVal(0)

		resultApartmentID, err := repo.ValidateAndConsumeInvitation(ctx, code)
		assert.Error(t, err)
		assert.Equal(t, 0, resultApartmentID)
		assert.Contains(t, err.Error(), "invitation not found or already used")
	})

	t.Run("redis error during deletion", func(t *testing.T) {
		expectedKey := "invitation:1:2:3"
		mock.ExpectSet(expectedKey, "1", 24*time.Hour).SetVal("OK")

		code, err := repo.CreateInvitation(ctx, userID, apartmentID, managerID)
		require.NoError(t, err)

		//mocking deletion error
		mock.ExpectDel(expectedKey).SetErr(redis.Nil)

		resultApartmentID, err := repo.ValidateAndConsumeInvitation(ctx, code)
		assert.Error(t, err)
		assert.Equal(t, 0, resultApartmentID)
		assert.Contains(t, err.Error(), "failed to access Redis")
	})

	t.Run("empty code", func(t *testing.T) {
		resultApartmentID, err := repo.ValidateAndConsumeInvitation(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, 0, resultApartmentID)
		assert.Contains(t, err.Error(), "invalid or tampered code")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInvitationLinkRepository_redisKey(t *testing.T) {
	db, _ := redismock.NewClientMock()
	defer db.Close()

	repo := &invitationLinkRepository{
		redisClient: db,
		expiration:  24 * time.Hour,
	}

	userID := 123
	apartmentID := 456
	managerID := 789

	expectedKey := "invitation:123:456:789"
	actualKey := repo.redisKey(userID, apartmentID, managerID)

	assert.Equal(t, expectedKey, actualKey)
}

func TestNewInvitationLinkRepository(t *testing.T) {
	db, _ := redismock.NewClientMock()
	defer db.Close()

	salt := "test-salt"
	repo := NewInvitationLinkRepository(db, salt)

	assert.NotNil(t, repo)

	var _ InviteLinkRepo = repo
}

func TestInvitationLinkRepository_Integration(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	repo := NewInvitationLinkRepository(db, "integration-test-salt")
	ctx := context.Background()

	userID := 1
	apartmentID := 2
	managerID := 3

	t.Run("create and validate invitation flow", func(t *testing.T) {
		expectedKey := "invitation:1:2:3"

		mock.ExpectSet(expectedKey, "1", 24*time.Hour).SetVal("OK")
		code, err := repo.CreateInvitation(ctx, userID, apartmentID, managerID)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)

		//validating and consuming invitation
		mock.ExpectDel(expectedKey).SetVal(1)
		resultApartmentID, err := repo.ValidateAndConsumeInvitation(ctx, code)
		assert.NoError(t, err)
		assert.Equal(t, apartmentID, resultApartmentID)
	})

	t.Run("double consumption should fail", func(t *testing.T) {
		expectedKey := "invitation:1:2:3"

		mock.ExpectSet(expectedKey, "1", 24*time.Hour).SetVal("OK")
		code, err := repo.CreateInvitation(ctx, userID, apartmentID, managerID)
		assert.NoError(t, err)

		//first consumption - success
		mock.ExpectDel(expectedKey).SetVal(1)
		_, err = repo.ValidateAndConsumeInvitation(ctx, code)
		assert.NoError(t, err)

		//second consumption - should fail (key already deleted)
		mock.ExpectDel(expectedKey).SetVal(0)
		resultApartmentID, err := repo.ValidateAndConsumeInvitation(ctx, code)
		assert.Error(t, err)
		assert.Equal(t, 0, resultApartmentID)
		assert.Contains(t, err.Error(), "invitation not found or already used")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
