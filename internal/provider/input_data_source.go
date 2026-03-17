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

var _ datasource.DataSource = &InputDataSource{}

func NewInputDataSource() datasource.DataSource {
	return &InputDataSource{}
}

type InputDataSource struct {
	client *client.Client
}

type InputDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Type          types.String `tfsdk:"type"`
	Global        types.Bool   `tfsdk:"global"`
	Node          types.String `tfsdk:"node"`
	Configuration types.String `tfsdk:"configuration"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (d *InputDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_input"
}

func (d *InputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single Graylog input by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The input ID.",
			},
			"title": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The input title.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The input type class name.",
			},
			"global": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the input runs on all nodes.",
			},
			"node": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The node ID if the input is local.",
			},
			"configuration": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The input configuration as a JSON string.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the input was created.",
			},
		},
	}
}

func (d *InputDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InputDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := d.client.GetInput(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read input", err.Error())
		return
	}

	data.Title = types.StringValue(input.Title)
	data.Type = types.StringValue(input.Type)
	data.Global = types.BoolValue(input.Global)
	data.CreatedAt = types.StringValue(input.CreatedAt)
	if input.Node != "" {
		data.Node = types.StringValue(input.Node)
	} else {
		data.Node = types.StringNull()
	}

	if input.Attributes != nil {
		configJSON, _ := json.Marshal(input.Attributes)
		data.Configuration = types.StringValue(string(configJSON))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// InputsDataSource lists all inputs.
var _ datasource.DataSource = &InputsDataSource{}

func NewInputsDataSource() datasource.DataSource {
	return &InputsDataSource{}
}

type InputsDataSource struct {
	client *client.Client
}

type InputsDataSourceModel struct {
	Inputs []InputDataSourceModel `tfsdk:"inputs"`
}

func (d *InputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inputs"
}

func (d *InputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog inputs.",
		Attributes: map[string]schema.Attribute{
			"inputs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"title":         schema.StringAttribute{Computed: true},
						"type":          schema.StringAttribute{Computed: true},
						"global":        schema.BoolAttribute{Computed: true},
						"node":          schema.StringAttribute{Computed: true},
						"configuration": schema.StringAttribute{Computed: true},
						"created_at":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *InputsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InputsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetInputs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list inputs", err.Error())
		return
	}

	var data InputsDataSourceModel
	for _, input := range result.Inputs {
		item := InputDataSourceModel{
			ID:        types.StringValue(input.ID),
			Title:     types.StringValue(input.Title),
			Type:      types.StringValue(input.Type),
			Global:    types.BoolValue(input.Global),
			CreatedAt: types.StringValue(input.CreatedAt),
		}
		if input.Node != "" {
			item.Node = types.StringValue(input.Node)
		} else {
			item.Node = types.StringNull()
		}
		if input.Attributes != nil {
			configJSON, _ := json.Marshal(input.Attributes)
			item.Configuration = types.StringValue(string(configJSON))
		}
		data.Inputs = append(data.Inputs, item)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// InputTypesDataSource lists available input types.
var _ datasource.DataSource = &InputTypesDataSource{}

func NewInputTypesDataSource() datasource.DataSource {
	return &InputTypesDataSource{}
}

type InputTypesDataSource struct {
	client *client.Client
}

type InputTypeModel struct {
	Type string `tfsdk:"type"`
	Name string `tfsdk:"name"`
}

type InputTypesDataSourceModel struct {
	Types []InputTypeModel `tfsdk:"types"`
}

func (d *InputTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_input_types"
}

func (d *InputTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all available Graylog input types.",
		Attributes: map[string]schema.Attribute{
			"types": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{Computed: true, MarkdownDescription: "The input type class name."},
						"name": schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable name of the input type."},
					},
				},
			},
		},
	}
}

func (d *InputTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InputTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetInputTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list input types", err.Error())
		return
	}

	var data InputTypesDataSourceModel
	for typeName, typeInfo := range result.Types {
		data.Types = append(data.Types, InputTypeModel{
			Type: typeName,
			Name: typeInfo.Name,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
