package lettr

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

const testCampaignID = "0193e6c0-1111-7d4f-a8f1-2222cccc3333"

func TestListCampaigns(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if pg := r.URL.Query().Get("page"); pg != "2" {
			t.Errorf("expected page=2, got %q", pg)
		}
		if pp := r.URL.Query().Get("per_page"); pp != "10" {
			t.Errorf("expected per_page=10, got %q", pp)
		}
		if st := r.URL.Query().Get("status"); st != CampaignStatusSent {
			t.Errorf("expected status=sent, got %q", st)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListCampaignsResponse{
			Message: "Campaigns retrieved successfully.",
			Data: ListCampaignsData{
				Campaigns: []Campaign{{
					ID:        testCampaignID,
					Name:      "Spring Sale",
					Subject:   strPtr("Spring Sale"),
					Status:    CampaignStatusSent,
					SentCount: 124,
					CreatedAt: "2026-05-01T09:00:00+00:00",
					Stats:     CampaignStats{Deliveries: 120, Opens: 80, UniqueOpens: 60},
				}},
				Pagination: PagePagination{Total: 42, PerPage: 10, CurrentPage: 2, LastPage: 5},
			},
		})
	})
	defer server.Close()

	resp, err := client.Campaigns.List(context.Background(), &ListCampaignsParams{Page: 2, PerPage: 10, Status: CampaignStatusSent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(resp.Data.Campaigns))
	}
	c := resp.Data.Campaigns[0]
	if c.Subject == nil || *c.Subject != "Spring Sale" {
		t.Errorf("expected subject %q, got %v", "Spring Sale", c.Subject)
	}
	if c.SentCount != 124 {
		t.Errorf("expected sent_count 124, got %d", c.SentCount)
	}
	if c.Stats.UniqueOpens != 60 {
		t.Errorf("expected unique_opens 60, got %d", c.Stats.UniqueOpens)
	}
	if resp.Data.Pagination.LastPage != 5 {
		t.Errorf("expected last_page 5, got %d", resp.Data.Pagination.LastPage)
	}
}

func TestListCampaignsNilParams(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListCampaignsResponse{Message: "ok"})
	})
	defer server.Close()

	if _, err := client.Campaigns.List(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCampaign(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns/"+testCampaignID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CampaignResponse{
			Message: "Campaign retrieved successfully.",
			Data: CampaignDetail{
				Campaign: Campaign{
					ID:        testCampaignID,
					Name:      "Spring Sale",
					Status:    CampaignStatusSent,
					SentCount: 124,
					CreatedAt: "2026-05-01T09:00:00+00:00",
					Stats:     CampaignStats{Deliveries: 120, Clicks: 30},
				},
				HTMLContent: strPtr("<h1>Hello</h1>"),
			},
		})
	})
	defer server.Close()

	resp, err := client.Campaigns.Get(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ID != testCampaignID {
		t.Errorf("expected id %q, got %q", testCampaignID, resp.Data.ID)
	}
	if resp.Data.HTMLContent == nil || *resp.Data.HTMLContent != "<h1>Hello</h1>" {
		t.Errorf("expected html_content, got %v", resp.Data.HTMLContent)
	}
	if resp.Data.Subject != nil {
		t.Errorf("expected nil subject, got %v", *resp.Data.Subject)
	}
	if resp.Data.Stats.Clicks != 30 {
		t.Errorf("expected clicks 30, got %d", resp.Data.Stats.Clicks)
	}
}

func TestListCampaignEvents(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns/"+testCampaignID+"/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		q := r.URL.Query()
		if got := q.Get("event_type"); got != CampaignEventTypeOpen {
			t.Errorf("expected event_type=open, got %q", got)
		}
		if got := q.Get("email"); got != "jane@example.com" {
			t.Errorf("expected email=jane@example.com, got %q", got)
		}
		if got := q.Get("start_date"); got != "2026-05-01T00:00:00Z" {
			t.Errorf("expected start_date, got %q", got)
		}
		if got := q.Get("end_date"); got != "2026-05-31T23:59:59Z" {
			t.Errorf("expected end_date, got %q", got)
		}
		if got := q.Get("limit"); got != "50" {
			t.Errorf("expected limit=50, got %q", got)
		}
		if got := q.Get("cursor"); got != "abc" {
			t.Errorf("expected cursor=abc, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListCampaignEventsResponse{
			Message: "Campaign events retrieved successfully.",
			Data: ListCampaignEventsData{
				Events: []CampaignEvent{{
					EventID:   "92356829",
					EventType: CampaignEventTypeOpen,
					Email:     "jane@example.com",
					Timestamp: "2026-05-01T12:30:00+00:00",
					UserAgent: strPtr("Mozilla/5.0"),
				}},
				NextCursor: strPtr("def"),
			},
		})
	})
	defer server.Close()

	resp, err := client.Campaigns.ListEvents(context.Background(), testCampaignID, &ListCampaignEventsParams{
		EventType: CampaignEventTypeOpen,
		Email:     "jane@example.com",
		StartDate: "2026-05-01T00:00:00Z",
		EndDate:   "2026-05-31T23:59:59Z",
		Limit:     50,
		Cursor:    "abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Data.Events))
	}
	if resp.Data.Events[0].EventType != CampaignEventTypeOpen {
		t.Errorf("expected event_type open, got %q", resp.Data.Events[0].EventType)
	}
	if resp.Data.NextCursor == nil || *resp.Data.NextCursor != "def" {
		t.Errorf("expected next_cursor def, got %v", resp.Data.NextCursor)
	}
}

func TestSendCampaign(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns/"+testCampaignID+"/send" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(CampaignActionResponse{
			Message: "Campaign queued for delivery.",
			Data:    &Campaign{ID: testCampaignID, Name: "Spring Sale", Status: CampaignStatusPreparing},
		})
	})
	defer server.Close()

	resp, err := client.Campaigns.Send(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil || resp.Data.Status != CampaignStatusPreparing {
		t.Errorf("expected status preparing, got %v", resp.Data)
	}
}

func TestSendCampaignValidationError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "The campaign is not in a sendable state.",
			"error_code": "campaign_not_sendable",
		})
	})
	defer server.Close()

	_, err := client.Campaigns.Send(context.Background(), testCampaignID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *lettr.Error, got %T", err)
	}
	if apiErr.ErrorCode != "campaign_not_sendable" {
		t.Errorf("expected error_code campaign_not_sendable, got %q", apiErr.ErrorCode)
	}
}

func TestScheduleCampaign(t *testing.T) {
	const scheduledAt = "2026-06-01T09:00:00+00:00"
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns/"+testCampaignID+"/schedule" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body ScheduleCampaignRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.ScheduledAt != scheduledAt {
			t.Errorf("expected scheduled_at %q, got %q", scheduledAt, body.ScheduledAt)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CampaignActionResponse{
			Message: "Campaign scheduled for delivery.",
			Data:    &Campaign{ID: testCampaignID, Status: CampaignStatusScheduled, ScheduledAt: strPtr(scheduledAt)},
		})
	})
	defer server.Close()

	resp, err := client.Campaigns.Schedule(context.Background(), testCampaignID, &ScheduleCampaignRequest{ScheduledAt: scheduledAt})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil || resp.Data.Status != CampaignStatusScheduled {
		t.Errorf("expected status scheduled, got %v", resp.Data)
	}
}

func TestScheduleCampaignNilParamsRejected(t *testing.T) {
	// The server must NOT be called; we don't even hand newTestClient a real handler.
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server was unexpectedly called: %s %s", r.Method, r.URL.Path)
	})
	defer server.Close()

	if _, err := client.Campaigns.Schedule(context.Background(), testCampaignID, nil); err == nil {
		t.Fatal("expected error for nil params, got nil")
	}
}

func TestUnscheduleCampaign(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/campaigns/"+testCampaignID+"/unschedule" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CampaignActionResponse{
			Message: "Campaign returned to draft.",
			Data:    &Campaign{ID: testCampaignID, Status: CampaignStatusDraft},
		})
	})
	defer server.Close()

	resp, err := client.Campaigns.Unschedule(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil || resp.Data.Status != CampaignStatusDraft {
		t.Errorf("expected status draft, got %v", resp.Data)
	}
}
