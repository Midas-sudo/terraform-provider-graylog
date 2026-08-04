// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &IndexSetDataSource{}

func NewIndexSetDataSource() datasource.DataSource {
	return &IndexSetDataSource{}
}

type IndexSetDataSource struct {
	client *client.Client
}

type IndexSetDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Title                  types.String `tfsdk:"title"`
	Description            types.String `tfsdk:"description"`
	IndexPrefix            types.String `tfsdk:"index_prefix"`
	IndexOptimizationMaxNumSegments types.Int64 `tfsdk:"index_optimization_max_num_segments"`
	IndexOptimizationDisabled       types.Bool  `tfsdk:"index_optimization_disabled"`
	FieldTypeRefreshInterval        types.Int64 `tfsdk:"field_type_refresh_interval"`
	Shards                 types.Int64  `tfsdk:"shards"`
	Replicas               types.Int64  `tfsdk:"replicas"`
	Default                types.Bool   `tfsdk:"default"`
	IndexAnalyzer          types.String `tfsdk:"index_analyzer"`
	RotationStrategyClass  types.String `tfsdk:"rotation_strategy_class"`
	RetentionStrategyClass types.String `tfsdk:"retention_strategy_class"`
	RotationStrategyType   types.String `tfsdk:"rotation_strategy_type"`
	RetentionStrategyType  types.String `tfsdk:"retention_strategy_type"`
}

func (d *IndexSetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index_set"
}

func (d *IndexSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single Graylog index set by ID.",
		Attributes: map[string]schema.Attribute{
			"id":                       schema.StringAttribute{Required: true, MarkdownDescription: "Index set ID."},
			"title":                    schema.StringAttribute{Computed: true},
			"description":              schema.StringAttribute{Computed: true},
			"index_prefix":             schema.StringAttribute{Computed: true},
			"index_optimization_max_num_segments": schema.Int64Attribute{Computed: true},
			"index_optimization_disabled":         schema.BoolAttribute{Computed: true},
			"field_type_refresh_interval":         schema.Int64Attribute{Computed: true},
			"shards":                   schema.Int64Attribute{Computed: true},
			"replicas":                 schema.Int64Attribute{Computed: true},
			"default":                  schema.BoolAttribute{Computed: true},
			"index_analyzer":           schema.StringAttribute{Computed: true},
			"rotation_strategy_class": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Rotation strategy class (short name when known, otherwise the value returned by Graylog).",
			},
			"retention_strategy_class": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Retention strategy class (short name when known, otherwise the value returned by Graylog).",
			},
			"rotation_strategy_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Rotation strategy config type (short name when known).",
			},
			"retention_strategy_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Retention strategy config type (short name when known).",
			},
		},
	}
}

func (d *IndexSetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IndexSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IndexSetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := d.client.GetIndexSet(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read index set", err.Error())
		return
	}
	mapIndexSetDataSource(item, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapIndexSetDataSource(src *client.IndexSet, dst *IndexSetDataSourceModel) {
	dst.ID = types.StringValue(src.ID)
	dst.Title = types.StringValue(src.Title)
	dst.Description = types.StringValue(src.Description)
	dst.IndexPrefix = types.StringValue(src.IndexPrefix)
	dst.IndexOptimizationMaxNumSegments = types.Int64Value(src.IndexOptimizationMaxNumSegments)
	dst.IndexOptimizationDisabled = types.BoolValue(src.IndexOptimizationDisabled)
	dst.FieldTypeRefreshInterval = types.Int64Value(src.FieldTypeRefreshInterval)
	dst.Shards = types.Int64Value(src.Shards)
	dst.Replicas = types.Int64Value(src.Replicas)
	dst.Default = types.BoolValue(src.Default)
	dst.IndexAnalyzer = types.StringValue(src.IndexAnalyzer)
	dst.RotationStrategyClass = types.StringValue(collapseRotationStrategyClass(src.RotationStrategyClass))
	dst.RetentionStrategyClass = types.StringValue(collapseRetentionStrategyClass(src.RetentionStrategyClass))
	dst.RotationStrategyType = types.StringValue(collapseRotationStrategyConfigType(strategyMapType(src.RotationStrategy)))
	dst.RetentionStrategyType = types.StringValue(collapseRetentionStrategyConfigType(strategyMapType(src.RetentionStrategy)))
}

// IndexSetsDataSource lists index sets.
var _ datasource.DataSource = &IndexSetsDataSource{}

func NewIndexSetsDataSource() datasource.DataSource {
	return &IndexSetsDataSource{}
}

type IndexSetsDataSource struct {
	client *client.Client
}

type IndexSetsDataSourceModel struct {
	IndexSets []IndexSetDataSourceModel `tfsdk:"index_sets"`
}

func (d *IndexSetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index_sets"
}

func (d *IndexSetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog index sets.",
		Attributes: map[string]schema.Attribute{
			"index_sets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{Computed: true},
						"title":                    schema.StringAttribute{Computed: true},
						"description":              schema.StringAttribute{Computed: true},
						"index_prefix":             schema.StringAttribute{Computed: true},
						"index_optimization_max_num_segments": schema.Int64Attribute{Computed: true},
						"index_optimization_disabled":         schema.BoolAttribute{Computed: true},
						"field_type_refresh_interval":         schema.Int64Attribute{Computed: true},
						"shards":                   schema.Int64Attribute{Computed: true},
						"replicas":                 schema.Int64Attribute{Computed: true},
						"default":                  schema.BoolAttribute{Computed: true},
						"index_analyzer":           schema.StringAttribute{Computed: true},
						"rotation_strategy_class": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Rotation strategy class (short name when known).",
						},
						"retention_strategy_class": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Retention strategy class (short name when known).",
						},
						"rotation_strategy_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Rotation strategy config type (short name when known).",
						},
						"retention_strategy_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Retention strategy config type (short name when known).",
						},
					},
				},
			},
		},
	}
}

func (d *IndexSetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IndexSetsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetIndexSets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list index sets", err.Error())
		return
	}
	var data IndexSetsDataSourceModel
	for _, item := range result.IndexSets {
		row := IndexSetDataSourceModel{}
		mapIndexSetDataSource(&item, &row)
		data.IndexSets = append(data.IndexSets, row)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// IndexTemplateDataSource gets a single index template by index set id.
var _ datasource.DataSource = &IndexTemplateDataSource{}

func NewIndexTemplateDataSource() datasource.DataSource {
	return &IndexTemplateDataSource{}
}

type IndexTemplateDataSource struct {
	client *client.Client
}

type IndexTemplateDataSourceModel struct {
	IndexSetID    types.String `tfsdk:"index_set_id"`
	Name          types.String `tfsdk:"name"`
	IndexPatterns types.List   `tfsdk:"index_patterns"`
	Order         types.Int64  `tfsdk:"order"`
	SettingsJSON  types.String `tfsdk:"settings_json"`
	MappingsJSON  types.String `tfsdk:"mappings_json"`
}

func (d *IndexTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index_template"
}

func (d *IndexTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets Graylog index template for a given index set ID.",
		Attributes: map[string]schema.Attribute{
			"index_set_id":   schema.StringAttribute{Required: true, MarkdownDescription: "Index set ID."},
			"name":           schema.StringAttribute{Computed: true},
			"index_patterns": schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"order":          schema.Int64Attribute{Computed: true},
			"settings_json":  schema.StringAttribute{Computed: true},
			"mappings_json":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *IndexTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IndexTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IndexTemplateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	item, err := d.client.GetIndexTemplate(ctx, data.IndexSetID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read index template", err.Error())
		return
	}
	data.Name = types.StringValue(item.Name)
	list, ldiags := types.ListValueFrom(ctx, types.StringType, item.Template.IndexPatterns)
	resp.Diagnostics.Append(ldiags...)
	data.IndexPatterns = list
	data.Order = types.Int64Value(item.Template.Order)
	settings, _ := json.Marshal(item.Template.Settings)
	mappings, _ := json.Marshal(item.Template.Mappings)
	data.SettingsJSON = types.StringValue(string(settings))
	data.MappingsJSON = types.StringValue(string(mappings))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
