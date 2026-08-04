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
	ID          types.String  `tfsdk:"id"`
	Title       types.String  `tfsdk:"title"`
	Description types.String  `tfsdk:"description"`
	Config      types.Dynamic `tfsdk:"config"`
}

type eventNotificationListItemModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Config      types.String `tfsdk:"config"`
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
			"config": schema.DynamicAttribute{
				Computed:            true,
				MarkdownDescription: "Event notification configuration object.",
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

	resp.Diagnostics.Append(mapEventNotificationToDataSourceModel(ctx, notification, &data)...)
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
	Notifications []eventNotificationListItemModel `tfsdk:"notifications"`
}

func (d *EventNotificationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_notifications"
}

func (d *EventNotificationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog event notifications. Nested `config` is a JSON string " +
			"(Plugin Framework limitation); use `graylog_event_notification` for a typed object.",
		Attributes: map[string]schema.Attribute{
			"notifications": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"title":       schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"config": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "JSON-encoded configuration object.",
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
		data.Notifications = append(data.Notifications, eventNotificationListItemModel{
			ID:          types.StringValue(notification.ID),
			Title:       types.StringValue(notification.Title),
			Description: types.StringValue(notification.Description),
			Config:      lookupConfigJSONString(notification.Config),
		})
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

// eventDefinitionListItemModel keeps Dynamic fields as JSON strings (framework limitation).
type eventDefinitionListItemModel struct {
	ID                   types.String                              `tfsdk:"id"`
	Title                types.String                              `tfsdk:"title"`
	Description          types.String                              `tfsdk:"description"`
	State                types.String                              `tfsdk:"state"`
	Priority             types.Int64                               `tfsdk:"priority"`
	Alert                types.Bool                                `tfsdk:"alert"`
	Config               types.String                              `tfsdk:"config"`
	FieldSpec            types.String                              `tfsdk:"field_spec"`
	KeySpec              []types.String                            `tfsdk:"key_spec"`
	NotificationSettings *eventDefinitionNotificationSettingsModel `tfsdk:"notification_settings"`
	Notifications        types.String                              `tfsdk:"notifications"`
	Storage              types.String                              `tfsdk:"storage"`
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
			"priority": schema.Int64Attribute{Computed: true},
			"alert":    schema.BoolAttribute{Computed: true},
			"config":   schema.DynamicAttribute{Computed: true},
			"field_spec": schema.DynamicAttribute{Computed: true},
			"key_spec":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"notification_settings": eventDefinitionNotificationSettingsSchema(false),
			"notifications":         schema.DynamicAttribute{Computed: true},
			"storage":               schema.DynamicAttribute{Computed: true},
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

	resp.Diagnostics.Append(mapEventDefinitionToDataSourceModel(ctx, definition, &data)...)
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
	EventDefinitions []eventDefinitionListItemModel `tfsdk:"event_definitions"`
}

func (d *EventDefinitionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_definitions"
}

func (d *EventDefinitionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog event definitions. Nested Dynamic fields are JSON strings " +
			"(Plugin Framework limitation); use `graylog_event_definition` for typed objects.",
		Attributes: map[string]schema.Attribute{
			"event_definitions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true},
						"title":                 schema.StringAttribute{Computed: true},
						"description":           schema.StringAttribute{Computed: true},
						"state":                 schema.StringAttribute{Computed: true},
						"priority":              schema.Int64Attribute{Computed: true},
						"alert":                 schema.BoolAttribute{Computed: true},
						"config":                schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded configuration object."},
						"field_spec":            schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded field spec object."},
						"key_spec":              schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"notification_settings": eventDefinitionNotificationSettingsSchema(false),
						"notifications":         schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded notifications array."},
						"storage":               schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded storage array."},
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
		data.EventDefinitions = append(data.EventDefinitions, mapEventDefinitionToListItemModel(&definition))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
