package lettr

import (
	"context"
	"fmt"
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
	Key      string          `json:"key"`
	Required bool            `json:"required"`
	Type     string          `json:"type,omitempty"`
	Children []MergeTagChild `json:"children,omitempty"`
}

// MergeTagChild represents a child merge tag within a loop block.
type MergeTagChild struct {
	Key  string `json:"key"`
	Type string `json:"type,omitempty"`
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

// TemplateDetail represents detailed information about a template,
// including version info and content.
type TemplateDetail struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	ProjectID     int    `json:"project_id"`
	FolderID      int    `json:"folder_id"`
	ActiveVersion int    `json:"active_version"`
	VersionsCount int    `json:"versions_count"`
	Html          string `json:"html,omitempty"`
	Json          string `json:"json,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// GetTemplateParams contains optional query parameters for getting a template.
type GetTemplateParams struct {
	// ProjectID is the project to look in. Uses team's default if not set.
	ProjectID int
}

// GetTemplateResponse is the response from getting a single template.
type GetTemplateResponse struct {
	Message string         `json:"message"`
	Data    TemplateDetail `json:"data"`
}

// Get retrieves details of a single template by its slug.
//
// Pass nil for params to use the team's default project.
//
// Example:
//
//	template, err := client.Templates.Get(ctx, "welcome-email", nil)
func (s *TemplateService) Get(ctx context.Context, slug string, params *GetTemplateParams) (*GetTemplateResponse, error) {
	path := fmt.Sprintf("templates/%s", url.PathEscape(slug))
	if params != nil {
		q := url.Values{}
		if params.ProjectID > 0 {
			q.Set("project_id", strconv.Itoa(params.ProjectID))
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetTemplateResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateTemplateRequest represents the request body for updating a template.
type UpdateTemplateRequest struct {
	// Name is the new name for the template.
	Name string `json:"name,omitempty"`

	// Html is new HTML content (creates a new active version).
	Html string `json:"html,omitempty"`

	// Json is new Topol editor JSON content (creates a new active version).
	Json string `json:"json,omitempty"`

	// ProjectID is the project containing the template.
	ProjectID *int `json:"project_id,omitempty"`
}

// UpdateTemplateData contains the result of updating a template.
type UpdateTemplateData struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	ProjectID     int        `json:"project_id"`
	FolderID      int        `json:"folder_id"`
	ActiveVersion int        `json:"active_version"`
	MergeTags     []MergeTag `json:"merge_tags"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

// UpdateTemplateResponse is the response from updating a template.
type UpdateTemplateResponse struct {
	Message string             `json:"message"`
	Data    UpdateTemplateData `json:"data"`
}

// Update modifies an existing template's name or content.
//
// Example:
//
//	updated, err := client.Templates.Update(ctx, "welcome-email", &lettr.UpdateTemplateRequest{
//	    Html: "<h1>Updated Hello {{FIRST_NAME}}!</h1>",
//	})
func (s *TemplateService) Update(ctx context.Context, slug string, params *UpdateTemplateRequest) (*UpdateTemplateResponse, error) {
	path := fmt.Sprintf("templates/%s", url.PathEscape(slug))

	req, err := s.client.newRequest(ctx, http.MethodPut, path, params)
	if err != nil {
		return nil, err
	}

	var resp UpdateTemplateResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteTemplateParams contains optional query parameters for deleting a template.
type DeleteTemplateParams struct {
	// ProjectID is the project containing the template.
	ProjectID int
}

// DeleteTemplateResponse is the response from deleting a template.
type DeleteTemplateResponse struct {
	Message string `json:"message"`
}

// Delete permanently removes a template.
//
// Pass nil for params to use the team's default project.
//
// Example:
//
//	err := client.Templates.Delete(ctx, "welcome-email", nil)
func (s *TemplateService) Delete(ctx context.Context, slug string, params *DeleteTemplateParams) error {
	path := fmt.Sprintf("templates/%s", url.PathEscape(slug))
	if params != nil {
		q := url.Values{}
		if params.ProjectID > 0 {
			q.Set("project_id", strconv.Itoa(params.ProjectID))
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}

// GetMergeTagsParams contains optional query parameters for getting merge tags.
type GetMergeTagsParams struct {
	// ProjectID is the project containing the template.
	ProjectID int

	// Version is the template version number. Uses active version if not set.
	Version int
}

// GetMergeTagsResponse is the response from getting merge tags.
type GetMergeTagsResponse struct {
	Message string           `json:"message"`
	Data    GetMergeTagsData `json:"data"`
}

// GetMergeTagsData contains merge tags for a template version.
type GetMergeTagsData struct {
	ProjectID    int        `json:"project_id"`
	TemplateSlug string     `json:"template_slug"`
	Version      int        `json:"version"`
	MergeTags    []MergeTag `json:"merge_tags"`
}

// GetMergeTags retrieves the merge tags for a template.
//
// Pass nil for params to use defaults (team's default project, active version).
//
// Example:
//
//	tags, err := client.Templates.GetMergeTags(ctx, "welcome-email", nil)
func (s *TemplateService) GetMergeTags(ctx context.Context, slug string, params *GetMergeTagsParams) (*GetMergeTagsResponse, error) {
	path := fmt.Sprintf("templates/%s/merge-tags", url.PathEscape(slug))
	if params != nil {
		q := url.Values{}
		if params.ProjectID > 0 {
			q.Set("project_id", strconv.Itoa(params.ProjectID))
		}
		if params.Version > 0 {
			q.Set("version", strconv.Itoa(params.Version))
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetMergeTagsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTemplateHtmlParams contains the query parameters for getting template HTML.
type GetTemplateHtmlParams struct {
	// ProjectID is the project containing the template (required).
	ProjectID int

	// Slug is the template slug identifier (required).
	Slug string
}

// GetTemplateHtmlResponse is the response from getting template HTML.
type GetTemplateHtmlResponse struct {
	Success bool                `json:"success"`
	Data    GetTemplateHtmlData `json:"data"`
}

// GetTemplateHtmlData contains the HTML content of a template.
type GetTemplateHtmlData struct {
	Html string `json:"html"`
}

// GetHtml retrieves the rendered HTML content of a template.
//
// Example:
//
//	html, err := client.Templates.GetHtml(ctx, &lettr.GetTemplateHtmlParams{
//	    ProjectID: 1,
//	    Slug:      "welcome-email",
//	})
func (s *TemplateService) GetHtml(ctx context.Context, params *GetTemplateHtmlParams) (*GetTemplateHtmlResponse, error) {
	path := "templates/html"
	if params != nil {
		q := url.Values{}
		if params.ProjectID > 0 {
			q.Set("project_id", strconv.Itoa(params.ProjectID))
		}
		if params.Slug != "" {
			q.Set("slug", params.Slug)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetTemplateHtmlResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
