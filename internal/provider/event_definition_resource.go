// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &EventDefinitionResource{}
	_ resource.ResourceWithImportState = &EventDefinitionResource{}
)

func NewEventDefinitionResource() resource.Resource {
	return &EventDefinitionResource{}
}

type EventDefinitionResource struct {
	client *client.Client
}

type EventDefinitionResourceModel struct {
	ID                       types.String   `tfsdk:"id"`
	Title                    types.String   `tfsdk:"title"`
	Description              types.String   `tfsdk:"description"`
	State                    types.String   `tfsdk:"state"`
	Priority                 types.Int64    `tfsdk:"priority"`
	Alert                    types.Bool     `tfsdk:"alert"`
	ConfigJSON               types.String   `tfsdk:"config_json"`
	FieldSpecJSON            types.String   `tfsdk:"field_spec_json"`
	KeySpec                  []types.String `tfsdk:"key_spec"`
	NotificationSettingsJSON types.String   `tfsdk:"notification_settings_json"`
	NotificationsJSON        types.String   `tfsdk:"notifications_json"`
	StorageJSON              types.String   `tfsdk:"storage_json"`
}

func (r *EventDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_definition"
}

func (r *EventDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog event definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Event definition title.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Event definition description.",
			},
			"state": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Event definition state reported by Graylog (`ENABLED`/`DISABLED`).",
			},
			"priority": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Event priority.",
			},
			"alert": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether the definition should trigger alerts.",
			},
			"config_json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON object with event definition config.",
			},
			"field_spec_json": schema.StringAttribute{
				Optional: true,
			},
			"key_spec": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"notification_settings_json": schema.StringAttribute{
				Optional: true,
			},
			"notifications_json": schema.StringAttribute{
				Optional: true,
			},
			"storage_json": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *EventDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *EventDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EventDefinitionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	definitionReq, diags := eventDefinitionFromModel(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateEventDefinition(ctx, definitionReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create event definition", err.Error())
		return
	}

	mapEventDefinitionToResourceModel(created, &data)
	populateEventDefinitionJSONFields(created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EventDefinitionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetEventDefinition(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read event definition", err.Error())
		return
	}

	mapEventDefinitionToResourceModel(current, &data)
	populateEventDefinitionJSONFields(current, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EventDefinitionResourceModel
	var state EventDefinitionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	definitionReq, diags := eventDefinitionFromModel(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateEventDefinition(ctx, state.ID.ValueString(), definitionReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update event definition", err.Error())
		return
	}

	mapEventDefinitionToResourceModel(updated, &data)
	populateEventDefinitionJSONFields(updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EventDefinitionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEventDefinition(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete event definition", err.Error())
	}
}

func (r *EventDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func eventDefinitionFromModel(data *EventDefinitionResourceModel) (*client.EventDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics

	def := &client.EventDefinition{
		Title:    data.Title.ValueString(),
		Priority: data.Priority.ValueInt64(),
		Alert:    data.Alert.ValueBool(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		def.Description = data.Description.ValueString()
	}
	if !data.State.IsNull() && !data.State.IsUnknown() {
		def.State = data.State.ValueString()
	}

	if err := json.Unmarshal([]byte(data.ConfigJSON.ValueString()), &def.Config); err != nil {
		diags.AddError("Invalid config_json", fmt.Sprintf("Failed to parse config_json: %v", err))
		return nil, diags
	}
	if !data.FieldSpecJSON.IsNull() && !data.FieldSpecJSON.IsUnknown() && data.FieldSpecJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.FieldSpecJSON.ValueString()), &def.FieldSpec); err != nil {
			diags.AddError("Invalid field_spec_json", fmt.Sprintf("Failed to parse field_spec_json: %v", err))
			return nil, diags
		}
	}
	if len(data.KeySpec) > 0 {
		def.KeySpec = make([]string, 0, len(data.KeySpec))
		for _, v := range data.KeySpec {
			def.KeySpec = append(def.KeySpec, v.ValueString())
		}
	}
	if !data.NotificationSettingsJSON.IsNull() && !data.NotificationSettingsJSON.IsUnknown() && data.NotificationSettingsJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.NotificationSettingsJSON.ValueString()), &def.NotificationSettings); err != nil {
			diags.AddError("Invalid notification_settings_json", fmt.Sprintf("Failed to parse notification_settings_json: %v", err))
			return nil, diags
		}
	}
	if !data.NotificationsJSON.IsNull() && !data.NotificationsJSON.IsUnknown() && data.NotificationsJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.NotificationsJSON.ValueString()), &def.Notifications); err != nil {
			diags.AddError("Invalid notifications_json", fmt.Sprintf("Failed to parse notifications_json: %v", err))
			return nil, diags
		}
	}
	if !data.StorageJSON.IsNull() && !data.StorageJSON.IsUnknown() && data.StorageJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.StorageJSON.ValueString()), &def.Storage); err != nil {
			diags.AddError("Invalid storage_json", fmt.Sprintf("Failed to parse storage_json: %v", err))
			return nil, diags
		}
	}

	return def, diags
}

func populateEventDefinitionJSONFields(def *client.EventDefinition, data *EventDefinitionResourceModel) {
	if data.ConfigJSON.IsNull() || data.ConfigJSON.IsUnknown() || data.ConfigJSON.ValueString() == "" {
		if def.Config == nil {
			data.ConfigJSON = types.StringValue("{}")
		} else if b, err := json.Marshal(def.Config); err == nil {
			data.ConfigJSON = types.StringValue(string(b))
		}
	}

	if data.FieldSpecJSON.IsNull() || data.FieldSpecJSON.IsUnknown() || data.FieldSpecJSON.ValueString() == "" {
		if def.FieldSpec != nil {
			if b, err := json.Marshal(def.FieldSpec); err == nil {
				data.FieldSpecJSON = types.StringValue(string(b))
			}
		} else {
			data.FieldSpecJSON = types.StringNull()
		}
	}

	if len(data.KeySpec) == 0 && def.KeySpec != nil {
		keys := make([]types.String, 0, len(def.KeySpec))
		for _, k := range def.KeySpec {
			keys = append(keys, types.StringValue(k))
		}
		data.KeySpec = keys
	}

	if data.NotificationSettingsJSON.IsNull() || data.NotificationSettingsJSON.IsUnknown() || data.NotificationSettingsJSON.ValueString() == "" {
		if def.NotificationSettings != nil {
			if b, err := json.Marshal(def.NotificationSettings); err == nil {
				data.NotificationSettingsJSON = types.StringValue(string(b))
			}
		} else {
			data.NotificationSettingsJSON = types.StringNull()
		}
	}
	if data.NotificationsJSON.IsNull() || data.NotificationsJSON.IsUnknown() || data.NotificationsJSON.ValueString() == "" {
		if def.Notifications != nil {
			if b, err := json.Marshal(def.Notifications); err == nil {
				data.NotificationsJSON = types.StringValue(string(b))
			}
		} else {
			data.NotificationsJSON = types.StringNull()
		}
	}
	if data.StorageJSON.IsNull() || data.StorageJSON.IsUnknown() || data.StorageJSON.ValueString() == "" {
		if def.Storage != nil {
			if b, err := json.Marshal(def.Storage); err == nil {
				data.StorageJSON = types.StringValue(string(b))
			}
		} else {
			data.StorageJSON = types.StringNull()
		}
	}
}
