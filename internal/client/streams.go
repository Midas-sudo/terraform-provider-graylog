// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type StreamCreateRequest struct {
	Title                          string       `json:"title"`
	Description                    string       `json:"description,omitempty"`
	IndexSetID                     string       `json:"index_set_id"`
	MatchingType                   string       `json:"matching_type,omitempty"`
	RemoveMatchesFromDefaultStream bool         `json:"remove_matches_from_default_stream"`
	Rules                          []StreamRule `json:"rules"`
}

type StreamUpdateRequest struct {
	Title                          string `json:"title"`
	Description                    string `json:"description,omitempty"`
	IndexSetID                     string `json:"index_set_id"`
	MatchingType                   string `json:"matching_type,omitempty"`
	RemoveMatchesFromDefaultStream bool   `json:"remove_matches_from_default_stream"`
}

type Stream struct {
	ID                             string       `json:"id"`
	Title                          string       `json:"title"`
	Description                    string       `json:"description"`
	CreatorUserID                  string       `json:"creator_user_id"`
	CreatedAt                      string       `json:"created_at"`
	Rules                          []StreamRule `json:"rules"`
	Disabled                       bool         `json:"disabled"`
	MatchingType                   string       `json:"matching_type"`
	RemoveMatchesFromDefaultStream bool         `json:"remove_matches_from_default_stream"`
	IsDefault                      bool         `json:"is_default"`
	IsEditable                     bool         `json:"is_editable"`
	IndexSetID                     string       `json:"index_set_id"`
	ContentPack                    string       `json:"content_pack"`
	Categories                     []string     `json:"categories"`
}

type StreamRule struct {
	ID          string `json:"id,omitempty"`
	StreamID    string `json:"stream_id,omitempty"`
	Field       string `json:"field,omitempty"`
	Value       string `json:"value,omitempty"`
	Type        int    `json:"type"`
	Inverted    bool   `json:"inverted"`
	Description string `json:"description,omitempty"`
}

type StreamListResponse struct {
	Total   int      `json:"total"`
	Streams []Stream `json:"streams"`
}

type StreamCreatedResponse struct {
	StreamID string `json:"stream_id"`
}

type StreamRuleCreatedResponse struct {
	StreamRuleID string `json:"streamrule_id"`
}

func (c *Client) GetStreams(ctx context.Context) (*StreamListResponse, error) {
	var result StreamListResponse
	err := c.Get(ctx, "/streams", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetStream(ctx context.Context, id string) (*Stream, error) {
	var result Stream
	err := c.Get(ctx, fmt.Sprintf("/streams/%s", id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// createStreamRequestBody wraps the stream in the "entity" key required by Graylog API.
type createStreamRequestBody struct {
	Entity *StreamCreateRequest `json:"entity"`
}

func (c *Client) CreateStream(ctx context.Context, stream *StreamCreateRequest) (*StreamCreatedResponse, error) {
	body := &createStreamRequestBody{Entity: stream}
	var result StreamCreatedResponse
	err := c.Post(ctx, "/streams", body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateStream(ctx context.Context, id string, stream *StreamUpdateRequest) (*Stream, error) {
	var result Stream
	err := c.Put(ctx, fmt.Sprintf("/streams/%s", id), stream, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteStream(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/streams/%s", id))
}

func (c *Client) ResumeStream(ctx context.Context, id string) error {
	return c.Post(ctx, fmt.Sprintf("/streams/%s/resume", id), nil, nil)
}

func (c *Client) PauseStream(ctx context.Context, id string) error {
	return c.Post(ctx, fmt.Sprintf("/streams/%s/pause", id), nil, nil)
}

func (c *Client) GetStreamRules(ctx context.Context, streamID string) ([]StreamRule, error) {
	var result struct {
		StreamRules []StreamRule `json:"stream_rules"`
	}
	err := c.Get(ctx, fmt.Sprintf("/streams/%s/rules", streamID), &result)
	if err != nil {
		return nil, err
	}
	return result.StreamRules, nil
}

func (c *Client) GetStreamRule(ctx context.Context, streamID, ruleID string) (*StreamRule, error) {
	var result StreamRule
	err := c.Get(ctx, fmt.Sprintf("/streams/%s/rules/%s", streamID, ruleID), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateStreamRule(ctx context.Context, streamID string, rule *StreamRule) (*StreamRuleCreatedResponse, error) {
	var result StreamRuleCreatedResponse
	err := c.Post(ctx, fmt.Sprintf("/streams/%s/rules", streamID), rule, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateStreamRule(ctx context.Context, streamID, ruleID string, rule *StreamRule) error {
	// Response body is typically {"streamrule_id":"..."} rather than the full rule.
	return c.Put(ctx, fmt.Sprintf("/streams/%s/rules/%s", streamID, ruleID), rule, nil)
}

func (c *Client) DeleteStreamRule(ctx context.Context, streamID, ruleID string) error {
	return c.Delete(ctx, fmt.Sprintf("/streams/%s/rules/%s", streamID, ruleID))
}
