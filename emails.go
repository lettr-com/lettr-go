package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

	// Subject is the email subject line (required).
	Subject string `json:"subject"`

	// Html is the HTML body content. At least one of Html or Text is required.
	Html string `json:"html,omitempty"`

	// Text is the plain text body content. At least one of Html or Text is required.
	Text string `json:"text,omitempty"`

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
	EventID              string                 `json:"event_id"`
	Type                 string                 `json:"type,omitempty"`
	Timestamp            string                 `json:"timestamp"`
	RequestID            string                 `json:"request_id"`
	MessageID            string                 `json:"message_id"`
	Subject              string                 `json:"subject"`
	FriendlyFrom         string                 `json:"friendly_from"`
	SendingDomain        string                 `json:"sending_domain"`
	RcptTo               string                 `json:"rcpt_to"`
	RawRcptTo            string                 `json:"raw_rcpt_to"`
	RecipientDomain      string                 `json:"recipient_domain"`
	MailboxProvider      string                 `json:"mailbox_provider"`
	MailboxProviderRegion string                `json:"mailbox_provider_region"`
	SendingIP            string                 `json:"sending_ip"`
	ClickTracking        bool                   `json:"click_tracking"`
	OpenTracking         bool                   `json:"open_tracking"`
	Transactional        bool                   `json:"transactional"`
	MsgSize              int                    `json:"msg_size"`
	InjectionTime        string                 `json:"injection_time"`
	Reason               *string                `json:"reason"`
	RawReason            *string                `json:"raw_reason"`
	ErrorCode            *string                `json:"error_code"`
	RcptMeta             map[string]interface{} `json:"rcpt_meta"`
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

// ListEmailsData contains the paginated list of email events.
type ListEmailsData struct {
	Results    []EmailEvent       `json:"results"`
	TotalCount int                `json:"total_count"`
	Pagination CursorPagination   `json:"pagination"`
}

// CursorPagination holds cursor-based pagination info.
type CursorPagination struct {
	NextCursor *string `json:"next_cursor"`
	PerPage    int     `json:"per_page"`
}

// GetEmailResponse is the response from getting email details.
type GetEmailResponse struct {
	Message string       `json:"message"`
	Data    GetEmailData `json:"data"`
}

// GetEmailData contains the events for a specific email.
type GetEmailData struct {
	Results    []EmailEvent `json:"results"`
	TotalCount int          `json:"total_count"`
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

// Get retrieves all events for a specific email by its request ID
// (the transmission ID returned when sending).
//
// Example:
//
//	details, err := client.Emails.Get(ctx, "12345678901234567890")
func (s *EmailService) Get(ctx context.Context, requestID string) (*GetEmailResponse, error) {
	path := fmt.Sprintf("emails/%s", url.PathEscape(requestID))

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
