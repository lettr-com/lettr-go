package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AudienceSegmentService handles communication with the audience-segment
// endpoints of the Lettr API.
type AudienceSegmentService struct {
	client *Client
}

// Segment condition operator values.
const (
	SegmentOperatorContains           = "contains"
	SegmentOperatorNotContains        = "not_contains"
	SegmentOperatorEquals             = "equals"
	SegmentOperatorNotEquals          = "not_equals"
	SegmentOperatorStartsWith         = "starts_with"
	SegmentOperatorNotStartsWith      = "not_starts_with"
	SegmentOperatorEndsWith           = "ends_with"
	SegmentOperatorNotEndsWith        = "not_ends_with"
	SegmentOperatorIsTrue             = "is_true"
	SegmentOperatorIsFalse            = "is_false"
	SegmentOperatorGreaterThan        = "greater_than"
	SegmentOperatorGreaterThanOrEqual = "greater_than_or_equal"
	SegmentOperatorLessThan           = "less_than"
	SegmentOperatorLessThanOrEqual    = "less_than_or_equal"
	SegmentOperatorBefore             = "before"
	SegmentOperatorAfter              = "after"
)

// SegmentCondition is a single field/operator/value match.
// Value is optional and not required when Operator is "is_true" or "is_false".
type SegmentCondition struct {
	Field    string  `json:"field"`
	Operator string  `json:"operator"`
	Value    *string `json:"value,omitempty"`
}

// SegmentConditionGroup is a set of conditions joined by OR. Multiple groups
// within a segment are joined by AND, i.e. (A OR B) AND (C OR D).
type SegmentConditionGroup struct {
	Conditions []SegmentCondition `json:"conditions"`
}

// SegmentConditionsInput is the request-side wrapper for segment conditions.
type SegmentConditionsInput struct {
	Groups []SegmentConditionGroup `json:"groups"`
}

// AudienceSegment represents an audience segment.
type AudienceSegment struct {
	ID                  string                  `json:"id"`
	Name                string                  `json:"name"`
	ListID              *string                 `json:"list_id"`
	ListName            *string                 `json:"list_name"`
	ConditionGroups     []SegmentConditionGroup `json:"condition_groups"`
	CachedContactsCount *int                    `json:"cached_contacts_count"`
	CreatedAt           string                  `json:"created_at"`
}

// ListAudienceSegmentsParams contains the query parameters for listing segments.
type ListAudienceSegmentsParams struct {
	PerPage int
	Page    int

	// ListID filters segments to those restricted to a specific list.
	ListID string
}

// ListAudienceSegmentsResponse is the response from listing segments.
type ListAudienceSegmentsResponse struct {
	Message string                   `json:"message"`
	Data    ListAudienceSegmentsData `json:"data"`
}

// ListAudienceSegmentsData contains the paginated list of segments.
type ListAudienceSegmentsData struct {
	Segments   []AudienceSegment `json:"segments"`
	Pagination PagePagination    `json:"pagination"`
}

// AudienceSegmentResponse is the response from endpoints that return a single segment.
type AudienceSegmentResponse struct {
	Message string          `json:"message"`
	Data    AudienceSegment `json:"data"`
}

// CreateAudienceSegmentRequest is the body for creating a segment.
type CreateAudienceSegmentRequest struct {
	Name       string                 `json:"name"`
	ListID     *string                `json:"list_id,omitempty"`
	Conditions SegmentConditionsInput `json:"conditions"`
}

// UpdateAudienceSegmentRequest is the body for partially updating a segment.
//
// ListID is a nullable PATCH field: leave it nil to omit, NewNullString(uuid)
// to restrict, or &NullString{} (zero value) to clear the list restriction
// (so the segment matches across all lists).
type UpdateAudienceSegmentRequest struct {
	Name       string                  `json:"name,omitempty"`
	ListID     *NullString             `json:"list_id,omitempty"`
	Conditions *SegmentConditionsInput `json:"conditions,omitempty"`
}

// List retrieves a paginated list of segments.
//
// Pass nil for params to use defaults.
func (s *AudienceSegmentService) List(ctx context.Context, params *ListAudienceSegmentsParams) (*ListAudienceSegmentsResponse, error) {
	path := "audience/segments"
	if params != nil {
		q := url.Values{}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.ListID != "" {
			q.Set("list_id", params.ListID)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListAudienceSegmentsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create creates a new segment.
func (s *AudienceSegmentService) Create(ctx context.Context, params *CreateAudienceSegmentRequest) (*AudienceSegmentResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/segments", params)
	if err != nil {
		return nil, err
	}

	var resp AudienceSegmentResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single segment by ID.
func (s *AudienceSegmentService) Get(ctx context.Context, segmentID string) (*AudienceSegmentResponse, error) {
	path := fmt.Sprintf("audience/segments/%s", url.PathEscape(segmentID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp AudienceSegmentResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update partially updates a segment.
func (s *AudienceSegmentService) Update(ctx context.Context, segmentID string, params *UpdateAudienceSegmentRequest) (*AudienceSegmentResponse, error) {
	path := fmt.Sprintf("audience/segments/%s", url.PathEscape(segmentID))

	req, err := s.client.newRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return nil, err
	}

	var resp AudienceSegmentResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete permanently deletes a segment.
func (s *AudienceSegmentService) Delete(ctx context.Context, segmentID string) error {
	path := fmt.Sprintf("audience/segments/%s", url.PathEscape(segmentID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
