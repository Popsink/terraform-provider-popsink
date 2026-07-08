package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TeamMemberBulkCreate is the body for POST /teams/{team_id}/members/bulk.
// Users listed in Owners are added with admin privileges; Members without.
type TeamMemberBulkCreate struct {
	Owners  []string `json:"owners"`
	Members []string `json:"members"`
}

// TeamMember is one membership entry as returned by the list endpoint. ID is
// the membership entry id (used for deletion), distinct from UserID.
type TeamMember struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Admin  bool   `json:"admin"`
}

type teamMemberPage struct {
	Items []TeamMember `json:"items"`
	Page  int          `json:"page"`
	Pages int          `json:"pages"`
	Size  int          `json:"size"`
	Total int          `json:"total"`
}

// BulkCreateTeamMembers adds users to a team (204). Already-present users are
// skipped server-side.
func (c *Client) BulkCreateTeamMembers(ctx context.Context, teamID string, body *TeamMemberBulkCreate) error {
	path := fmt.Sprintf("/teams/%s/members/bulk", teamID)
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	return checkResponse(resp)
}

// ListTeamMembers returns all members of a team, following pagination.
func (c *Client) ListTeamMembers(ctx context.Context, teamID string) ([]TeamMember, error) {
	var all []TeamMember
	page := 1
	for {
		path := fmt.Sprintf("/teams/%s/members?page=%d&size=100", teamID, page)
		resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			return nil, nil
		}
		if err := checkResponse(resp); err != nil {
			_ = resp.Body.Close()
			return nil, err
		}

		raw, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var p teamMemberPage
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		all = append(all, p.Items...)
		if page >= p.Pages || len(p.Items) == 0 {
			break
		}
		page++
	}
	return all, nil
}

// GetTeamMember finds a single membership entry by its id, or (nil, nil) if the
// user is no longer a member.
func (c *Client) GetTeamMember(ctx context.Context, teamID, memberID string) (*TeamMember, error) {
	members, err := c.ListTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	for i := range members {
		if members[i].ID == memberID {
			return &members[i], nil
		}
	}
	return nil, nil
}

// FindTeamMemberByUser finds a membership entry by user id, or (nil, nil).
func (c *Client) FindTeamMemberByUser(ctx context.Context, teamID, userID string) (*TeamMember, error) {
	members, err := c.ListTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	for i := range members {
		if members[i].UserID == userID {
			return &members[i], nil
		}
	}
	return nil, nil
}

// DeleteTeamMember removes a membership entry (204). A missing entry is treated
// as already deleted.
func (c *Client) DeleteTeamMember(ctx context.Context, teamID, memberID string) error {
	path := fmt.Sprintf("/teams/%s/members/%s", teamID, memberID)
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
