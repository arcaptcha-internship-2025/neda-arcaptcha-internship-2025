package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Notification interface {
	SendNotification(ctx context.Context, username string, message string) error
	SendInvitation(ctx context.Context, inv models.InvitationLink) error
}

type notificationImpl struct {
	httpClient HTTPClient
	botToken   string
	baseURL    string
	appBaseURL string // base url for app
}

func NewNotification(cfg config.TelegramConfig, appBaseURL string) Notification {
	return &notificationImpl{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		botToken:   cfg.BotToken,
		baseURL:    fmt.Sprintf("https://api.telegram.org/bot%s/", cfg.BotToken),
		appBaseURL: appBaseURL,
	}
}

func (n *notificationImpl) SendNotification(ctx context.Context, username string, message string) error {
	endpoint := n.baseURL + "sendMessage"
	data := url.Values{}
	data.Set("chat_id", "@"+username) //@username format
	data.Set("text", message)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func (n *notificationImpl) SendInvitation(ctx context.Context, inv models.InvitationLink) error {
	//creating the invitation url with token as query param
	inviteURL := fmt.Sprintf("%s/join?token=%s", n.appBaseURL, inv.Token)

	message := fmt.Sprintf(
		"You've been invited to join apartment %d!\n\n"+
			"Click this link to accept the invitation:\n%s\n\n"+
			"This invitation expires at: %s",
		inv.ApartmentID,
		inviteURL,
		inv.ExpiresAt.Format("2006-01-02 15:04:05"),
	)

	endpoint := n.baseURL + "sendMessage"
	data := url.Values{}
	data.Set("chat_id", "@"+inv.ReceiverUsername) //@username format
	data.Set("text", message)
	data.Set("parse_mode", "HTML") //html formatting

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create invitation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send invitation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-200 status for invitation: %d", resp.StatusCode)
	}

	return nil
}
