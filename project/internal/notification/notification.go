// /project/internal/notification/notification.go
package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

type Notification interface {
	SendNotification(ctx context.Context, chatID int64, message string) error
	SendInvitation(ctx context.Context, inv models.InvitationLink) error
}
type notificationImpl struct {
	httpClient *http.Client
	botToken   string
	baseURL    string
}

func NewNotification(cfg config.TelegramConfig) Notification {
	return &notificationImpl{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		botToken:   cfg.BotToken,
		baseURL:    fmt.Sprintf("https://api.telegram.org/bot%s/", cfg.BotToken),
	}
}

func (n *notificationImpl) SendNotification(ctx context.Context, chatID int64, message string) error {
	//assuming that chatid = user name(which is right when the chat is already started)
	endpoint := n.baseURL + "sendMessage"
	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%d", chatID))
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
	message := fmt.Sprintf(
		"You've been invited to join apartment %d!\n\n"+
			"Click this link to accept: http://yourapp.com/join?token=%s",
		inv.ApartmentID, inv.Token,
	)

	endpoint := n.baseURL + "sendMessage"
	data := url.Values{}

	//trying to use chat id if available, otherwise fall back to username
	if inv.ChatID != nil {
		data.Set("chat_id", strconv.FormatInt(*inv.ChatID, 10))
	} else {
		data.Set("chat_id", "@"+inv.ReceiverUsername)
	}

	data.Set("text", message)

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
