// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

func lookupDataAdapterFromPayload(payload string) (*client.LookupDataAdapter, diag.Diagnostics) {
	var diags diag.Diagnostics
	var adapter client.LookupDataAdapter
	if err := json.Unmarshal([]byte(payload), &adapter); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &adapter, diags
}

func lookupCacheFromPayload(payload string) (*client.LookupCache, diag.Diagnostics) {
	var diags diag.Diagnostics
	var cache client.LookupCache
	if err := json.Unmarshal([]byte(payload), &cache); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &cache, diags
}

func lookupTableFromPayload(payload string) (*client.LookupTable, diag.Diagnostics) {
	var diags diag.Diagnostics
	var table client.LookupTable
	if err := json.Unmarshal([]byte(payload), &table); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &table, diags
}

func marshalLookupDataAdapterJSON(adapter *client.LookupDataAdapter) string {
	clone := *adapter
	clone.ID = ""
	clone.Scope = ""
	clone.ContentPack = nil
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func marshalLookupCacheJSON(cache *client.LookupCache) string {
	clone := *cache
	clone.ID = ""
	clone.Scope = ""
	clone.ContentPack = nil
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func marshalLookupTableJSON(table *client.LookupTable) string {
	clone := *table
	clone.ID = ""
	clone.Scope = ""
	clone.ContentPack = nil
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func mapLookupDataAdapterToResourceModel(adapter *client.LookupDataAdapter, data *LookupDataAdapterResourceModel) {
	data.ID = types.StringValue(adapter.ID)
	data.Title = types.StringValue(adapter.Title)
	data.Name = types.StringValue(adapter.Name)
	data.Description = types.StringValue(adapter.Description)
	if adapter.CustomErrorTTLEnabled != nil {
		data.CustomErrorTTLEnabled = types.BoolValue(*adapter.CustomErrorTTLEnabled)
	} else {
		data.CustomErrorTTLEnabled = types.BoolNull()
	}
	if adapter.CustomErrorTTL != nil {
		data.CustomErrorTTL = types.Int64Value(*adapter.CustomErrorTTL)
	} else {
		data.CustomErrorTTL = types.Int64Null()
	}
	if adapter.CustomErrorTTLUnit != nil {
		data.CustomErrorTTLUnit = types.StringValue(*adapter.CustomErrorTTLUnit)
	} else {
		data.CustomErrorTTLUnit = types.StringNull()
	}
}

func mapLookupCacheToResourceModel(cache *client.LookupCache, data *LookupCacheResourceModel) {
	data.ID = types.StringValue(cache.ID)
	data.Title = types.StringValue(cache.Title)
	data.Name = types.StringValue(cache.Name)
	data.Description = types.StringValue(cache.Description)
}

func mapLookupTableToResourceModel(table *client.LookupTable, data *LookupTableResourceModel) {
	data.ID = types.StringValue(table.ID)
	data.Title = types.StringValue(table.Title)
	data.Name = types.StringValue(table.Name)
	data.Description = types.StringValue(table.Description)
	data.CacheID = types.StringValue(table.CacheID)
	data.DataAdapterID = types.StringValue(table.DataAdapterID)
	data.DefaultSingleValue = types.StringValue(table.DefaultSingleValue)
	data.DefaultSingleValueType = types.StringValue(table.DefaultSingleType)
	data.DefaultMultiValue = types.StringValue(table.DefaultMultiValue)
	data.DefaultMultiValueType = types.StringValue(table.DefaultMultiType)
}

func mapLookupDataAdapterToDataSourceModel(adapter *client.LookupDataAdapter, data *LookupDataAdapterDataSourceModel) {
	data.ID = types.StringValue(adapter.ID)
	data.Title = types.StringValue(adapter.Title)
	data.Name = types.StringValue(adapter.Name)
	data.Description = types.StringValue(adapter.Description)
	if adapter.Config != nil {
		if b, err := json.Marshal(adapter.Config); err == nil {
			data.ConfigJSON = types.StringValue(string(b))
		} else {
			data.ConfigJSON = types.StringValue("{}")
		}
	} else {
		data.ConfigJSON = types.StringValue("{}")
	}
	if adapter.CustomErrorTTLEnabled != nil {
		data.CustomErrorTTLEnabled = types.BoolValue(*adapter.CustomErrorTTLEnabled)
	} else {
		data.CustomErrorTTLEnabled = types.BoolNull()
	}
	if adapter.CustomErrorTTL != nil {
		data.CustomErrorTTL = types.Int64Value(*adapter.CustomErrorTTL)
	} else {
		data.CustomErrorTTL = types.Int64Null()
	}
	if adapter.CustomErrorTTLUnit != nil {
		data.CustomErrorTTLUnit = types.StringValue(*adapter.CustomErrorTTLUnit)
	} else {
		data.CustomErrorTTLUnit = types.StringNull()
	}
}

func mapLookupCacheToDataSourceModel(cache *client.LookupCache, data *LookupCacheDataSourceModel) {
	data.ID = types.StringValue(cache.ID)
	data.Title = types.StringValue(cache.Title)
	data.Name = types.StringValue(cache.Name)
	data.Description = types.StringValue(cache.Description)
	if cache.Config != nil {
		if b, err := json.Marshal(cache.Config); err == nil {
			data.ConfigJSON = types.StringValue(string(b))
		} else {
			data.ConfigJSON = types.StringValue("{}")
		}
	} else {
		data.ConfigJSON = types.StringValue("{}")
	}
}

func mapLookupTableToDataSourceModel(table *client.LookupTable, data *LookupTableDataSourceModel) {
	data.ID = types.StringValue(table.ID)
	data.Title = types.StringValue(table.Title)
	data.Name = types.StringValue(table.Name)
	data.Description = types.StringValue(table.Description)
	data.CacheID = types.StringValue(table.CacheID)
	data.DataAdapterID = types.StringValue(table.DataAdapterID)
	data.DefaultSingleValue = types.StringValue(table.DefaultSingleValue)
	data.DefaultSingleValueType = types.StringValue(table.DefaultSingleType)
	data.DefaultMultiValue = types.StringValue(table.DefaultMultiValue)
	data.DefaultMultiValueType = types.StringValue(table.DefaultMultiType)
}
