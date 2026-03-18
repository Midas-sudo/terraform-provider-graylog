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
	_ resource.Resource                = &LookupTableResource{}
	_ resource.ResourceWithImportState = &LookupTableResource{}
)

func NewLookupTableResource() resource.Resource {
	return &LookupTableResource{}
}

type LookupTableResource struct {
	client *client.Client
}

type LookupTableResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Title                  types.String `tfsdk:"title"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	CacheID                types.String `tfsdk:"cache_id"`
	DataAdapterID          types.String `tfsdk:"data_adapter_id"`
	DefaultSingleValue     types.String `tfsdk:"default_single_value"`
	DefaultSingleValueType types.String `tfsdk:"default_single_value_type"`
	DefaultMultiValue      types.String `tfsdk:"default_multi_value"`
	DefaultMultiValueType  types.String `tfsdk:"default_multi_value_type"`
}

func (r *LookupTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_table"
}

func (r *LookupTableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog lookup table.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title":       schema.StringAttribute{Required: true},
			"name":        schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true},
			"cache_id":    schema.StringAttribute{Required: true},
			"data_adapter_id": schema.StringAttribute{
				Required: true,
			},
			"default_single_value": schema.StringAttribute{
				Optional: true,
			},
			"default_single_value_type": schema.StringAttribute{
				Optional: true,
			},
			"default_multi_value": schema.StringAttribute{
				Optional: true,
			},
			"default_multi_value_type": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *LookupTableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *LookupTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LookupTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tableReq := lookupTableFromModel(&data)

	created, err := r.client.CreateLookupTable(ctx, tableReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create lookup table", err.Error())
		return
	}

	mapLookupTableToResourceModel(created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LookupTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetLookupTable(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read lookup table", err.Error())
		return
	}

	mapLookupTableToResourceModel(current, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LookupTableResourceModel
	var state LookupTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tableReq := lookupTableFromModel(&data)

	updated, err := r.client.UpdateLookupTable(ctx, state.ID.ValueString(), tableReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update lookup table", err.Error())
		return
	}

	mapLookupTableToResourceModel(updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LookupTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLookupTable(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete lookup table", err.Error())
	}
}

func (r *LookupTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func lookupTableFromModel(data *LookupTableResourceModel) *client.LookupTable {
	table := &client.LookupTable{
		Title:         data.Title.ValueString(),
		Name:          data.Name.ValueString(),
		CacheID:       data.CacheID.ValueString(),
		DataAdapterID: data.DataAdapterID.ValueString(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		table.Description = data.Description.ValueString()
	}
	if !data.DefaultSingleValue.IsNull() && !data.DefaultSingleValue.IsUnknown() {
		table.DefaultSingleValue = data.DefaultSingleValue.ValueString()
	}
	if !data.DefaultSingleValueType.IsNull() && !data.DefaultSingleValueType.IsUnknown() {
		table.DefaultSingleType = data.DefaultSingleValueType.ValueString()
	}
	if !data.DefaultMultiValue.IsNull() && !data.DefaultMultiValue.IsUnknown() {
		table.DefaultMultiValue = data.DefaultMultiValue.ValueString()
	}
	if !data.DefaultMultiValueType.IsNull() && !data.DefaultMultiValueType.IsUnknown() {
		table.DefaultMultiType = data.DefaultMultiValueType.ValueString()
	}
	return table
}
