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
	SubstitutionData map[string]interface{} `json:"substitution_data,omitempty"`

	// Metadata contains custom key-value pairs stored with the email.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

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
}

// EmailEvent represents a single event in an email's lifecycle
// (injection, delivery, bounce, open, click, etc).
type EmailEvent struct {
	EventID               string                 `json:"event_id"`
	Type                  string                 `json:"type,omitempty"`
	Timestamp             string                 `json:"timestamp"`
	RequestID             string                 `json:"request_id"`
	MessageID             string                 `json:"message_id"`
	Subject               string                 `json:"subject"`
	FriendlyFrom          string                 `json:"friendly_from"`
	SendingDomain         string                 `json:"sending_domain"`
	RcptTo                string                 `json:"rcpt_to"`
	RawRcptTo             string                 `json:"raw_rcpt_to"`
	RecipientDomain       string                 `json:"recipient_domain"`
	MailboxProvider       string                 `json:"mailbox_provider"`
	MailboxProviderRegion string                 `json:"mailbox_provider_region"`
	SendingIP             string                 `json:"sending_ip"`
	ClickTracking         bool                   `json:"click_tracking"`
	OpenTracking          bool                   `json:"open_tracking"`
	Transactional         bool                   `json:"transactional"`
	MsgSize               int                    `json:"msg_size"`
	InjectionTime         string                 `json:"injection_time"`
	Reason                *string                `json:"reason"`
	RawReason             *string                `json:"raw_reason"`
	ErrorCode             *string                `json:"error_code"`
	BounceClass           *int                   `json:"bounce_class,omitempty"`
	// RcptMeta is polymorphic per spec: an object (in /emails list items)
	// or an array (in event-stream payloads like /emails/events), or null.
	// Type-assert to map[string]interface{} or []interface{} as appropriate.
	RcptMeta              interface{}            `json:"rcpt_meta"`
	TemplateID            *string                `json:"template_id,omitempty"`
	TemplateVersion       *string                `json:"template_version,omitempty"`
	DelvMethod            *string                `json:"delv_method,omitempty"`
	RecvMethod            *string                `json:"recv_method,omitempty"`
	RoutingDomain         *string                `json:"routing_domain,omitempty"`
	ScheduledTime         *string                `json:"scheduled_time,omitempty"`
	CampaignID            *string                `json:"campaign_id,omitempty"`
	AbTestID              *string                `json:"ab_test_id,omitempty"`
	AbTestVersion         *string                `json:"ab_test_version,omitempty"`
	AmpEnabled            *bool                  `json:"amp_enabled,omitempty"`
	RcptType              *string                `json:"rcpt_type,omitempty"`
	RcptTags              []string               `json:"rcpt_tags,omitempty"`
	IpPool                *string                `json:"ip_pool,omitempty"`
	DnsProvider           *string                `json:"dns_provider,omitempty"`
	TargetLinkURL         *string                `json:"target_link_url,omitempty"`
	TargetLinkName        *string                `json:"target_link_name,omitempty"`
	UserAgent             *string                `json:"user_agent,omitempty"`
	UserAgentParsed       *UserAgentParsed       `json:"user_agent_parsed,omitempty"`
	GeoIp                 *GeoIp                 `json:"geo_ip,omitempty"`
	IpAddress             *string                `json:"ip_address,omitempty"`
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
	Success bool           `json:"success"`
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
// The data shape matches ShowScheduledTransmissionResponse — transmission
// metadata plus the full list of delivery events.
type GetEmailResponse struct {
	Message string                `json:"message"`
	Data    ScheduledTransmission `json:"data"`
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
func (s *EmailService) Send(ctx context.Context, params *SendEmailRequest) (*SendEmailResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "emails", params)
	if err != nil {
		return nil, err
	}

	var resp SendEmailResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
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
	Recipients string

	// Transmissions filters by transmission IDs.
	Transmissions []string

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
		if params.Recipients != "" {
			q.Set("recipients", params.Recipients)
		}
		if len(params.Transmissions) > 0 {
			q.Set("transmissions", strings.Join(params.Transmissions, ","))
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
	// Must be at least 5 minutes in the future and at most 3 days.
	ScheduledAt string `json:"scheduled_at"`
}

// ScheduleEmailResponse is the response from scheduling an email.
type ScheduleEmailResponse struct {
	Message string            `json:"message"`
	Data    ScheduleEmailData `json:"data"`
}

// ScheduleEmailData contains the result of scheduling an email.
type ScheduleEmailData struct {
	TransmissionID string `json:"transmission_id"`
}

// GetScheduledEmailResponse is the response from getting a scheduled email.
type GetScheduledEmailResponse struct {
	Message string                 `json:"message"`
	Data    ScheduledTransmission  `json:"data"`
}

// ScheduledTransmission represents a scheduled email transmission.
type ScheduledTransmission struct {
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

// Schedule queues an email for future delivery.
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

// GetScheduled retrieves details of a scheduled email transmission.
//
// Example:
//
//	scheduled, err := client.Emails.GetScheduled(ctx, "transmission-123")
func (s *EmailService) GetScheduled(ctx context.Context, transmissionID string) (*GetScheduledEmailResponse, error) {
	path := fmt.Sprintf("emails/scheduled/%s", url.PathEscape(transmissionID))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetScheduledEmailResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CancelScheduled cancels a pending scheduled email transmission.
//
// Example:
//
//	err := client.Emails.CancelScheduled(ctx, "transmission-123")
func (s *EmailService) CancelScheduled(ctx context.Context, transmissionID string) error {
	path := fmt.Sprintf("emails/scheduled/%s", url.PathEscape(transmissionID))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
