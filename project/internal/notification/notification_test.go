package notification

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
	}, nil
}

func TestNewNotification(t *testing.T) {
	cfg := config.TelegramConfig{
		BotToken: "test_token_123",
		Timeout:  30 * time.Second,
	}

	notification := NewNotification(cfg)

	impl, ok := notification.(*notificationImpl)
	if !ok {
		t.Fatal("NewNotification should return *notificationImpl")
	}

	if impl.botToken != cfg.BotToken {
		t.Errorf("expected botToken %s, got %s", cfg.BotToken, impl.botToken)
	}

	expectedBaseURL := fmt.Sprintf("https://api.telegram.org/bot%s/", cfg.BotToken)
	if impl.baseURL != expectedBaseURL {
		t.Errorf("Expected baseURL %s, got %s", expectedBaseURL, impl.baseURL)
	}
}

func TestSendNotification_Success(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			if req.Method != "POST" {
				t.Errorf("Expected POST method, got %s", req.Method)
			}

			expectedURL := "https://api.telegram.org/bot123456:ABC/sendMessage"
			if req.URL.String() != expectedURL {
				t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
			}

			if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", req.Header.Get("Content-Type"))
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}

			data, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}

			if data.Get("chat_id") != "12345" {
				t.Errorf("Expected chat_id 12345, got %s", data.Get("chat_id"))
			}

			if data.Get("text") != "Hello World" {
				t.Errorf("Expected text 'Hello World', got %s", data.Get("text"))
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			}, nil
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	ctx := context.Background()
	err := notification.SendNotification(ctx, 12345, "Hello World")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendNotification_HTTPClientError(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	ctx := context.Background()
	err := notification.SendNotification(ctx, 12345, "Hello World")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	expectedError := "failed to send notification: network error"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}

func TestSendNotification_Non200Status(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"ok": false, "description": "Bad Request"}`)),
			}, nil
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	ctx := context.Background()
	err := notification.SendNotification(ctx, 12345, "Hello World")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	expectedError := "telegram API returned non-200 status: 400"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}

func TestSendInvitation_WithChatID_Success(t *testing.T) {
	chatID := int64(67890)
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			if req.Method != "POST" {
				t.Errorf("Expected POST method, got %s", req.Method)
			}

			expectedURL := "https://api.telegram.org/bot123456:ABC/sendMessage"
			if req.URL.String() != expectedURL {
				t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}

			data, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}

			if data.Get("chat_id") != "67890" {
				t.Errorf("Expected chat_id 67890, got %s", data.Get("chat_id"))
			}

			expectedMessage := "You've been invited to join apartment 1!\n\nClick this link to accept: http://yourapp.com/join?token=abc123"
			if data.Get("text") != expectedMessage {
				t.Errorf("Expected text %s, got %s", expectedMessage, data.Get("text"))
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			}, nil
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	inv := models.InvitationLink{
		SenderID:         1,
		ReceiverUsername: "testuser",
		ChatID:           &chatID,
		ApartmentID:      1,
		Token:            "abc123",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Status:           models.InvitationStatusPending,
	}

	ctx := context.Background()
	err := notification.SendInvitation(ctx, inv)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendInvitation_WithUsername_Success(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}

			data, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}

			if data.Get("chat_id") != "@testuser" {
				t.Errorf("Expected chat_id @testuser, got %s", data.Get("chat_id"))
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			}, nil
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	inv := models.InvitationLink{
		SenderID:         1,
		ReceiverUsername: "testuser",
		ChatID:           nil, // no chat id, should use username
		ApartmentID:      1,
		Token:            "abc123",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Status:           models.InvitationStatusPending,
	}

	ctx := context.Background()
	err := notification.SendInvitation(ctx, inv)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendInvitation_HTTPClientError(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	inv := models.InvitationLink{
		SenderID:         1,
		ReceiverUsername: "testuser",
		ChatID:           nil,
		ApartmentID:      1,
		Token:            "abc123",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Status:           models.InvitationStatusPending,
	}

	ctx := context.Background()
	err := notification.SendInvitation(ctx, inv)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	expectedError := "failed to send invitation: network error"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}

func TestSendInvitation_Non200Status(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"ok": false, "description": "Unauthorized"}`)),
			}, nil
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	inv := models.InvitationLink{
		SenderID:         1,
		ReceiverUsername: "testuser",
		ChatID:           nil,
		ApartmentID:      1,
		Token:            "abc123",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Status:           models.InvitationStatusPending,
	}

	ctx := context.Background()
	err := notification.SendInvitation(ctx, inv)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	expectedError := "telegram API returned non-200 status for invitation: 401"
	if err.Error() != expectedError {
		t.Errorf("Expected error %s, got %s", expectedError, err.Error())
	}
}

func TestSendNotification_ContextCancellation(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			//checking if context is cancelled
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			default:
				//simulating slow response
				time.Sleep(50 * time.Millisecond)

				//checking again after delay
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				default:
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
					}, nil
				}
			}
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := notification.SendNotification(ctx, 12345, "Hello World")

	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context error, got %v", err)
	}
}

func TestSendInvitation_MessageFormat(t *testing.T) {
	var capturedMessage string

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}

			data, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}

			capturedMessage = data.Get("text")

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			}, nil
		},
	}

	notification := &notificationImpl{
		httpClient: mockClient,
		botToken:   "123456:ABC",
		baseURL:    "https://api.telegram.org/bot123456:ABC/",
	}

	inv := models.InvitationLink{
		SenderID:         1,
		ReceiverUsername: "testuser",
		ChatID:           nil,
		ApartmentID:      42,
		Token:            "xyz789",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Status:           models.InvitationStatusPending,
	}

	ctx := context.Background()
	err := notification.SendInvitation(ctx, inv)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedMessage := "You've been invited to join apartment 42!\n\nClick this link to accept: http://yourapp.com/join?token=xyz789"
	if capturedMessage != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, capturedMessage)
	}
}
