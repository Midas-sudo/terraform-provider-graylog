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
	_ resource.Resource                 = &EventDefinitionResource{}
	_ resource.ResourceWithImportState  = &EventDefinitionResource{}
	_ resource.ResourceWithUpgradeState = &EventDefinitionResource{}
)

func NewEventDefinitionResource() resource.Resource {
	return &EventDefinitionResource{}
}

type EventDefinitionResource struct {
	client *client.Client
}

type eventDefinitionNotificationSettingsModel struct {
	GracePeriodMs types.Int64 `tfsdk:"grace_period_ms"`
	BacklogSize   types.Int64 `tfsdk:"backlog_size"`
}

type EventDefinitionResourceModel struct {
	ID                   types.String                              `tfsdk:"id"`
	Title                types.String                              `tfsdk:"title"`
	Description          types.String                              `tfsdk:"description"`
	State                types.String                              `tfsdk:"state"`
	Priority             types.Int64                               `tfsdk:"priority"`
	Alert                types.Bool                                `tfsdk:"alert"`
	Config               types.Dynamic                             `tfsdk:"config"`
	FieldSpec            types.Dynamic                             `tfsdk:"field_spec"`
	KeySpec              []types.String                            `tfsdk:"key_spec"`
	NotificationSettings *eventDefinitionNotificationSettingsModel `tfsdk:"notification_settings"`
	Notifications        types.Dynamic                             `tfsdk:"notifications"`
	Storage              types.Dynamic                             `tfsdk:"storage"`
}

type eventDefinitionResourceModelV0 struct {
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

func eventDefinitionNotificationSettingsSchema(optional bool) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:            optional,
		Computed:            !optional,
		MarkdownDescription: "Notification timing settings.",
		Attributes: map[string]schema.Attribute{
			"grace_period_ms": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"backlog_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *EventDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
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
			"config": schema.DynamicAttribute{
				Required:            true,
				MarkdownDescription: "Event definition configuration object.",
			},
			"field_spec": schema.DynamicAttribute{
				Optional: true,
			},
			"key_spec": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"notification_settings": eventDefinitionNotificationSettingsSchema(true),
			"notifications": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "List of notification binding objects.",
			},
			"storage": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "List of storage configuration objects.",
			},
		},
	}
}

func (r *EventDefinitionResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                         schema.StringAttribute{Computed: true},
					"title":                      schema.StringAttribute{Required: true},
					"description":                schema.StringAttribute{Optional: true},
					"state":                      schema.StringAttribute{Optional: true},
					"priority":                   schema.Int64Attribute{Required: true},
					"alert":                      schema.BoolAttribute{Required: true},
					"config_json":                schema.StringAttribute{Required: true},
					"field_spec_json":            schema.StringAttribute{Optional: true},
					"key_spec":                   schema.ListAttribute{Optional: true, ElementType: types.StringType},
					"notification_settings_json": schema.StringAttribute{Optional: true},
					"notifications_json":         schema.StringAttribute{Optional: true},
					"storage_json":               schema.StringAttribute{Optional: true},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior eventDefinitionResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				configDyn, err := upgradeJSONStringAttr(prior.ConfigJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade config", err.Error())
					return
				}
				fieldSpecDyn, err := upgradeJSONStringAttr(prior.FieldSpecJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade field_spec", err.Error())
					return
				}
				notificationsDyn, err := upgradeJSONStringAttr(prior.NotificationsJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade notifications", err.Error())
					return
				}
				storageDyn, err := upgradeJSONStringAttr(prior.StorageJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade storage", err.Error())
					return
				}

				var settings *eventDefinitionNotificationSettingsModel
				if !prior.NotificationSettingsJSON.IsNull() && prior.NotificationSettingsJSON.ValueString() != "" {
					var raw map[string]interface{}
					if err := json.Unmarshal([]byte(prior.NotificationSettingsJSON.ValueString()), &raw); err != nil {
						resp.Diagnostics.AddError("Failed to upgrade notification_settings", err.Error())
						return
					}
					settings = &eventDefinitionNotificationSettingsModel{
						GracePeriodMs: types.Int64Value(jsonNumberAsInt64(raw["grace_period_ms"])),
						BacklogSize:   types.Int64Value(jsonNumberAsInt64(raw["backlog_size"])),
					}
				}

				upgraded := EventDefinitionResourceModel{
					ID:                   prior.ID,
					Title:                prior.Title,
					Description:          prior.Description,
					State:                prior.State,
					Priority:             prior.Priority,
					Alert:                prior.Alert,
					Config:               configDyn,
					FieldSpec:            fieldSpecDyn,
					KeySpec:              prior.KeySpec,
					NotificationSettings: settings,
					Notifications:        notificationsDyn,
					Storage:              storageDyn,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func jsonNumberAsInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
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
	plannedConfig := data.Config
	plannedFieldSpec := data.FieldSpec
	plannedNotifications := data.Notifications
	plannedStorage := data.Storage
	plannedSettings := data.NotificationSettings

	definitionReq, diags := eventDefinitionFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateEventDefinition(ctx, definitionReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create event definition", err.Error())
		return
	}

	resp.Diagnostics.Append(mapEventDefinitionToResourceModel(ctx, created, &data)...)
	data.Config = plannedConfig
	data.FieldSpec = plannedFieldSpec
	data.Notifications = plannedNotifications
	data.Storage = plannedStorage
	data.NotificationSettings = plannedSettings
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

	resp.Diagnostics.Append(mapEventDefinitionToResourceModel(ctx, current, &data)...)
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
	plannedConfig := data.Config
	plannedFieldSpec := data.FieldSpec
	plannedNotifications := data.Notifications
	plannedStorage := data.Storage
	plannedSettings := data.NotificationSettings

	definitionReq, diags := eventDefinitionFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateEventDefinition(ctx, state.ID.ValueString(), definitionReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update event definition", err.Error())
		return
	}

	resp.Diagnostics.Append(mapEventDefinitionToResourceModel(ctx, updated, &data)...)
	data.Config = plannedConfig
	data.FieldSpec = plannedFieldSpec
	data.Notifications = plannedNotifications
	data.Storage = plannedStorage
	data.NotificationSettings = plannedSettings
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

func eventDefinitionFromModel(ctx context.Context, data *EventDefinitionResourceModel) (*client.EventDefinition, diag.Diagnostics) {
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

	cfg, d := dynamicToMap(ctx, data.Config)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	def.Config = cfg

	if !data.FieldSpec.IsNull() && !data.FieldSpec.IsUnknown() {
		fs, d := dynamicToMap(ctx, data.FieldSpec)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		def.FieldSpec = fs
	}

	if len(data.KeySpec) > 0 {
		def.KeySpec = make([]string, 0, len(data.KeySpec))
		for _, v := range data.KeySpec {
			def.KeySpec = append(def.KeySpec, v.ValueString())
		}
	}

	if data.NotificationSettings != nil {
		def.NotificationSettings = map[string]interface{}{
			"grace_period_ms": data.NotificationSettings.GracePeriodMs.ValueInt64(),
			"backlog_size":    data.NotificationSettings.BacklogSize.ValueInt64(),
		}
	}

	if !data.Notifications.IsNull() && !data.Notifications.IsUnknown() {
		raw, d := dynamicToSlice(ctx, data.Notifications)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		b, err := json.Marshal(raw)
		if err != nil {
			diags.AddError("Invalid notifications", err.Error())
			return nil, diags
		}
		if err := json.Unmarshal(b, &def.Notifications); err != nil {
			diags.AddError("Invalid notifications", err.Error())
			return nil, diags
		}
	}

	if !data.Storage.IsNull() && !data.Storage.IsUnknown() {
		storage, d := dynamicToSliceOfMaps(ctx, data.Storage)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		def.Storage = storage
	}

	return def, diags
}
