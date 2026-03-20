// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &InputResource{}
	_ resource.ResourceWithImportState = &InputResource{}
)

func NewInputResource() resource.Resource {
	return &InputResource{}
}

type InputResource struct {
	client *client.Client
}

type InputResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Type          types.String `tfsdk:"type"`
	Global        types.Bool   `tfsdk:"global"`
	Node          types.String `tfsdk:"node"`
	Configuration types.String `tfsdk:"configuration"`
	StaticFields  types.Map    `tfsdk:"static_fields"`
}

func (r *InputResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_input"
}

func (r *InputResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog input. Inputs define how Graylog receives log messages " +
			"(e.g. Syslog UDP, GELF TCP, Beats, etc.).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The input ID assigned by Graylog.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A descriptive name for the input.",
			},
			"type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Graylog input type: use the short class name (e.g. `SyslogUDPInput`, `GELFUDPInput`) " +
					"or the full Java type (`org.graylog2.inputs....`). Short names are expanded on create/update " +
					"and collapsed in state when they match a known built-in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"global": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the input runs on all nodes (`true`) or a specific node (`false`). Defaults to `true`.",
			},
			"node": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Node ID to run a local input on. Required when `global` is `false`.",
			},
			"configuration": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Input configuration as a JSON string. The schema depends on the input `type`.",
			},
			"static_fields": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Static fields to add to every message received by this input.",
			},
		},
	}
}

func (r *InputResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InputResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InputResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, diags := parseJSONConfig(data.Configuration.ValueString())
	if diags != nil {
		resp.Diagnostics.AddError("Invalid configuration JSON", diags.Error())
		return
	}

	createReq := &client.InputCreateRequest{
		Title:         data.Title.ValueString(),
		Type:          expandInputType(data.Type.ValueString()),
		Global:        data.Global.ValueBool(),
		Node:          data.Node.ValueString(),
		Configuration: config,
	}

	result, err := r.client.CreateInput(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create input", err.Error())
		return
	}

	data.ID = types.StringValue(result.ID)

	input, err := r.client.GetInput(ctx, result.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created input", err.Error())
		return
	}

	mapInputToModel(input, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InputResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InputResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := r.client.GetInput(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read input", err.Error())
		return
	}

	mapInputToModel(input, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InputResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InputResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, diags := parseJSONConfig(data.Configuration.ValueString())
	if diags != nil {
		resp.Diagnostics.AddError("Invalid configuration JSON", diags.Error())
		return
	}

	updateReq := &client.InputUpdateRequest{
		Title:         data.Title.ValueString(),
		Type:          expandInputType(data.Type.ValueString()),
		Global:        data.Global.ValueBool(),
		Node:          data.Node.ValueString(),
		Configuration: config,
	}

	err := r.client.UpdateInput(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update input", err.Error())
		return
	}

	input, err := r.client.GetInput(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated input", err.Error())
		return
	}

	mapInputToModel(input, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InputResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InputResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteInput(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete input", err.Error())
	}
}

func (r *InputResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapInputToModel(input *client.Input, data *InputResourceModel) {
	data.ID = types.StringValue(input.ID)
	data.Title = types.StringValue(input.Title)
	data.Type = types.StringValue(collapseInputType(input.Type))
	data.Global = types.BoolValue(input.Global)

	if input.Node != "" {
		data.Node = types.StringValue(input.Node)
	}

	if input.Attributes != nil {
		configJSON, _ := json.Marshal(input.Attributes)
		data.Configuration = types.StringValue(string(configJSON))
	}

	if len(input.StaticFields) > 0 {
		elems := make(map[string]types.String, len(input.StaticFields))
		for k, v := range input.StaticFields {
			elems[k] = types.StringValue(v)
		}
		// Preserve existing static_fields even when empty from API
	}
}

func parseJSONConfig(s string) (map[string]interface{}, error) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(s), &config); err != nil {
		return nil, fmt.Errorf("configuration must be valid JSON: %w", err)
	}
	return config, nil
}
