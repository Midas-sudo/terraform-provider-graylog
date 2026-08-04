// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type View struct {
	ID          string                 `json:"id,omitempty"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	SearchID    string                 `json:"search_id"`
	Properties  []interface{}          `json:"properties,omitempty"`
	Requires    map[string]interface{} `json:"requires,omitempty"`
	State       map[string]interface{} `json:"state"`
}

type ViewsResponse struct {
	Total int    `json:"total"`
	Views []View `json:"views"`
}

type viewEntityRequest struct {
	Entity *View `json:"entity"`
}

func (c *Client) GetViews(ctx context.Context) (*ViewsResponse, error) {
	var result ViewsResponse
	err := c.Get(ctx, "/views", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetDashboards(ctx context.Context) (*ViewsResponse, error) {
	// Graylog 7 `/dashboards` returns paginated `elements`, while `/views?type=DASHBOARD`
	// returns the same shape as `/views` (`views` array). Prefer the views endpoint.
	var result ViewsResponse
	err := c.Get(ctx, "/views?type=DASHBOARD", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetView(ctx context.Context, id string) (*View, error) {
	var result View
	err := c.Get(ctx, fmt.Sprintf("/views/%s", id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateView(ctx context.Context, view *View) (*View, error) {
	var result View
	err := c.Post(ctx, "/views", &viewEntityRequest{Entity: view}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateView(ctx context.Context, id string, view *View) (*View, error) {
	var result View
	err := c.Put(ctx, fmt.Sprintf("/views/%s", id), &viewEntityRequest{Entity: view}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteView(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/views/%s", id))
}
