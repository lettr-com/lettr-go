package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AudienceListService handles communication with the audience-list endpoints
// of the Lettr API.
type AudienceListService struct {
	client *Client
}

// AudienceList represents an audience list.
type AudienceList struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContactsCount int    `json:"contacts_count"`
}

// ListAudienceListsParams contains the query parameters for listing audience lists.
type ListAudienceListsParams struct {
	// PerPage is the number of results per page (1-100, default 20).
	PerPage int

	// Page is the page number (default 1).
	Page int
}

// ListAudienceListsResponse is the response from listing audience lists.
type ListAudienceListsResponse struct {
	Message string                `json:"message"`
	Data    ListAudienceListsData `json:"data"`
}

// ListAudienceListsData contains the paginated list of audience lists.
type ListAudienceListsData struct {
	Lists      []AudienceList `json:"lists"`
	Pagination PagePagination `json:"pagination"`
}

// AudienceListResponse is the response from endpoints that return a single
// audience list (Create, Get, Update).
type AudienceListResponse struct {
	Message string       `json:"message"`
	Data    AudienceList `json:"data"`
}

// CreateAudienceListRequest is the body for creating an audience list.
type CreateAudienceListRequest struct {
	// Name must be unique within the team (max 255 chars).
	Name string `json:"name"`
}

// UpdateAudienceListRequest is the body for partially updating an audience list.
type UpdateAudienceListRequest struct {
	Name string `json:"name,omitempty"`
}

// BulkDeleteAudienceListsRequest is the body for bulk-deleting audience lists.
type BulkDeleteAudienceListsRequest struct {
	// ListIDs is 1-50 list IDs to delete.
	ListIDs []string `json:"list_ids"`
}

// BulkDeleteAudienceListsResponse is the response from a bulk delete.
type BulkDeleteAudienceListsResponse struct {
	Message string                      `json:"message"`
	Data    BulkDeleteAudienceListsData `json:"data"`
}

// BulkDeleteAudienceListsData holds the counts from a bulk-delete operation.
type BulkDeleteAudienceListsData struct {
	Deleted int `json:"deleted"`
}

// List retrieves a paginated list of audience lists.
//
// Pass nil for params to use defaults.
func (s *AudienceListService) List(ctx context.Context, params *ListAudienceListsParams) (*ListAudienceListsResponse, error) {
	path := "audience/lists"
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

	var resp ListAudienceListsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create creates a new audience list.
func (s *AudienceListService) Create(ctx context.Context, params *CreateAudienceListRequest) (*AudienceListResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/lists", params)
	if err != nil {
		return nil, err
	}

	var resp AudienceListResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single audience list by ID.
func (s *AudienceListService) Get(ctx context.Context, listID string) (*AudienceListResponse, error) {
	path := fmt.Sprintf("audience/lists/%s", url.PathEscape(listID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp AudienceListResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update partially updates an audience list.
func (s *AudienceListService) Update(ctx context.Context, listID string, params *UpdateAudienceListRequest) (*AudienceListResponse, error) {
	path := fmt.Sprintf("audience/lists/%s", url.PathEscape(listID))

	req, err := s.client.newRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return nil, err
	}

	var resp AudienceListResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete permanently deletes an audience list.
func (s *AudienceListService) Delete(ctx context.Context, listID string) error {
	path := fmt.Sprintf("audience/lists/%s", url.PathEscape(listID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}

// BulkDelete deletes up to 50 audience lists in a single request.
func (s *AudienceListService) BulkDelete(ctx context.Context, listIDs []string) (*BulkDeleteAudienceListsResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodDelete, "audience/lists/bulk", &BulkDeleteAudienceListsRequest{ListIDs: listIDs})
	if err != nil {
		return nil, err
	}

	var resp BulkDeleteAudienceListsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
