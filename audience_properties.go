package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AudiencePropertyService handles communication with the audience-property
// endpoints of the Lettr API.
type AudiencePropertyService struct {
	client *Client
}

// Data-type values for audience properties.
const (
	PropertyTypeString  = "string"
	PropertyTypeNumber  = "number"
	PropertyTypeBoolean = "boolean"
	PropertyTypeDate    = "date"
	PropertyTypeJSON    = "json"
)

// AudienceProperty represents a custom property definition that contacts can hold.
type AudienceProperty struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	FallbackValue *string `json:"fallback_value"`
	CreatedAt     string  `json:"created_at"`
}

// ListAudiencePropertiesParams contains the query parameters for listing properties.
type ListAudiencePropertiesParams struct {
	PerPage int
	Page    int
}

// ListAudiencePropertiesResponse is the response from listing properties.
type ListAudiencePropertiesResponse struct {
	Message string                     `json:"message"`
	Data    ListAudiencePropertiesData `json:"data"`
}

// ListAudiencePropertiesData contains the paginated list of properties.
type ListAudiencePropertiesData struct {
	Properties []AudienceProperty `json:"properties"`
	Pagination PagePagination     `json:"pagination"`
}

// AudiencePropertyResponse is the response from endpoints that return a single property.
type AudiencePropertyResponse struct {
	Message string           `json:"message"`
	Data    AudienceProperty `json:"data"`
}

// CreateAudiencePropertyRequest is the body for creating a property.
// Name must match the pattern ^[a-z][a-z0-9_]*$.
type CreateAudiencePropertyRequest struct {
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	FallbackValue *string `json:"fallback_value,omitempty"`
}

// UpdateAudiencePropertyRequest is the body for updating a property's fallback
// value. The name and type fields are immutable.
//
// FallbackValue is a nullable PATCH field: leave it nil to omit,
// NewNullString("x") to replace, or &NullString{} (zero value) to clear the
// existing fallback server-side.
type UpdateAudiencePropertyRequest struct {
	FallbackValue *NullString `json:"fallback_value,omitempty"`
}

// List retrieves a paginated list of property definitions.
//
// Pass nil for params to use defaults.
func (s *AudiencePropertyService) List(ctx context.Context, params *ListAudiencePropertiesParams) (*ListAudiencePropertiesResponse, error) {
	path := "audience/properties"
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

	var resp ListAudiencePropertiesResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create creates a new property definition.
func (s *AudiencePropertyService) Create(ctx context.Context, params *CreateAudiencePropertyRequest) (*AudiencePropertyResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/properties", params)
	if err != nil {
		return nil, err
	}

	var resp AudiencePropertyResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single property by ID.
func (s *AudiencePropertyService) Get(ctx context.Context, propertyID string) (*AudiencePropertyResponse, error) {
	path := fmt.Sprintf("audience/properties/%s", url.PathEscape(propertyID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp AudiencePropertyResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update updates a property's fallback value.
func (s *AudiencePropertyService) Update(ctx context.Context, propertyID string, params *UpdateAudiencePropertyRequest) (*AudiencePropertyResponse, error) {
	path := fmt.Sprintf("audience/properties/%s", url.PathEscape(propertyID))

	req, err := s.client.newRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return nil, err
	}

	var resp AudiencePropertyResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete permanently deletes a property definition.
func (s *AudiencePropertyService) Delete(ctx context.Context, propertyID string) error {
	path := fmt.Sprintf("audience/properties/%s", url.PathEscape(propertyID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
