package lettr

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestFoldersList(t *testing.T) {
	var gotPath string

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "Folders retrieved successfully.",
			"data": map[string]any{
				"folders": []map[string]any{
					{
						"id":              10,
						"name":            "Emails",
						"project_id":      5,
						"purpose":         "transactional",
						"templates_count": 12,
						"created_at":      "2026-01-15T10:00:00+00:00",
						"updated_at":      "2026-01-20T14:30:00+00:00",
					},
					{
						"id":              11,
						"name":            "Campaigns",
						"project_id":      5,
						"purpose":         "campaign",
						"templates_count": 3,
						"created_at":      "2026-01-15T10:00:00+00:00",
						"updated_at":      "2026-01-20T14:30:00+00:00",
					},
				},
				"pagination": map[string]any{
					"total": 2, "per_page": 25, "current_page": 1, "last_page": 1,
				},
			},
		})
	})
	defer server.Close()

	resp, err := client.Folders.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/folders" {
		t.Errorf("expected path /folders, got %q", gotPath)
	}
	if len(resp.Data.Folders) != 2 {
		t.Fatalf("expected 2 folders, got %d", len(resp.Data.Folders))
	}

	// The id is the whole point: it is what CreateTemplateRequest.FolderID
	// wants, and nothing else in the SDK returns one.
	if resp.Data.Folders[0].ID != 10 {
		t.Errorf("expected folder id 10, got %d", resp.Data.Folders[0].ID)
	}
	if !resp.Data.Folders[1].Purpose.IsCampaign() {
		t.Errorf("expected the second folder to be a campaign folder, got %q", resp.Data.Folders[1].Purpose)
	}
	if resp.Data.Folders[0].TemplatesCount != 12 {
		t.Errorf("expected templates_count 12, got %d", resp.Data.Folders[0].TemplatesCount)
	}
}

func TestFoldersListSendsEveryFilter(t *testing.T) {
	var gotQuery string

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"folders":    []map[string]any{},
				"pagination": map[string]any{"total": 0, "per_page": 50, "current_page": 2, "last_page": 2},
			},
		})
	})
	defer server.Close()

	_, err := client.Folders.List(context.Background(), &ListFoldersParams{
		ProjectID: 5,
		Purpose:   PurposeCampaign,
		PerPage:   50,
		Page:      2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"project_id=5", "purpose=campaign", "per_page=50", "page=2"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("expected query to contain %q, got %q", want, gotQuery)
		}
	}
}

func TestFolderPurposeDefaultsWhenAbsent(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// An API deployment that predates the field.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"folders": []map[string]any{
					{
						"id": 10, "name": "Emails", "project_id": 5,
						"created_at": "2026-01-15T10:00:00+00:00",
						"updated_at": "2026-01-20T14:30:00+00:00",
					},
				},
				"pagination": map[string]any{"total": 1, "per_page": 25, "current_page": 1, "last_page": 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Folders.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Data.Folders[0].Purpose.IsCampaign() {
		t.Error("a folder without a purpose must not read as a campaign folder")
	}
	if resp.Data.Folders[0].TemplatesCount != 0 {
		t.Errorf("expected a zero templates count, got %d", resp.Data.Folders[0].TemplatesCount)
	}
}
