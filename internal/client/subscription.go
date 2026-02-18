package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SubscriptionCreate represents the request to create a subscription
type SubscriptionCreate struct {
	Name              string `json:"name"`
	DatamodelID       string `json:"datamodel_id"`
	TargetConnectorID string `json:"target_connector_id"`
	TargetTableName   string `json:"target_table_name,omitempty"`
	Backfill          bool   `json:"backfill"`
	ErrorTableEnabled bool   `json:"error_table_enabled"`
	ErrorTableName    string `json:"error_table_name,omitempty"`
}

// SubscriptionUpdate represents the request to update a subscription
type SubscriptionUpdate struct {
	Name              *string `json:"name,omitempty"`
	TargetTableName   *string `json:"target_table_name,omitempty"`
	Backfill          *bool   `json:"backfill,omitempty"`
	ErrorTableEnabled *bool   `json:"error_table_enabled,omitempty"`
	ErrorTableName    *string `json:"error_table_name,omitempty"`
}

// SubscriptionCreateResponse represents the response from creating a subscription
type SubscriptionCreateResponse struct {
	ID string `json:"id"`
}

// SubscriptionRead represents a subscription detail response
type SubscriptionRead struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	DatamodelID       string `json:"datamodel_id"`
	TargetConnectorID string `json:"targetId"`
	TargetTableName   string `json:"target_table_name"`
	Backfill          bool   `json:"backfill"`
	ErrorTableEnabled bool   `json:"error_table_enabled"`
	ErrorTableName    string `json:"error_table_name"`
	Enabled           bool   `json:"enabled"`
	Status            string `json:"status"`
}

// CreateSubscription creates a new subscription
func (c *Client) CreateSubscription(ctx context.Context, sub *SubscriptionCreate) (*SubscriptionRead, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/subscriptions/", sub)
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

	var createResp SubscriptionCreateResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// POST only returns the ID, fetch the full detail
	return c.GetSubscription(ctx, createResp.ID)
}

// GetSubscription retrieves a subscription by ID
func (c *Client) GetSubscription(ctx context.Context, subscriptionID string) (*SubscriptionRead, error) {
	path := fmt.Sprintf("/subscriptions/%s", subscriptionID)
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

	var result SubscriptionRead
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// UpdateSubscription updates an existing subscription
func (c *Client) UpdateSubscription(ctx context.Context, subscriptionID string, sub *SubscriptionUpdate) (*SubscriptionRead, error) {
	path := fmt.Sprintf("/subscriptions/%s", subscriptionID)
	resp, err := c.doRequest(ctx, http.MethodPatch, path, sub)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	// PATCH returns {id}, fetch the full detail
	return c.GetSubscription(ctx, subscriptionID)
}

// DeleteSubscription deletes a subscription by ID
func (c *Client) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	path := fmt.Sprintf("/subscriptions/%s", subscriptionID)
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
