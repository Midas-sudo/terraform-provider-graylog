// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type ContentPack struct {
	ID          string                   `json:"id"`
	Rev         int64                    `json:"rev"`
	V           string                   `json:"v"`
	Name        string                   `json:"name"`
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description,omitempty"`
	Vendor      string                   `json:"vendor,omitempty"`
	URL         string                   `json:"url,omitempty"`
	CreatedAt   string                   `json:"created_at,omitempty"`
	Parameters  []map[string]interface{} `json:"parameters"`
	Entities    []map[string]interface{} `json:"entities"`
}

type ContentPacksResponse struct {
	Total        int           `json:"total"`
	ContentPacks []ContentPack `json:"content_packs"`
}

type contentPackRevisionResponse struct {
	ContentPack *ContentPack `json:"content_pack"`
}

type ContentPackInstallationRequest struct {
	Comment    string                 `json:"comment,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type ContentPackInstallation struct {
	ID                  string                 `json:"_id"`
	ContentPackID       string                 `json:"content_pack_id"`
	ContentPackRevision int64                  `json:"content_pack_revision"`
	Comment             string                 `json:"comment,omitempty"`
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt           string                 `json:"created_at,omitempty"`
	CreatedBy           string                 `json:"created_by,omitempty"`
}

type ContentPackInstallationsResponse struct {
	Total         int                       `json:"total"`
	Installations []ContentPackInstallation `json:"installations"`
}

type contentPackInstallationCreateRequest struct {
	Entity *ContentPackInstallationRequest `json:"entity"`
}

func (c *Client) GetContentPacks(ctx context.Context) (*ContentPacksResponse, error) {
	var result ContentPacksResponse
	if err := c.Get(ctx, "/system/content_packs/latest", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetContentPack(ctx context.Context, id string, rev int64) (*ContentPack, error) {
	var result contentPackRevisionResponse
	if err := c.Get(ctx, fmt.Sprintf("/system/content_packs/%s/%d", id, rev), &result); err != nil {
		return nil, err
	}
	if result.ContentPack == nil {
		return nil, &APIError{
			StatusCode: 404,
			Message:    fmt.Sprintf("content pack %s/%d not found", id, rev),
		}
	}
	return result.ContentPack, nil
}

func (c *Client) CreateContentPack(ctx context.Context, contentPack *ContentPack) error {
	if contentPack.Parameters == nil {
		contentPack.Parameters = []map[string]interface{}{}
	}
	if contentPack.Entities == nil {
		contentPack.Entities = []map[string]interface{}{}
	}
	return c.Post(ctx, "/system/content_packs", contentPack, nil)
}

func (c *Client) DeleteContentPack(ctx context.Context, id string, rev int64) error {
	return c.Delete(ctx, fmt.Sprintf("/system/content_packs/%s/%d", id, rev))
}

func (c *Client) GetContentPackInstallations(ctx context.Context, contentPackID string) (*ContentPackInstallationsResponse, error) {
	var result ContentPackInstallationsResponse
	if err := c.Get(ctx, fmt.Sprintf("/system/content_packs/%s/installations", contentPackID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) InstallContentPack(ctx context.Context, contentPackID string, rev int64, req *ContentPackInstallationRequest) (*ContentPackInstallation, error) {
	var result ContentPackInstallation
	createReq := &contentPackInstallationCreateRequest{
		Entity: req,
	}
	if err := c.Post(ctx, fmt.Sprintf("/system/content_packs/%s/%d/installations", contentPackID, rev), createReq, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteContentPackInstallation(ctx context.Context, contentPackID string, installationID string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/content_packs/%s/installations/%s", contentPackID, installationID))
}
