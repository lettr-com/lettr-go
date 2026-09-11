package lettr

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestTemplatesListSendsFolderAndPurposeFilters(t *testing.T) {
	var gotQuery string

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"templates":  []map[string]any{},
				"pagination": map[string]any{"total": 0, "per_page": 100, "current_page": 1, "last_page": 1},
			},
		})
	})
	defer server.Close()

	_, err := client.Templates.List(context.Background(), &ListTemplatesParams{
		FolderID: 10,
		Purpose:  PurposeCampaign,
		PerPage:  100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"folder_id=10", "purpose=campaign", "per_page=100"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("expected query to contain %q, got %q", want, gotQuery)
		}
	}
}

// An existing caller's request is unchanged.
func TestTemplatesListOmitsTheNewFiltersWhenUnset(t *testing.T) {
	var gotQuery string

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"templates":  []map[string]any{},
				"pagination": map[string]any{"total": 0, "per_page": 25, "current_page": 1, "last_page": 1},
			},
		})
	})
	defer server.Close()

	_, err := client.Templates.List(context.Background(), &ListTemplatesParams{ProjectID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(gotQuery, "folder_id") || strings.Contains(gotQuery, "purpose") {
		t.Errorf("expected no folder_id or purpose in the query, got %q", gotQuery)
	}
}

func TestTemplatesListReportsPreparationStatus(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"templates": []map[string]any{
					{"id": 1, "name": "Ready", "slug": "ready-one", "project_id": 5, "folder_id": 10,
						"purpose": "transactional", "preparation_status": "ready",
						"created_at": "2026-01-15T10:00:00+00:00", "updated_at": "2026-01-20T14:30:00+00:00"},
					{"id": 2, "name": "Working", "slug": "still-working", "project_id": 5, "folder_id": 10,
						"purpose": "campaign", "preparation_status": "pending",
						"created_at": "2026-01-15T10:00:00+00:00", "updated_at": "2026-01-20T14:30:00+00:00"},
					{"id": 3, "name": "Gave up", "slug": "gave-up", "project_id": 5, "folder_id": 10,
						"purpose": "transactional", "preparation_status": "failed",
						"created_at": "2026-01-15T10:00:00+00:00", "updated_at": "2026-01-20T14:30:00+00:00"},
				},
				"pagination": map[string]any{"total": 3, "per_page": 25, "current_page": 1, "last_page": 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Templates.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// One list call is the point: a bulk import reconciles everything here
	// instead of a detail call each, all dragging the full HTML payload.
	want := []TemplatePreparationStatus{PreparationReady, PreparationPending, PreparationFailed}
	for i, status := range want {
		if resp.Data.Templates[i].PreparationStatus != status {
			t.Errorf("template %d: expected %q, got %q", i, status, resp.Data.Templates[i].PreparationStatus)
		}
	}

	if !resp.Data.Templates[0].PreparationStatus.Settled() {
		t.Error("a ready template should be settled")
	}
	if resp.Data.Templates[1].PreparationStatus.Settled() {
		t.Error("a pending template is not settled")
	}
	if !resp.Data.Templates[1].Purpose.IsCampaign() {
		t.Error("expected the second template to be a campaign template")
	}
}

// An API deployment that predates the fields had every template with HTML
// simply usable, so an empty status must count as settled. Treating it as
// pending would make an old API look like a stalled queue.
func TestPreparationStatusEmptyCountsAsSettled(t *testing.T) {
	if !TemplatePreparationStatus("").Settled() {
		t.Error("an empty preparation status must count as settled")
	}
	if TemplatePurpose("").IsCampaign() {
		t.Error("an empty purpose must not read as a campaign template")
	}
}

func TestTemplatesCreateSendsPurposeOnlyWhenSet(t *testing.T) {
	var gotBody map[string]any

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id": 1, "name": "October Newsletter", "slug": "october-newsletter",
				"project_id": 5, "folder_id": 11,
				"purpose": "campaign", "preparation_status": "pending",
				"active_version": 1, "merge_tags": []any{},
				"created_at": "2026-01-15T10:00:00+00:00",
			},
		})
	})
	defer server.Close()

	resp, err := client.Templates.Create(context.Background(), &CreateTemplateRequest{
		Name:    "October Newsletter",
		Json:    "{}",
		Purpose: PurposeCampaign,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["purpose"] != "campaign" {
		t.Errorf("expected purpose in the body, got %v", gotBody["purpose"])
	}
	// A JSON import has no HTML until the background job renders it.
	if resp.Data.PreparationStatus != PreparationPending {
		t.Errorf("expected pending, got %q", resp.Data.PreparationStatus)
	}
}

func TestTemplatesCreateOmitsPurposeWhenUnset(t *testing.T) {
	var gotBody map[string]any

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id": 1, "name": "Welcome", "slug": "welcome", "project_id": 5, "folder_id": 10,
				"active_version": 1, "merge_tags": []any{}, "created_at": "2026-01-15T10:00:00+00:00",
			},
		})
	})
	defer server.Close()

	_, err := client.Templates.Create(context.Background(), &CreateTemplateRequest{
		Name: "Welcome",
		Html: "<p>Hi</p>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, present := gotBody["purpose"]; present {
		t.Error("purpose must be omitted from the body when unset")
	}
}
