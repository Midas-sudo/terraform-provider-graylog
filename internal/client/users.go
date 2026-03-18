// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type User struct {
	ID               string   `json:"id,omitempty"`
	Username         string   `json:"username"`
	Email            string   `json:"email,omitempty"`
	FullName         string   `json:"full_name,omitempty"`
	FirstName        string   `json:"first_name,omitempty"`
	LastName         string   `json:"last_name,omitempty"`
	Roles            []string `json:"roles,omitempty"`
	Permissions      []string `json:"permissions,omitempty"`
	Timezone         string   `json:"timezone,omitempty"`
	SessionTimeoutMS int64    `json:"session_timeout_ms,omitempty"`
	ServiceAccount   bool     `json:"service_account,omitempty"`
	AccountStatus    string   `json:"account_status,omitempty"`
	ReadOnly         bool     `json:"read_only,omitempty"`
	External         bool     `json:"external,omitempty"`
}

type CreateUserRequest struct {
	Username         string   `json:"username"`
	Password         string   `json:"password"`
	Email            string   `json:"email,omitempty"`
	FullName         string   `json:"full_name,omitempty"`
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	Roles            []string `json:"roles"`
	Permissions      []string `json:"permissions"`
	Timezone         string   `json:"timezone,omitempty"`
	SessionTimeoutMS int64    `json:"session_timeout_ms,omitempty"`
	ServiceAccount   bool     `json:"service_account,omitempty"`
}

type UpdateUserRequest struct {
	Email            string   `json:"email,omitempty"`
	FullName         string   `json:"full_name,omitempty"`
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	Roles            []string `json:"roles"`
	Permissions      []string `json:"permissions"`
	Timezone         string   `json:"timezone,omitempty"`
	SessionTimeoutMS int64    `json:"session_timeout_ms,omitempty"`
	ServiceAccount   bool     `json:"service_account,omitempty"`
}

type UsersResponse struct {
	Users []User `json:"users"`
}

func (c *Client) GetUsers(ctx context.Context) (*UsersResponse, error) {
	var result UsersResponse
	err := c.Get(ctx, "/users", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var result User
	err := c.Get(ctx, fmt.Sprintf("/users/id/%s", id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var result User
	err := c.Get(ctx, fmt.Sprintf("/users/%s", username), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) error {
	err := c.Post(ctx, "/users", req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) error {
	err := c.Put(ctx, fmt.Sprintf("/users/%s", id), req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteUser(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/users/id/%s", id))
}
