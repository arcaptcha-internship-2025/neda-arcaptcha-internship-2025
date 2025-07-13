package notification

import (
	"context"
	"net/http"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
)

type Notification interface {
	SendNotification(ctx context.Context, message string) error
}

type notificationImpl struct {
	baleClient  *http.Client
	baleToken   string
	baleBaseUrl string
}

func NewNotification(cfg config.BaleConfig) Notification {
	return &notificationImpl{
		baleClient:  &http.Client{Timeout: cfg.Timeout},
		baleToken:   cfg.ApiToken,
		baleBaseUrl: cfg.BaseUrl,
	}
}

func (n *notificationImpl) SendNotification(ctx context.Context, message string) error {
	n.baleClient.Post(n.baleBaseUrl+"/sendmessage", "", nil)
	return nil
}
