package lettr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
				Results:    []EmailEvent{{EventID: "evt-1", Subject: "Test"}},
				TotalCount: 1,
				Pagination: CursorPagination{PerPage: 10},
			},
		})
	})
	defer server.Close()

	resp, err := client.Emails.List(context.Background(), &ListEmailsParams{PerPage: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", resp.Data.TotalCount)
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
			Data: GetEmailData{
				Results:    []EmailEvent{{EventID: "evt-1", Type: "delivery"}},
				TotalCount: 1,
			},
		})
	})
	defer server.Close()

	resp, err := client.Emails.Get(context.Background(), "req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Results[0].Type != "delivery" {
		t.Errorf("expected type %q, got %q", "delivery", resp.Data.Results[0].Type)
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

	_, err := client.Emails.Get(context.Background(), "nonexistent")
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
