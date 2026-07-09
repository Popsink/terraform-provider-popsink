package client

import (
	"context"
	"fmt"
	"net/http"
)

// DataModelRead is the subset of a datamodel the provider adopts and manages.
// `Enabled` reflects the start/stop lifecycle; `State` is the worker state.
type DataModelRead struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Enabled bool   `json:"enabled"`
}

// GetDataModel retrieves a datamodel by ID. Returns (nil, nil) when not found.
func (c *Client) GetDataModel(ctx context.Context, id string) (*DataModelRead, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/datamodels/%s", id), nil)
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

	var result DataModelRead
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StartDataModel enables a datamodel and returns its refreshed detail.
func (c *Client) StartDataModel(ctx context.Context, id string) (*DataModelRead, error) {
	return c.dataModelLifecycle(ctx, id, "start")
}

// StopDataModel disables a datamodel and returns its refreshed detail.
func (c *Client) StopDataModel(ctx context.Context, id string) (*DataModelRead, error) {
	return c.dataModelLifecycle(ctx, id, "stop")
}

func (c *Client) dataModelLifecycle(ctx context.Context, id, action string) (*DataModelRead, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/datamodels/%s/%s", id, action), nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result DataModelRead
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
