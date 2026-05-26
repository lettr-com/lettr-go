package lettr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const testListID = "0193e6a8-1f3a-7c2a-b9e2-1aa1d2e5d3f0"
const testContactID = "0193e6b0-9c1d-7d4f-a8f1-cef9a1b2d3e4"
const testTopicID = "0193e6b1-aaaa-7d4f-a8f1-bbbbbbbbbbbb"
const testPropertyID = "0193e6b2-cccc-7d4f-a8f1-dddddddddddd"
const testSegmentID = "0193e6b3-eeee-7d4f-a8f1-ffffffffffff"

func TestNewClientHasAudienceService(t *testing.T) {
	client := NewClient("key")
	if client.Audience == nil {
		t.Fatal("expected Audience service to be initialized")
	}
	if client.Audience.Lists == nil {
		t.Error("expected Audience.Lists to be initialized")
	}
	if client.Audience.Contacts == nil {
		t.Error("expected Audience.Contacts to be initialized")
	}
	if client.Audience.Topics == nil {
		t.Error("expected Audience.Topics to be initialized")
	}
	if client.Audience.Properties == nil {
		t.Error("expected Audience.Properties to be initialized")
	}
	if client.Audience.Segments == nil {
		t.Error("expected Audience.Segments to be initialized")
	}
}

// --- Lists ---

func TestListAudienceLists(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/lists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if pp := r.URL.Query().Get("per_page"); pp != "10" {
			t.Errorf("expected per_page=10, got %q", pp)
		}
		if pg := r.URL.Query().Get("page"); pg != "2" {
			t.Errorf("expected page=2, got %q", pg)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListAudienceListsResponse{
			Message: "Lists retrieved successfully.",
			Data: ListAudienceListsData{
				Lists:      []AudienceList{{ID: testListID, Name: "Newsletter", ContactsCount: 42}},
				Pagination: PagePagination{Total: 42, PerPage: 10, CurrentPage: 2, LastPage: 5},
			},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Lists.List(context.Background(), &ListAudienceListsParams{PerPage: 10, Page: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Lists) != 1 {
		t.Fatalf("expected 1 list, got %d", len(resp.Data.Lists))
	}
	if resp.Data.Lists[0].ContactsCount != 42 {
		t.Errorf("expected contacts_count 42, got %d", resp.Data.Lists[0].ContactsCount)
	}
	if resp.Data.Pagination.LastPage != 5 {
		t.Errorf("expected last_page 5, got %d", resp.Data.Pagination.LastPage)
	}
}

func TestListAudienceListsNilParams(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListAudienceListsResponse{Message: "ok"})
	})
	defer server.Close()

	_, err := client.Audience.Lists.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateAudienceList(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/lists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body CreateAudienceListRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Newsletter" {
			t.Errorf("expected name %q, got %q", "Newsletter", body.Name)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceListResponse{
			Message: "List created successfully.",
			Data:    AudienceList{ID: testListID, Name: "Newsletter", ContactsCount: 0},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Lists.Create(context.Background(), &CreateAudienceListRequest{Name: "Newsletter"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ID != testListID {
		t.Errorf("expected ID %q, got %q", testListID, resp.Data.ID)
	}
}

func TestGetAudienceList(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/lists/"+testListID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceListResponse{
			Message: "List retrieved successfully.",
			Data:    AudienceList{ID: testListID, Name: "Newsletter"},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Lists.Get(context.Background(), testListID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "Newsletter" {
		t.Errorf("expected name %q, got %q", "Newsletter", resp.Data.Name)
	}
}

func TestGetAudienceListEscapesID(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// A list ID containing a slash must be percent-encoded by the SDK.
		if r.URL.EscapedPath() != "/audience/lists/foo%2Fbar" {
			t.Errorf("expected escaped path, got %s", r.URL.EscapedPath())
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceListResponse{Message: "ok"})
	})
	defer server.Close()

	_, err := client.Audience.Lists.Get(context.Background(), "foo/bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudienceList(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/lists/"+testListID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		var body UpdateAudienceListRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Renamed" {
			t.Errorf("expected name %q, got %q", "Renamed", body.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceListResponse{
			Message: "List updated successfully.",
			Data:    AudienceList{ID: testListID, Name: "Renamed"},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Lists.Update(context.Background(), testListID, &UpdateAudienceListRequest{Name: "Renamed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Name != "Renamed" {
		t.Errorf("expected name %q, got %q", "Renamed", resp.Data.Name)
	}
}

func TestDeleteAudienceList(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/lists/"+testListID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Audience.Lists.Delete(context.Background(), testListID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBulkDeleteAudienceLists(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/lists/bulk" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		var body BulkDeleteAudienceListsRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.ListIDs) != 2 {
			t.Errorf("expected 2 list_ids, got %d", len(body.ListIDs))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BulkDeleteAudienceListsResponse{
			Message: "Lists deleted successfully.",
			Data:    BulkDeleteAudienceListsData{Deleted: 2},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Lists.BulkDelete(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", resp.Data.Deleted)
	}
}

// --- Contacts ---

func TestListAudienceContacts(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("search") != "jane" {
			t.Errorf("expected search=jane, got %q", q.Get("search"))
		}
		if q.Get("status") != "subscribed" {
			t.Errorf("expected status=subscribed, got %q", q.Get("status"))
		}
		if q.Get("list_id") != testListID {
			t.Errorf("expected list_id=%q, got %q", testListID, q.Get("list_id"))
		}
		if q.Get("segment_id") != testSegmentID {
			t.Errorf("expected segment_id=%q, got %q", testSegmentID, q.Get("segment_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ListAudienceContactsResponse{
			Message: "Contacts retrieved successfully.",
			Data: ListAudienceContactsData{
				Contacts: []AudienceContact{{
					ID:    testContactID,
					Email: "jane@example.com",
					Status: ContactStatusSubscribed,
					Properties: ContactProperties{"first_name": "Jane"},
					Lists:  []AudienceContactListLink{{ID: testListID, Name: "Newsletter"}},
					Topics: []AudienceContactTopicLink{},
				}},
				Pagination: PagePagination{Total: 1, PerPage: 20, CurrentPage: 1, LastPage: 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.List(context.Background(), &ListAudienceContactsParams{
		Search:    "jane",
		Status:    ContactStatusSubscribed,
		ListID:    testListID,
		SegmentID: testSegmentID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(resp.Data.Contacts))
	}
	c := resp.Data.Contacts[0]
	if c.Email != "jane@example.com" {
		t.Errorf("expected email %q, got %q", "jane@example.com", c.Email)
	}
	if c.Properties["first_name"] != "Jane" {
		t.Errorf("expected first_name=Jane, got %v", c.Properties)
	}
	if len(c.Lists) != 1 || c.Lists[0].ID != testListID {
		t.Errorf("expected one list link, got %+v", c.Lists)
	}
}

func TestCreateAudienceContact(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		var body CreateAudienceContactRequest
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("failed to decode: %v", err)
		} else {
			if body.Email != "jane@example.com" {
				t.Errorf("unexpected email: %s", body.Email)
			}
			if body.DoubleOptIn == nil {
				t.Errorf("expected double_opt_in to be set")
			} else if body.DoubleOptIn.TemplateSlug != "double-opt-in" {
				t.Errorf("unexpected template_slug: %s", body.DoubleOptIn.TemplateSlug)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AudienceContactResponse{
			Message: "Contact created successfully.",
			Data: AudienceContact{
				ID:     testContactID,
				Email:  "jane@example.com",
				Status: ContactStatusUnverified,
			},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.Create(context.Background(), &CreateAudienceContactRequest{
		Email: "jane@example.com",
		DoubleOptIn: &DoubleOptInConfig{
			From:         "no-reply@example.com",
			Subject:      "Please confirm",
			TemplateSlug: "double-opt-in",
			RedirectURL:  "https://example.com/confirmed",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Status != ContactStatusUnverified {
		t.Errorf("expected status %q, got %q", ContactStatusUnverified, resp.Data.Status)
	}
}

func TestGetAudienceContact(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts/"+testContactID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceContactResponse{
			Message: "Contact retrieved successfully.",
			Data:    AudienceContact{ID: testContactID, Email: "jane@example.com", Status: ContactStatusSubscribed},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.Get(context.Background(), testContactID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Email != "jane@example.com" {
		t.Errorf("expected email %q, got %q", "jane@example.com", resp.Data.Email)
	}
}

func TestUpdateAudienceContactWithNullProperty(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts/"+testContactID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		// A nil-pointer value should be encoded as JSON null to signal removal.
		if !bytes.Contains(raw, []byte(`"country":null`)) {
			t.Errorf("expected country to be null, got body: %s", raw)
		}
		if !bytes.Contains(raw, []byte(`"first_name":"Jane"`)) {
			t.Errorf("expected first_name=Jane, got body: %s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceContactResponse{
			Message: "Contact updated successfully.",
			Data:    AudienceContact{ID: testContactID, Email: "jane@example.com", Status: ContactStatusSubscribed},
		})
	})
	defer server.Close()

	name := "Jane"
	_, err := client.Audience.Contacts.Update(context.Background(), testContactID, &UpdateAudienceContactRequest{
		Properties: map[string]*string{
			"first_name": &name,
			"country":    nil,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAudienceContact(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts/"+testContactID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Audience.Contacts.Delete(context.Background(), testContactID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBulkCreateAudienceContacts(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts/bulk" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body BulkCreateAudienceContactsRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.Emails) != 2 {
			t.Errorf("expected 2 emails, got %d", len(body.Emails))
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BulkCreateAudienceContactsResponse{
			Message: "Contacts created.",
			Data:    BulkCreateAudienceContactsData{Created: 1, AlreadyExisted: 1},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.BulkCreate(context.Background(), &BulkCreateAudienceContactsRequest{
		Emails: []string{"jane@example.com", "joe@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Created != 1 || resp.Data.AlreadyExisted != 1 {
		t.Errorf("unexpected counts: created=%d, already=%d", resp.Data.Created, resp.Data.AlreadyExisted)
	}
}

func TestAttachContactToList(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		expected := "/audience/contacts/" + testContactID + "/lists/" + testListID
		if r.URL.Path != expected {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AttachContactResponse{Message: "Contact attached to list."})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.AttachToList(context.Background(), testContactID, testListID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "Contact attached to list." {
		t.Errorf("unexpected message: %q", resp.Message)
	}
}

func TestDetachContactFromList(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		expected := "/audience/contacts/" + testContactID + "/lists/" + testListID
		if r.URL.Path != expected {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Audience.Contacts.DetachFromList(context.Background(), testContactID, testListID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBulkAttachContactsToLists(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts/lists/bulk" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body BulkContactsListsRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.ContactIDs) != 1 || len(body.ListIDs) != 1 {
			t.Errorf("unexpected body shape: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BulkAttachContactsToListsResponse{
			Message: "Contacts attached.",
			Data:    BulkAttachContactsToListsData{Attached: 1, AlreadyAttached: 0, TotalPairs: 1},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.BulkAttachToLists(context.Background(), &BulkContactsListsRequest{
		ContactIDs: []string{testContactID},
		ListIDs:    []string{testListID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Attached != 1 {
		t.Errorf("expected 1 attached, got %d", resp.Data.Attached)
	}
}

func TestBulkDetachContactsFromLists(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/contacts/lists/bulk" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		var body BulkContactsListsRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.ContactIDs) != 1 || body.ContactIDs[0] != testContactID {
			t.Errorf("unexpected contact_ids: %v", body.ContactIDs)
		}
		if len(body.ListIDs) != 1 || body.ListIDs[0] != testListID {
			t.Errorf("unexpected list_ids: %v", body.ListIDs)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BulkDetachContactsFromListsResponse{
			Message: "Contacts detached.",
			Data:    BulkDetachContactsFromListsData{Detached: 1, NotPresent: 0, TotalPairs: 1},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.BulkDetachFromLists(context.Background(), &BulkContactsListsRequest{
		ContactIDs: []string{testContactID},
		ListIDs:    []string{testListID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Detached != 1 {
		t.Errorf("expected 1 detached, got %d", resp.Data.Detached)
	}
}

func TestSubscribeContactToTopic(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		expected := "/audience/contacts/" + testContactID + "/topics/" + testTopicID
		if r.URL.Path != expected {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AttachContactResponse{Message: "Contact subscribed to topic."})
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.SubscribeToTopic(context.Background(), testContactID, testTopicID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Message != "Contact subscribed to topic." {
		t.Errorf("unexpected message: %q", resp.Message)
	}
}

func TestUnsubscribeContactFromTopic(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		expected := "/audience/contacts/" + testContactID + "/topics/" + testTopicID
		if r.URL.Path != expected {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Audience.Contacts.UnsubscribeFromTopic(context.Background(), testContactID, testTopicID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Topics ---

func TestListAudienceTopics(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/topics" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		desc := "Monthly product updates"
		json.NewEncoder(w).Encode(ListAudienceTopicsResponse{
			Message: "Topics retrieved successfully.",
			Data: ListAudienceTopicsData{
				Topics: []AudienceTopic{{
					ID:                  testTopicID,
					Name:                "Product updates",
					Description:         &desc,
					DefaultSubscription: TopicDefaultSubscriptionOptIn,
					Visibility:          TopicVisibilityPublic,
					ContactsCount:       12,
				}},
				Pagination: PagePagination{Total: 1, PerPage: 20, CurrentPage: 1, LastPage: 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Topics.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Topics[0].Visibility != TopicVisibilityPublic {
		t.Errorf("expected visibility %q, got %q", TopicVisibilityPublic, resp.Data.Topics[0].Visibility)
	}
	if resp.Data.Topics[0].Description == nil || *resp.Data.Topics[0].Description != "Monthly product updates" {
		t.Errorf("unexpected description: %v", resp.Data.Topics[0].Description)
	}
}

func TestCreateAudienceTopic(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body CreateAudienceTopicRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Product updates" {
			t.Errorf("expected name %q, got %q", "Product updates", body.Name)
		}
		if body.Visibility != TopicVisibilityPublic {
			t.Errorf("expected visibility %q, got %q", TopicVisibilityPublic, body.Visibility)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceTopicResponse{
			Message: "Topic created.",
			Data:    AudienceTopic{ID: testTopicID, Name: "Product updates"},
		})
	})
	defer server.Close()

	_, err := client.Audience.Topics.Create(context.Background(), &CreateAudienceTopicRequest{
		Name:       "Product updates",
		Visibility: TopicVisibilityPublic,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudienceTopic(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/topics/"+testTopicID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceTopicResponse{
			Message: "Topic updated.",
			Data:    AudienceTopic{ID: testTopicID, Name: "Renamed", Visibility: TopicVisibilityPrivate},
		})
	})
	defer server.Close()

	_, err := client.Audience.Topics.Update(context.Background(), testTopicID, &UpdateAudienceTopicRequest{
		Name:       "Renamed",
		Visibility: TopicVisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAudienceTopic(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/topics/"+testTopicID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Audience.Topics.Delete(context.Background(), testTopicID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAudienceTopic(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/topics/"+testTopicID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		// Verifies the SDK accepts a null description.
		w.Write([]byte(`{"message":"Topic retrieved.","data":{"id":"` + testTopicID + `","name":"X","description":null,"default_subscription":"opt_in","visibility":"private","contacts_count":0,"created_at":null}}`))
	})
	defer server.Close()

	resp, err := client.Audience.Topics.Get(context.Background(), testTopicID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Description != nil {
		t.Errorf("expected nil description, got %v", *resp.Data.Description)
	}
	if resp.Data.CreatedAt != nil {
		t.Errorf("expected nil created_at, got %v", *resp.Data.CreatedAt)
	}
}

// --- Properties ---

func TestListAudienceProperties(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/properties" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fallback := "Friend"
		json.NewEncoder(w).Encode(ListAudiencePropertiesResponse{
			Message: "Properties retrieved.",
			Data: ListAudiencePropertiesData{
				Properties: []AudienceProperty{{
					ID:            testPropertyID,
					Name:          "first_name",
					Type:          PropertyTypeString,
					FallbackValue: &fallback,
					CreatedAt:     "2024-01-15T10:30:00Z",
				}},
				Pagination: PagePagination{Total: 1, PerPage: 20, CurrentPage: 1, LastPage: 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Properties.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Properties[0].Type != PropertyTypeString {
		t.Errorf("expected type %q, got %q", PropertyTypeString, resp.Data.Properties[0].Type)
	}
	if resp.Data.Properties[0].FallbackValue == nil || *resp.Data.Properties[0].FallbackValue != "Friend" {
		t.Errorf("unexpected fallback_value: %v", resp.Data.Properties[0].FallbackValue)
	}
}

func TestCreateAudienceProperty(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body CreateAudiencePropertyRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "first_name" {
			t.Errorf("expected name first_name, got %s", body.Name)
		}
		if body.Type != PropertyTypeString {
			t.Errorf("expected type string, got %s", body.Type)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudiencePropertyResponse{
			Message: "Property created.",
			Data:    AudienceProperty{ID: testPropertyID, Name: "first_name", Type: PropertyTypeString},
		})
	})
	defer server.Close()

	_, err := client.Audience.Properties.Create(context.Background(), &CreateAudiencePropertyRequest{
		Name: "first_name",
		Type: PropertyTypeString,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudienceProperty(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/properties/"+testPropertyID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		if !bytes.Contains(raw, []byte(`"fallback_value":"Friend"`)) {
			t.Errorf("expected fallback_value in body, got: %s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		fallback := "Friend"
		json.NewEncoder(w).Encode(AudiencePropertyResponse{
			Message: "Property updated.",
			Data:    AudienceProperty{ID: testPropertyID, Name: "first_name", Type: PropertyTypeString, FallbackValue: &fallback},
		})
	})
	defer server.Close()

	_, err := client.Audience.Properties.Update(context.Background(), testPropertyID, &UpdateAudiencePropertyRequest{
		FallbackValue: NewNullString("Friend"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudiencePropertyClearFallback(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if !bytes.Contains(raw, []byte(`"fallback_value":null`)) {
			t.Errorf("expected fallback_value:null in body, got: %s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudiencePropertyResponse{
			Message: "Property updated.",
			Data:    AudienceProperty{ID: testPropertyID, Name: "first_name", Type: PropertyTypeString},
		})
	})
	defer server.Close()

	_, err := client.Audience.Properties.Update(context.Background(), testPropertyID, &UpdateAudiencePropertyRequest{
		FallbackValue: &NullString{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudiencePropertyOmitFallback(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if bytes.Contains(raw, []byte(`"fallback_value"`)) {
			t.Errorf("expected fallback_value to be omitted, got: %s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudiencePropertyResponse{Message: "ok"})
	})
	defer server.Close()

	_, err := client.Audience.Properties.Update(context.Background(), testPropertyID, &UpdateAudiencePropertyRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudienceTopicClearDescription(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if !bytes.Contains(raw, []byte(`"description":null`)) {
			t.Errorf("expected description:null in body, got: %s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceTopicResponse{Message: "ok"})
	})
	defer server.Close()

	_, err := client.Audience.Topics.Update(context.Background(), testTopicID, &UpdateAudienceTopicRequest{
		Description: &NullString{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudienceSegmentClearListID(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if !bytes.Contains(raw, []byte(`"list_id":null`)) {
			t.Errorf("expected list_id:null in body, got: %s", raw)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceSegmentResponse{Message: "ok"})
	})
	defer server.Close()

	_, err := client.Audience.Segments.Update(context.Background(), testSegmentID, &UpdateAudienceSegmentRequest{
		ListID: &NullString{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAudienceProperty(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/properties/"+testPropertyID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		// Verify nullable fallback_value decodes correctly.
		w.Write([]byte(`{"message":"ok","data":{"id":"` + testPropertyID + `","name":"first_name","type":"string","fallback_value":null,"created_at":"2024-01-15T10:30:00Z"}}`))
	})
	defer server.Close()

	resp, err := client.Audience.Properties.Get(context.Background(), testPropertyID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.FallbackValue != nil {
		t.Errorf("expected nil fallback_value, got %v", *resp.Data.FallbackValue)
	}
}

func TestDeleteAudienceProperty(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/properties/"+testPropertyID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Audience.Properties.Delete(context.Background(), testPropertyID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Segments ---

func TestListAudienceSegments(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/segments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if lid := r.URL.Query().Get("list_id"); lid != testListID {
			t.Errorf("expected list_id=%q, got %q", testListID, lid)
		}
		w.Header().Set("Content-Type", "application/json")
		listID := testListID
		listName := "Newsletter"
		count := 7
		json.NewEncoder(w).Encode(ListAudienceSegmentsResponse{
			Message: "Segments retrieved.",
			Data: ListAudienceSegmentsData{
				Segments: []AudienceSegment{{
					ID:                  testSegmentID,
					Name:                "Engaged",
					ListID:              &listID,
					ListName:            &listName,
					ConditionGroups:     []SegmentConditionGroup{{Conditions: []SegmentCondition{{Field: "email", Operator: SegmentOperatorContains, Value: strPtr("@example.com")}}}},
					CachedContactsCount: &count,
					CreatedAt:           "2024-01-15T10:30:00Z",
				}},
				Pagination: PagePagination{Total: 1, PerPage: 20, CurrentPage: 1, LastPage: 1},
			},
		})
	})
	defer server.Close()

	resp, err := client.Audience.Segments.List(context.Background(), &ListAudienceSegmentsParams{ListID: testListID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(resp.Data.Segments))
	}
	seg := resp.Data.Segments[0]
	if seg.CachedContactsCount == nil || *seg.CachedContactsCount != 7 {
		t.Errorf("unexpected cached_contacts_count: %v", seg.CachedContactsCount)
	}
	if len(seg.ConditionGroups) != 1 || len(seg.ConditionGroups[0].Conditions) != 1 {
		t.Fatalf("unexpected condition groups: %+v", seg.ConditionGroups)
	}
	cond := seg.ConditionGroups[0].Conditions[0]
	if cond.Operator != SegmentOperatorContains {
		t.Errorf("expected operator %q, got %q", SegmentOperatorContains, cond.Operator)
	}
}

func TestCreateAudienceSegment(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body CreateAudienceSegmentRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Engaged" {
			t.Errorf("expected name Engaged, got %s", body.Name)
		}
		if len(body.Conditions.Groups) != 1 {
			t.Errorf("expected 1 group, got %d", len(body.Conditions.Groups))
		}
		if body.Conditions.Groups[0].Conditions[0].Operator != SegmentOperatorEquals {
			t.Errorf("unexpected operator: %s", body.Conditions.Groups[0].Conditions[0].Operator)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceSegmentResponse{
			Message: "Segment created.",
			Data:    AudienceSegment{ID: testSegmentID, Name: "Engaged"},
		})
	})
	defer server.Close()

	_, err := client.Audience.Segments.Create(context.Background(), &CreateAudienceSegmentRequest{
		Name: "Engaged",
		Conditions: SegmentConditionsInput{
			Groups: []SegmentConditionGroup{{
				Conditions: []SegmentCondition{{Field: "status", Operator: SegmentOperatorEquals, Value: strPtr("subscribed")}},
			}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAudienceSegment(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/segments/"+testSegmentID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AudienceSegmentResponse{
			Message: "Segment updated.",
			Data:    AudienceSegment{ID: testSegmentID, Name: "Renamed"},
		})
	})
	defer server.Close()

	_, err := client.Audience.Segments.Update(context.Background(), testSegmentID, &UpdateAudienceSegmentRequest{Name: "Renamed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAudienceSegment(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/segments/"+testSegmentID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		// Verifies the SDK accepts a null list_id and null cached_contacts_count.
		w.Write([]byte(`{"message":"ok","data":{"id":"` + testSegmentID + `","name":"X","list_id":null,"list_name":null,"condition_groups":[],"cached_contacts_count":null,"created_at":"2024-01-15T10:30:00Z"}}`))
	})
	defer server.Close()

	resp, err := client.Audience.Segments.Get(context.Background(), testSegmentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ListID != nil {
		t.Errorf("expected nil list_id, got %v", *resp.Data.ListID)
	}
	if resp.Data.CachedContactsCount != nil {
		t.Errorf("expected nil cached_contacts_count, got %v", *resp.Data.CachedContactsCount)
	}
}

func TestDeleteAudienceSegment(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audience/segments/"+testSegmentID {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := client.Audience.Segments.Delete(context.Background(), testSegmentID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- ContactProperties decoding ---

// Verifies the PHP/Laravel empty-associative-array quirk: when a contact has
// no properties, the server may serialize `properties` as `[]` instead of
// `{}`. The SDK must accept both.
func TestContactPropertiesDecodesEmptyArrayAndObject(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty_array", `{"id":"c1","email":"a@b","status":"subscribed","properties":[],"created_at":"t","lists":[],"topics":[]}`},
		{"empty_object", `{"id":"c2","email":"a@b","status":"subscribed","properties":{},"created_at":"t","lists":[],"topics":[]}`},
		{"populated", `{"id":"c3","email":"a@b","status":"subscribed","properties":{"first_name":"Jane"},"created_at":"t","lists":[],"topics":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c AudienceContact
			if err := json.Unmarshal([]byte(tc.body), &c); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if tc.name == "populated" {
				if c.Properties["first_name"] != "Jane" {
					t.Errorf("expected first_name=Jane, got %v", c.Properties)
				}
			} else {
				if len(c.Properties) != 0 {
					t.Errorf("expected empty properties, got %v", c.Properties)
				}
			}
		})
	}
}

func TestListAudienceContactsAcceptsEmptyArrayProperties(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Server returns `properties: []` for a contact with no properties.
		w.Write([]byte(`{"message":"ok","data":{"contacts":[{"id":"` + testContactID + `","email":"jane@example.com","status":"subscribed","properties":[],"created_at":"2024-01-15T10:30:00Z","lists":[],"topics":[]}],"pagination":{"total":1,"per_page":20,"current_page":1,"last_page":1}}}`))
	})
	defer server.Close()

	resp, err := client.Audience.Contacts.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(resp.Data.Contacts))
	}
	if len(resp.Data.Contacts[0].Properties) != 0 {
		t.Errorf("expected empty properties, got %v", resp.Data.Contacts[0].Properties)
	}
}

// --- Errors ---

// Verifies the same PHP/Laravel empty-associative-array quirk for the
// top-level Error.Errors field: a 422 response carrying `"errors": []` (no
// field-level errors) must decode cleanly instead of surfacing a confusing
// "cannot unmarshal array into Go struct field" message.
func TestErrorDecodesEmptyArrayErrors(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty_array", `{"message":"Validation failed.","error_code":"validation_error","errors":[]}`},
		{"null", `{"message":"Validation failed.","error_code":"validation_error","errors":null}`},
		{"missing", `{"message":"Validation failed.","error_code":"validation_error"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var e Error
			if err := json.Unmarshal([]byte(tc.body), &e); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if e.Message != "Validation failed." {
				t.Errorf("expected message preserved, got %q", e.Message)
			}
			if e.ErrorCode != "validation_error" {
				t.Errorf("expected error_code preserved, got %q", e.ErrorCode)
			}
			if len(e.Errors) != 0 {
				t.Errorf("expected no field errors, got %v", e.Errors)
			}
		})
	}
}

func TestErrorDecodesPopulatedErrors(t *testing.T) {
	body := `{"message":"Validation failed.","error_code":"validation_error","errors":{"name":["required"],"email":["invalid","too long"]}}`
	var e Error
	if err := json.Unmarshal([]byte(body), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msgs, ok := e.Errors["name"]; !ok || len(msgs) != 1 || msgs[0] != "required" {
		t.Errorf("unexpected name errors: %v", e.Errors["name"])
	}
	if msgs, ok := e.Errors["email"]; !ok || len(msgs) != 2 {
		t.Errorf("unexpected email errors: %v", e.Errors["email"])
	}
}

func TestAudienceValidationError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Error{
			Message:   "Validation failed.",
			ErrorCode: "validation_error",
			Errors: map[string][]string{
				"name": {"The name field is required."},
			},
		})
	})
	defer server.Close()

	_, err := client.Audience.Lists.Create(context.Background(), &CreateAudienceListRequest{})
	if !IsValidationError(err) {
		t.Fatalf("expected validation error, got: %v", err)
	}
	apiErr, _ := err.(*Error)
	if msgs, exists := apiErr.Errors["name"]; !exists || len(msgs) == 0 {
		t.Error("expected name field error")
	}
}
