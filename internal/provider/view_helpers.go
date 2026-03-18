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

func viewFromPayload(payload string) (*client.View, diag.Diagnostics) {
	var diags diag.Diagnostics
	var view client.View
	if err := json.Unmarshal([]byte(payload), &view); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &view, diags
}

func viewFromModel(data *ViewResourceModel, defaultType string) (*client.View, diag.Diagnostics) {
	var diags diag.Diagnostics
	view := &client.View{
		Type:     defaultType,
		Title:    data.Title.ValueString(),
		SearchID: data.SearchID.ValueString(),
	}
	if !data.Summary.IsNull() && !data.Summary.IsUnknown() {
		view.Summary = data.Summary.ValueString()
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		view.Description = data.Description.ValueString()
	}
	if !data.StateJSON.IsNull() && !data.StateJSON.IsUnknown() && data.StateJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.StateJSON.ValueString()), &view.State); err != nil {
			diags.AddError("Invalid state_json", fmt.Sprintf("Failed to parse state_json: %v", err))
			return nil, diags
		}
	}
	if !data.PropertiesJSON.IsNull() && !data.PropertiesJSON.IsUnknown() && data.PropertiesJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.PropertiesJSON.ValueString()), &view.Properties); err != nil {
			diags.AddError("Invalid properties_json", fmt.Sprintf("Failed to parse properties_json: %v", err))
			return nil, diags
		}
	}
	if !data.RequiresJSON.IsNull() && !data.RequiresJSON.IsUnknown() && data.RequiresJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.RequiresJSON.ValueString()), &view.Requires); err != nil {
			diags.AddError("Invalid requires_json", fmt.Sprintf("Failed to parse requires_json: %v", err))
			return nil, diags
		}
	}
	return view, diags
}

func marshalViewJSON(view *client.View) string {
	clone := *view
	clone.ID = ""
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func mapViewToModel(view *client.View, data *ViewResourceModel) {
	data.ID = types.StringValue(view.ID)
	data.Type = types.StringValue(view.Type)
	data.Title = types.StringValue(view.Title)
	data.Summary = types.StringValue(view.Summary)
	data.Description = types.StringValue(view.Description)
	data.SearchID = types.StringValue(view.SearchID)
}

func populateViewJSONFields(view *client.View, data *ViewResourceModel) {
	if data.StateJSON.IsNull() || data.StateJSON.IsUnknown() || data.StateJSON.ValueString() == "" {
		if view.State != nil {
			if b, err := json.Marshal(view.State); err == nil {
				data.StateJSON = types.StringValue(string(b))
			}
		} else {
			data.StateJSON = types.StringNull()
		}
	}
	if data.PropertiesJSON.IsNull() || data.PropertiesJSON.IsUnknown() || data.PropertiesJSON.ValueString() == "" {
		if view.Properties != nil {
			if b, err := json.Marshal(view.Properties); err == nil {
				data.PropertiesJSON = types.StringValue(string(b))
			}
		} else {
			data.PropertiesJSON = types.StringNull()
		}
	}
	if data.RequiresJSON.IsNull() || data.RequiresJSON.IsUnknown() || data.RequiresJSON.ValueString() == "" {
		if view.Requires != nil {
			if b, err := json.Marshal(view.Requires); err == nil {
				data.RequiresJSON = types.StringValue(string(b))
			}
		} else {
			data.RequiresJSON = types.StringNull()
		}
	}
}

func mapDashboardToModel(view *client.View, data *DashboardResourceModel) {
	data.ID = types.StringValue(view.ID)
	data.Type = types.StringValue(view.Type)
	data.Title = types.StringValue(view.Title)
	data.Summary = types.StringValue(view.Summary)
	data.Description = types.StringValue(view.Description)
	data.SearchID = types.StringValue(view.SearchID)
}
