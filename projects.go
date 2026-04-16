package lettr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ProjectService handles communication with the project-related endpoints
// of the Lettr API.
type ProjectService struct {
	client *Client
}

// Project represents a project.
type Project struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	TeamID    int    `json:"team_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Emoji     string `json:"emoji,omitempty"`
}

// ListProjectsParams contains the query parameters for listing projects.
type ListProjectsParams struct {
	// PerPage is the number of results per page (1-100, default 25).
	PerPage int

	// Page is the page number (default 1).
	Page int
}

// ListProjectsResponse is the response from listing projects.
type ListProjectsResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    ListProjectsData `json:"data"`
}

// ListProjectsData contains the paginated list of projects.
type ListProjectsData struct {
	Projects   []Project      `json:"projects"`
	Pagination PagePagination `json:"pagination"`
}

// List retrieves a paginated list of projects associated with the team.
//
// Pass nil for params to use defaults.
//
// Example:
//
//	projects, err := client.Projects.List(ctx, nil)
func (s *ProjectService) List(ctx context.Context, params *ListProjectsParams) (*ListProjectsResponse, error) {
	path := "projects"
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

	var resp ListProjectsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
