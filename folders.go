package lettr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// FolderService handles communication with the folder-related endpoints of the
// Lettr API.
//
// Read-only: creating, renaming and deleting folders stay in the app, because
// deleting one moves or deletes the templates inside it.
type FolderService struct {
	client *Client
}

// Folder is a folder templates are filed into.
//
// ID is what CreateTemplateRequest.FolderID expects, so listing folders is how
// a caller picks where a template lands instead of hardcoding an integer read
// out of an app URL.
type Folder struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	ProjectID      int             `json:"project_id"`
	Purpose        TemplatePurpose `json:"purpose"`
	TemplatesCount int             `json:"templates_count"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

// ListFoldersParams contains the query parameters for listing folders.
type ListFoldersParams struct {
	// ProjectID is the project to list folders from. Uses the team's default
	// project if not set, the same way Templates.List resolves it.
	ProjectID int

	// Purpose narrows the list to one module. Both are returned if not set.
	Purpose TemplatePurpose

	// PerPage is the number of results per page (1-100, default 25).
	PerPage int

	// Page is the page number (default 1).
	Page int
}

// ListFoldersResponse is the response from listing folders.
type ListFoldersResponse struct {
	Message string          `json:"message"`
	Data    ListFoldersData `json:"data"`
}

// ListFoldersData contains the paginated list of folders.
type ListFoldersData struct {
	Folders    []Folder       `json:"folders"`
	Pagination PagePagination `json:"pagination"`
}

// List retrieves the folders templates are filed into.
//
// Pass nil for params to use defaults.
//
// Example:
//
//	folders, err := client.Folders.List(ctx, &lettr.ListFoldersParams{
//	    Purpose: lettr.PurposeCampaign,
//	})
func (s *FolderService) List(ctx context.Context, params *ListFoldersParams) (*ListFoldersResponse, error) {
	path := "folders"
	if params != nil {
		q := url.Values{}
		if params.ProjectID > 0 {
			q.Set("project_id", strconv.Itoa(params.ProjectID))
		}
		if params.Purpose != "" {
			q.Set("purpose", string(params.Purpose))
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

	var resp ListFoldersResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
