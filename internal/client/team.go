package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TeamCreate represents the request to create a team
type TeamCreate struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	EnvID       *string `json:"env_id,omitempty"`
}

// TeamUpdate represents the request to update a team
type TeamUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	EnvID       *string `json:"env_id,omitempty"`
}

// TeamRead represents a team response
type TeamRead struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	EnvID       *string `json:"env_id"`
}

// CreateTeam creates a new team
func (c *Client) CreateTeam(ctx context.Context, team *TeamCreate) (*TeamRead, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/teams/", team)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result TeamRead
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetTeam retrieves a team by ID
func (c *Client) GetTeam(ctx context.Context, teamID string) (*TeamRead, error) {
	path := fmt.Sprintf("/teams/%s", teamID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result TeamRead
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// UpdateTeam updates an existing team
func (c *Client) UpdateTeam(ctx context.Context, teamID string, team *TeamUpdate) (*TeamRead, error) {
	path := fmt.Sprintf("/teams/%s", teamID)
	resp, err := c.doRequest(ctx, http.MethodPatch, path, team)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result TeamRead
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// DeleteTeam deletes a team by ID
func (c *Client) DeleteTeam(ctx context.Context, teamID string) error {
	path := fmt.Sprintf("/teams/%s", teamID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := checkResponse(resp); err != nil {
		return err
	}

	return nil
}
