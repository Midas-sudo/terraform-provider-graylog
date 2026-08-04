// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
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
	_ resource.Resource                 = &EventNotificationResource{}
	_ resource.ResourceWithImportState  = &EventNotificationResource{}
	_ resource.ResourceWithUpgradeState = &EventNotificationResource{}
)

func NewEventNotificationResource() resource.Resource {
	return &EventNotificationResource{}
}

type EventNotificationResource struct {
	client *client.Client
}

type EventNotificationResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Title       types.String  `tfsdk:"title"`
	Description types.String  `tfsdk:"description"`
	Config      types.Dynamic `tfsdk:"config"`
}

type eventNotificationResourceModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	ConfigJSON  types.String `tfsdk:"config_json"`
}

func (r *EventNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_notification"
}

func (r *EventNotificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a Graylog event notification.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Notification title.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Notification description.",
			},
			"config": schema.DynamicAttribute{
				Required:            true,
				MarkdownDescription: "Notification-specific configuration object.",
			},
		},
	}
}

func (r *EventNotificationResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":          schema.StringAttribute{Computed: true},
					"title":       schema.StringAttribute{Required: true},
					"description": schema.StringAttribute{Optional: true},
					"config_json": schema.StringAttribute{Required: true},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior eventNotificationResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				dyn, err := upgradeJSONStringAttr(prior.ConfigJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade event notification config", err.Error())
					return
				}
				upgraded := EventNotificationResourceModel{
					ID:          prior.ID,
					Title:       prior.Title,
					Description: prior.Description,
					Config:      dyn,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *EventNotificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EventNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plannedConfig := data.Config

	notificationReq, diags := eventNotificationFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateEventNotification(ctx, notificationReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create event notification", err.Error())
		return
	}

	resp.Diagnostics.Append(mapEventNotificationToResourceModel(ctx, created, &data)...)
	data.Config = plannedConfig
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EventNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetEventNotification(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read event notification", err.Error())
		return
	}

	resp.Diagnostics.Append(mapEventNotificationToResourceModel(ctx, current, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EventNotificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plannedConfig := data.Config

	notificationReq, diags := eventNotificationFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateEventNotification(ctx, data.ID.ValueString(), notificationReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update event notification", err.Error())
		return
	}

	resp.Diagnostics.Append(mapEventNotificationToResourceModel(ctx, updated, &data)...)
	data.Config = plannedConfig
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EventNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEventNotification(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete event notification", err.Error())
	}
}

func (r *EventNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func eventNotificationFromModel(ctx context.Context, data *EventNotificationResourceModel) (*client.EventNotification, diag.Diagnostics) {
	cfg, diags := dynamicToMap(ctx, data.Config)
	if diags.HasError() {
		return nil, diags
	}

	notification := &client.EventNotification{
		Title:  data.Title.ValueString(),
		Config: cfg,
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		notification.Description = data.Description.ValueString()
	}
	return notification, diags
}
