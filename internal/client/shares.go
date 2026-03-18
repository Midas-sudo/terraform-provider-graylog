// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type EntityShareRequest struct {
	SelectedGranteeCapabilities map[string]string `json:"selected_grantee_capabilities,omitempty"`
}

type ActiveShare struct {
	Capability string `json:"capability"`
	Grantee    string `json:"grantee"`
	Grant      string `json:"grant"`
}

type EntityShareResponse struct {
	Entity                      string            `json:"entity"`
	ActiveShares                []ActiveShare     `json:"active_shares"`
	SelectedGranteeCapabilities map[string]string `json:"selected_grantee_capabilities"`
}

func (c *Client) GetEntityShares(ctx context.Context, entityGRN string) (*EntityShareResponse, error) {
	var result EntityShareResponse
	err := c.Get(ctx, fmt.Sprintf("/authz/shares/entities/%s", entityGRN), &result)
	if err == nil {
		return &result, nil
	}

	if apiErr, ok := err.(*APIError); ok && (apiErr.StatusCode == 404 || apiErr.StatusCode == 405) {
		prepareReq := &EntityShareRequest{SelectedGranteeCapabilities: map[string]string{}}
		err = c.Post(ctx, fmt.Sprintf("/authz/shares/entities/%s/prepare", entityGRN), prepareReq, &result)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	return nil, err
}

func (c *Client) UpdateEntityShares(ctx context.Context, entityGRN string, req *EntityShareRequest) (*EntityShareResponse, error) {
	var result EntityShareResponse
	err := c.Post(ctx, fmt.Sprintf("/authz/shares/entities/%s", entityGRN), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
