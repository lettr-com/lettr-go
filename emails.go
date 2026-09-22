package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// EmailService handles communication with the email-related endpoints
// of the Lettr API.
type EmailService struct {
	client *Client
}

// SendEmailRequest represents the request body for sending an email.
type SendEmailRequest struct {
	// From is the sender email address (required).
	From string `json:"from"`

	// FromName is the sender display name (optional).
	FromName string `json:"from_name,omitempty"`

	// To is the list of recipient email addresses (required, max 50).
	To []string `json:"to"`

	// Cc is the list of carbon copy recipient email addresses (optional).
	Cc []string `json:"cc,omitempty"`

	// Bcc is the list of blind carbon copy recipient email addresses (optional).
	Bcc []string `json:"bcc,omitempty"`

	// Subject is the email subject line (required unless using template_slug).
	Subject string `json:"subject,omitempty"`

	// Html is the HTML body content. At least one of Html or Text is required.
	Html string `json:"html,omitempty"`

	// Text is the plain text body content. At least one of Html or Text is required.
	Text string `json:"text,omitempty"`

	// AmpHtml is the AMP HTML content for supported email clients (optional).
	AmpHtml string `json:"amp_html,omitempty"`

	// ReplyTo is the reply-to email address (optional).
	ReplyTo string `json:"reply_to,omitempty"`

	// ReplyToName is the reply-to display name (optional).
	ReplyToName string `json:"reply_to_name,omitempty"`

	// TemplateSlug is the slug of a pre-defined template to use.
	TemplateSlug string `json:"template_slug,omitempty"`

	// TemplateVersion is the specific version of the template to use.
	TemplateVersion *int `json:"template_version,omitempty"`

	// ProjectID is the project to source the template from.
	ProjectID *int `json:"project_id,omitempty"`

	// Attachments is a list of file attachments (base64-encoded).
	Attachments []Attachment `json:"attachments,omitempty"`

	// SubstitutionData contains key-value pairs for template variable replacement.
	SubstitutionData map[string]string `json:"substitution_data,omitempty"`

	// Metadata contains custom key-value pairs stored with the email.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Tag is a tag for tracking and analytics (optional).
	Tag string `json:"tag,omitempty"`

	// Headers contains custom email headers (up to 10, optional).
	Headers map[string]string `json:"headers,omitempty"`

	// Options contains tracking and delivery options.
	Options *SendEmailOptions `json:"options,omitempty"`
}

// SendEmailOptions contains optional send settings.
type SendEmailOptions struct {
	// ClickTracking enables or disables click tracking.
	ClickTracking *bool `json:"click_tracking,omitempty"`

	// OpenTracking enables or disables open tracking.
	OpenTracking *bool `json:"open_tracking,omitempty"`

	// Transactional marks the email as transactional.
	Transactional *bool `json:"transactional,omitempty"`

	// InlineCss enables inlining CSS styles in HTML content.
	InlineCss *bool `json:"inline_css,omitempty"`

	// PerformSubstitutions enables variable substitutions in content.
	PerformSubstitutions *bool `json:"perform_substitutions,omitempty"`
}

// Attachment represents a file attachment on an email.
type Attachment struct {
	// Name is the filename of the attachment.
	Name string `json:"name"`

	// Type is the MIME type of the attachment (e.g. "application/pdf").
	Type string `json:"type"`

	// Data is the base64-encoded content of the attachment.
	Data string `json:"data"`
}

// SendEmailResponse is the response from sending an email.
type SendEmailResponse struct {
	Message string        `json:"message"`
	Data    SendEmailData `json:"data"`
}

// SendEmailData contains the result of a send operation.
type SendEmailData struct {
	// RequestID is the unique transmission ID for the sent email.
	RequestID string `json:"request_id"`

	// Accepted is the number of recipients that were accepted.
	Accepted int `json:"accepted"`

	// Rejected is the number of recipients that were rejected.
	Rejected int `json:"rejected"`

	// Replayed is true when this response replayed an earlier send under the
	// same idempotency key - no second email went out. It is still a success.
	//
	// Read from the Idempotency-Replayed response header rather than the body,
	// so it is only ever set when WithIdempotencyKey was used.
	Replayed bool `json:"-"`
}

// EmailEvent represents a single event in an email's lifecycle
// (injection, delivery, bounce, open, click, etc).
type EmailEvent struct {
	EventID               string  `json:"event_id"`
	Type                  string  `json:"type,omitempty"`
	Timestamp             string  `json:"timestamp"`
	RequestID             *string `json:"request_id"`
	MessageID             *string `json:"message_id"`
	Subject               *string `json:"subject"`
	FriendlyFrom          *string `json:"friendly_from"`
	SendingDomain         *string `json:"sending_domain"`
	RcptTo                *string `json:"rcpt_to"`
	RawRcptTo             *string `json:"raw_rcpt_to"`
	RecipientDomain       *string `json:"recipient_domain"`
	MailboxProvider       *string `json:"mailbox_provider"`
	MailboxProviderRegion *string `json:"mailbox_provider_region"`
	SendingIP             *string `json:"sending_ip"`
	ClickTracking         *bool   `json:"click_tracking"`
	OpenTracking          *bool   `json:"open_tracking"`
	Transactional         *bool   `json:"transactional"`
	MsgSize               *int    `json:"msg_size"`
	InjectionTime         *string `json:"injection_time"`
	Reason                *string `json:"reason"`
	RawReason             *string `json:"raw_reason"`
	ErrorCode             *string `json:"error_code"`
	BounceClass           *int    `json:"bounce_class,omitempty"`
	// RcptMeta is polymorphic per spec: an object (in /emails list items)
	// or an array (in event-stream payloads like /emails/events), or null.
	// Type-assert to map[string]interface{} or []interface{} as appropriate.
	RcptMeta        interface{}      `json:"rcpt_meta"`
	TemplateID      *string          `json:"template_id,omitempty"`
	TemplateVersion *string          `json:"template_version,omitempty"`
	DelvMethod      *string          `json:"delv_method,omitempty"`
	RecvMethod      *string          `json:"recv_method,omitempty"`
	RoutingDomain   *string          `json:"routing_domain,omitempty"`
	ScheduledTime   *string          `json:"scheduled_time,omitempty"`
	CampaignID      *string          `json:"campaign_id,omitempty"`
	AbTestID        *string          `json:"ab_test_id,omitempty"`
	AbTestVersion   *string          `json:"ab_test_version,omitempty"`
	AmpEnabled      *bool            `json:"amp_enabled,omitempty"`
	RcptType        *string          `json:"rcpt_type,omitempty"`
	RcptTags        []string         `json:"rcpt_tags,omitempty"`
	IpPool          *string          `json:"ip_pool,omitempty"`
	MsgFrom         *string          `json:"msg_from,omitempty"`
	QueueTime       *int             `json:"queue_time,omitempty"`
	OutboundTls     *string          `json:"outbound_tls,omitempty"`
	InitialPixel    *bool            `json:"initial_pixel,omitempty"`
	NumRetries      *int             `json:"num_retries,omitempty"`
	DeviceToken     *string          `json:"device_token,omitempty"`
	TargetLinkURL   *string          `json:"target_link_url,omitempty"`
	TargetLinkName  *string          `json:"target_link_name,omitempty"`
	UserAgent       *string          `json:"user_agent,omitempty"`
	UserAgentParsed *UserAgentParsed `json:"user_agent_parsed,omitempty"`
	GeoIp           *GeoIp           `json:"geo_ip,omitempty"`
	IpAddress       *string          `json:"ip_address,omitempty"`
}

// UserAgentParsed contains parsed user agent information from open/click events.
type UserAgentParsed struct {
	AgentFamily  string `json:"agent_family,omitempty"`
	OsFamily     string `json:"os_family,omitempty"`
	OsVersion    string `json:"os_version,omitempty"`
	DeviceFamily string `json:"device_family,omitempty"`
	DeviceBrand  string `json:"device_brand,omitempty"`
	IsMobile     bool   `json:"is_mobile,omitempty"`
	IsProxy      bool   `json:"is_proxy,omitempty"`
	IsPrefetched bool   `json:"is_prefetched,omitempty"`
}

// GeoIp contains geolocation data derived from IP address of open/click events.
type GeoIp struct {
	Country    string  `json:"country,omitempty"`
	Region     string  `json:"region,omitempty"`
	City       string  `json:"city,omitempty"`
	PostalCode string  `json:"postal_code,omitempty"`
	Zip        string  `json:"zip,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

// ListEmailsParams contains the query parameters for listing emails.
type ListEmailsParams struct {
	// PerPage is the number of results per page (1-100, default 25).
	PerPage int

	// Cursor is the pagination cursor from a previous response.
	Cursor string

	// Recipients filters by recipient email address.
	Recipients string

	// From filters emails sent on or after this date (ISO 8601, e.g. "2024-01-15").
	From string

	// To filters emails sent on or before this date (ISO 8601, e.g. "2024-01-31").
	To string
}

// ListEmailsResponse is the response from listing emails.
type ListEmailsResponse struct {
	Message string         `json:"message"`
	Data    ListEmailsData `json:"data"`
}

// ListEmailsData wraps the paginated email events returned by the API.
type ListEmailsData struct {
	Events ListEmailsEvents `json:"events"`
}

// ListEmailsEvents contains the paginated list of email events plus
// the query date range echoed back by the API.
type ListEmailsEvents struct {
	Data       []EmailEvent     `json:"data"`
	TotalCount int              `json:"total_count"`
	From       string           `json:"from"`
	To         string           `json:"to"`
	Pagination CursorPagination `json:"pagination"`
}

// CursorPagination holds cursor-based pagination info.
type CursorPagination struct {
	NextCursor *string `json:"next_cursor"`
	PerPage    int     `json:"per_page"`
}

// GetEmailResponse is the response from getting email details.
type GetEmailResponse struct {
	Message string      `json:"message"`
	Data    EmailDetail `json:"data"`
}

// EmailDetail is an already-sent email, reconstructed from its delivery events.
//
// State here is derived from the events that arrived ("delivered", "bounced",
// "failed"), which is a different vocabulary from ScheduledEmail.State: that
// one is Lettr's own lifecycle and is known before anything is delivered.
type EmailDetail struct {
	// TransmissionID is the provider's id, the same value that appears on
	// webhook events for this email.
	TransmissionID string       `json:"transmission_id"`
	State          string       `json:"state"`
	ScheduledAt    *string      `json:"scheduled_at"`
	From           string       `json:"from"`
	FromName       *string      `json:"from_name"`
	Subject        string       `json:"subject"`
	Recipients     []string     `json:"recipients"`
	NumRecipients  int          `json:"num_recipients"`
	Events         []EmailEvent `json:"events"`
}

// Send sends an email with the given parameters.
//
// Example:
//
//	resp, err := client.Emails.Send(ctx, &lettr.SendEmailRequest{
//	    From:    "sender@example.com",
//	    To:      []string{"recipient@example.com"},
//	    Subject: "Hello from Lettr",
//	    Html:    "<h1>Hello!</h1>",
//	})
//
// Pass WithIdempotencyKey to make a retry safe:
//
//	resp, err := client.Emails.Send(ctx, params,
//	    lettr.WithIdempotencyKey("order-confirmation-12345"))
//
//	resp.Data.Replayed // true → this replayed an earlier send
//
// The options are variadic, so existing two-argument calls are unaffected.
func (s *EmailService) Send(ctx context.Context, params *SendEmailRequest, opts ...SendOption) (*SendEmailResponse, error) {
	var cfg sendConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	// Checked here so a malformed key fails locally instead of costing a round
	// trip and a 422.
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, "emails", params)
	if err != nil {
		return nil, err
	}

	if cfg.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", cfg.idempotencyKey)
	}

	var resp SendEmailResponse
	httpResp, err := s.client.do(req, &resp)
	if err != nil {
		return nil, err
	}

	if httpResp != nil {
		resp.Data.Replayed = strings.EqualFold(httpResp.Header.Get("Idempotency-Replayed"), "true")
	}

	return &resp, nil
}

// List retrieves a paginated list of sent emails.
//
// Pass nil for params to use defaults.
//
// Example:
//
//	emails, err := client.Emails.List(ctx, &lettr.ListEmailsParams{
//	    PerPage: 10,
//	})
func (s *EmailService) List(ctx context.Context, params *ListEmailsParams) (*ListEmailsResponse, error) {
	path := "emails"
	if params != nil {
		q := url.Values{}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Recipients != "" {
			q.Set("recipients", params.Recipients)
		}
		if params.From != "" {
			q.Set("from", params.From)
		}
		if params.To != "" {
			q.Set("to", params.To)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListEmailsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEmailParams contains optional query parameters for getting email details.
type GetEmailParams struct {
	// From is the start date for event search range (ISO 8601). Defaults to 10 days ago.
	From string

	// To is the end date for event search range (ISO 8601). Defaults to now.
	To string
}

// Get retrieves all events for a specific email by its request ID
// (the transmission ID returned when sending).
//
// Example:
//
//	details, err := client.Emails.Get(ctx, "12345678901234567890", nil)
func (s *EmailService) Get(ctx context.Context, requestID string, params *GetEmailParams) (*GetEmailResponse, error) {
	path := fmt.Sprintf("emails/%s", url.PathEscape(requestID))
	if params != nil {
		q := url.Values{}
		if params.From != "" {
			q.Set("from", params.From)
		}
		if params.To != "" {
			q.Set("to", params.To)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetEmailResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListEmailEventsParams contains the query parameters for listing email events.
type ListEmailEventsParams struct {
	// Events filters by event types (e.g. "delivery", "bounce", "open", "click").
	Events []string

	// Recipients filters by recipient email addresses.
	Recipients []string

	// Transmissions filters by transmission ID.
	Transmissions string

	// BounceClasses filters by bounce classification codes.
	BounceClasses []int

	// From is the start date for events (ISO 8601). Defaults to 10 days ago.
	From string

	// To is the end date for events (ISO 8601). Defaults to now.
	To string

	// PerPage is the number of events per page.
	PerPage int

	// Cursor is the pagination cursor from a previous response.
	Cursor string
}

// ListEmailEventsResponse is the response from listing email events.
type ListEmailEventsResponse struct {
	Message string              `json:"message"`
	Data    ListEmailEventsData `json:"data"`
}

// ListEmailEventsData wraps the paginated email events returned by the API.
type ListEmailEventsData struct {
	Events ListEmailEventsEvents `json:"events"`
}

// ListEmailEventsEvents contains the paginated list of email events plus
// the query date range echoed back by the API.
type ListEmailEventsEvents struct {
	Data       []EmailEvent     `json:"data"`
	TotalCount int              `json:"total_count"`
	From       string           `json:"from"`
	To         string           `json:"to"`
	Pagination CursorPagination `json:"pagination"`
}

// ListEvents retrieves email delivery events (opens, bounces, clicks, etc.)
// with optional filtering.
//
// Pass nil for params to use defaults.
//
// Example:
//
//	events, err := client.Emails.ListEvents(ctx, &lettr.ListEmailEventsParams{
//	    Events:  []string{"delivery", "bounce"},
//	    PerPage: 50,
//	})
func (s *EmailService) ListEvents(ctx context.Context, params *ListEmailEventsParams) (*ListEmailEventsResponse, error) {
	path := "emails/events"
	if params != nil {
		q := url.Values{}
		if len(params.Events) > 0 {
			q.Set("events", strings.Join(params.Events, ","))
		}
		if len(params.Recipients) > 0 {
			q.Set("recipients", strings.Join(params.Recipients, ","))
		}
		if params.Transmissions != "" {
			q.Set("transmissions", params.Transmissions)
		}
		if len(params.BounceClasses) > 0 {
			bc := make([]string, len(params.BounceClasses))
			for i, v := range params.BounceClasses {
				bc[i] = strconv.Itoa(v)
			}
			q.Set("bounce_classes", strings.Join(bc, ","))
		}
		if params.From != "" {
			q.Set("from", params.From)
		}
		if params.To != "" {
			q.Set("to", params.To)
		}
		if params.PerPage > 0 {
			q.Set("per_page", strconv.Itoa(params.PerPage))
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

	var resp ListEmailEventsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ScheduleEmailRequest represents the request body for scheduling an email
// for future delivery.
type ScheduleEmailRequest struct {
	SendEmailRequest

	// ScheduledAt is the time to send the email (ISO 8601).
	// Must be at least 5 minutes in the future and at most 30 days.
	ScheduledAt string `json:"scheduled_at"`
}

// ScheduledEmailState is the lifecycle of a scheduled email.
//
// Lettr holds the email in its own store until it is due, so these states are
// Lettr's and are authoritative from the moment of scheduling. The provider
// never sees a future-dated message, which is why a cancelled email can be
// told apart from one that never existed.
type ScheduledEmailState string

const (
	// ScheduledStateScheduled is waiting for its delivery time. The only state
	// CancelScheduled accepts.
	ScheduledStateScheduled ScheduledEmailState = "scheduled"

	// ScheduledStateSending is being handed to the provider right now.
	ScheduledStateSending ScheduledEmailState = "sending"

	// ScheduledStateSent has been handed over. TransmissionID is set from here
	// on, and delivery detail comes from the events API.
	ScheduledStateSent ScheduledEmailState = "sent"

	// ScheduledStateCancelled was cancelled before hand-off, so nothing was sent.
	ScheduledStateCancelled ScheduledEmailState = "cancelled"

	// ScheduledStateFailed gave up trying to hand the email over. FailureReason
	// says why.
	ScheduledStateFailed ScheduledEmailState = "failed"

	// The states below are the provider's, not Lettr's. GetScheduled reports
	// them only on the legacy read path, which answers from delivery events in
	// the provider's own vocabulary - never for a sch_ id.
	//
	// Deprecated: legacy provider state, from a numeric transmission id.
	ScheduledStateSubmitted ScheduledEmailState = "submitted"
	// Deprecated: legacy provider state, from a numeric transmission id.
	ScheduledStateGenerating ScheduledEmailState = "generating"
	// Deprecated: legacy provider state, from a numeric transmission id.
	ScheduledStateDelivered ScheduledEmailState = "delivered"
	// Deprecated: legacy provider state, from a numeric transmission id.
	ScheduledStateBounced ScheduledEmailState = "bounced"
)

// IsCancellable reports whether CancelScheduled would still be honoured.
//
// Once the email is with the provider there is no per-message recall, so
// cancelling anything past "scheduled" returns a 409 rather than stopping it.
func (s ScheduledEmailState) IsCancellable() bool {
	return s == ScheduledStateScheduled
}

// ScheduledEmail is an email Lettr is holding until its delivery time.
//
// It carries two ids, and they are not interchangeable:
//
//   - RequestID ("sch_…") is Lettr's own id. It is what GetScheduled and
//     CancelScheduled take.
//   - TransmissionID is the provider's id, nil until the email is actually
//     handed over. It is the value webhook events carry, so it is what
//     correlates this email with the webhooks it produces.
type ScheduledEmail struct {
	RequestID      string              `json:"request_id"`
	TransmissionID *string             `json:"transmission_id"`
	State          ScheduledEmailState `json:"state"`

	// ScheduledAt is the delivery time (ISO 8601). It is nil only on the
	// legacy read path described on GetScheduled.
	ScheduledAt *string `json:"scheduled_at"`

	From          string   `json:"from"`
	FromName      *string  `json:"from_name"`
	Subject       *string  `json:"subject"`
	Recipients    []string `json:"recipients"`
	NumRecipients int      `json:"num_recipients"`

	// Accepted and Rejected are the provider's counts once the email has been
	// sent, and before that describe what Lettr took on for delivery — so a
	// cancelled email reads back as 0 accepted, not 1.
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`

	Tag *string `json:"tag"`

	// FailureReason is set only in ScheduledStateFailed.
	FailureReason *string `json:"failure_reason"`

	// Events are the delivery events, which only exist once the email has been
	// handed over.
	Events []EmailEvent `json:"events"`
}

// ScheduledTransmission is the former name of ScheduledEmail.
//
// Deprecated: renamed to ScheduledEmail. If you used it for Emails.Get rather
// than the scheduled endpoints, that response now has its own type, EmailDetail.
type ScheduledTransmission = ScheduledEmail

// ScheduleEmailResponse is the response from scheduling an email.
type ScheduleEmailResponse struct {
	Message string         `json:"message"`
	Data    ScheduledEmail `json:"data"`
}

// GetScheduledEmailResponse is the response from getting a scheduled email.
type GetScheduledEmailResponse struct {
	Message string         `json:"message"`
	Data    ScheduledEmail `json:"data"`
}

// CancelScheduledResponse is the response from cancelling a scheduled email.
type CancelScheduledResponse struct {
	Message string         `json:"message"`
	Data    ScheduledEmail `json:"data"`
}

// ListScheduledEmailsParams contains the query parameters for listing
// scheduled emails.
type ListScheduledEmailsParams struct {
	// Status narrows the list to one state. All states are returned if not set.
	Status ScheduledEmailState

	// PerPage is the number of results per page (1-100, default 25).
	PerPage int

	// Page is the page number (default 1).
	Page int
}

// ListScheduledEmailsResponse is the response from listing scheduled emails.
type ListScheduledEmailsResponse struct {
	Message string                  `json:"message"`
	Data    ListScheduledEmailsData `json:"data"`
}

// ListScheduledEmailsData contains the paginated list of scheduled emails.
type ListScheduledEmailsData struct {
	ScheduledEmails []ScheduledEmail `json:"scheduled_emails"`
	Pagination      PagePagination   `json:"pagination"`
}

// Schedule queues an email for future delivery.
//
// The delivery time must be at least 5 minutes and at most 30 days out.
// The response is the scheduled email itself, so there is no need to read it
// back to learn its state.
//
// Example:
//
//	resp, err := client.Emails.Schedule(ctx, &lettr.ScheduleEmailRequest{
//	    SendEmailRequest: lettr.SendEmailRequest{
//	        From:    "sender@example.com",
//	        To:      []string{"recipient@example.com"},
//	        Subject: "Scheduled Hello",
//	        Html:    "<h1>Hello!</h1>",
//	    },
//	    ScheduledAt: "2024-12-25T10:00:00Z",
//	})
//
//	resp.Data.RequestID // "sch_…" — pass this to GetScheduled and CancelScheduled
func (s *EmailService) Schedule(ctx context.Context, params *ScheduleEmailRequest) (*ScheduleEmailResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "emails/scheduled", params)
	if err != nil {
		return nil, err
	}

	var resp ScheduleEmailResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListScheduled retrieves a paginated list of scheduled emails.
//
// Pass nil for params to use defaults.
//
// Example:
//
//	resp, err := client.Emails.ListScheduled(ctx, &lettr.ListScheduledEmailsParams{
//	    Status: lettr.ScheduledStateScheduled,
//	})
func (s *EmailService) ListScheduled(ctx context.Context, params *ListScheduledEmailsParams) (*ListScheduledEmailsResponse, error) {
	path := "emails/scheduled"
	if params != nil {
		q := url.Values{}
		if params.Status != "" {
			q.Set("status", string(params.Status))
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

	var resp ListScheduledEmailsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetScheduled retrieves a scheduled email by its RequestID.
//
// A provider transmission id stored before Lettr owned the schedule still
// resolves, answered from delivery events. That older shape carries no
// request_id, so RequestID is filled in from the id you asked about and always
// holds the id that addresses this email.
//
// Example:
//
//	scheduled, err := client.Emails.GetScheduled(ctx, "sch_01M322YMWVCZ4RNYXHMSSMDTM1")
func (s *EmailService) GetScheduled(ctx context.Context, id string) (*GetScheduledEmailResponse, error) {
	path := fmt.Sprintf("emails/scheduled/%s", url.PathEscape(id))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetScheduledEmailResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}

	if resp.Data.RequestID == "" {
		resp.Data.RequestID = id
	}
	return &resp, nil
}

// CancelScheduled cancels a scheduled email before it is sent, and returns it
// in its cancelled state.
//
// Only an email still in ScheduledStateScheduled can be cancelled; past that
// it is with the provider, which offers no per-message recall, and the call
// fails with a 409.
//
// Example:
//
//	resp, err := client.Emails.CancelScheduled(ctx, "sch_01M322YMWVCZ4RNYXHMSSMDTM1")
//	resp.Data.State // lettr.ScheduledStateCancelled
func (s *EmailService) CancelScheduled(ctx context.Context, id string) (*CancelScheduledResponse, error) {
	path := fmt.Sprintf("emails/scheduled/%s", url.PathEscape(id))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	var resp CancelScheduledResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
