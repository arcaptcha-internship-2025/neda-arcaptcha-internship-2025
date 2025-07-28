package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	goredis "github.com/redis/go-redis/v9"
)

type InviteLinkFlagRepo interface {
	CreateInvitation(ctx context.Context, inv models.InvitationLink) error
	GetInvitationByToken(ctx context.Context, token string) (*models.InvitationLink, error)
}

type invitationLinkRepository struct {
	redisClient *goredis.Client
	expiration  time.Duration
}

func NewInvitationLinkRepository(redisClient *goredis.Client) InviteLinkFlagRepo {
	return &invitationLinkRepository{
		redisClient: redisClient,
		expiration:  24 * time.Hour,
	}
}

func (r *invitationLinkRepository) CreateInvitation(ctx context.Context, inv models.InvitationLink) error {
	if inv.Status == "" {
		inv.Status = models.InvitationStatusPending
	}

	invData, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal invitation: %w", err)
	}

	pipe := r.redisClient.TxPipeline()

	//storing main invitation data
	pipe.Set(ctx, "invitation:"+inv.Token, invData, r.expiration)

	//creating indexes
	pipe.SAdd(ctx, "user_invitations:"+inv.ReceiverUsername, inv.Token)
	pipe.Expire(ctx, "user_invitations:"+inv.ReceiverUsername, r.expiration)

	pipe.SAdd(ctx, "apartment_invitations:"+strconv.Itoa(inv.ApartmentID), inv.Token)
	pipe.Expire(ctx, "apartment_invitations:"+strconv.Itoa(inv.ApartmentID), r.expiration)

	//add to pending invitations set if status is pending
	if inv.Status == models.InvitationStatusPending {
		pipe.SAdd(ctx, "pending_invitations", inv.Token)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *invitationLinkRepository) GetInvitationByToken(ctx context.Context, token string) (*models.InvitationLink, error) {
	result, err := r.redisClient.Get(ctx, "invitation:"+token).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, errors.New("invitation not found or expired")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	var inv models.InvitationLink
	if err := json.Unmarshal([]byte(result), &inv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation: %w", err)
	}

	if time.Now().After(inv.ExpiresAt) && inv.Status == models.InvitationStatusPending {
		if err := r.MarkInvitationExpired(ctx, token); err != nil {
			return nil, err
		}
		inv.Status = models.InvitationStatusExpired
	}

	return &inv, nil
}

func (r *invitationLinkRepository) MarkInvitationExpired(ctx context.Context, token string) error {
	return r.updateInvitationStatus(ctx, token, models.InvitationStatusExpired)
}

func (r *invitationLinkRepository) updateInvitationStatus(ctx context.Context, token string, status models.InvitationStatus) error {
	inv, err := r.GetInvitationByToken(ctx, token)
	if err != nil {
		return err
	}

	inv.Status = status
	return r.saveInvitation(ctx, token, inv)
}

func (r *invitationLinkRepository) saveInvitation(ctx context.Context, token string, inv *models.InvitationLink) error {
	invData, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal invitation: %w", err)
	}

	ttl := r.redisClient.TTL(ctx, "invitation:"+token).Val()
	if ttl < 0 {
		ttl = r.expiration
	}

	pipe := r.redisClient.TxPipeline()
	pipe.Set(ctx, "invitation:"+token, invData, ttl)

	if inv.Status == models.InvitationStatusPending {
		pipe.SAdd(ctx, "pending_invitations", token)
	} else {
		pipe.SRem(ctx, "pending_invitations", token)
	}

	_, err = pipe.Exec(ctx)
	return err
}
