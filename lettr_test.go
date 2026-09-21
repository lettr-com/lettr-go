package lettr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func strPtr(s string) *string { return &s }

// newTestClient creates a Client pointing to a test server.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := NewClient("test-api-key")
	if err := client.SetBaseURL(server.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return client, server
}

func TestNewClient(t *testing.T) {
	client := NewClient("my-api-key")
	if client.apiKey != "my-api-key" {
		t.Errorf("expected api key %q, got %q", "my-api-key", client.apiKey)
	}
	if client.Emails == nil {
		t.Error("expected Emails service to be initialized")
	}
	if client.Domains == nil {
		t.Error("expected Domains service to be initialized")
	}
	if client.Webhooks == nil {
		t.Error("expected Webhooks service to be initialized")
	}
	if client.Templates == nil {
		t.Error("expected Templates service to be initialized")
	}
	if client.Projects == nil {
		t.Error("expected Projects service to be initialized")
	}
	if client.Audience == nil {
		t.Error("expected Audience service to be initialized")
	}
	if client.Campaigns == nil {
		t.Error("expected Campaigns service to be initialized")
	}
}

func TestHealthCheck(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Health check should not have an Authorization header.
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("expected no Authorization header, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthCheckResponse{
			Message: "Health check passed.",
			Data:    HealthCheckData{Status: "ok", Timestamp: "2024-01-15T10:30:00.000Z"},
		})
	})
	defer server.Close()

	resp, err := client.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Status != "ok" {
		t.Errorf("expected status %q, got %q", "ok", resp.Data.Status)
	}
}

func TestValidateAPIKey(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/check" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-api-key" {
			t.Errorf("unexpected Authorization header: %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthCheckResponse{
			Message: "API key is valid.",
			Data:    AuthCheckData{TeamID: 123, Timestamp: "2024-01-15T10:30:00.000Z"},
		})
	})
	defer server.Close()

	resp, err := client.ValidateAPIKey(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.TeamID != 123 {
		t.Errorf("expected team ID 123, got %d", resp.Data.TeamID)
	}
}

func TestSendEmail(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body SendEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.From != "sender@example.com" {
			t.Errorf("expected from %q, got %q", "sender@example.com", body.From)
		}
		if len(body.To) != 1 || body.To[0] != "recipient@example.com" {
			t.Errorf("unexpected to: %v", body.To)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SendEmailResponse{
			Message: "Email queued for delivery.",
			Data: SendEmailData{
				RequestID: "req-123",
				Accepted:  1,
				Rejected:  0,
			},
		})
	})
	defer server.Close()

	resp, err := client.Emails.Send(context.Background(), &SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Hello",
		Html:    "<h1>Hello!</h1>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.RequestID != "req-123" {
		t.Errorf("expected request ID %q, got %q", "req-123", resp.Data.RequestID)
	}
	if resp.Data.Accepted != 1 {
		t.Errorf("expected 1 accepted, got %d", resp.Data.Accepted)
	}
}

func TestListEmails(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if pp := r.URL.Query().Get("per_page"); pp != "10" {
			t.Errorf("expected per_page=10, got %q", pp)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListEmailsResponse{
			Message: "Emails retrieved successfully.",
			Data: ListEmailsData{
				Events: ListEmailsEvents{
					Data:       []EmailEvent{{EventID: "evt-1", Subject: strPtr("Test")}},
					TotalCount: 1,
					Pagination: CursorPagination{PerPage: 10},
				},
			},
		})
	})
	defer server.Close()

	resp, err := client.Emails.List(context.Background(), &ListEmailsParams{PerPage: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Events.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", resp.Data.Events.TotalCount)
	}
	if len(resp.Data.Events.Data) != 1 || resp.Data.Events.Data[0].EventID != "evt-1" {
		t.Errorf("expected 1 event with id evt-1, got %+v", resp.Data.Events.Data)
	}
}

func TestGetEmail(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/req-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetEmailResponse{
			Message: "Email retrieved successfully.",
			Data: EmailDetail{
				TransmissionID: "req-123",
				State:          "delivered",
				From:           "sender@example.com",
				Subject:        "Hello",
				Recipients:     []string{"recipient@example.com"},
				NumRecipients:  1,
				Events:         []EmailEvent{{EventID: "evt-1", Type: "delivery"}},
			},
		})
	})
	defer server.Close()

	resp, err := client.Emails.Get(context.Background(), "req-123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.TransmissionID != "req-123" {
		t.Errorf("expected transmission id %q, got %q", "req-123", resp.Data.TransmissionID)
	}
	if len(resp.Data.Events) != 1 || resp.Data.Events[0].Type != "delivery" {
		t.Errorf("expected 1 delivery event, got %+v", resp.Data.Events)
	}
}

func TestListDomains(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListDomainsResponse{
			Message: "Domains retrieved successfully.",
			Data: ListDomainsData{
				Domains: []Domain{{Domain: "example.com", Status: "approved", CanSend: true}},
			},
		})
	})
	defer server.Close()

	resp, err := client.Domains.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(resp.Data.Domains))
	}
	if resp.Data.Domains[0].Domain != "example.com" {
		t.Errorf("expected domain %q, got %q", "example.com", resp.Data.Domains[0].Domain)
	}
}

func TestCreateDomain(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateDomainResponse{
			Message: "Domain created successfully.",
			Data: CreateDomainData{
				Domain:      "example.com",
				Status:      "pending",
				StatusLabel: "Pending Review",
			},
		})
	})
	defer server.Close()

	resp, err := client.Domains.Create(context.Background(), &CreateDomainRequest{
		Domain: "example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Status != "pending" {
		t.Errorf("expected status %q, got %q", "pending", resp.Data.Status)
	}
}

func TestDeleteDomain(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains/example.com" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.Domains.Delete(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListWebhooks(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListWebhooksResponse{
			Message: "Webhooks retrieved successfully.",
			Data: ListWebhooksData{
				Webhooks: []Webhook{{ID: "wh-1", Name: "Test", Enabled: true}},
			},
		})
	})
	defer server.Close()

	resp, err := client.Webhooks.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Webhooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(resp.Data.Webhooks))
	}
}

func TestListTemplates(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListTemplatesResponse{
			Message: "Templates retrieved successfully.",
			Data: ListTemplatesData{
				Templates:  []Template{{ID: 1, Name: "Welcome", Slug: "welcome"}},
				Pagination: PagePagination{Total: 1, PerPage: 25, CurrentPage: 1, LastPage: 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Templates.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(resp.Data.Templates))
	}
	if resp.Data.Templates[0].Slug != "welcome" {
		t.Errorf("expected slug %q, got %q", "welcome", resp.Data.Templates[0].Slug)
	}
}

func TestErrorHandling(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Error{
			Message:   "Validation failed.",
			ErrorCode: "validation_error",
			Errors: map[string][]string{
				"from": {"The sender email address is required."},
			},
		})
	})
	defer server.Close()

	_, err := client.Emails.Send(context.Background(), &SendEmailRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsValidationError(err) {
		t.Errorf("expected validation error, got: %v", err)
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.ErrorCode != "validation_error" {
		t.Errorf("expected error code %q, got %q", "validation_error", apiErr.ErrorCode)
	}
	if msgs, exists := apiErr.Errors["from"]; !exists || len(msgs) == 0 {
		t.Error("expected 'from' field error")
	}
}

func TestUnauthorizedError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Error{
			Message:   "Invalid API key.",
			ErrorCode: "unauthorized",
		})
	})
	defer server.Close()

	_, err := client.ValidateAPIKey(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got: %v", err)
	}
}

func TestNotFoundError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{
			Message:   "Email not found.",
			ErrorCode: "not_found",
		})
	})
	defer server.Close()

	_, err := client.Emails.Get(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestUserAgentHeader(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "lettr-go/"+Version {
			t.Errorf("expected User-Agent %q, got %q", "lettr-go/"+Version, ua)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthCheckResponse{
			Message: "Health check passed.",
			Data:    HealthCheckData{Status: "ok"},
		})
	})
	defer server.Close()

	client.HealthCheck(context.Background())
}

func TestSetBaseURL(t *testing.T) {
	client := NewClient("key")
	err := client.SetBaseURL("https://custom.example.com/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL.String() != "https://custom.example.com/api/" {
		t.Errorf("expected base URL %q, got %q", "https://custom.example.com/api/", client.baseURL.String())
	}
}

func TestSendEmailWithCcBcc(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body SendEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if len(body.Cc) != 1 || body.Cc[0] != "cc@example.com" {
			t.Errorf("unexpected cc: %v", body.Cc)
		}
		if len(body.Bcc) != 1 || body.Bcc[0] != "bcc@example.com" {
			t.Errorf("unexpected bcc: %v", body.Bcc)
		}
		if body.ReplyTo != "reply@example.com" {
			t.Errorf("unexpected reply_to: %s", body.ReplyTo)
		}
		if body.Tag != "welcome" {
			t.Errorf("unexpected tag: %s", body.Tag)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SendEmailResponse{
			Message: "Email queued.",
			Data:    SendEmailData{RequestID: "req-cc", Accepted: 3, Rejected: 0},
		})
	})
	defer server.Close()

	resp, err := client.Emails.Send(context.Background(), &SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"bcc@example.com"},
		Subject: "Hello",
		Html:    "<h1>Hello!</h1>",
		ReplyTo: "reply@example.com",
		Tag:     "welcome",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Accepted != 3 {
		t.Errorf("expected 3 accepted, got %d", resp.Data.Accepted)
	}
}

func TestListEmailEvents(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if ev := r.URL.Query().Get("events"); ev != "delivery,bounce" {
			t.Errorf("expected events=delivery,bounce, got %q", ev)
		}
		if pp := r.URL.Query().Get("per_page"); pp != "10" {
			t.Errorf("expected per_page=10, got %q", pp)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListEmailEventsResponse{
			Message: "Events retrieved.",
			Data: ListEmailEventsData{
				Events: ListEmailEventsEvents{
					Data:       []EmailEvent{{EventID: "evt-1", Type: "delivery"}},
					TotalCount: 1,
					Pagination: CursorPagination{PerPage: 10},
				},
			},
		})
	})
	defer server.Close()

	resp, err := client.Emails.ListEvents(context.Background(), &ListEmailEventsParams{
		Events:  []string{"delivery", "bounce"},
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Events.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", resp.Data.Events.TotalCount)
	}
	if resp.Data.Events.Data[0].Type != "delivery" {
		t.Errorf("expected type %q, got %q", "delivery", resp.Data.Events.Data[0].Type)
	}
}

// scheduledEmailJSON is the wire shape of a scheduled email, written out by
// hand rather than encoded from the struct so these tests would catch a
// renamed or dropped JSON tag.
const scheduledEmailJSON = `{
	"request_id": "sch_01M322YMWVCZ4RNYXHMSSMDTM1",
	"transmission_id": null,
	"state": "scheduled",
	"scheduled_at": "2026-09-21T15:37:10Z",
	"from": "hello@dev.uselettr.com",
	"from_name": null,
	"subject": "sdk audit probe",
	"recipients": ["vojta@ecomail.cz"],
	"num_recipients": 1,
	"accepted": 1,
	"rejected": 0,
	"tag": null,
	"failure_reason": null,
	"events": []
}`

func TestScheduleEmail(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/scheduled" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["scheduled_at"] != "2026-12-25T10:00:00Z" {
			t.Errorf("unexpected scheduled_at: %v", body["scheduled_at"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"Email scheduled for delivery.","data":` + scheduledEmailJSON + `}`))
	})
	defer server.Close()

	resp, err := client.Emails.Schedule(context.Background(), &ScheduleEmailRequest{
		SendEmailRequest: SendEmailRequest{
			From:    "sender@example.com",
			To:      []string{"recipient@example.com"},
			Subject: "Scheduled",
			Html:    "<h1>Hello!</h1>",
		},
		ScheduledAt: "2026-12-25T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.RequestID != "sch_01M322YMWVCZ4RNYXHMSSMDTM1" {
		t.Errorf("expected request ID %q, got %q", "sch_01M322YMWVCZ4RNYXHMSSMDTM1", resp.Data.RequestID)
	}
	if resp.Data.State != ScheduledStateScheduled {
		t.Errorf("expected state %q, got %q", ScheduledStateScheduled, resp.Data.State)
	}
	if resp.Data.Accepted != 1 {
		t.Errorf("expected 1 accepted, got %d", resp.Data.Accepted)
	}
	if resp.Data.Subject == nil || *resp.Data.Subject != "sdk audit probe" {
		t.Errorf("expected subject %q, got %v", "sdk audit probe", resp.Data.Subject)
	}
}

func TestGetScheduledEmail(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/scheduled/sch_01M322YMWVCZ4RNYXHMSSMDTM1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Scheduled transmission retrieved successfully.","data":` + scheduledEmailJSON + `}`))
	})
	defer server.Close()

	resp, err := client.Emails.GetScheduled(context.Background(), "sch_01M322YMWVCZ4RNYXHMSSMDTM1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.State != ScheduledStateScheduled {
		t.Errorf("expected state %q, got %q", ScheduledStateScheduled, resp.Data.State)
	}
	if resp.Data.NumRecipients != 1 {
		t.Errorf("expected num_recipients 1, got %d", resp.Data.NumRecipients)
	}
	// Nothing has been handed to the provider yet, so there is no id to
	// correlate webhooks with.
	if resp.Data.TransmissionID != nil {
		t.Errorf("expected nil transmission id while scheduled, got %q", *resp.Data.TransmissionID)
	}
	if resp.Data.ScheduledAt == nil || *resp.Data.ScheduledAt != "2026-09-21T15:37:10Z" {
		t.Errorf("expected scheduled_at to decode, got %v", resp.Data.ScheduledAt)
	}
	if resp.Data.Tag != nil || resp.Data.FailureReason != nil || resp.Data.FromName != nil {
		t.Errorf("expected the null fields to decode as nil, got %+v", resp.Data)
	}
}

// Once the email is sent the provider's id appears, and that - not the
// request id - is what webhook events carry.
func TestGetScheduledEmailSentCarriesTransmissionID(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Scheduled transmission retrieved successfully.","data":{
			"request_id": "sch_01M322YMWVCZ4RNYXHMSSMDTM1",
			"transmission_id": "112233445566778899",
			"state": "sent",
			"scheduled_at": "2026-09-21T15:37:10Z",
			"from": "hello@dev.uselettr.com",
			"from_name": null,
			"subject": "sdk audit probe",
			"recipients": ["vojta@ecomail.cz"],
			"num_recipients": 1,
			"accepted": 1,
			"rejected": 0,
			"tag": "welcome",
			"failure_reason": null,
			"events": [{"event_id":"evt-1","type":"delivery"}]
		}}`))
	})
	defer server.Close()

	resp, err := client.Emails.GetScheduled(context.Background(), "sch_01M322YMWVCZ4RNYXHMSSMDTM1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.TransmissionID == nil || *resp.Data.TransmissionID != "112233445566778899" {
		t.Errorf("expected the provider id to decode, got %v", resp.Data.TransmissionID)
	}
	if resp.Data.State != ScheduledStateSent {
		t.Errorf("expected state %q, got %q", ScheduledStateSent, resp.Data.State)
	}
	if resp.Data.Tag == nil || *resp.Data.Tag != "welcome" {
		t.Errorf("expected tag %q, got %v", "welcome", resp.Data.Tag)
	}
	if len(resp.Data.Events) != 1 || resp.Data.Events[0].Type != "delivery" {
		t.Errorf("expected 1 delivery event, got %+v", resp.Data.Events)
	}
}

// An id issued before Lettr owned the schedule is answered from delivery
// events, in a shape that has no request_id at all. Without the fallback the
// caller would get an empty RequestID back for an email they just asked about
// by id.
func TestGetScheduledEmailLegacyShapeFillsRequestID(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/scheduled/112233445566778899" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Scheduled transmission retrieved successfully.","data":{
			"transmission_id": "112233445566778899",
			"state": "delivered",
			"scheduled_at": null,
			"from": "hello@dev.uselettr.com",
			"from_name": null,
			"subject": "legacy probe",
			"recipients": ["vojta@ecomail.cz"],
			"num_recipients": 1,
			"events": [{"event_id":"evt-1","type":"delivery"}]
		}}`))
	})
	defer server.Close()

	resp, err := client.Emails.GetScheduled(context.Background(), "112233445566778899")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.RequestID != "112233445566778899" {
		t.Errorf("expected RequestID to fall back to the id asked for, got %q", resp.Data.RequestID)
	}
	if resp.Data.TransmissionID == nil || *resp.Data.TransmissionID != "112233445566778899" {
		t.Errorf("expected the provider id to decode, got %v", resp.Data.TransmissionID)
	}
	if resp.Data.ScheduledAt != nil {
		t.Errorf("expected nil scheduled_at on the legacy shape, got %q", *resp.Data.ScheduledAt)
	}
	// The legacy shape carries no counts, so these must stay zero rather than
	// being reported as a rejection.
	if resp.Data.Accepted != 0 || resp.Data.Rejected != 0 {
		t.Errorf("expected zero counts on the legacy shape, got %d/%d", resp.Data.Accepted, resp.Data.Rejected)
	}
}

func TestScheduledEmailStates(t *testing.T) {
	states := map[string]ScheduledEmailState{
		"scheduled": ScheduledStateScheduled,
		"sending":   ScheduledStateSending,
		"sent":      ScheduledStateSent,
		"cancelled": ScheduledStateCancelled,
		"failed":    ScheduledStateFailed,
	}

	for raw, want := range states {
		var email ScheduledEmail
		if err := json.Unmarshal([]byte(`{"state":"`+raw+`"}`), &email); err != nil {
			t.Fatalf("unexpected error decoding %q: %v", raw, err)
		}
		if email.State != want {
			t.Errorf("expected state %q, got %q", want, email.State)
		}
		// The type is a defined string, so a caller comparing against a plain
		// literal keeps compiling.
		if email.State != ScheduledEmailState(raw) || string(email.State) != raw {
			t.Errorf("state %q did not round-trip", raw)
		}
		if got := email.State.IsCancellable(); got != (raw == "scheduled") {
			t.Errorf("state %q reported IsCancellable() = %v", raw, got)
		}
	}
}

func TestCancelScheduledEmail(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/scheduled/sch_01M322YMWVCZ4RNYXHMSSMDTM1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Scheduled transmission cancelled successfully.","data":{
			"request_id": "sch_01M322YMWVCZ4RNYXHMSSMDTM1",
			"transmission_id": null,
			"state": "cancelled",
			"scheduled_at": "2026-09-21T15:37:10Z",
			"from": "hello@dev.uselettr.com",
			"from_name": null,
			"subject": "sdk audit probe",
			"recipients": ["vojta@ecomail.cz"],
			"num_recipients": 1,
			"accepted": 0,
			"rejected": 0,
			"tag": null,
			"failure_reason": null,
			"events": []
		}}`))
	})
	defer server.Close()

	resp, err := client.Emails.CancelScheduled(context.Background(), "sch_01M322YMWVCZ4RNYXHMSSMDTM1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "Scheduled transmission cancelled successfully." {
		t.Errorf("unexpected message %q", resp.Message)
	}
	if resp.Data.State != ScheduledStateCancelled {
		t.Errorf("expected state %q, got %q", ScheduledStateCancelled, resp.Data.State)
	}
	// Nothing was handed over, so the email must not read back as accepted.
	if resp.Data.Accepted != 0 {
		t.Errorf("expected 0 accepted after cancelling, got %d", resp.Data.Accepted)
	}
	if resp.Data.State.IsCancellable() {
		t.Error("a cancelled email should not report itself as cancellable")
	}
}

func TestListScheduledEmails(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails/scheduled" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		q := r.URL.Query()
		if q.Get("status") != "scheduled" {
			t.Errorf("expected status %q, got %q", "scheduled", q.Get("status"))
		}
		if q.Get("per_page") != "2" {
			t.Errorf("expected per_page %q, got %q", "2", q.Get("per_page"))
		}
		if q.Get("page") != "1" {
			t.Errorf("expected page %q, got %q", "1", q.Get("page"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Scheduled emails retrieved successfully.","data":{
			"scheduled_emails": [` + scheduledEmailJSON + `],
			"pagination": {"total":2,"per_page":2,"current_page":1,"last_page":1}
		}}`))
	})
	defer server.Close()

	resp, err := client.Emails.ListScheduled(context.Background(), &ListScheduledEmailsParams{
		Status:  ScheduledStateScheduled,
		PerPage: 2,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.ScheduledEmails) != 1 {
		t.Fatalf("expected 1 scheduled email, got %d", len(resp.Data.ScheduledEmails))
	}
	if resp.Data.ScheduledEmails[0].RequestID != "sch_01M322YMWVCZ4RNYXHMSSMDTM1" {
		t.Errorf("unexpected request id %q", resp.Data.ScheduledEmails[0].RequestID)
	}
	if resp.Data.Pagination.Total != 2 || resp.Data.Pagination.LastPage != 1 {
		t.Errorf("unexpected pagination %+v", resp.Data.Pagination)
	}
}

// Nil params must send a bare path, so the API applies its own defaults
// instead of the SDK pinning page 1 of 0 results.
func TestListScheduledEmailsNilParams(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Scheduled emails retrieved successfully.","data":{"scheduled_emails":[],"pagination":{"total":0,"per_page":25,"current_page":1,"last_page":1}}}`))
	})
	defer server.Close()

	resp, err := client.Emails.ListScheduled(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Pagination.PerPage != 25 {
		t.Errorf("expected the API default per_page 25, got %d", resp.Data.Pagination.PerPage)
	}
}

func TestVerifyDomain(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/domains/example.com/verify" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VerifyDomainResponse{
			Message: "Verification completed.",
			Data: DomainVerificationView{
				Domain:      "example.com",
				DkimStatus:  "valid",
				CnameStatus: "valid",
				DmarcStatus: "valid",
				SpfStatus:   "valid",
			},
		})
	})
	defer server.Close()

	resp, err := client.Domains.Verify(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.DkimStatus != "valid" {
		t.Errorf("expected dkim_status %q, got %q", "valid", resp.Data.DkimStatus)
	}
}

func TestCreateWebhook(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body CreateWebhookRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "My Webhook" {
			t.Errorf("expected name %q, got %q", "My Webhook", body.Name)
		}
		if body.EventsMode != "all" {
			t.Errorf("expected events_mode %q, got %q", "all", body.EventsMode)
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateWebhookResponse{
			Message: "Webhook created.",
			Data:    Webhook{ID: "wh-new", Name: "My Webhook", Enabled: true},
		})
	})
	defer server.Close()

	resp, err := client.Webhooks.Create(context.Background(), &CreateWebhookRequest{
		Name:       "My Webhook",
		URL:        "https://example.com/webhook",
		AuthType:   "none",
		EventsMode: "all",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ID != "wh-new" {
		t.Errorf("expected ID %q, got %q", "wh-new", resp.Data.ID)
	}
}

func TestUpdateWebhook(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/wh-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		raw, _ := io.ReadAll(r.Body)
		if !bytes.Contains(raw, []byte(`"url":"https://example.com/new"`)) {
			t.Errorf("expected url field in body, got: %s", raw)
		}
		if bytes.Contains(raw, []byte(`"target"`)) {
			t.Errorf("did not expect target field in body, got: %s", raw)
		}

		var body UpdateWebhookRequest
		json.Unmarshal(raw, &body)
		if body.Name != "Updated" {
			t.Errorf("expected name %q, got %q", "Updated", body.Name)
		}
		if body.URL != "https://example.com/new" {
			t.Errorf("expected URL %q, got %q", "https://example.com/new", body.URL)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UpdateWebhookResponse{
			Message: "Webhook updated.",
			Data:    Webhook{ID: "wh-123", Name: "Updated", Enabled: true},
		})
	})
	defer server.Close()

	resp, err := client.Webhooks.Update(context.Background(), "wh-123", &UpdateWebhookRequest{
		Name: "Updated",
		URL:  "https://example.com/new",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "Updated" {
		t.Errorf("expected name %q, got %q", "Updated", resp.Data.Name)
	}
}

func TestUpdateWebhookTargetDeprecated(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if !bytes.Contains(raw, []byte(`"target":"https://example.com/legacy"`)) {
			t.Errorf("expected legacy target field in body, got: %s", raw)
		}
		if bytes.Contains(raw, []byte(`"url"`)) {
			t.Errorf("did not expect url field when only Target is set, got: %s", raw)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UpdateWebhookResponse{
			Message: "Webhook updated.",
			Data:    Webhook{ID: "wh-123", Name: "Legacy", Enabled: true},
		})
	})
	defer server.Close()

	_, err := client.Webhooks.Update(context.Background(), "wh-123", &UpdateWebhookRequest{
		Target: "https://example.com/legacy",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteWebhook(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/wh-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Webhook deleted."}`))
	})
	defer server.Close()

	resp, err := client.Webhooks.Delete(context.Background(), "wh-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "Webhook deleted." {
		t.Errorf("expected message %q, got %q", "Webhook deleted.", resp.Message)
	}
}

func TestGetTemplate(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates/welcome" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		activeVersion := 2
		json.NewEncoder(w).Encode(GetTemplateResponse{
			Message: "Template retrieved.",
			Data: TemplateDetail{
				ID:            1,
				Name:          "Welcome",
				Slug:          "welcome",
				ActiveVersion: &activeVersion,
				VersionsCount: 2,
			},
		})
	})
	defer server.Close()

	resp, err := client.Templates.Get(context.Background(), "welcome", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ActiveVersion == nil || *resp.Data.ActiveVersion != 2 {
		t.Errorf("expected active version 2, got %v", resp.Data.ActiveVersion)
	}
}

func TestUpdateTemplate(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates/welcome" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body UpdateTemplateRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Html != "<h1>Updated</h1>" {
			t.Errorf("unexpected html: %s", body.Html)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UpdateTemplateResponse{
			Message: "Template updated.",
			Data: UpdateTemplateData{
				ID:            1,
				Name:          "Welcome",
				Slug:          "welcome",
				ActiveVersion: 3,
			},
		})
	})
	defer server.Close()

	resp, err := client.Templates.Update(context.Background(), "welcome", &UpdateTemplateRequest{
		Html: "<h1>Updated</h1>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ActiveVersion != 3 {
		t.Errorf("expected active version 3, got %d", resp.Data.ActiveVersion)
	}
}

func TestDeleteTemplate(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates/welcome" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Template deleted."}`))
	})
	defer server.Close()

	resp, err := client.Templates.Delete(context.Background(), "welcome", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "Template deleted." {
		t.Errorf("expected message %q, got %q", "Template deleted.", resp.Message)
	}
}

func TestGetMergeTags(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates/welcome/merge-tags" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetMergeTagsResponse{
			Message: "Merge tags retrieved.",
			Data: GetMergeTagsData{
				ProjectID:    1,
				TemplateSlug: "welcome",
				Version:      1,
				MergeTags: []MergeTag{
					{Key: "FIRST_NAME", Required: true, Type: "text"},
				},
			},
		})
	})
	defer server.Close()

	resp, err := client.Templates.GetMergeTags(context.Background(), "welcome", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.MergeTags) != 1 {
		t.Fatalf("expected 1 merge tag, got %d", len(resp.Data.MergeTags))
	}
	if resp.Data.MergeTags[0].Key != "FIRST_NAME" {
		t.Errorf("expected key %q, got %q", "FIRST_NAME", resp.Data.MergeTags[0].Key)
	}
}

func TestGetTemplateHtml(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/templates/html" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if pid := r.URL.Query().Get("project_id"); pid != "1" {
			t.Errorf("expected project_id=1, got %q", pid)
		}
		if slug := r.URL.Query().Get("slug"); slug != "welcome" {
			t.Errorf("expected slug=welcome, got %q", slug)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetTemplateHtmlResponse{
			Success: true,
			Data:    GetTemplateHtmlData{Html: "<h1>Hello!</h1>"},
		})
	})
	defer server.Close()

	resp, err := client.Templates.GetHtml(context.Background(), &GetTemplateHtmlParams{
		ProjectID: 1,
		Slug:      "welcome",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Html != "<h1>Hello!</h1>" {
		t.Errorf("expected html %q, got %q", "<h1>Hello!</h1>", resp.Data.Html)
	}
}

func TestListProjects(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListProjectsResponse{
			Message: "Projects retrieved.",
			Data: ListProjectsData{
				Projects:   []Project{{ID: 1, Name: "Default", TeamID: 10}},
				Pagination: PagePagination{Total: 1, PerPage: 25, CurrentPage: 1, LastPage: 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Projects.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(resp.Data.Projects))
	}
	if resp.Data.Projects[0].Name != "Default" {
		t.Errorf("expected name %q, got %q", "Default", resp.Data.Projects[0].Name)
	}
}

func TestEmailEventRcptMetaPolymorphic(t *testing.T) {
	// Per spec: rcpt_meta is object|null for list items and array|null
	// for event-stream payloads. The SDK must decode both shapes.

	// Object form (from GET /emails).
	objJSON := `{"event_id":"e1","rcpt_meta":{"user_id":"42","plan":"pro"}}`
	var ev1 EmailEvent
	if err := json.Unmarshal([]byte(objJSON), &ev1); err != nil {
		t.Fatalf("object form failed to decode: %v", err)
	}
	m, ok := ev1.RcptMeta.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", ev1.RcptMeta)
	}
	if m["user_id"] != "42" {
		t.Errorf("expected user_id=42, got %v", m["user_id"])
	}

	// Array form (from GET /emails/events).
	arrJSON := `{"event_id":"e2","rcpt_meta":[{"user_id":"42"},{"plan":"pro"}]}`
	var ev2 EmailEvent
	if err := json.Unmarshal([]byte(arrJSON), &ev2); err != nil {
		t.Fatalf("array form failed to decode: %v", err)
	}
	arr, ok := ev2.RcptMeta.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", ev2.RcptMeta)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 array items, got %d", len(arr))
	}

	// Null form.
	nullJSON := `{"event_id":"e3","rcpt_meta":null}`
	var ev3 EmailEvent
	if err := json.Unmarshal([]byte(nullJSON), &ev3); err != nil {
		t.Fatalf("null form failed to decode: %v", err)
	}
	if ev3.RcptMeta != nil {
		t.Errorf("expected nil, got %v", ev3.RcptMeta)
	}
}

func TestWebhookNullEventTypes(t *testing.T) {
	data := `{"id":"wh-1","name":"Test","url":"https://example.com","enabled":true,"event_types":null,"auth_type":"none","has_auth_credentials":false}`
	var wh Webhook
	if err := json.Unmarshal([]byte(data), &wh); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if wh.EventTypes != nil {
		t.Error("expected nil EventTypes for null JSON value")
	}

	events := []string{"message.delivery", "message.bounce"}
	data2 := `{"id":"wh-2","name":"Test2","url":"https://example.com","enabled":true,"event_types":["message.delivery","message.bounce"],"auth_type":"none","has_auth_credentials":false}`
	var wh2 Webhook
	if err := json.Unmarshal([]byte(data2), &wh2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if wh2.EventTypes == nil {
		t.Fatal("expected non-nil EventTypes")
	}
	if len(*wh2.EventTypes) != len(events) {
		t.Errorf("expected %d event types, got %d", len(events), len(*wh2.EventTypes))
	}
}
