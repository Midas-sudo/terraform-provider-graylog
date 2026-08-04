// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type InputCreateRequest struct {
	Title         string                 `json:"title"`
	Type          string                 `json:"type"`
	Global        bool                   `json:"global"`
	Node          string                 `json:"node,omitempty"`
	Configuration map[string]interface{} `json:"configuration"`
}

type InputUpdateRequest struct {
	Title         string                 `json:"title"`
	Type          string                 `json:"type"`
	Global        bool                   `json:"global"`
	Node          string                 `json:"node,omitempty"`
	Configuration map[string]interface{} `json:"configuration"`
}

type Input struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Type          string                 `json:"type"`
	Global        bool                   `json:"global"`
	Node          string                 `json:"node"`
	CreatorUserID string                 `json:"creator_user_id"`
	CreatedAt     string                 `json:"created_at"`
	Attributes    map[string]interface{} `json:"attributes"`
	StaticFields  map[string]string      `json:"static_fields"`
	ContentPack   string                 `json:"content_pack"`
	Name          string                 `json:"name"`
}

type InputsList struct {
	Total  int     `json:"total"`
	Inputs []Input `json:"inputs"`
}

type InputCreatedResponse struct {
	ID string `json:"id"`
}

// InputType is the human-readable label for an input type class name.
// Graylog 6+/7 `/system/inputs/types` returns types as map[className]humanName.
type InputType string

type InputTypesResponse struct {
	Types map[string]InputType `json:"types"`
}

func (c *Client) GetInputs(ctx context.Context) (*InputsList, error) {
	var result InputsList
	err := c.Get(ctx, "/system/inputs", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetInput(ctx context.Context, id string) (*Input, error) {
	var result Input
	err := c.Get(ctx, fmt.Sprintf("/system/inputs/%s", id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateInput(ctx context.Context, input *InputCreateRequest) (*InputCreatedResponse, error) {
	var result InputCreatedResponse
	err := c.Post(ctx, "/system/inputs", input, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateInput(ctx context.Context, id string, input *InputUpdateRequest) error {
	return c.Put(ctx, fmt.Sprintf("/system/inputs/%s", id), input, nil)
}

func (c *Client) DeleteInput(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/inputs/%s", id))
}

func (c *Client) GetInputTypes(ctx context.Context) (*InputTypesResponse, error) {
	var result InputTypesResponse
	err := c.Get(ctx, "/system/inputs/types", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
