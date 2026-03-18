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

// -- Single pipeline data source --

var _ datasource.DataSource = &PipelineDataSource{}

func NewPipelineDataSource() datasource.DataSource {
	return &PipelineDataSource{}
}

type PipelineDataSource struct {
	client *client.Client
}

type PipelineDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ModifiedAt  types.String `tfsdk:"modified_at"`
}

func (d *PipelineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (d *PipelineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single Graylog pipeline by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true, MarkdownDescription: "The pipeline ID."},
			"title":       schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"source":      schema.StringAttribute{Computed: true},
			"created_at":  schema.StringAttribute{Computed: true},
			"modified_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *PipelineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PipelineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PipelineDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipeline, err := d.client.GetPipeline(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read pipeline", err.Error())
		return
	}

	data.Title = types.StringValue(pipeline.Title)
	data.Description = types.StringValue(pipeline.Description)
	data.Source = types.StringValue(pipeline.Source)
	data.CreatedAt = types.StringValue(pipeline.CreatedAt)
	data.ModifiedAt = types.StringValue(pipeline.ModifiedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -- Pipelines list data source --

var _ datasource.DataSource = &PipelinesDataSource{}

func NewPipelinesDataSource() datasource.DataSource {
	return &PipelinesDataSource{}
}

type PipelinesDataSource struct {
	client *client.Client
}

type PipelinesDataSourceModel struct {
	Pipelines []PipelineDataSourceModel `tfsdk:"pipelines"`
}

func (d *PipelinesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipelines"
}

func (d *PipelinesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog processing pipelines.",
		Attributes: map[string]schema.Attribute{
			"pipelines": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"title":       schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"source":      schema.StringAttribute{Computed: true},
						"created_at":  schema.StringAttribute{Computed: true},
						"modified_at": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *PipelinesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PipelinesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	pipelines, err := d.client.GetPipelines(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list pipelines", err.Error())
		return
	}

	var data PipelinesDataSourceModel
	for _, p := range pipelines {
		data.Pipelines = append(data.Pipelines, PipelineDataSourceModel{
			ID:          types.StringValue(p.ID),
			Title:       types.StringValue(p.Title),
			Description: types.StringValue(p.Description),
			Source:      types.StringValue(p.Source),
			CreatedAt:   types.StringValue(p.CreatedAt),
			ModifiedAt:  types.StringValue(p.ModifiedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -- Single pipeline rule data source --

var _ datasource.DataSource = &PipelineRuleDataSource{}

func NewPipelineRuleDataSource() datasource.DataSource {
	return &PipelineRuleDataSource{}
}

type PipelineRuleDataSource struct {
	client *client.Client
}

func (d *PipelineRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_rule"
}

func (d *PipelineRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single Graylog pipeline rule by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true, MarkdownDescription: "The pipeline rule ID."},
			"title":       schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"source":      schema.StringAttribute{Computed: true},
			"created_at":  schema.StringAttribute{Computed: true},
			"modified_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *PipelineRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PipelineRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PipelineDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := d.client.GetPipelineRule(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read pipeline rule", err.Error())
		return
	}

	data.Title = types.StringValue(rule.Title)
	data.Description = types.StringValue(rule.Description)
	data.Source = types.StringValue(rule.Source)
	data.CreatedAt = types.StringValue(rule.CreatedAt)
	data.ModifiedAt = types.StringValue(rule.ModifiedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// -- Pipeline rules list data source --

var _ datasource.DataSource = &PipelineRulesDataSource{}

func NewPipelineRulesDataSource() datasource.DataSource {
	return &PipelineRulesDataSource{}
}

type PipelineRulesDataSource struct {
	client *client.Client
}

type PipelineRulesDataSourceModel struct {
	Rules []PipelineDataSourceModel `tfsdk:"rules"`
}

func (d *PipelineRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_rules"
}

func (d *PipelineRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog pipeline rules.",
		Attributes: map[string]schema.Attribute{
			"rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"title":       schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"source":      schema.StringAttribute{Computed: true},
						"created_at":  schema.StringAttribute{Computed: true},
						"modified_at": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *PipelineRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PipelineRulesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	rules, err := d.client.GetPipelineRules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list pipeline rules", err.Error())
		return
	}

	var data PipelineRulesDataSourceModel
	for _, r := range rules {
		data.Rules = append(data.Rules, PipelineDataSourceModel{
			ID:          types.StringValue(r.ID),
			Title:       types.StringValue(r.Title),
			Description: types.StringValue(r.Description),
			Source:      types.StringValue(r.Source),
			CreatedAt:   types.StringValue(r.CreatedAt),
			ModifiedAt:  types.StringValue(r.ModifiedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
