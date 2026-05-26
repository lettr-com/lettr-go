package lettr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AudienceContactService handles communication with the audience-contact
// endpoints of the Lettr API.
type AudienceContactService struct {
	client *Client
}

// Contact subscription status values returned by the API.
const (
	ContactStatusSubscribed   = "subscribed"
	ContactStatusUnsubscribed = "unsubscribed"
	ContactStatusBounced      = "bounced"
	ContactStatusComplained   = "complained"
	ContactStatusUnverified   = "unverified"
)

// AudienceContact represents a contact in the team's audience.
type AudienceContact struct {
	ID         string                     `json:"id"`
	Email      string                     `json:"email"`
	Status     string                     `json:"status"`
	Properties ContactProperties          `json:"properties"`
	CreatedAt  string                     `json:"created_at"`
	Lists      []AudienceContactListLink  `json:"lists"`
	Topics     []AudienceContactTopicLink `json:"topics"`
}

// ContactProperties holds a contact's custom property key-value pairs. It is
// a map[string]string with a tolerant UnmarshalJSON: when the API returns an
// empty associative array as `[]` (a PHP/Laravel serialization quirk for
// empty maps), it decodes to an empty ContactProperties instead of erroring.
type ContactProperties map[string]string

// UnmarshalJSON implements json.Unmarshaler. It accepts both a JSON object
// (`{...}`) and an empty JSON array (`[]`) for the empty-map case.
func (p *ContactProperties) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 2 && trimmed[0] == '[' && trimmed[1] == ']' {
		*p = ContactProperties{}
		return nil
	}
	// Unmarshal through the underlying map type to avoid recursion.
	m := map[string]string{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*p = m
	return nil
}

// AudienceContactListLink is a short reference to a list the contact belongs to.
type AudienceContactListLink struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AudienceContactTopicLink is a short reference to a topic the contact is subscribed to.
type AudienceContactTopicLink struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListAudienceContactsParams contains the query parameters for listing contacts.
type ListAudienceContactsParams struct {
	// PerPage is the number of results per page (1-100, default 20).
	PerPage int

	// Page is the page number (default 1).
	Page int

	// Search filters by email or contact name (max 255 chars).
	Search string

	// Status filters by contact status. Use the ContactStatus* constants.
	Status string

	// ListID filters to contacts in a specific list.
	ListID string

	// SegmentID filters to contacts matching a specific segment.
	SegmentID string
}

// ListAudienceContactsResponse is the response from listing audience contacts.
type ListAudienceContactsResponse struct {
	Message string                   `json:"message"`
	Data    ListAudienceContactsData `json:"data"`
}

// ListAudienceContactsData contains the paginated list of contacts.
type ListAudienceContactsData struct {
	Contacts   []AudienceContact `json:"contacts"`
	Pagination PagePagination    `json:"pagination"`
}

// AudienceContactResponse is the response from endpoints that return a single
// contact (Create, Get, Update).
type AudienceContactResponse struct {
	Message string          `json:"message"`
	Data    AudienceContact `json:"data"`
}

// DoubleOptInConfig configures the confirmation email sent when a contact is
// created with double opt-in. When provided, the contact is created in
// "unverified" status until they click the confirmation link.
type DoubleOptInConfig struct {
	From        string  `json:"from"`
	FromName    *string `json:"from_name,omitempty"`
	Subject     string  `json:"subject"`
	TemplateSlug string `json:"template_slug"`
	RedirectURL string  `json:"redirect_url"`
}

// CreateAudienceContactRequest is the body for creating a single contact.
type CreateAudienceContactRequest struct {
	Email       string             `json:"email"`
	ListID      *string            `json:"list_id,omitempty"`
	Properties  map[string]string  `json:"properties,omitempty"`
	DoubleOptIn *DoubleOptInConfig `json:"double_opt_in,omitempty"`
}

// UpdateAudienceContactRequest is the body for partially updating a contact.
// A property set to a nil pointer in Properties is sent as JSON null and
// removes that property from the contact server-side.
type UpdateAudienceContactRequest struct {
	Email      string             `json:"email,omitempty"`
	Status     string             `json:"status,omitempty"`
	Properties map[string]*string `json:"properties,omitempty"`
}

// BulkCreateAudienceContactsRequest is the body for bulk-creating contacts.
type BulkCreateAudienceContactsRequest struct {
	// Emails is 1-1000 email addresses. They are normalized and deduplicated server-side.
	Emails     []string          `json:"emails"`
	ListID     *string           `json:"list_id,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

// BulkCreateAudienceContactsResponse is the response from a bulk create.
type BulkCreateAudienceContactsResponse struct {
	Message string                         `json:"message"`
	Data    BulkCreateAudienceContactsData `json:"data"`
}

// BulkCreateAudienceContactsData holds the counts from a bulk-create operation.
type BulkCreateAudienceContactsData struct {
	Created        int `json:"created"`
	AlreadyExisted int `json:"already_existed"`
}

// BulkContactsListsRequest is the body for bulk attach/detach between contacts
// and lists. Every contact × list combination is processed.
type BulkContactsListsRequest struct {
	// ContactIDs is 1-1000 contact IDs.
	ContactIDs []string `json:"contact_ids"`
	// ListIDs is 1-50 list IDs.
	ListIDs []string `json:"list_ids"`
}

// BulkAttachContactsToListsResponse is the response from bulk attach.
type BulkAttachContactsToListsResponse struct {
	Message string                        `json:"message"`
	Data    BulkAttachContactsToListsData `json:"data"`
}

// BulkAttachContactsToListsData holds the counts from a bulk-attach operation.
type BulkAttachContactsToListsData struct {
	Attached        int `json:"attached"`
	AlreadyAttached int `json:"already_attached"`
	TotalPairs      int `json:"total_pairs"`
}

// BulkDetachContactsFromListsResponse is the response from bulk detach.
type BulkDetachContactsFromListsResponse struct {
	Message string                          `json:"message"`
	Data    BulkDetachContactsFromListsData `json:"data"`
}

// BulkDetachContactsFromListsData holds the counts from a bulk-detach operation.
type BulkDetachContactsFromListsData struct {
	Detached   int `json:"detached"`
	NotPresent int `json:"not_present"`
	TotalPairs int `json:"total_pairs"`
}

// AttachContactResponse is the response from attaching a contact to a list or
// subscribing a contact to a topic. The endpoint returns only a status message;
// no payload data is sent.
type AttachContactResponse struct {
	Message string `json:"message"`
}

// List retrieves a paginated, filterable list of audience contacts.
//
// Pass nil for params to use defaults.
func (s *AudienceContactService) List(ctx context.Context, params *ListAudienceContactsParams) (*ListAudienceContactsResponse, error) {
	path := "audience/contacts"
	if params != nil {
		q := url.Values{}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.Search != "" {
			q.Set("search", params.Search)
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if params.ListID != "" {
			q.Set("list_id", params.ListID)
		}
		if params.SegmentID != "" {
			q.Set("segment_id", params.SegmentID)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListAudienceContactsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create creates a single contact. When params.DoubleOptIn is set, the contact
// is created in "unverified" status and the confirmation email is sent.
func (s *AudienceContactService) Create(ctx context.Context, params *CreateAudienceContactRequest) (*AudienceContactResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/contacts", params)
	if err != nil {
		return nil, err
	}

	var resp AudienceContactResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single contact by ID.
func (s *AudienceContactService) Get(ctx context.Context, contactID string) (*AudienceContactResponse, error) {
	path := fmt.Sprintf("audience/contacts/%s", url.PathEscape(contactID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp AudienceContactResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update partially updates a contact.
func (s *AudienceContactService) Update(ctx context.Context, contactID string, params *UpdateAudienceContactRequest) (*AudienceContactResponse, error) {
	path := fmt.Sprintf("audience/contacts/%s", url.PathEscape(contactID))

	req, err := s.client.newRequest(ctx, http.MethodPatch, path, params)
	if err != nil {
		return nil, err
	}

	var resp AudienceContactResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete permanently deletes a contact.
func (s *AudienceContactService) Delete(ctx context.Context, contactID string) error {
	path := fmt.Sprintf("audience/contacts/%s", url.PathEscape(contactID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}

// BulkCreate creates up to 1000 contacts in a single request.
func (s *AudienceContactService) BulkCreate(ctx context.Context, params *BulkCreateAudienceContactsRequest) (*BulkCreateAudienceContactsResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/contacts/bulk", params)
	if err != nil {
		return nil, err
	}

	var resp BulkCreateAudienceContactsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AttachToList adds a contact to a list. Returns 201 when the contact is newly
// attached or 200 when the contact was already in the list.
func (s *AudienceContactService) AttachToList(ctx context.Context, contactID, listID string) (*AttachContactResponse, error) {
	path := fmt.Sprintf("audience/contacts/%s/lists/%s", url.PathEscape(contactID), url.PathEscape(listID))

	req, err := s.client.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var resp AttachContactResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DetachFromList removes a contact from a list. The operation is idempotent.
func (s *AudienceContactService) DetachFromList(ctx context.Context, contactID, listID string) error {
	path := fmt.Sprintf("audience/contacts/%s/lists/%s", url.PathEscape(contactID), url.PathEscape(listID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}

// BulkAttachToLists attaches every combination of contact_ids × list_ids
// (up to 1000 contacts × 50 lists).
func (s *AudienceContactService) BulkAttachToLists(ctx context.Context, params *BulkContactsListsRequest) (*BulkAttachContactsToListsResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/contacts/lists/bulk", params)
	if err != nil {
		return nil, err
	}

	var resp BulkAttachContactsToListsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BulkDetachFromLists detaches every combination of contact_ids × list_ids.
func (s *AudienceContactService) BulkDetachFromLists(ctx context.Context, params *BulkContactsListsRequest) (*BulkDetachContactsFromListsResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodDelete, "audience/contacts/lists/bulk", params)
	if err != nil {
		return nil, err
	}

	var resp BulkDetachContactsFromListsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SubscribeToTopic subscribes a contact to a topic. Returns 201 when the
// subscription is newly created or 200 when it already existed.
func (s *AudienceContactService) SubscribeToTopic(ctx context.Context, contactID, topicID string) (*AttachContactResponse, error) {
	path := fmt.Sprintf("audience/contacts/%s/topics/%s", url.PathEscape(contactID), url.PathEscape(topicID))

	req, err := s.client.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var resp AttachContactResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsubscribeFromTopic removes a contact's subscription from a topic.
// The operation is idempotent.
func (s *AudienceContactService) UnsubscribeFromTopic(ctx context.Context, contactID, topicID string) error {
	path := fmt.Sprintf("audience/contacts/%s/topics/%s", url.PathEscape(contactID), url.PathEscape(topicID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
