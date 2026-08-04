// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type DataTieringConfig struct {
	Type             string `json:"type"`
	IndexLifetimeMin string `json:"index_lifetime_min,omitempty"`
	IndexLifetimeMax string `json:"index_lifetime_max,omitempty"`
}

type IndexSet struct {
	ID                              string                 `json:"id,omitempty"`
	Title                           string                 `json:"title,omitempty"`
	Description                     string                 `json:"description,omitempty"`
	IndexPrefix                     string                 `json:"index_prefix,omitempty"`
	IndexOptimizationMaxNumSegments int64                  `json:"index_optimization_max_num_segments,omitempty"`
	IndexOptimizationDisabled       bool                   `json:"index_optimization_disabled"`
	FieldTypeRefreshInterval        int64                  `json:"field_type_refresh_interval,omitempty"`
	Shards                          int64                  `json:"shards,omitempty"`
	Replicas                        int64                  `json:"replicas"`
	Writable                        bool                   `json:"writable,omitempty"`
	Default                         bool                   `json:"default,omitempty"`
	IndexAnalyzer                   string                 `json:"index_analyzer,omitempty"`
	UseLegacyRotation               bool                   `json:"use_legacy_rotation"`
	IndexTemplateType               string                 `json:"index_template_type,omitempty"`
	RotationStrategyClass           string                 `json:"rotation_strategy_class,omitempty"`
	RotationStrategy                map[string]interface{} `json:"rotation_strategy"`
	RetentionStrategyClass          string                 `json:"retention_strategy_class,omitempty"`
	RetentionStrategy               map[string]interface{} `json:"retention_strategy"`
	DataTiering                     *DataTieringConfig     `json:"data_tiering,omitempty"`
}

type IndexSetsResponse struct {
	Total     int        `json:"total"`
	IndexSets []IndexSet `json:"index_sets"`
}

type IndexTemplate struct {
	Settings      map[string]interface{} `json:"settings,omitempty"`
	Mappings      map[string]interface{} `json:"mappings,omitempty"`
	IndexPatterns []string               `json:"index_patterns,omitempty"`
	Order         int64                  `json:"order,omitempty"`
}

type IndexTemplateResponse struct {
	Name     string        `json:"name"`
	Template IndexTemplate `json:"template"`
}

func (c *Client) GetIndexSets(ctx context.Context) (*IndexSetsResponse, error) {
	var result IndexSetsResponse
	if err := c.Get(ctx, "/system/indices/index_sets", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetIndexSet(ctx context.Context, id string) (*IndexSet, error) {
	var result IndexSet
	if err := c.Get(ctx, fmt.Sprintf("/system/indices/index_sets/%s", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateIndexSet(ctx context.Context, req *IndexSet) (*IndexSet, error) {
	var result IndexSet
	if err := c.Post(ctx, "/system/indices/index_sets", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateIndexSet(ctx context.Context, id string, req *IndexSet) (*IndexSet, error) {
	var result IndexSet
	if err := c.Put(ctx, fmt.Sprintf("/system/indices/index_sets/%s", id), req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteIndexSet(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/indices/index_sets/%s", id))
}

func (c *Client) SetDefaultIndexSet(ctx context.Context, id string) error {
	return c.Put(ctx, fmt.Sprintf("/system/indices/index_sets/%s/default", id), map[string]interface{}{}, nil)
}

func (c *Client) GetIndexTemplate(ctx context.Context, indexSetID string) (*IndexTemplateResponse, error) {
	var result IndexTemplateResponse
	if err := c.Get(ctx, fmt.Sprintf("/system/indexer/indices/templates/%s", indexSetID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetIndexTemplates(ctx context.Context) ([]IndexTemplateResponse, error) {
	var result []IndexTemplateResponse
	if err := c.Get(ctx, "/system/indexer/indices/templates", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) SyncIndexTemplate(ctx context.Context, indexSetID string) (*IndexTemplateResponse, error) {
	var result IndexTemplateResponse
	if err := c.Post(ctx, fmt.Sprintf("/system/indexer/indices/templates/%s/update", indexSetID), map[string]interface{}{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
