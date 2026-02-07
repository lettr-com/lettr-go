package lettr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// TemplateService handles communication with the template-related endpoints
// of the Lettr API.
type TemplateService struct {
	client *Client
}

// Template represents an email template.
type Template struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	ProjectID int    `json:"project_id"`
	FolderID  int    `json:"folder_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MergeTag represents a merge tag extracted from template content.
type MergeTag struct {
	Key      string `json:"key"`
	Required bool   `json:"required"`
}

// ListTemplatesParams contains the query parameters for listing templates.
type ListTemplatesParams struct {
	// ProjectID is the project to retrieve templates from. Uses the team's
	// default project if not set.
	ProjectID int

	// PerPage is the number of results per page (1-100, default 25).
	PerPage int

	// Page is the page number (default 1).
	Page int
}

// ListTemplatesResponse is the response from listing templates.
type ListTemplatesResponse struct {
	Message string            `json:"message"`
	Data    ListTemplatesData `json:"data"`
}

// ListTemplatesData contains the paginated list of templates.
type ListTemplatesData struct {
	Templates  []Template     `json:"templates"`
	Pagination PagePagination `json:"pagination"`
}

// PagePagination holds page-based pagination info.
type PagePagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
}

// CreateTemplateRequest represents the request body for creating a template.
type CreateTemplateRequest struct {
	// Name is the template name (required).
	Name string `json:"name"`

	// Html is the HTML content for the template. Mutually exclusive with Json.
	Html string `json:"html,omitempty"`

	// Json is the Topol editor JSON content. Mutually exclusive with Html.
	Json string `json:"json,omitempty"`

	// ProjectID specifies which project to create the template in.
	ProjectID *int `json:"project_id,omitempty"`

	// FolderID specifies which folder within the project.
	FolderID *int `json:"folder_id,omitempty"`
}

// CreateTemplateResponse is the response from creating a template.
type CreateTemplateResponse struct {
	Message string             `json:"message"`
	Data    CreateTemplateData `json:"data"`
}

// CreateTemplateData contains the result of creating a template.
type CreateTemplateData struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	ProjectID     int        `json:"project_id"`
	FolderID      int        `json:"folder_id"`
	ActiveVersion int        `json:"active_version"`
	MergeTags     []MergeTag `json:"merge_tags"`
	CreatedAt     string     `json:"created_at"`
}

// List retrieves a paginated list of email templates.
//
// Pass nil for params to use defaults.
//
// Example:
//
//	templates, err := client.Templates.List(ctx, nil)
func (s *TemplateService) List(ctx context.Context, params *ListTemplatesParams) (*ListTemplatesResponse, error) {
	path := "templates"
	if params != nil {
		q := url.Values{}
		if params.ProjectID > 0 {
			q.Set("project_id", strconv.Itoa(params.ProjectID))
		}
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

	var resp ListTemplatesResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create creates a new email template with HTML or Topol JSON content.
//
// Example:
//
//	template, err := client.Templates.Create(ctx, &lettr.CreateTemplateRequest{
//	    Name: "Welcome Email",
//	    Html: "<h1>Hello {{FIRST_NAME}}!</h1>",
//	})
func (s *TemplateService) Create(ctx context.Context, params *CreateTemplateRequest) (*CreateTemplateResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "templates", params)
	if err != nil {
		return nil, err
	}

	var resp CreateTemplateResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
