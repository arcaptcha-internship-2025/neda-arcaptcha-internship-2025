package notification

import (
	"context"
	"net/http"
)

type Notification interface {
	SendNotification(ctx context.Context, message string) error
}

type notificationImpl struct {
	baleClient  *http.Client
	baleToken   string
	baleBaseUrl string
}

func NewNotification(baleClient *http.Client, baleToken string, baleBaseUrl string) Notification {
	return &notificationImpl{
		baleClient:  baleClient,
		baleToken:   baleToken,
		baleBaseUrl: baleBaseUrl,
	}
}

func (n *notificationImpl) SendNotification(ctx context.Context, message string) error {
	n.baleClient.Post(n.baleBaseUrl+"/sendmessage", "", nil)
	return nil
}
