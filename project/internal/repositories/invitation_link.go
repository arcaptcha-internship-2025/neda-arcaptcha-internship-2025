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
	MarkInvitationUsed(ctx context.Context, token string) error
	MarkInvitationRejected(ctx context.Context, token string) error
	GetInvitationsByUser(ctx context.Context, username string) ([]*models.InvitationLink, error)
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
	//serializing invitation data as json
	invData, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal invitation: %w", err)
	}

	//storing with pipeline for atomic operations
	pipe := r.redisClient.TxPipeline()

	//storing main invitation data
	pipe.Set(ctx, "invitation:"+inv.Token, invData, r.expiration)

	//creating index by username for tracking user's invitations
	pipe.SAdd(ctx, "user_invitations:"+inv.ReceiverUsername, inv.Token)
	pipe.Expire(ctx, "user_invitations:"+inv.ReceiverUsername, r.expiration)

	//creating index by apartment for tracking apartment invitations
	pipe.SAdd(ctx, "apartment_invitations:"+strconv.Itoa(inv.ApartmentID), inv.Token)
	pipe.Expire(ctx, "apartment_invitations:"+strconv.Itoa(inv.ApartmentID), r.expiration)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *invitationLinkRepository) GetInvitationByToken(ctx context.Context, token string) (*models.InvitationLink, error) {
	//getting invitation data
	result, err := r.redisClient.Get(ctx, "invitation:"+token).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, errors.New("invitation not found or expired")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	//unmarshal json data
	var inv models.InvitationLink
	if err := json.Unmarshal([]byte(result), &inv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation: %w", err)
	}

	if time.Now().After(inv.ExpiresAt) {
		r.MarkInvitationExpired(ctx, token)
		return nil, errors.New("invitation has expired")
	}

	return &inv, nil
}

func (r *invitationLinkRepository) MarkInvitationUsed(ctx context.Context, token string) error {
	//current invitation
	inv, err := r.GetInvitationByToken(ctx, token)
	if err != nil {
		return err
	}

	inv.Status = models.InvitationStatusAccepted

	//marshal and store updated data
	invData, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal updated invitation: %w", err)
	}

	//storing updated invitation with remaining TTL
	ttl := r.redisClient.TTL(ctx, "invitation:"+token).Val()
	err = r.redisClient.Set(ctx, "invitation:"+token, invData, ttl).Err()
	return err
}

func (r *invitationLinkRepository) MarkInvitationRejected(ctx context.Context, token string) error {
	//current invitation
	inv, err := r.GetInvitationByToken(ctx, token)
	if err != nil {
		return err
	}

	inv.Status = models.InvitationStatusRejected

	//marshaling and storing updated data
	invData, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal updated invitation: %w", err)
	}

	//storing updated invitation with remaining ttl
	ttl := r.redisClient.TTL(ctx, "invitation:"+token).Val()
	err = r.redisClient.Set(ctx, "invitation:"+token, invData, ttl).Err()
	return err
}

func (r *invitationLinkRepository) MarkInvitationExpired(ctx context.Context, token string) error {
	//current invitation
	inv, err := r.GetInvitationByToken(ctx, token)
	if err != nil {
		return err
	}

	inv.Status = models.InvitationStatusExpired

	invData, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal updated invitation: %w", err)
	}

	ttl := r.redisClient.TTL(ctx, "invitation:"+token).Val()
	err = r.redisClient.Set(ctx, "invitation:"+token, invData, ttl).Err()
	return err
}

func (r *invitationLinkRepository) GetInvitationsByUser(ctx context.Context, username string) ([]*models.InvitationLink, error) {
	//getting all invitation tokens for the user
	tokens, err := r.redisClient.SMembers(ctx, "user_invitations:"+username).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user invitations: %w", err)
	}

	var invitations []*models.InvitationLink
	for _, token := range tokens {
		inv, err := r.GetInvitationByToken(ctx, token)
		if err != nil {
			//skipping invalid/expired invitations butt log the error
			continue
		}
		invitations = append(invitations, inv)
	}

	return invitations, nil
}
