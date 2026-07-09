package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Lookup types capture only the fields the by-name data sources expose. They
// unmarshal from the corresponding `filter-one` responses (extra fields are
// ignored), keeping the data-source layer decoupled from the resource read
// structs.

// TeamLookup is a team resolved by name.
type TeamLookup struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	EnvID       *string `json:"env_id"`
}

// ConnectorLookup is a connector resolved by name.
type ConnectorLookup struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ConnectorType string `json:"connector_type"`
	TeamID        string `json:"team_id"`
}

// EnvLookup is an environment resolved by name.
type EnvLookup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PipelineLookup is a pipeline resolved by name.
type PipelineLookup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FindTeamByName resolves a team via GET /teams/filter-one. Returns (nil, nil)
// when no team matches.
func (c *Client) FindTeamByName(ctx context.Context, name string) (*TeamLookup, error) {
	var out TeamLookup
	found, err := c.filterOne(ctx, "/teams/filter-one", name, &out)
	if err != nil || !found {
		return nil, err
	}
	return &out, nil
}

// FindConnectorByName resolves a connector via GET /connectors/filter-one.
// Returns (nil, nil) when no connector matches.
func (c *Client) FindConnectorByName(ctx context.Context, name string) (*ConnectorLookup, error) {
	var out ConnectorLookup
	found, err := c.filterOne(ctx, "/connectors/filter-one", name, &out)
	if err != nil || !found {
		return nil, err
	}
	return &out, nil
}

// FindEnvByName resolves an environment via GET /envs/filter-one. Returns
// (nil, nil) when no environment matches.
func (c *Client) FindEnvByName(ctx context.Context, name string) (*EnvLookup, error) {
	var out EnvLookup
	found, err := c.filterOne(ctx, "/envs/filter-one", name, &out)
	if err != nil || !found {
		return nil, err
	}
	return &out, nil
}

// FindPipelineByName resolves a pipeline via GET /pipelines/filter-one. Returns
// (nil, nil) when no pipeline matches.
func (c *Client) FindPipelineByName(ctx context.Context, name string) (*PipelineLookup, error) {
	var out PipelineLookup
	found, err := c.filterOne(ctx, "/pipelines/filter-one", name, &out)
	if err != nil || !found {
		return nil, err
	}
	return &out, nil
}

// filterOne performs a `filter-one` lookup by name, decoding into out. It
// returns found=false (with no error) on 404 so callers can emit a friendly
// "not found" diagnostic.
func (c *Client) filterOne(ctx context.Context, path, name string, out any) (bool, error) {
	q := url.Values{}
	q.Set("name", name)
	fullPath := fmt.Sprintf("%s?%s", path, q.Encode())

	resp, err := c.doRequest(ctx, http.MethodGet, fullPath, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if err := checkResponse(resp); err != nil {
		return false, err
	}
	if err := decodeJSONBody(resp.Body, out); err != nil {
		return false, err
	}
	return true, nil
}

func decodeJSONBody(body io.Reader, out any) error {
	raw, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
