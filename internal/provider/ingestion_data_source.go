package provider

import (
	"context"
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
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Type        types.String `tfsdk:"type"`
	PayloadJSON types.String `tfsdk:"payload_json"`
}

func (d *OutputDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_output"
}

func (d *OutputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog output by ID.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Required: true},
			"title":        schema.StringAttribute{Computed: true},
			"type":         schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{Computed: true},
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

	mapOutputToDataSourceModel(output, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &OutputsDataSource{}

func NewOutputsDataSource() datasource.DataSource {
	return &OutputsDataSource{}
}

type OutputsDataSource struct {
	client *client.Client
}

type OutputsDataSourceModel struct {
	Outputs []OutputDataSourceModel `tfsdk:"outputs"`
}

func (d *OutputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outputs"
}

func (d *OutputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog outputs.",
		Attributes: map[string]schema.Attribute{
			"outputs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"title":        schema.StringAttribute{Computed: true},
						"type":         schema.StringAttribute{Computed: true},
						"payload_json": schema.StringAttribute{Computed: true},
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
		row := OutputDataSourceModel{}
		mapOutputToDataSourceModel(&output, &row)
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
	ID            types.String `tfsdk:"id"`
	InputID       types.String `tfsdk:"input_id"`
	Title         types.String `tfsdk:"title"`
	ExtractorType types.String `tfsdk:"extractor_type"`
	PayloadJSON   types.String `tfsdk:"payload_json"`
}

func (d *ExtractorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extractor"
}

func (d *ExtractorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves an extractor by input and extractor IDs.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Required: true},
			"input_id":       schema.StringAttribute{Required: true},
			"title":          schema.StringAttribute{Computed: true},
			"extractor_type": schema.StringAttribute{Computed: true},
			"payload_json":   schema.StringAttribute{Computed: true},
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

	mapExtractorToDataSourceModel(extractor, &data)
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
	InputID    types.String               `tfsdk:"input_id"`
	Extractors []ExtractorDataSourceModel `tfsdk:"extractors"`
}

func (d *ExtractorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extractors"
}

func (d *ExtractorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists extractors for a specific Graylog input.",
		Attributes: map[string]schema.Attribute{
			"input_id": schema.StringAttribute{Required: true},
			"extractors": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"input_id":       schema.StringAttribute{Computed: true},
						"title":          schema.StringAttribute{Computed: true},
						"extractor_type": schema.StringAttribute{Computed: true},
						"payload_json":   schema.StringAttribute{Computed: true},
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

	var rows []ExtractorDataSourceModel
	for _, extractor := range result.Extractors {
		row := ExtractorDataSourceModel{
			InputID: data.InputID,
		}
		mapExtractorToDataSourceModel(&extractor, &row)
		row.InputID = data.InputID
		rows = append(rows, row)
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
