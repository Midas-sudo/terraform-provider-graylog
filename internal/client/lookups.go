package client

import (
	"context"
	"fmt"
)

type LookupDataAdapter struct {
	ID                    string                 `json:"id,omitempty"`
	Scope                 string                 `json:"_scope,omitempty"`
	Title                 string                 `json:"title"`
	Description           string                 `json:"description,omitempty"`
	Name                  string                 `json:"name"`
	Config                map[string]interface{} `json:"config"`
	CustomErrorTTLEnabled *bool                  `json:"custom_error_ttl_enabled,omitempty"`
	CustomErrorTTL        *int64                 `json:"custom_error_ttl,omitempty"`
	CustomErrorTTLUnit    *string                `json:"custom_error_ttl_unit,omitempty"`
	ContentPack           *string                `json:"content_pack,omitempty"`
}

type LookupCache struct {
	ID          string                 `json:"id,omitempty"`
	Scope       string                 `json:"_scope,omitempty"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	ContentPack *string                `json:"content_pack,omitempty"`
}

type LookupTable struct {
	ID                 string  `json:"id,omitempty"`
	Scope              string  `json:"_scope,omitempty"`
	Title              string  `json:"title"`
	Description        string  `json:"description,omitempty"`
	Name               string  `json:"name"`
	CacheID            string  `json:"cache_id"`
	DataAdapterID      string  `json:"data_adapter_id"`
	ContentPack        *string `json:"content_pack,omitempty"`
	DefaultSingleValue string  `json:"default_single_value"`
	DefaultSingleType  string  `json:"default_single_value_type"`
	DefaultMultiValue  string  `json:"default_multi_value"`
	DefaultMultiType   string  `json:"default_multi_value_type"`
}

type LookupDataAdaptersResponse struct {
	Total        int                 `json:"total"`
	Page         int                 `json:"page"`
	PerPage      int                 `json:"per_page"`
	Count        int                 `json:"count"`
	DataAdapters []LookupDataAdapter `json:"data_adapters"`
}

type LookupCachesResponse struct {
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
	Count   int           `json:"count"`
	Caches  []LookupCache `json:"caches"`
}

type LookupTablesResponse struct {
	Total        int           `json:"total"`
	Page         int           `json:"page"`
	PerPage      int           `json:"per_page"`
	Count        int           `json:"count"`
	LookupTables []LookupTable `json:"lookup_tables"`
}

func (c *Client) GetLookupDataAdapters(ctx context.Context) (*LookupDataAdaptersResponse, error) {
	var result LookupDataAdaptersResponse
	if err := c.Get(ctx, "/system/lookup/adapters", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetLookupDataAdapter(ctx context.Context, idOrName string) (*LookupDataAdapter, error) {
	var result LookupDataAdapter
	if err := c.Get(ctx, fmt.Sprintf("/system/lookup/adapters/%s", idOrName), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateLookupDataAdapter(ctx context.Context, adapter *LookupDataAdapter) (*LookupDataAdapter, error) {
	var result LookupDataAdapter
	if err := c.Post(ctx, "/system/lookup/adapters", adapter, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateLookupDataAdapter(ctx context.Context, id string, adapter *LookupDataAdapter) (*LookupDataAdapter, error) {
	adapter.ID = id
	var result LookupDataAdapter
	if err := c.Put(ctx, fmt.Sprintf("/system/lookup/adapters/%s", id), adapter, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteLookupDataAdapter(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/lookup/adapters/%s", id))
}

func (c *Client) GetLookupCaches(ctx context.Context) (*LookupCachesResponse, error) {
	var result LookupCachesResponse
	if err := c.Get(ctx, "/system/lookup/caches", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetLookupCache(ctx context.Context, idOrName string) (*LookupCache, error) {
	var result LookupCache
	if err := c.Get(ctx, fmt.Sprintf("/system/lookup/caches/%s", idOrName), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateLookupCache(ctx context.Context, cache *LookupCache) (*LookupCache, error) {
	var result LookupCache
	if err := c.Post(ctx, "/system/lookup/caches", cache, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateLookupCache(ctx context.Context, id string, cache *LookupCache) (*LookupCache, error) {
	cache.ID = id
	var result LookupCache
	if err := c.Put(ctx, fmt.Sprintf("/system/lookup/caches/%s", id), cache, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteLookupCache(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/lookup/caches/%s", id))
}

func (c *Client) GetLookupTables(ctx context.Context) (*LookupTablesResponse, error) {
	var result LookupTablesResponse
	if err := c.Get(ctx, "/system/lookup/tables", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetLookupTable(ctx context.Context, idOrName string) (*LookupTable, error) {
	path := fmt.Sprintf("/system/lookup/tables/%s", idOrName)

	var result LookupTable
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.ID != "" {
		return &result, nil
	}

	// Some Graylog versions return a list envelope on this endpoint.
	var envelope LookupTablesResponse
	if err := c.Get(ctx, path, &envelope); err != nil {
		return nil, err
	}
	for _, table := range envelope.LookupTables {
		if table.ID == idOrName || table.Name == idOrName {
			return &table, nil
		}
	}

	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("lookup table %q not found", idOrName),
	}
}

func (c *Client) CreateLookupTable(ctx context.Context, table *LookupTable) (*LookupTable, error) {
	var result LookupTable
	if err := c.Post(ctx, "/system/lookup/tables", table, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateLookupTable(ctx context.Context, id string, table *LookupTable) (*LookupTable, error) {
	table.ID = id
	var result LookupTable
	if err := c.Put(ctx, fmt.Sprintf("/system/lookup/tables/%s", id), table, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteLookupTable(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/lookup/tables/%s", id))
}
