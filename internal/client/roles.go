package client

import (
	"context"
	"fmt"
)

type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	ReadOnly    bool     `json:"read_only,omitempty"`
}

type RolesResponse struct {
	Total int    `json:"total"`
	Roles []Role `json:"roles"`
}

func (c *Client) GetRoles(ctx context.Context) (*RolesResponse, error) {
	var result RolesResponse
	err := c.Get(ctx, "/roles", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetRole(ctx context.Context, name string) (*Role, error) {
	var result Role
	err := c.Get(ctx, fmt.Sprintf("/roles/%s", name), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateRole(ctx context.Context, req *Role) (*Role, error) {
	var result Role
	err := c.Post(ctx, "/roles", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateRole(ctx context.Context, name string, req *Role) (*Role, error) {
	var result Role
	err := c.Put(ctx, fmt.Sprintf("/roles/%s", name), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteRole(ctx context.Context, name string) error {
	return c.Delete(ctx, fmt.Sprintf("/roles/%s", name))
}
