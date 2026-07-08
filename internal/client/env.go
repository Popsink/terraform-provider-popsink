package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BrokerConfiguration is the Kafka broker retention configuration carried by an
// environment. Credential fields (sasl_username, sasl_password, ca_cert, cert,
// key) are accepted on write but stripped from read responses by the API (see
// BrokerConfigurationPublic in the data-plane): reads only return
// bootstrap_server, security_protocol, sasl_mechanism and group_id.
type BrokerConfiguration struct {
	BootstrapServer  string  `json:"bootstrap_server"`
	SecurityProtocol string  `json:"security_protocol,omitempty"`
	SaslMechanism    string  `json:"sasl_mechanism,omitempty"`
	SaslUsername     *string `json:"sasl_username"`
	SaslPassword     *string `json:"sasl_password"`
	CaCert           string  `json:"ca_cert,omitempty"`
	Cert             string  `json:"cert,omitempty"`
	Key              string  `json:"key,omitempty"`
	GroupID          *string `json:"group_id,omitempty"`
}

// EnvCreate represents the request to create an environment.
type EnvCreate struct {
	Name                   string              `json:"name"`
	RetentionConfiguration BrokerConfiguration `json:"retention_configuration"`
}

// EnvUpdate represents a partial update to an environment.
type EnvUpdate struct {
	Name                   *string              `json:"name,omitempty"`
	RetentionConfiguration *BrokerConfiguration `json:"retention_configuration,omitempty"`
}

// EnvRead represents an environment response. RetentionConfiguration carries
// only the non-credential fields the API returns.
type EnvRead struct {
	ID                     string              `json:"id"`
	Name                   string              `json:"name"`
	RetentionConfiguration BrokerConfiguration `json:"retention_configuration"`
}

// CreateEnv creates a new environment.
func (c *Client) CreateEnv(ctx context.Context, env *EnvCreate) (*EnvRead, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/envs/", env)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	return decodeEnv(resp.Body)
}

// GetEnv retrieves an environment by ID. Returns (nil, nil) when not found.
func (c *Client) GetEnv(ctx context.Context, envID string) (*EnvRead, error) {
	path := fmt.Sprintf("/envs/%s", envID)
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

	return decodeEnv(resp.Body)
}

// UpdateEnv applies a partial update to an environment.
func (c *Client) UpdateEnv(ctx context.Context, envID string, env *EnvUpdate) (*EnvRead, error) {
	path := fmt.Sprintf("/envs/%s", envID)
	resp, err := c.doRequest(ctx, http.MethodPatch, path, env)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	return decodeEnv(resp.Body)
}

// DeleteEnv deletes an environment by ID. A missing environment is treated as
// already deleted.
func (c *Client) DeleteEnv(ctx context.Context, envID string) error {
	path := fmt.Sprintf("/envs/%s", envID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return checkResponse(resp)
}

func decodeEnv(body io.Reader) (*EnvRead, error) {
	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result EnvRead
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
