package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AudienceTopicService handles communication with the audience-topic endpoints
// of the Lettr API.
type AudienceTopicService struct {
	client *Client
}

// Default-subscription values for audience topics.
const (
	TopicDefaultSubscriptionOptIn  = "opt_in"
	TopicDefaultSubscriptionOptOut = "opt_out"
)

// Visibility values for audience topics.
const (
	TopicVisibilityPrivate = "private"
	TopicVisibilityPublic  = "public"
)

// AudienceTopic represents a topic that contacts can subscribe to.
type AudienceTopic struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Description         *string `json:"description"`
	DefaultSubscription string  `json:"default_subscription"`
	Visibility          string  `json:"visibility"`
	ContactsCount       int     `json:"contacts_count"`
	CreatedAt           *string `json:"created_at"`
}

// ListAudienceTopicsParams contains the query parameters for listing topics.
type ListAudienceTopicsParams struct {
	PerPage int
	Page    int
}

// ListAudienceTopicsResponse is the response from listing topics.
type ListAudienceTopicsResponse struct {
	Message string                 `json:"message"`
	Data    ListAudienceTopicsData `json:"data"`
}

// ListAudienceTopicsData contains the paginated list of topics.
type ListAudienceTopicsData struct {
	Topics     []AudienceTopic `json:"topics"`
	Pagination PagePagination  `json:"pagination"`
}

// AudienceTopicResponse is the response from endpoints that return a single topic.
type AudienceTopicResponse struct {
	Message string        `json:"message"`
	Data    AudienceTopic `json:"data"`
}

// CreateAudienceTopicRequest is the body for creating a topic.
type CreateAudienceTopicRequest struct {
	Name string `json:"name"`
	// Description is optional. Set to a pointer to an empty string for empty,
	// or omit to leave unset.
	Description *string `json:"description,omitempty"`
	// DefaultSubscription defaults to "opt_in" server-side when omitted.
	DefaultSubscription string `json:"default_subscription,omitempty"`
	// Visibility defaults to "private" server-side when omitted.
	Visibility string `json:"visibility,omitempty"`
}

// UpdateAudienceTopicRequest is the body for partially updating a topic.
// The default_subscription field is immutable and cannot be sent.
//
// Description is a nullable PATCH field: leave it nil to omit,
// NewNullString("x") to replace, or &NullString{} (zero value) to clear
// the existing description server-side.
type UpdateAudienceTopicRequest struct {
	Name        string      `json:"name,omitempty"`
	Description *NullString `json:"description,omitempty"`
	Visibility  string      `json:"visibility,omitempty"`
}

// List retrieves a paginated list of topics.
//
// Pass nil for params to use defaults.
func (s *AudienceTopicService) List(ctx context.Context, params *ListAudienceTopicsParams) (*ListAudienceTopicsResponse, error) {
	path := "audience/topics"
	if params != nil {
		q := url.Values{}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListAudienceTopicsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create creates a new topic.
func (s *AudienceTopicService) Create(ctx context.Context, params *CreateAudienceTopicRequest) (*AudienceTopicResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/topics", params)
	if err != nil {
		return nil, err
	}

	var resp AudienceTopicResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single topic by ID.
func (s *AudienceTopicService) Get(ctx context.Context, topicID string) (*AudienceTopicResponse, error) {
	path := fmt.Sprintf("audience/topics/%s", url.PathEscape(topicID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp AudienceTopicResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update partially updates a topic.
func (s *AudienceTopicService) Update(ctx context.Context, topicID string, params *UpdateAudienceTopicRequest) (*AudienceTopicResponse, error) {
	path := fmt.Sprintf("audience/topics/%s", url.PathEscape(topicID))

	req, err := s.client.newRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return nil, err
	}

	var resp AudienceTopicResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete permanently deletes a topic.
func (s *AudienceTopicService) Delete(ctx context.Context, topicID string) error {
	path := fmt.Sprintf("audience/topics/%s", url.PathEscape(topicID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
