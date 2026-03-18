// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type PipelineSource struct {
	ID            string `json:"id,omitempty"`
	Title         string `json:"title,omitempty"`
	Description   string `json:"description,omitempty"`
	Source        string `json:"source"`
	CreatedAt     string `json:"created_at,omitempty"`
	ModifiedAt    string `json:"modified_at,omitempty"`
}

type PipelineRule struct {
	ID            string `json:"id,omitempty"`
	Title         string `json:"title,omitempty"`
	Description   string `json:"description,omitempty"`
	Source        string `json:"source"`
	CreatedAt     string `json:"created_at,omitempty"`
	ModifiedAt    string `json:"modified_at,omitempty"`
}

type PipelineConnection struct {
	ID          string   `json:"id,omitempty"`
	StreamID    string   `json:"stream_id"`
	PipelineIDs []string `json:"pipeline_ids"`
}

type PipelineConnectionRequest struct {
	StreamID    string   `json:"stream_id"`
	PipelineIDs []string `json:"pipeline_ids"`
}

func (c *Client) GetPipelines(ctx context.Context) ([]PipelineSource, error) {
	var result []PipelineSource
	err := c.Get(ctx, "/system/pipelines/pipeline", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetPipeline(ctx context.Context, id string) (*PipelineSource, error) {
	var result PipelineSource
	err := c.Get(ctx, fmt.Sprintf("/system/pipelines/pipeline/%s", id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreatePipeline(ctx context.Context, pipeline *PipelineSource) (*PipelineSource, error) {
	var result PipelineSource
	err := c.Post(ctx, "/system/pipelines/pipeline", pipeline, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdatePipeline(ctx context.Context, id string, pipeline *PipelineSource) (*PipelineSource, error) {
	var result PipelineSource
	err := c.Put(ctx, fmt.Sprintf("/system/pipelines/pipeline/%s", id), pipeline, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeletePipeline(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/pipelines/pipeline/%s", id))
}

func (c *Client) GetPipelineRules(ctx context.Context) ([]PipelineRule, error) {
	var result []PipelineRule
	err := c.Get(ctx, "/system/pipelines/rule", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetPipelineRule(ctx context.Context, id string) (*PipelineRule, error) {
	var result PipelineRule
	err := c.Get(ctx, fmt.Sprintf("/system/pipelines/rule/%s", id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreatePipelineRule(ctx context.Context, rule *PipelineRule) (*PipelineRule, error) {
	var result PipelineRule
	err := c.Post(ctx, "/system/pipelines/rule", rule, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdatePipelineRule(ctx context.Context, id string, rule *PipelineRule) (*PipelineRule, error) {
	var result PipelineRule
	err := c.Put(ctx, fmt.Sprintf("/system/pipelines/rule/%s", id), rule, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeletePipelineRule(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/pipelines/rule/%s", id))
}

func (c *Client) GetPipelineConnections(ctx context.Context) ([]PipelineConnection, error) {
	var result []PipelineConnection
	err := c.Get(ctx, "/system/pipelines/connections", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetPipelineConnectionsForStream(ctx context.Context, streamID string) (*PipelineConnection, error) {
	var result PipelineConnection
	err := c.Get(ctx, fmt.Sprintf("/system/pipelines/connections/%s", streamID), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ConnectPipelinesToStream(ctx context.Context, conn *PipelineConnectionRequest) (*PipelineConnection, error) {
	var result PipelineConnection
	err := c.Post(ctx, "/system/pipelines/connections/to_stream", conn, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
