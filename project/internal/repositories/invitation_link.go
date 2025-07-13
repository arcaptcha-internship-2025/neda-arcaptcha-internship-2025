package repositories

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type InviteLinkFlagRepo interface {
	Set(ctx context.Context, userID, apartmentID string) error
	GetAndDelete(ctx context.Context, userID, apartmentID string) error
}

type invitationLinkRepository struct {
	redisClient *goredis.Client
}

func NewInvitationLinkRepository(redisClient *goredis.Client) InviteLinkFlagRepo {
	return &invitationLinkRepository{
		redisClient: redisClient,
	}
}

func (r *invitationLinkRepository) Set(ctx context.Context, userID, apartmentID string) error {
	err := r.redisClient.Set(ctx, generateKey(userID, apartmentID), "1", 24*time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *invitationLinkRepository) GetAndDelete(ctx context.Context, userID string, apartmentID string) error {
	_, err := r.redisClient.GetDel(ctx, generateKey(userID, apartmentID)).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil
		}
		return err
	}
	return nil
}

func generateKey(userID string, apartemantID string) string {
	return "invitelink:" + userID + ":" + apartemantID
}
