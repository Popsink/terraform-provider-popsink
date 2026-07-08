package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ConnectorCreate represents the request to create a connector
type ConnectorCreate struct {
	Name              string         `json:"name"`
	ConnectorType     string         `json:"connector_type"`
	JsonConfiguration map[string]any `json:"json_configuration"`
	TeamID            string         `json:"team_id"`
}

// ConnectorUpdate represents the request to update a connector
type ConnectorUpdate struct {
	Name              *string         `json:"name,omitempty"`
	ConnectorType     *string         `json:"connector_type,omitempty"`
	JsonConfiguration *map[string]any `json:"json_configuration,omitempty"`
	TeamID            *string         `json:"team_id,omitempty"`
}

// ConnectorRead represents a connector response
type ConnectorRead struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	ConnectorType     string         `json:"connector_type"`
	JsonConfiguration map[string]any `json:"json_configuration"`
	TeamID            string         `json:"team_id"`
	ItemsCount        int            `json:"items_count"`
	Status            string         `json:"status"`
}

// CreateConnector creates a new connector
func (c *Client) CreateConnector(ctx context.Context, connector *ConnectorCreate) (*ConnectorRead, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/connectors/", connector)
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

	var result ConnectorRead
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetConnector retrieves a connector by ID
func (c *Client) GetConnector(ctx context.Context, connectorID string) (*ConnectorRead, error) {
	path := fmt.Sprintf("/connectors/%s", connectorID)
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

	var result ConnectorRead
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// UpdateConnector updates an existing connector
func (c *Client) UpdateConnector(ctx context.Context, connectorID string, connector *ConnectorUpdate) (*ConnectorRead, error) {
	path := fmt.Sprintf("/connectors/%s", connectorID)
	resp, err := c.doRequest(ctx, http.MethodPatch, path, connector)
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

	var result ConnectorRead
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// StartConnectorWorker starts a connector's worker. The endpoint is
// asynchronous (202): callers should poll GetConnector until the status
// converges to "live" (or "error").
func (c *Client) StartConnectorWorker(ctx context.Context, connectorID string) error {
	return c.connectorWorkerAction(ctx, connectorID, "start")
}

// StopConnectorWorker stops a connector's worker. The endpoint is asynchronous
// (202): callers should poll GetConnector until the status converges to
// "paused".
func (c *Client) StopConnectorWorker(ctx context.Context, connectorID string) error {
	return c.connectorWorkerAction(ctx, connectorID, "stop")
}

func (c *Client) connectorWorkerAction(ctx context.Context, connectorID, action string) error {
	path := fmt.Sprintf("/connectors/%s/%s", connectorID, action)
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	return checkResponse(resp)
}

// DeleteConnector deletes a connector by ID
func (c *Client) DeleteConnector(ctx context.Context, connectorID string) error {
	path := fmt.Sprintf("/connectors/%s", connectorID)
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
