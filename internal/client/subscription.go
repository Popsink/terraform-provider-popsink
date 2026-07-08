package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SubscriptionCreate is the request body to create a subscription. The
// data-plane create endpoint accepts the SMT chain under `smt_config`.
type SubscriptionCreate struct {
	Name              string           `json:"name"`
	DatamodelID       string           `json:"datamodel_id"`
	TargetConnectorID string           `json:"target_connector_id"`
	SmtConfig         []map[string]any `json:"smt_config,omitempty"`
	ConsumerID        *string          `json:"consumer_id,omitempty"`
	ErrorTableEnabled bool             `json:"error_table_enabled"`
	ErrorTableName    string           `json:"error_table_name"`
	TargetTableName   string           `json:"target_table_name"`
	Backfill          bool             `json:"backfill"`
}

// SubscriptionUpdate is the partial update body. Note the data-plane update
// endpoint carries the SMT chain under `mapper_config` (not `smt_config`); the
// resource maps its opaque smt_config attribute onto this field.
type SubscriptionUpdate struct {
	Name               *string          `json:"name,omitempty"`
	MapperConfig       []map[string]any `json:"mapper_config,omitempty"`
	ConsumerID         *string          `json:"consumer_id,omitempty"`
	ErrorTableEnabled  *bool            `json:"error_table_enabled,omitempty"`
	ErrorTableName     *string          `json:"error_table_name,omitempty"`
	ErrorTableTargetID *string          `json:"error_table_target_id,omitempty"`
	TargetTableName    *string          `json:"target_table_name,omitempty"`
	Backfill           *bool            `json:"backfill,omitempty"`
}

// SubscriptionRead is the detail response. Field tags mirror
// SubscriptionDetailDTO (a mix of snake_case config fields and camelCase
// display fields).
type SubscriptionRead struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Status             string           `json:"status"`
	DatamodelID        *string          `json:"datamodel_id"`
	TargetID           *string          `json:"targetId"`
	TargetTableName    string           `json:"target_table_name"`
	SmtConfig          []map[string]any `json:"smt_config"`
	ConsumerID         *string          `json:"consumer_id"`
	ErrorTableEnabled  bool             `json:"error_table_enabled"`
	ErrorTableName     string           `json:"error_table_name"`
	ErrorTableTargetID *string          `json:"error_table_target_id"`
	Backfill           bool             `json:"backfill"`
	Enabled            bool             `json:"enabled"`
	ConfigHash         string           `json:"config_hash"`
}

type subscriptionIDResponse struct {
	ID string `json:"id"`
}

// CreateSubscription creates a subscription and returns its new ID (the create
// endpoint returns only the id).
func (c *Client) CreateSubscription(ctx context.Context, sub *SubscriptionCreate) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/subscriptions/", sub)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return "", err
	}

	var result subscriptionIDResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return "", err
	}
	return result.ID, nil
}

// GetSubscription retrieves a subscription by ID. Returns (nil, nil) when not found.
func (c *Client) GetSubscription(ctx context.Context, id string) (*SubscriptionRead, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/subscriptions/%s", id), nil)
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

	var result SubscriptionRead
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSubscription applies a partial update. The endpoint returns only the id,
// so callers should GetSubscription afterwards to refresh state.
func (c *Client) UpdateSubscription(ctx context.Context, id string, sub *SubscriptionUpdate) error {
	resp, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/subscriptions/%s", id), sub)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	return checkResponse(resp)
}

// DeleteSubscription deletes a subscription by ID. A missing subscription is
// treated as already deleted.
func (c *Client) DeleteSubscription(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/subscriptions/%s", id), nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return checkResponse(resp)
}

// StartSubscription enables a subscription for processing and returns its
// refreshed detail.
func (c *Client) StartSubscription(ctx context.Context, id string) (*SubscriptionRead, error) {
	return c.subscriptionLifecycle(ctx, id, "start")
}

// PauseSubscription disables a subscription from processing and returns its
// refreshed detail.
func (c *Client) PauseSubscription(ctx context.Context, id string) (*SubscriptionRead, error) {
	return c.subscriptionLifecycle(ctx, id, "pause")
}

func (c *Client) subscriptionLifecycle(ctx context.Context, id, action string) (*SubscriptionRead, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/subscriptions/%s/%s", id, action), nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result SubscriptionRead
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func decodeJSON(body io.Reader, v any) error {
	raw, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
