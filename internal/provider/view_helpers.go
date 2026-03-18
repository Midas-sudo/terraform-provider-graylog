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

func mapDashboardToModel(view *client.View, data *DashboardResourceModel) {
	data.ID = types.StringValue(view.ID)
	data.Type = types.StringValue(view.Type)
	data.Title = types.StringValue(view.Title)
	data.Summary = types.StringValue(view.Summary)
	data.Description = types.StringValue(view.Description)
	data.SearchID = types.StringValue(view.SearchID)
}
