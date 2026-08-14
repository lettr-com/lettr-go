package lettr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

// Topic subscription states for a write request. These say what a request
// should *do* with a topic, and are distinct from a topic's
// DefaultSubscription, which describes how the topic behaves for a contact that
// says nothing.
//
// TopicOptOut also cancels the auto-subscription a topic whose default is
// opt-out would otherwise give a newly created contact, so a create and an
// unsubscribe fit in one request.
const (
	TopicOptIn  = "opt_in"
	TopicOptOut = "opt_out"
)

// AudienceTopicSubscription is a topic and the subscription state to apply to
// it. Used batch-wide on BulkCreateAudienceContactsRequest and per row on
// BulkAudienceContactRow. A row-level opt-out wins over a batch-level opt-in
// for that contact.
//
// Build them with SubscribeTopic / UnsubscribeTopic for readability at the
// call site.
type AudienceTopicSubscription struct {
	ID string `json:"id"`

	// Subscription is TopicOptIn or TopicOptOut. Empty means opt-in server-side.
	Subscription string `json:"subscription,omitempty"`
}

// SubscribeTopic returns a subscription that opts the contact in to topicID.
func SubscribeTopic(topicID string) AudienceTopicSubscription {
	return AudienceTopicSubscription{ID: topicID, Subscription: TopicOptIn}
}

// UnsubscribeTopic returns a subscription that opts the contact out of topicID,
// including a topic that would otherwise auto-subscribe new contacts.
func UnsubscribeTopic(topicID string) AudienceTopicSubscription {
	return AudienceTopicSubscription{ID: topicID, Subscription: TopicOptOut}
}

// BulkAudienceContactRow is one contact in a bulk-create payload.
//
// ListIDs and Topics here are applied on top of the batch-wide ones on
// BulkCreateAudienceContactsRequest; a Properties key here overrides the
// batch-wide value for the same key.
//
// A row that fails validation is skipped rather than failing the request — it
// comes back in BulkCreateAudienceContactsData.Errors.
type BulkAudienceContactRow struct {
	Email string `json:"email"`

	// Properties keys must each match a property defined for the team.
	Properties map[string]string `json:"properties,omitempty"`

	// ListIDs is up to 50 list IDs for this row, on top of the batch-wide ones.
	ListIDs []string `json:"list_ids,omitempty"`

	// Topics is up to 50 topic subscriptions for this row.
	Topics []AudienceTopicSubscription `json:"topics,omitempty"`
}

// BulkCreateAudienceContactsRequest is the body for bulk-creating contacts.
//
// Two shapes are supported, and exactly one of them must be filled in:
//
//   - Emails — a flat list of addresses that all share the batch-wide ListID /
//     ListIDs, Properties and Topics. The original shape, unchanged.
//   - Contacts — one BulkAudienceContactRow per contact, each with its own
//     properties, lists and topic subscriptions.
//
// Batch-wide ListIDs and Topics are unioned into every row; a row-level
// property key or opt-out wins over the batch-wide value.
type BulkCreateAudienceContactsRequest struct {
	// Emails is 1-1000 email addresses. They are normalized and deduplicated
	// server-side. Leave nil when using Contacts.
	Emails []string `json:"emails,omitempty"`

	// ListID is a single batch-wide list, folded into ListIDs server-side.
	ListID *string `json:"list_id,omitempty"`

	// Properties applies to every contact in the batch; a row's own key wins.
	Properties map[string]string `json:"properties,omitempty"`

	// Contacts is 1-1000 rows. Alternative to Emails.
	Contacts []BulkAudienceContactRow `json:"contacts,omitempty"`

	// ListIDs is up to 50 batch-wide lists.
	ListIDs []string `json:"list_ids,omitempty"`

	// Topics is up to 50 batch-wide topic subscriptions.
	Topics []AudienceTopicSubscription `json:"topics,omitempty"`

	// UpdateExisting merges properties into contacts that already exist
	// (submitted keys overwrite, absent keys are preserved) and allows dropping
	// a subscription via an opt-out. Defaults to false, in which case existing
	// contacts keep their properties but are still attached to the requested
	// lists. Omitted from the payload when false, so a legacy request stays
	// byte-identical on the wire.
	UpdateExisting bool `json:"update_existing,omitempty"`
}

// Reasons a single row was skipped during a bulk create. These are per-row
// codes reported inside a 201 body — not the top-level Error.ErrorCode of a
// failed request.
const (
	BulkContactErrorMissingEmail             = "missing_email"
	BulkContactErrorInvalidEmail             = "invalid_email"
	BulkContactErrorInvalidPropertyValue     = "invalid_property_value"
	BulkContactErrorUnknownPropertyKey       = "unknown_property_key"
	BulkContactErrorUnknownList              = "unknown_list"
	BulkContactErrorUnknownTopic             = "unknown_topic"
	BulkContactErrorInvalidTopicSubscription = "invalid_topic_subscription"
)

// BulkAudienceContactError is a row that was skipped during a bulk create.
type BulkAudienceContactError struct {
	// Index is the zero-based position of the row in the submitted slice.
	Index int `json:"index"`

	Email string `json:"email"`

	// ErrorCode is one of the BulkContactError* constants. It is a plain string
	// so a code added server-side is still readable here.
	ErrorCode string `json:"error_code"`

	Error string `json:"error"`
}

// BulkAudienceContactRef identifies a contact that exists after a bulk create,
// so a caller can chain into the bulk list and topic endpoints without a
// follow-up lookup.
type BulkAudienceContactRef struct {
	ID    string `json:"id"`
	Email string `json:"email"`

	// Created is true when this request created the contact, false when it
	// already existed.
	Created bool `json:"created"`
}

// BulkCreateAudienceContactsResponse is the response from a bulk create.
type BulkCreateAudienceContactsResponse struct {
	Message string                         `json:"message"`
	Data    BulkCreateAudienceContactsData `json:"data"`
}

// BulkCreateAudienceContactsData holds the results of a bulk-create operation.
//
// A bulk create can partially succeed: rows that fail validation are skipped
// and reported in Errors while the rest of the batch is written, and the call
// still returns HTTP 201. A nil error from BulkCreate therefore does not mean
// every row landed — check HasErrors.
//
// AlreadyExisted and Updated overlap by design. They answer different questions
// ("was the address already in the audience?" vs "did this request change the
// contact?"), so they do not sum to the row count: a contact that already
// existed and got attached to a list is counted in both.
type BulkCreateAudienceContactsData struct {
	Created        int `json:"created"`
	AlreadyExisted int `json:"already_existed"`

	// Updated counts existing contacts this request changed — properties
	// merged, a list or topic attached, or a subscription dropped.
	Updated int `json:"updated"`

	// ErrorCount is the number of skipped rows.
	ErrorCount int `json:"error_count"`

	Errors []BulkAudienceContactError `json:"errors"`

	// Contacts lists every contact that exists after the request, in
	// submission order.
	Contacts []BulkAudienceContactRef `json:"contacts"`
}

// HasErrors reports whether any row was skipped. Always check this — a bulk
// create reports partial failures in the body, not in the HTTP status.
func (d BulkCreateAudienceContactsData) HasErrors() bool {
	return len(d.Errors) > 0
}

// ContactIDs returns the IDs of every contact that exists after the request, in
// submission order — ready to pass to BulkAttachToLists or BulkSubscribeToTopics.
func (d BulkCreateAudienceContactsData) ContactIDs() []string {
	ids := make([]string, 0, len(d.Contacts))
	for _, c := range d.Contacts {
		ids = append(ids, c.ID)
	}
	return ids
}

// IDFor looks up the ID for a submitted address, reporting whether it was
// found. Matching is case-insensitive because the API normalizes addresses
// before storing them.
func (d BulkCreateAudienceContactsData) IDFor(email string) (string, bool) {
	needle := strings.ToLower(strings.TrimSpace(email))
	for _, c := range d.Contacts {
		if strings.ToLower(c.Email) == needle {
			return c.ID, true
		}
	}
	return "", false
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

// BulkContactsTopicsRequest is the body for bulk subscribe/unsubscribe between
// contacts and topics. Every contact × topic combination is processed.
type BulkContactsTopicsRequest struct {
	// ContactIDs is 1-1000 contact IDs.
	ContactIDs []string `json:"contact_ids"`
	// TopicIDs is 1-50 topic IDs.
	TopicIDs []string `json:"topic_ids"`
}

// BulkSubscribeContactsToTopicsResponse is the response from a bulk subscribe.
type BulkSubscribeContactsToTopicsResponse struct {
	Message string                            `json:"message"`
	Data    BulkSubscribeContactsToTopicsData `json:"data"`
}

// BulkSubscribeContactsToTopicsData holds the counts from a bulk-subscribe operation.
type BulkSubscribeContactsToTopicsData struct {
	Subscribed        int `json:"subscribed"`
	AlreadySubscribed int `json:"already_subscribed"`
	TotalPairs        int `json:"total_pairs"`
}

// BulkUnsubscribeContactsFromTopicsResponse is the response from a bulk unsubscribe.
type BulkUnsubscribeContactsFromTopicsResponse struct {
	Message string                                `json:"message"`
	Data    BulkUnsubscribeContactsFromTopicsData `json:"data"`
}

// BulkUnsubscribeContactsFromTopicsData holds the counts from a bulk-unsubscribe
// operation. Pairs that did not exist are ignored, so Unsubscribed can be lower
// than TotalPairs.
type BulkUnsubscribeContactsFromTopicsData struct {
	Unsubscribed int `json:"unsubscribed"`
	TotalPairs   int `json:"total_pairs"`
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
//
// An email already in the team's audience comes back as a 409 with error code
// "resource_already_exists" — use IsContactAlreadyExists to detect it. That is
// a client-correctable condition, not an outage: do not retry it. Update the
// existing contact with Update, or use BulkCreate with UpdateExisting set.
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
//
// Rows that fail validation are skipped, not fatal: the call still returns HTTP
// 201 and reports them in Data.Errors. A nil error does not mean every row
// landed — check Data.HasErrors.
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

// BulkSubscribeToTopics subscribes every combination of contact_ids × topic_ids
// (up to 1000 contacts × 50 topics).
//
// Pass BulkCreateAudienceContactsData.ContactIDs from a bulk create — no ID
// lookup needed.
func (s *AudienceContactService) BulkSubscribeToTopics(ctx context.Context, params *BulkContactsTopicsRequest) (*BulkSubscribeContactsToTopicsResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "audience/contacts/topics/bulk", params)
	if err != nil {
		return nil, err
	}

	var resp BulkSubscribeContactsToTopicsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BulkUnsubscribeFromTopics unsubscribes every combination of
// contact_ids × topic_ids. Pairs that do not exist are ignored.
//
// This is a DELETE carrying a request body, as BulkDetachFromLists already is.
func (s *AudienceContactService) BulkUnsubscribeFromTopics(ctx context.Context, params *BulkContactsTopicsRequest) (*BulkUnsubscribeContactsFromTopicsResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodDelete, "audience/contacts/topics/bulk", params)
	if err != nil {
		return nil, err
	}

	var resp BulkUnsubscribeContactsFromTopicsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
