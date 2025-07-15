package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type InviteLinkFlagRepo interface {
	Set(ctx context.Context, username, apartmentID, token string) error
	VerifyToken(ctx context.Context, token string) (int, error)
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

func (r *invitationLinkRepository) Set(ctx context.Context, username, apartmentID, token string) error {
	//store token -> apartmentID mapping
	err := r.redisClient.Set(ctx, "invite:token:"+token, apartmentID, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	//store username -> token mapping (for tracking)
	err = r.redisClient.Set(ctx, "invite:user:"+username+":"+apartmentID, token, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *invitationLinkRepository) VerifyToken(ctx context.Context, token string) (int, error) {
	apartmentIDStr, err := r.redisClient.Get(ctx, "invite:token:"+token).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, errors.New("invalid or expired token")
		}
		return 0, err
	}

	//deleting the token after verification
	_, _ = r.redisClient.Del(ctx, "invite:token:"+token).Result()

	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		return 0, errors.New("invalid apartment ID")
	}

	return apartmentID, nil
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
