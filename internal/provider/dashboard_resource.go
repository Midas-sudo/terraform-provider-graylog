// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &DashboardResource{}
	_ resource.ResourceWithImportState = &DashboardResource{}
)

func NewDashboardResource() resource.Resource {
	return &DashboardResource{}
}

type DashboardResource struct {
	client *client.Client
}

type DashboardResourceModel struct {
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

func (r *DashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *DashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog dashboard object.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Dashboard type reported by Graylog (`DASHBOARD`).",
			},
			"title": schema.StringAttribute{
				Required: true,
			},
			"summary": schema.StringAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"search_id": schema.StringAttribute{
				Required: true,
			},
			"state_json": schema.StringAttribute{
				Optional: true,
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

func (r *DashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	viewReq, diags := dashboardFromModel(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateView(ctx, viewReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create dashboard", err.Error())
		return
	}

	mapDashboardToModel(created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DashboardResourceModel
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
		resp.Diagnostics.AddError("Failed to read dashboard", err.Error())
		return
	}

	mapDashboardToModel(current, &data)
	populateDashboardJSONFields(current, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DashboardResourceModel
	var state DashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	viewReq, diags := dashboardFromModel(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateView(ctx, state.ID.ValueString(), viewReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update dashboard", err.Error())
		return
	}

	mapDashboardToModel(updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteView(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete dashboard", err.Error())
	}
}

func (r *DashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func dashboardFromModel(data *DashboardResourceModel) (*client.View, diag.Diagnostics) {
	converted := &ViewResourceModel{
		Title:          data.Title,
		Summary:        data.Summary,
		Description:    data.Description,
		SearchID:       data.SearchID,
		StateJSON:      data.StateJSON,
		PropertiesJSON: data.PropertiesJSON,
		RequiresJSON:   data.RequiresJSON,
	}
	return viewFromModel(converted, "DASHBOARD")
}

func populateDashboardJSONFields(view *client.View, data *DashboardResourceModel) {
	converted := &ViewResourceModel{
		StateJSON:      data.StateJSON,
		PropertiesJSON: data.PropertiesJSON,
		RequiresJSON:   data.RequiresJSON,
	}
	populateViewJSONFields(view, converted)
	data.StateJSON = converted.StateJSON
	data.PropertiesJSON = converted.PropertiesJSON
	data.RequiresJSON = converted.RequiresJSON
}
