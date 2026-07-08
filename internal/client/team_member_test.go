package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBulkCreateTeamMembers(t *testing.T) {
	var got TeamMemberBulkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/teams/team-1/members/bulk" {
			t.Errorf("expected POST /teams/team-1/members/bulk, got %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := NewClient(server.URL, "tok", false).BulkCreateTeamMembers(context.Background(), "team-1",
		&TeamMemberBulkCreate{Owners: []string{"user-9"}, Members: []string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Owners) != 1 || got.Owners[0] != "user-9" {
		t.Errorf("expected owner user-9 to be sent, got %+v", got)
	}
}

func TestListTeamMembersPaginates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "1":
			_ = json.NewEncoder(w).Encode(teamMemberPage{
				Items: []TeamMember{{ID: "m1", UserID: "u1", Email: "a@x.io", Admin: true}},
				Page:  1, Pages: 2, Size: 100, Total: 2,
			})
		case "2":
			_ = json.NewEncoder(w).Encode(teamMemberPage{
				Items: []TeamMember{{ID: "m2", UserID: "u2", Email: "b@x.io", Admin: false}},
				Page:  2, Pages: 2, Size: 100, Total: 2,
			})
		default:
			t.Errorf("unexpected page %q", page)
		}
	}))
	defer server.Close()

	members, err := NewClient(server.URL, "tok", false).ListTeamMembers(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members across pages, got %d", len(members))
	}
	if members[0].ID != "m1" || members[1].ID != "m2" {
		t.Errorf("unexpected members: %+v", members)
	}
}

func TestGetTeamMemberByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(teamMemberPage{
			Items: []TeamMember{
				{ID: "m1", UserID: "u1", Admin: true},
				{ID: "m2", UserID: "u2", Admin: false},
			},
			Page: 1, Pages: 1, Size: 100, Total: 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok", false)

	m, err := c.GetTeamMember(context.Background(), "team-1", "m2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil || m.UserID != "u2" {
		t.Fatalf("expected member u2, got %+v", m)
	}

	missing, err := c.GetTeamMember(context.Background(), "team-1", "does-not-exist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for missing member, got %+v", missing)
	}

	byUser, err := c.FindTeamMemberByUser(context.Background(), "team-1", "u1")
	if err != nil || byUser == nil || byUser.ID != "m1" {
		t.Errorf("expected to find m1 by user u1, got (%+v,%v)", byUser, err)
	}
}

func TestDeleteTeamMember(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/teams/team-1/members/m1" {
			t.Errorf("expected DELETE /teams/team-1/members/m1, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := NewClient(server.URL, "tok", false).DeleteTeamMember(context.Background(), "team-1", "m1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTeamMemberConflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"detail":"last admin cannot leave"}`))
	}))
	defer server.Close()

	if err := NewClient(server.URL, "tok", false).DeleteTeamMember(context.Background(), "team-1", "m1"); err == nil {
		t.Fatal("expected error on 409, got nil")
	}
}
