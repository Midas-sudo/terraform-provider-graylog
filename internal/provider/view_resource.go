// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &ViewResource{}
	_ resource.ResourceWithImportState = &ViewResource{}
)

func NewViewResource() resource.Resource {
	return &ViewResource{}
}

type ViewResource struct {
	client *client.Client
}

type ViewResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	Title          types.String `tfsdk:"title"`
	Summary        types.String `tfsdk:"summary"`
	Description    types.String `tfsdk:"description"`
	SearchID       types.String `tfsdk:"search_id"`
	StateJSON      types.String `tfsdk:"state_json"`
	PropertiesJSON types.String `tfsdk:"properties_json"`
	RequiresJSON   types.String `tfsdk:"requires_json"`
}

func (r *ViewResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_view"
}

func (r *ViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog view object.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "View type reported by Graylog (e.g. `SEARCH`).",
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "View title.",
			},
			"summary": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "View summary.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "View description.",
			},
			"search_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Associated search ID.",
			},
			"state_json": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JSON object describing the view state.",
			},
			"properties_json": schema.StringAttribute{
				Optional: true,
			},
			"requires_json": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *ViewResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	viewReq, diags := viewFromModel(&data, "SEARCH")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateView(ctx, viewReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create view", err.Error())
		return
	}

	mapViewToModel(created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetView(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read view", err.Error())
		return
	}

	mapViewToModel(current, &data)
	populateViewJSONFields(current, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ViewResourceModel
	var state ViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	viewReq, diags := viewFromModel(&data, "SEARCH")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateView(ctx, state.ID.ValueString(), viewReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update view", err.Error())
		return
	}

	mapViewToModel(updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteView(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete view", err.Error())
	}
}

func (r *ViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
