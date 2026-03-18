// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &EventNotificationDataSource{}

func NewEventNotificationDataSource() datasource.DataSource {
	return &EventNotificationDataSource{}
}

type EventNotificationDataSource struct {
	client *client.Client
}

type EventNotificationDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	ConfigJSON  types.String `tfsdk:"config_json"`
}

func (d *EventNotificationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_notification"
}

func (d *EventNotificationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog event notification by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"title": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"config_json": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON object with event notification configuration.",
			},
		},
	}
}

func (d *EventNotificationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *EventNotificationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EventNotificationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	notification, err := d.client.GetEventNotification(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read event notification", err.Error())
		return
	}

	mapEventNotificationToDataSourceModel(notification, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &EventNotificationsDataSource{}

func NewEventNotificationsDataSource() datasource.DataSource {
	return &EventNotificationsDataSource{}
}

type EventNotificationsDataSource struct {
	client *client.Client
}

type EventNotificationsDataSourceModel struct {
	Notifications []EventNotificationDataSourceModel `tfsdk:"notifications"`
}

func (d *EventNotificationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_notifications"
}

func (d *EventNotificationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog event notifications.",
		Attributes: map[string]schema.Attribute{
			"notifications": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"title":       schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"config_json": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *EventNotificationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *EventNotificationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetEventNotifications(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list event notifications", err.Error())
		return
	}

	var data EventNotificationsDataSourceModel
	for _, notification := range result.Notifications {
		row := EventNotificationDataSourceModel{}
		mapEventNotificationToDataSourceModel(&notification, &row)
		data.Notifications = append(data.Notifications, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &EventDefinitionDataSource{}

func NewEventDefinitionDataSource() datasource.DataSource {
	return &EventDefinitionDataSource{}
}

type EventDefinitionDataSource struct {
	client *client.Client
}

type EventDefinitionDataSourceModel struct {
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

func (d *EventDefinitionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_definition"
}

func (d *EventDefinitionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog event definition by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"title": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"priority":                   schema.Int64Attribute{Computed: true},
			"alert":                      schema.BoolAttribute{Computed: true},
			"config_json":                schema.StringAttribute{Computed: true},
			"field_spec_json":            schema.StringAttribute{Computed: true},
			"key_spec":                   schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"notification_settings_json": schema.StringAttribute{Computed: true},
			"notifications_json":         schema.StringAttribute{Computed: true},
			"storage_json":               schema.StringAttribute{Computed: true},
		},
	}
}

func (d *EventDefinitionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *EventDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EventDefinitionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	definition, err := d.client.GetEventDefinition(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read event definition", err.Error())
		return
	}

	mapEventDefinitionToDataSourceModel(definition, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &EventDefinitionsDataSource{}

func NewEventDefinitionsDataSource() datasource.DataSource {
	return &EventDefinitionsDataSource{}
}

type EventDefinitionsDataSource struct {
	client *client.Client
}

type EventDefinitionsDataSourceModel struct {
	EventDefinitions []EventDefinitionDataSourceModel `tfsdk:"event_definitions"`
}

func (d *EventDefinitionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_definitions"
}

func (d *EventDefinitionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog event definitions.",
		Attributes: map[string]schema.Attribute{
			"event_definitions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"title":       schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"state":                      schema.StringAttribute{Computed: true},
						"priority":                   schema.Int64Attribute{Computed: true},
						"alert":                      schema.BoolAttribute{Computed: true},
						"config_json":                schema.StringAttribute{Computed: true},
						"field_spec_json":            schema.StringAttribute{Computed: true},
						"key_spec":                   schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"notification_settings_json": schema.StringAttribute{Computed: true},
						"notifications_json":         schema.StringAttribute{Computed: true},
						"storage_json":               schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *EventDefinitionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *EventDefinitionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetEventDefinitions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list event definitions", err.Error())
		return
	}

	var data EventDefinitionsDataSourceModel
	for _, definition := range result.EventDefinitions {
		row := EventDefinitionDataSourceModel{}
		mapEventDefinitionToDataSourceModel(&definition, &row)
		data.EventDefinitions = append(data.EventDefinitions, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
