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
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	URL                string    `json:"url"`
	Enabled            bool      `json:"enabled"`
	EventTypes         *[]string `json:"event_types"`
	AuthType           string    `json:"auth_type"`
	HasAuthCredentials bool      `json:"has_auth_credentials"`
	LastSuccessfulAt   *string   `json:"last_successful_at"`
	LastFailureAt      *string   `json:"last_failure_at"`
	LastStatus         *string   `json:"last_status"`
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

// CreateWebhookRequest represents the request body for creating a webhook.
type CreateWebhookRequest struct {
	Name              string   `json:"name"`
	URL               string   `json:"url"`
	AuthType          string   `json:"auth_type"`
	AuthUsername      string   `json:"auth_username,omitempty"`
	AuthPassword      string   `json:"auth_password,omitempty"`
	OAuthClientID     string   `json:"oauth_client_id,omitempty"`
	OAuthClientSecret string   `json:"oauth_client_secret,omitempty"`
	OAuthTokenURL     string   `json:"oauth_token_url,omitempty"`
	EventsMode        string   `json:"events_mode"`
	Events            []string `json:"events,omitempty"`
}

// UpdateWebhookRequest represents the request body for updating a webhook.
type UpdateWebhookRequest struct {
	Name              string   `json:"name,omitempty"`
	Target            string   `json:"target,omitempty"`
	AuthType          string   `json:"auth_type,omitempty"`
	AuthUsername      string   `json:"auth_username,omitempty"`
	AuthPassword      string   `json:"auth_password,omitempty"`
	OAuthClientID     string   `json:"oauth_client_id,omitempty"`
	OAuthClientSecret string   `json:"oauth_client_secret,omitempty"`
	OAuthTokenURL     string   `json:"oauth_token_url,omitempty"`
	Events            []string `json:"events,omitempty"`
	Active            *bool    `json:"active,omitempty"`
}

// CreateWebhookResponse is the response from creating a webhook.
type CreateWebhookResponse struct {
	Message string  `json:"message"`
	Data    Webhook `json:"data"`
}

// UpdateWebhookResponse is the response from updating a webhook.
type UpdateWebhookResponse struct {
	Message string  `json:"message"`
	Data    Webhook `json:"data"`
}

// Create creates a new webhook for event notifications.
//
// Example:
//
//	webhook, err := client.Webhooks.Create(ctx, &lettr.CreateWebhookRequest{
//	    Name:       "My Webhook",
//	    URL:        "https://example.com/webhook",
//	    AuthType:   "none",
//	    EventsMode: "all",
//	})
func (s *WebhookService) Create(ctx context.Context, params *CreateWebhookRequest) (*CreateWebhookResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "webhooks", params)
	if err != nil {
		return nil, err
	}

	var resp CreateWebhookResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update modifies an existing webhook's settings.
//
// Example:
//
//	active := false
//	updated, err := client.Webhooks.Update(ctx, "webhook-abc123", &lettr.UpdateWebhookRequest{
//	    Active: &active,
//	})
func (s *WebhookService) Update(ctx context.Context, webhookID string, params *UpdateWebhookRequest) (*UpdateWebhookResponse, error) {
	path := fmt.Sprintf("webhooks/%s", url.PathEscape(webhookID))

	req, err := s.client.newRequest(ctx, http.MethodPut, path, params)
	if err != nil {
		return nil, err
	}

	var resp UpdateWebhookResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteWebhookResponse is the response from deleting a webhook.
type DeleteWebhookResponse struct {
	Message string `json:"message"`
}

// Delete removes a webhook.
//
// Example:
//
//	resp, err := client.Webhooks.Delete(ctx, "webhook-abc123")
func (s *WebhookService) Delete(ctx context.Context, webhookID string) (*DeleteWebhookResponse, error) {
	path := fmt.Sprintf("webhooks/%s", url.PathEscape(webhookID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	var resp DeleteWebhookResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
