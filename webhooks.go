package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// WebhookService handles communication with the webhook-related endpoints
// of the Lettr API.
type WebhookService struct {
	client *Client
}

// Webhook represents a webhook configuration.
type Webhook struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	URL                string   `json:"url"`
	Enabled            bool     `json:"enabled"`
	EventTypes         []string `json:"event_types"`
	AuthType           string   `json:"auth_type"`
	HasAuthCredentials bool     `json:"has_auth_credentials"`
	LastSuccessfulAt   *string  `json:"last_successful_at"`
	LastFailureAt      *string  `json:"last_failure_at"`
	LastStatus         *string  `json:"last_status"`
}

// ListWebhooksResponse is the response from listing webhooks.
type ListWebhooksResponse struct {
	Message string           `json:"message"`
	Data    ListWebhooksData `json:"data"`
}

// ListWebhooksData contains the list of webhooks.
type ListWebhooksData struct {
	Webhooks []Webhook `json:"webhooks"`
}

// GetWebhookResponse is the response from getting a single webhook.
type GetWebhookResponse struct {
	Message string  `json:"message"`
	Data    Webhook `json:"data"`
}

// List retrieves all webhooks configured for your account.
//
// Example:
//
//	webhooks, err := client.Webhooks.List(ctx)
func (s *WebhookService) List(ctx context.Context) (*ListWebhooksResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, "webhooks", nil)
	if err != nil {
		return nil, err
	}

	var resp ListWebhooksResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves details of a single webhook.
//
// Example:
//
//	webhook, err := client.Webhooks.Get(ctx, "webhook-abc123")
func (s *WebhookService) Get(ctx context.Context, webhookID string) (*GetWebhookResponse, error) {
	path := fmt.Sprintf("webhooks/%s", url.PathEscape(webhookID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetWebhookResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
