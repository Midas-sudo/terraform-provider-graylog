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

var _ datasource.DataSource = &OutputDataSource{}

func NewOutputDataSource() datasource.DataSource {
	return &OutputDataSource{}
}

type OutputDataSource struct {
	client *client.Client
}

type OutputDataSourceModel struct {
	ID            types.String  `tfsdk:"id"`
	Title         types.String  `tfsdk:"title"`
	Type          types.String  `tfsdk:"type"`
	Configuration types.Dynamic `tfsdk:"configuration"`
}

func (d *OutputDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_output"
}

func (d *OutputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog output by ID.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true},
			"title":         schema.StringAttribute{Computed: true},
			"type":          schema.StringAttribute{Computed: true},
			"configuration": schema.DynamicAttribute{Computed: true},
		},
	}
}

func (d *OutputDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *OutputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OutputDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := d.client.GetOutput(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read output", err.Error())
		return
	}

	resp.Diagnostics.Append(mapOutputToDataSourceModel(ctx, output, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &OutputsDataSource{}

func NewOutputsDataSource() datasource.DataSource {
	return &OutputsDataSource{}
}

type OutputsDataSource struct {
	client *client.Client
}

// outputListItemModel uses a JSON string for configuration because the Plugin
// Framework does not allow DynamicAttribute inside nested collections.
type outputListItemModel struct {
	ID            types.String `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Type          types.String `tfsdk:"type"`
	Configuration types.String `tfsdk:"configuration"`
}

type OutputsDataSourceModel struct {
	Outputs []outputListItemModel `tfsdk:"outputs"`
}

func (d *OutputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outputs"
}

func (d *OutputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog outputs. Nested `configuration` is a JSON string " +
			"(Plugin Framework limitation); use `graylog_output` for a typed object.",
		Attributes: map[string]schema.Attribute{
			"outputs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"title":         schema.StringAttribute{Computed: true},
						"type":          schema.StringAttribute{Computed: true},
						"configuration": schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded configuration object."},
					},
				},
			},
		},
	}
}

func (d *OutputsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *OutputsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetOutputs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list outputs", err.Error())
		return
	}

	var data OutputsDataSourceModel
	for _, output := range result.Outputs {
		row := outputListItemModel{
			ID:    types.StringValue(output.ID),
			Title: types.StringValue(output.Title),
			Type:  types.StringValue(collapseOutputType(output.Type)),
		}
		if output.Configuration != nil {
			b, _ := json.Marshal(output.Configuration)
			row.Configuration = types.StringValue(string(b))
		} else {
			row.Configuration = types.StringValue("{}")
		}
		data.Outputs = append(data.Outputs, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &ExtractorDataSource{}

func NewExtractorDataSource() datasource.DataSource {
	return &ExtractorDataSource{}
}

type ExtractorDataSource struct {
	client *client.Client
}

type ExtractorDataSourceModel struct {
	ID              types.String  `tfsdk:"id"`
	InputID         types.String  `tfsdk:"input_id"`
	Title           types.String  `tfsdk:"title"`
	ExtractorType   types.String  `tfsdk:"extractor_type"`
	CursorStrategy  types.String  `tfsdk:"cursor_strategy"`
	SourceField     types.String  `tfsdk:"source_field"`
	TargetField     types.String  `tfsdk:"target_field"`
	ConditionType   types.String  `tfsdk:"condition_type"`
	ConditionValue  types.String  `tfsdk:"condition_value"`
	Order           types.Int64   `tfsdk:"order"`
	ExtractorConfig types.Dynamic `tfsdk:"extractor_config"`
	Converters      types.Dynamic `tfsdk:"converters"`
}

type extractorListItemModel struct {
	ID              types.String `tfsdk:"id"`
	InputID         types.String `tfsdk:"input_id"`
	Title           types.String `tfsdk:"title"`
	ExtractorType   types.String `tfsdk:"extractor_type"`
	CursorStrategy  types.String `tfsdk:"cursor_strategy"`
	SourceField     types.String `tfsdk:"source_field"`
	TargetField     types.String `tfsdk:"target_field"`
	ConditionType   types.String `tfsdk:"condition_type"`
	ConditionValue  types.String `tfsdk:"condition_value"`
	Order           types.Int64  `tfsdk:"order"`
	ExtractorConfig types.String `tfsdk:"extractor_config"`
	Converters      types.String `tfsdk:"converters"`
}

func (d *ExtractorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extractor"
}

func (d *ExtractorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves an extractor by input and extractor IDs.",
		Attributes: map[string]schema.Attribute{
			"id":                    schema.StringAttribute{Required: true},
			"input_id":              schema.StringAttribute{Required: true},
			"title":                 schema.StringAttribute{Computed: true},
			"extractor_type":        schema.StringAttribute{Computed: true},
			"cursor_strategy":       schema.StringAttribute{Computed: true},
			"source_field":          schema.StringAttribute{Computed: true},
			"target_field":          schema.StringAttribute{Computed: true},
			"condition_type":        schema.StringAttribute{Computed: true},
			"condition_value":       schema.StringAttribute{Computed: true},
			"order":                 schema.Int64Attribute{Computed: true},
			"extractor_config": schema.DynamicAttribute{Computed: true},
			"converters":       schema.DynamicAttribute{Computed: true},
		},
	}
}

func (d *ExtractorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *ExtractorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExtractorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	extractor, err := d.client.GetExtractor(ctx, data.InputID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read extractor", err.Error())
		return
	}

	resp.Diagnostics.Append(mapExtractorToDataSourceModel(ctx, extractor, &data)...)
	data.InputID = types.StringValue(data.InputID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &ExtractorsDataSource{}

func NewExtractorsDataSource() datasource.DataSource {
	return &ExtractorsDataSource{}
}

type ExtractorsDataSource struct {
	client *client.Client
}

type ExtractorsDataSourceModel struct {
	InputID    types.String             `tfsdk:"input_id"`
	Extractors []extractorListItemModel `tfsdk:"extractors"`
}

func (d *ExtractorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extractors"
}

func (d *ExtractorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists extractors for a specific Graylog input. Nested `extractor_config`/`converters` " +
			"are JSON strings (Plugin Framework limitation); use `graylog_extractor` for typed objects.",
		Attributes: map[string]schema.Attribute{
			"input_id": schema.StringAttribute{Required: true},
			"extractors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"input_id":         schema.StringAttribute{Computed: true},
						"title":            schema.StringAttribute{Computed: true},
						"extractor_type":   schema.StringAttribute{Computed: true},
						"cursor_strategy":  schema.StringAttribute{Computed: true},
						"source_field":     schema.StringAttribute{Computed: true},
						"target_field":     schema.StringAttribute{Computed: true},
						"condition_type":   schema.StringAttribute{Computed: true},
						"condition_value":  schema.StringAttribute{Computed: true},
						"order":            schema.Int64Attribute{Computed: true},
						"extractor_config": schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded configuration object."},
						"converters":       schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded converters array."},
					},
				},
			},
		},
	}
}

func (d *ExtractorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *ExtractorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExtractorsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetExtractors(ctx, data.InputID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list extractors", err.Error())
		return
	}

	var rows []extractorListItemModel
	for _, extractor := range result.Extractors {
		extractorType := extractor.Type
		if extractor.ExtractorType != "" {
			extractorType = extractor.ExtractorType
		}
		rows = append(rows, extractorListItemModel{
			ID:              types.StringValue(extractor.ID),
			InputID:         data.InputID,
			Title:           types.StringValue(extractor.Title),
			ExtractorType:   types.StringValue(extractorType),
			CursorStrategy:  types.StringValue(extractor.CursorStrategy),
			SourceField:     types.StringValue(extractor.SourceField),
			TargetField:     types.StringValue(extractor.TargetField),
			ConditionType:   types.StringValue(extractor.ConditionType),
			ConditionValue:  types.StringValue(extractor.ConditionValue),
			Order:           types.Int64Value(extractor.Order),
			ExtractorConfig: extractorConfigJSONString(extractor.ExtractorConfig),
			Converters:      extractorConvertersJSONString(extractor.Converters),
		})
	}
	data.Extractors = rows
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &GrokPatternsDataSource{}

func NewGrokPatternsDataSource() datasource.DataSource {
	return &GrokPatternsDataSource{}
}

type GrokPatternsDataSource struct {
	client *client.Client
}

type GrokPatternDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Pattern types.String `tfsdk:"pattern"`
}

type GrokPatternsDataSourceModel struct {
	Patterns []GrokPatternDataSourceModel `tfsdk:"patterns"`
}

func (d *GrokPatternsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grok_patterns"
}

func (d *GrokPatternsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog grok patterns.",
		Attributes: map[string]schema.Attribute{
			"patterns": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"pattern": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *GrokPatternsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *GrokPatternsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetGrokPatterns(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list grok patterns", err.Error())
		return
	}

	var data GrokPatternsDataSourceModel
	for _, p := range result.Patterns {
		row := GrokPatternDataSourceModel{}
		mapGrokPatternToDataSourceModel(&p, &row)
		data.Patterns = append(data.Patterns, row)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
