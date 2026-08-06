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

var _ datasource.DataSource = &InputDataSource{}

func NewInputDataSource() datasource.DataSource {
	return &InputDataSource{}
}

type InputDataSource struct {
	client *client.Client
}

type InputDataSourceModel struct {
	ID            types.String  `tfsdk:"id"`
	Title         types.String  `tfsdk:"title"`
	Type          types.String  `tfsdk:"type"`
	Global        types.Bool    `tfsdk:"global"`
	Node          types.String  `tfsdk:"node"`
	Configuration types.Dynamic `tfsdk:"configuration"`
	CreatedAt     types.String  `tfsdk:"created_at"`
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
				MarkdownDescription: "The input type: short alias when known (e.g. `SyslogUDPInput`), otherwise the full Java type.",
			},
			"global": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the input runs on all nodes.",
			},
			"node": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The node ID if the input is local.",
			},
			"configuration": schema.DynamicAttribute{
				Computed:            true,
				MarkdownDescription: "The input configuration object.",
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
	data.Type = types.StringValue(collapseInputType(input.Type))
	data.Global = types.BoolValue(input.Global)
	data.CreatedAt = types.StringValue(input.CreatedAt)
	if input.Node != "" {
		data.Node = types.StringValue(input.Node)
	} else {
		data.Node = types.StringNull()
	}

	if input.Attributes != nil {
		dyn, d := interfaceToDynamic(ctx, input.Attributes)
		resp.Diagnostics.Append(d...)
		data.Configuration = dyn
	} else {
		empty, d := interfaceToDynamic(ctx, map[string]interface{}{})
		resp.Diagnostics.Append(d...)
		data.Configuration = empty
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

// inputListItemModel uses a JSON string for configuration because the Plugin
// Framework does not allow DynamicAttribute inside nested collections.
type inputListItemModel struct {
	ID            types.String `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Type          types.String `tfsdk:"type"`
	Global        types.Bool   `tfsdk:"global"`
	Node          types.String `tfsdk:"node"`
	Configuration types.String `tfsdk:"configuration"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

type InputsDataSourceModel struct {
	Inputs []inputListItemModel `tfsdk:"inputs"`
}

func (d *InputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inputs"
}

func (d *InputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog inputs. Nested `configuration` is a JSON string " +
			"(Plugin Framework limitation); use `graylog_input` for a typed object.",
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
						"configuration": schema.StringAttribute{Computed: true, MarkdownDescription: "JSON-encoded configuration object."},
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
		item := inputListItemModel{
			ID:        types.StringValue(input.ID),
			Title:     types.StringValue(input.Title),
			Type:      types.StringValue(collapseInputType(input.Type)),
			Global:    types.BoolValue(input.Global),
			CreatedAt: types.StringValue(input.CreatedAt),
		}
		if input.Node != "" {
			item.Node = types.StringValue(input.Node)
		} else {
			item.Node = types.StringNull()
		}
		if input.Attributes != nil {
			b, _ := json.Marshal(input.Attributes)
			item.Configuration = types.StringValue(string(b))
		} else {
			item.Configuration = types.StringValue("{}")
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

type InputTypesDataSourceModel struct {
	Types []typeDescriptorModel `tfsdk:"types"`
}

func (d *InputTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_input_types"
}

func (d *InputTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available Graylog input types and their `requested_configuration` fields. " +
			"Use this to discover required/optional keys for `graylog_input.configuration`.",
		Attributes: map[string]schema.Attribute{
			"types": typeDescriptorsListSchema("Available input types."),
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
	result, err := d.client.GetInputTypesAll(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list input types", err.Error())
		return
	}

	types, err := mapTypeDescriptors(ctx, client.SortedTypeDescriptors(result), collapseInputType)
	if err != nil {
		resp.Diagnostics.AddError("Failed to map input types", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &InputTypesDataSourceModel{Types: types})...)
}
