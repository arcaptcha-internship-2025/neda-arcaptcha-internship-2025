package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	goredis "github.com/redis/go-redis/v9"
)

type InviteLinkFlagRepo interface {
	CreateInvitation(ctx context.Context, inv models.InvitationLink) error
	GetInvitationByToken(ctx context.Context, token string) (*models.InvitationLink, error)
	MarkInvitationUsed(ctx context.Context, token string, chatID int64) error
}

type invitationLinkRepository struct {
	redisClient *goredis.Client
	expiration  time.Duration
}

func NewInvitationLinkRepository(redisClient *goredis.Client) InviteLinkFlagRepo {
	return &invitationLinkRepository{
		redisClient: redisClient,
		expiration:  24 * time.Hour, // Set default expiration
	}
}

func (r *invitationLinkRepository) CreateInvitation(ctx context.Context, inv models.InvitationLink) error {
	//serializ invitation data
	data := map[string]interface{}{
		"sender_id":         inv.SenderID,
		"receiver_username": inv.ReceiverUsername,
		"apartment_id":      inv.ApartmentID,
		"status":            inv.Status,
		"expires_at":        inv.ExpiresAt.Format(time.RFC3339),
	}

	if inv.ChatID != nil {
		data["chat_id"] = *inv.ChatID
	}

	//store using pipeline for atomic operations
	pipe := r.redisClient.TxPipeline()

	//stores main invitation data
	pipe.HSet(ctx, "invitation:"+inv.Token, data)

	pipe.Expire(ctx, "invitation:"+inv.Token, r.expiration)

	//creating index by username:apartment for tracking
	pipe.Set(ctx, "invitation_index:"+inv.ReceiverUsername+":"+strconv.Itoa(inv.ApartmentID),
		inv.Token, r.expiration)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *invitationLinkRepository) GetInvitationByToken(ctx context.Context, token string) (*models.InvitationLink, error) {
	//getting all fields from hash
	result, err := r.redisClient.HGetAll(ctx, "invitation:"+token).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("invitation not found")
	}

	//parsing fields
	inv := &models.InvitationLink{
		Token:            token,
		ReceiverUsername: result["receiver_username"],
		Status:           models.InvitationStatus(result["status"]),
	}

	//parsing numeric fields
	if senderID, err := strconv.Atoi(result["sender_id"]); err == nil {
		inv.SenderID = senderID
	}
	if apartmentID, err := strconv.Atoi(result["apartment_id"]); err == nil {
		inv.ApartmentID = apartmentID
	}
	if chatIDStr, ok := result["chat_id"]; ok {
		if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
			inv.ChatID = &chatID
		}
	}
	if expiresAt, err := time.Parse(time.RFC3339, result["expires_at"]); err == nil {
		inv.ExpiresAt = expiresAt
	}

	return inv, nil
}

func (r *invitationLinkRepository) MarkInvitationUsed(ctx context.Context, token string, chatID int64) error {
	//updating status and chat id in a transaction
	_, err := r.redisClient.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.HSet(ctx, "invitation:"+token,
			"status", string(models.InvitationStatusAccepted),
			"chat_id", chatID)
		return nil
	})
	return err
}
