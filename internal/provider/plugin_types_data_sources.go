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

// --- graylog_output_types ---

var _ datasource.DataSource = &OutputTypesDataSource{}

func NewOutputTypesDataSource() datasource.DataSource { return &OutputTypesDataSource{} }

type OutputTypesDataSource struct {
	client *client.Client
}

type outputTypesDataSourceModel struct {
	Types []typeDescriptorModel `tfsdk:"types"`
}

func (d *OutputTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_output_types"
}

func (d *OutputTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available Graylog output types and their `requested_configuration` fields. " +
			"Use this to discover required/optional keys for `graylog_output.configuration`.",
		Attributes: map[string]schema.Attribute{
			"types": typeDescriptorsListSchema("Available output types."),
		},
	}
}

func (d *OutputTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OutputTypesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetAvailableOutputs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list output types", err.Error())
		return
	}
	types, err := mapTypeDescriptors(ctx, client.SortedTypeDescriptors(result), collapseOutputType)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map output types", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &outputTypesDataSourceModel{Types: types})...)
}

// --- graylog_lookup_adapter_types ---

var _ datasource.DataSource = &LookupAdapterTypesDataSource{}

func NewLookupAdapterTypesDataSource() datasource.DataSource { return &LookupAdapterTypesDataSource{} }

type LookupAdapterTypesDataSource struct {
	client *client.Client
}

type lookupAdapterTypesDataSourceModel struct {
	Types []typeDescriptorModel `tfsdk:"types"`
}

func (d *LookupAdapterTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_adapter_types"
}

func (d *LookupAdapterTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available Graylog lookup data adapter types. " +
			"Graylog exposes `default_config` (not full UI field metadata); keys are synthesized into `requested_configuration`. " +
			"Use this when configuring `graylog_lookup_data_adapter.config`.",
		Attributes: map[string]schema.Attribute{
			"types": typeDescriptorsListSchema("Available lookup data adapter types."),
		},
	}
}

func (d *LookupAdapterTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupAdapterTypesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetLookupAdapterTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list lookup adapter types", err.Error())
		return
	}
	types, err := mapTypeDescriptors(ctx, client.SortedTypeDescriptors(result), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map lookup adapter types", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &lookupAdapterTypesDataSourceModel{Types: types})...)
}

// --- graylog_lookup_cache_types ---

var _ datasource.DataSource = &LookupCacheTypesDataSource{}

func NewLookupCacheTypesDataSource() datasource.DataSource { return &LookupCacheTypesDataSource{} }

type LookupCacheTypesDataSource struct {
	client *client.Client
}

type lookupCacheTypesDataSourceModel struct {
	Types []typeDescriptorModel `tfsdk:"types"`
}

func (d *LookupCacheTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_cache_types"
}

func (d *LookupCacheTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available Graylog lookup cache types. " +
			"Graylog exposes `default_config` (not full UI field metadata); keys are synthesized into `requested_configuration`. " +
			"Use this when configuring `graylog_lookup_cache.config`.",
		Attributes: map[string]schema.Attribute{
			"types": typeDescriptorsListSchema("Available lookup cache types."),
		},
	}
}

func (d *LookupCacheTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupCacheTypesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetLookupCacheTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list lookup cache types", err.Error())
		return
	}
	types, err := mapTypeDescriptors(ctx, client.SortedTypeDescriptors(result), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map lookup cache types", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &lookupCacheTypesDataSourceModel{Types: types})...)
}

// --- graylog_index_set_strategy_types ---

var _ datasource.DataSource = &IndexSetStrategyTypesDataSource{}

func NewIndexSetStrategyTypesDataSource() datasource.DataSource {
	return &IndexSetStrategyTypesDataSource{}
}

type IndexSetStrategyTypesDataSource struct {
	client *client.Client
}

type indexSetStrategyTypesDataSourceModel struct {
	Rotation  []typeDescriptorModel `tfsdk:"rotation"`
	Retention []typeDescriptorModel `tfsdk:"retention"`
}

func (d *IndexSetStrategyTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index_set_strategy_types"
}

func (d *IndexSetStrategyTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available Graylog index set rotation and retention strategies, " +
			"including `json_schema` properties as `requested_configuration` and `default_config`. " +
			"Use this when configuring `graylog_index_set.rotation_strategy` / `retention_strategy`.",
		Attributes: map[string]schema.Attribute{
			"rotation":  typeDescriptorsListSchema("Available rotation strategies."),
			"retention": typeDescriptorsListSchema("Available retention strategies."),
		},
	}
}

func (d *IndexSetStrategyTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IndexSetStrategyTypesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	rotation, err := d.client.GetRotationStrategies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list rotation strategies", err.Error())
		return
	}
	retention, err := d.client.GetRetentionStrategies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list retention strategies", err.Error())
		return
	}
	rotModels, err := mapTypeDescriptors(ctx, rotation, collapseRotationStrategyClass)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map rotation strategies", err.Error())
		return
	}
	retModels, err := mapTypeDescriptors(ctx, retention, collapseRetentionStrategyClass)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map retention strategies", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &indexSetStrategyTypesDataSourceModel{
		Rotation:  rotModels,
		Retention: retModels,
	})...)
}

// --- graylog_event_notification_types ---

var _ datasource.DataSource = &EventNotificationTypesDataSource{}

func NewEventNotificationTypesDataSource() datasource.DataSource {
	return &EventNotificationTypesDataSource{}
}

type EventNotificationTypesDataSource struct {
	client *client.Client
}

type eventNotificationTypesDataSourceModel struct {
	Types []typeDescriptorModel `tfsdk:"types"`
}

func (d *EventNotificationTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_notification_types"
}

func (d *EventNotificationTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog event notification types and configuration fields. " +
			"Includes built-in modern types (`http-notification-v1`, `email-notification-v1`, …) plus " +
			"legacy alarm callback types from `/events/notifications/legacy/types`. " +
			"Use this when configuring `graylog_event_notification.config`.",
		Attributes: map[string]schema.Attribute{
			"types": typeDescriptorsListSchema("Available event notification types."),
		},
	}
}

func (d *EventNotificationTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EventNotificationTypesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	legacy, err := d.client.GetLegacyEventNotificationTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list legacy event notification types", err.Error())
		return
	}
	merged := append([]client.TypeDescriptor{}, client.ModernEventNotificationTypes()...)
	merged = append(merged, client.SortedTypeDescriptors(legacy)...)
	types, err := mapTypeDescriptors(ctx, merged, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map event notification types", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &eventNotificationTypesDataSourceModel{Types: types})...)
}

// --- graylog_event_entity_types ---

var _ datasource.DataSource = &EventEntityTypesDataSource{}

func NewEventEntityTypesDataSource() datasource.DataSource { return &EventEntityTypesDataSource{} }

type EventEntityTypesDataSource struct {
	client *client.Client
}

type eventEntityTypesDataSourceModel struct {
	ProcessorTypes       []types.String `tfsdk:"processor_types"`
	FieldProviderTypes   []types.String `tfsdk:"field_provider_types"`
	StorageHandlerTypes  []types.String `tfsdk:"storage_handler_types"`
	AggregationFunctions []types.String `tfsdk:"aggregation_functions"`
}

func (d *EventEntityTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_entity_types"
}

func (d *EventEntityTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog event definition entity type catalogs " +
			"(processor, field provider, storage handler, aggregation functions). " +
			"Use this when configuring `graylog_event_definition` Dynamic attributes.",
		Attributes: map[string]schema.Attribute{
			"processor_types": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Event definition processor/config types (e.g. aggregation-v1).",
			},
			"field_provider_types": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Field provider types for event definition field_spec values.",
			},
			"storage_handler_types": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Storage handler types for event definition storage entries.",
			},
			"aggregation_functions": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Aggregation functions available to aggregation-v1 configs.",
			},
		},
	}
}

func (d *EventEntityTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EventEntityTypesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetEventEntityTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list event entity types", err.Error())
		return
	}
	data := eventEntityTypesDataSourceModel{
		ProcessorTypes:       stringListValues(result.ProcessorTypes),
		FieldProviderTypes:   stringListValues(result.FieldProviderTypes),
		StorageHandlerTypes:  stringListValues(result.StorageHandlerTypes),
		AggregationFunctions: stringListValues(result.AggregationFunctions),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func stringListValues(in []string) []types.String {
	out := make([]types.String, len(in))
	for i, s := range in {
		out[i] = types.StringValue(s)
	}
	return out
}
