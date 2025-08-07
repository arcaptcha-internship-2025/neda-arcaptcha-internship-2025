package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/speps/go-hashids/v2"
)

type InviteLinkRepo interface {
	CreateInvitation(ctx context.Context, userID, apartmentID, managerID int) (string, error)
	ValidateAndConsumeInvitation(ctx context.Context, code string) (int, error)
}

type invitationLinkRepository struct {
	redisClient *goredis.Client
	expiration  time.Duration
	hashID      *hashids.HashID
}

func NewInvitationLinkRepository(redisClient *goredis.Client, salt string) InviteLinkRepo {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 8
	hashID, _ := hashids.NewWithData(hd)

	return &invitationLinkRepository{
		redisClient: redisClient,
		expiration:  24 * time.Hour,
		hashID:      hashID,
	}
}

func (r *invitationLinkRepository) CreateInvitation(ctx context.Context, userID, apartmentID, managerID int) (string, error) {
	code, err := r.hashID.Encode([]int{userID, apartmentID, managerID})
	if err != nil {
		return "", fmt.Errorf("failed to encode invitation: %w", err)
	}

	key := r.redisKey(userID, apartmentID, managerID)
	err = r.redisClient.Set(ctx, key, "1", r.expiration).Err()
	if err != nil {
		return "", fmt.Errorf("failed to save invitation: %w", err)
	}

	return code, nil
}

func (r *invitationLinkRepository) ValidateAndConsumeInvitation(ctx context.Context, code string) (int, error) {
	ids, err := r.hashID.DecodeWithError(code)
	if err != nil || len(ids) != 3 {
		return 0, errors.New("invalid or tampered code")
	}

	userID, apartmentID, managerID := ids[0], ids[1], ids[2]
	key := r.redisKey(userID, apartmentID, managerID)

	deleted, err := r.redisClient.Del(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to access Redis: %w", err)
	}
	if deleted == 0 {
		return 0, errors.New("invitation not found or already used")
	}

	return apartmentID, nil
}

func (r *invitationLinkRepository) redisKey(userID, apartmentID, managerID int) string {
	return fmt.Sprintf("invitation:%d:%d:%d", userID, apartmentID, managerID)
}
