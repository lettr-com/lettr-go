package lettr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// CampaignService handles communication with the campaign endpoints of the
// Lettr API.
type CampaignService struct {
	client *Client
}

// Campaign status values.
const (
	CampaignStatusDraft     = "draft"
	CampaignStatusScheduled = "scheduled"
	CampaignStatusPreparing = "preparing"
	CampaignStatusInReview  = "in_review"
	CampaignStatusSending   = "sending"
	CampaignStatusSent      = "sent"
	CampaignStatusFailed    = "failed"
)

// Campaign engagement event types.
const (
	CampaignEventTypeInjection       = "injection"
	CampaignEventTypeDelivery        = "delivery"
	CampaignEventTypeBounce          = "bounce"
	CampaignEventTypeSpamComplaint   = "spam_complaint"
	CampaignEventTypeOpen            = "open"
	CampaignEventTypeClick           = "click"
	CampaignEventTypeListUnsubscribe = "list_unsubscribe"
)

// CampaignStats holds aggregated engagement statistics for a campaign.
type CampaignStats struct {
	Injections     int `json:"injections"`
	Deliveries     int `json:"deliveries"`
	Bounces        int `json:"bounces"`
	SpamComplaints int `json:"spam_complaints"`
	Opens          int `json:"opens"`
	UniqueOpens    int `json:"unique_opens"`
	Clicks         int `json:"clicks"`
	UniqueClicks   int `json:"unique_clicks"`
	Unsubscribes   int `json:"unsubscribes"`
}

// Campaign is a campaign summary with embedded engagement stats. Nullable
// fields are represented as pointers.
type Campaign struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Subject         *string       `json:"subject"`
	FromEmail       *string       `json:"from_email"`
	FromName        *string       `json:"from_name"`
	ReplyTo         *string       `json:"reply_to"`
	Status          string        `json:"status"`
	ScheduledAt     *string       `json:"scheduled_at"`
	TotalRecipients *int          `json:"total_recipients"`
	SentCount       int           `json:"sent_count"`
	SentAt          *string       `json:"sent_at"`
	CreatedAt       string        `json:"created_at"`
	Stats           CampaignStats `json:"stats"`
}

// CampaignDetail is the full campaign representation, extending Campaign with
// the rendered HTML content of the campaign email.
type CampaignDetail struct {
	Campaign
	HTMLContent *string `json:"html_content"`
}

// CampaignEvent is a single campaign engagement event. Fields that only apply
// to certain event types are represented as pointers.
type CampaignEvent struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Email     string `json:"email"`
	Timestamp string `json:"timestamp"`
	// BounceClass is the SparkPost bounce classification code, serialised as
	// a string by the campaigns API (note this differs from EmailEvent's *int
	// BounceClass — see emails.go — which is the same underlying code but
	// kept as the type the /emails/events endpoint emits).
	BounceClass   *string `json:"bounce_class"`
	Reason        *string `json:"reason"`
	TargetLinkURL *string `json:"target_link_url"`
	UserAgent     *string `json:"user_agent"`
}

// ListCampaignsParams contains the query parameters for listing campaigns.
type ListCampaignsParams struct {
	// Page is the page number (default 1).
	Page int

	// PerPage is the number of results per page (1-100, default 20).
	PerPage int

	// Status filters by campaign status (e.g. CampaignStatusSent). Optional.
	Status string
}

// ListCampaignsResponse is the response from listing campaigns.
type ListCampaignsResponse struct {
	Message string            `json:"message"`
	Data    ListCampaignsData `json:"data"`
}

// ListCampaignsData contains the paginated list of campaigns.
type ListCampaignsData struct {
	Campaigns  []Campaign     `json:"campaigns"`
	Pagination PagePagination `json:"pagination"`
}

// CampaignResponse is the response from Get, returning a single campaign with
// its rendered HTML content.
type CampaignResponse struct {
	Message string         `json:"message"`
	Data    CampaignDetail `json:"data"`
}

// ListCampaignEventsParams contains the query parameters for listing campaign
// engagement events.
type ListCampaignEventsParams struct {
	// EventType filters by event type (e.g. CampaignEventTypeOpen). Optional.
	EventType string

	// Email filters events to a single recipient address. Optional.
	Email string

	// StartDate is the start of the date range (ISO 8601). A date-only value
	// is treated as the start of that day in UTC. Optional.
	StartDate string

	// EndDate is the end of the date range (ISO 8601). A date-only value
	// covers the whole day in UTC. Optional.
	EndDate string

	// Limit is the number of events per page (1-100, default 25).
	Limit int

	// Cursor is the pagination cursor returned as NextCursor by a prior call.
	Cursor string
}

// ListCampaignEventsResponse is the response from listing campaign events.
type ListCampaignEventsResponse struct {
	Message string                 `json:"message"`
	Data    ListCampaignEventsData `json:"data"`
}

// ListCampaignEventsData contains a page of campaign events. An empty Events
// slice together with a non-nil NextCursor means more pages exist; NextCursor
// is nil when there are no more events.
//
// Note: this endpoint does not nest its cursor under a "pagination" envelope
// (and does not echo per_page) the way /emails/events does, so it cannot
// directly reuse CursorPagination — the wire shape genuinely differs.
type ListCampaignEventsData struct {
	Events     []CampaignEvent `json:"events"`
	NextCursor *string         `json:"next_cursor"`
}

// CampaignActionResponse is the response from the send, schedule, and
// unschedule actions. Data carries the updated campaign on success; it is nil
// in the rare case the campaign cannot be re-read immediately after the action
// (e.g. it was concurrently deleted), in which case the server omits the key.
type CampaignActionResponse struct {
	Message string    `json:"message"`
	Data    *Campaign `json:"data"`
}

// ScheduleCampaignRequest is the body for scheduling a campaign.
type ScheduleCampaignRequest struct {
	// ScheduledAt is the future delivery time (ISO 8601). Include a timezone
	// offset (e.g. "+02:00" or "Z"); a value without an offset is interpreted
	// as UTC. Must be in the future.
	ScheduledAt string `json:"scheduled_at"`
}

// List retrieves a paginated list of campaigns with embedded engagement stats.
//
// Pass nil for params to use defaults.
func (s *CampaignService) List(ctx context.Context, params *ListCampaignsParams) (*ListCampaignsResponse, error) {
	path := "campaigns"
	if params != nil {
		q := url.Values{}
		if params.Page > 0 {
			q.Set("page", strconv.Itoa(params.Page))
		}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListCampaignsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single campaign by ID, including its rendered HTML content.
func (s *CampaignService) Get(ctx context.Context, campaignID string) (*CampaignResponse, error) {
	path := fmt.Sprintf("campaigns/%s", url.PathEscape(campaignID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp CampaignResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListEvents retrieves a page of engagement events (opens, clicks, bounces,
// etc.) for a campaign using cursor-based pagination.
//
// Pass nil for params to use defaults.
func (s *CampaignService) ListEvents(ctx context.Context, campaignID string, params *ListCampaignEventsParams) (*ListCampaignEventsResponse, error) {
	path := fmt.Sprintf("campaigns/%s/events", url.PathEscape(campaignID))
	if params != nil {
		q := url.Values{}
		if params.EventType != "" {
			q.Set("event_type", params.EventType)
		}
		if params.Email != "" {
			q.Set("email", params.Email)
		}
		if params.StartDate != "" {
			q.Set("start_date", params.StartDate)
		}
		if params.EndDate != "" {
			q.Set("end_date", params.EndDate)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListCampaignEventsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Send immediately dispatches a draft campaign. The campaign must have a
// subject, sender email, and content. Sending is asynchronous; the campaign
// transitions to the "preparing" status. Not available to sandbox keys.
func (s *CampaignService) Send(ctx context.Context, campaignID string) (*CampaignActionResponse, error) {
	path := fmt.Sprintf("campaigns/%s/send", url.PathEscape(campaignID))

	req, err := s.client.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var resp CampaignActionResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Schedule schedules a campaign for future delivery, or reschedules one that is
// already scheduled. The campaign is dispatched automatically at the given
// time. Not available to sandbox keys.
func (s *CampaignService) Schedule(ctx context.Context, campaignID string, params *ScheduleCampaignRequest) (*CampaignActionResponse, error) {
	if params == nil {
		return nil, errors.New("lettr: Schedule requires a non-nil *ScheduleCampaignRequest")
	}

	path := fmt.Sprintf("campaigns/%s/schedule", url.PathEscape(campaignID))

	req, err := s.client.newRequest(ctx, http.MethodPost, path, params)
	if err != nil {
		return nil, err
	}

	var resp CampaignActionResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Unschedule cancels a scheduled send and returns the campaign to the "draft"
// status. The campaign must currently be scheduled. Not available to sandbox
// keys.
func (s *CampaignService) Unschedule(ctx context.Context, campaignID string) (*CampaignActionResponse, error) {
	path := fmt.Sprintf("campaigns/%s/unschedule", url.PathEscape(campaignID))

	req, err := s.client.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var resp CampaignActionResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
