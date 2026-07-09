package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFindTeamByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/teams/filter-one" {
			t.Errorf("expected path /teams/filter-one, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "data-engineering" {
			t.Errorf("expected name query data-engineering, got %q", got)
		}
		env := "env-1"
		_ = json.NewEncoder(w).Encode(TeamLookup{ID: "team-1", Name: "data-engineering", Description: "DE", EnvID: &env})
	}))
	defer server.Close()

	team, err := NewClient(server.URL, "tok", false).FindTeamByName(context.Background(), "data-engineering")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team == nil || team.ID != "team-1" {
		t.Fatalf("expected team-1, got %+v", team)
	}
	if team.EnvID == nil || *team.EnvID != "env-1" {
		t.Errorf("expected env_id env-1, got %v", team.EnvID)
	}
}

func TestFindByNameNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok", false)
	if team, err := c.FindTeamByName(context.Background(), "missing"); err != nil || team != nil {
		t.Errorf("team: expected (nil,nil), got (%v,%v)", team, err)
	}
	if conn, err := c.FindConnectorByName(context.Background(), "missing"); err != nil || conn != nil {
		t.Errorf("connector: expected (nil,nil), got (%v,%v)", conn, err)
	}
	if env, err := c.FindEnvByName(context.Background(), "missing"); err != nil || env != nil {
		t.Errorf("env: expected (nil,nil), got (%v,%v)", env, err)
	}
	if p, err := c.FindPipelineByName(context.Background(), "missing"); err != nil || p != nil {
		t.Errorf("pipeline: expected (nil,nil), got (%v,%v)", p, err)
	}
}

func TestFindConnectorByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/connectors/filter-one" {
			t.Errorf("expected /connectors/filter-one, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(ConnectorLookup{ID: "conn-1", Name: "pg", ConnectorType: "POSTGRES_SOURCE", TeamID: "team-1"})
	}))
	defer server.Close()

	conn, err := NewClient(server.URL, "tok", false).FindConnectorByName(context.Background(), "pg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil || conn.ConnectorType != "POSTGRES_SOURCE" {
		t.Fatalf("unexpected connector: %+v", conn)
	}
}
