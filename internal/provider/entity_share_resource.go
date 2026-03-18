// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &EntityShareResource{}
	_ resource.ResourceWithImportState = &EntityShareResource{}
)

func NewEntityShareResource() resource.Resource {
	return &EntityShareResource{}
}

type EntityShareResource struct {
	client *client.Client
}

type EntityShareResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	EntityGRN           types.String `tfsdk:"entity_grn"`
	GranteeCapabilities types.Map    `tfsdk:"grantee_capabilities"`
}

func (r *EntityShareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity_share"
}

func (r *EntityShareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Graylog entity shares using grantee capabilities.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier (same as entity_grn).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entity_grn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Entity GRN to share, e.g. `grn::::stream:<id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"grantee_capabilities": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
				MarkdownDescription: "Map of grantee GRN to capability (`view`, `manage`, or `own`).",
			},
		},
	}
}

func (r *EntityShareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EntityShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntityShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	caps, diags := mapToStringMap(ctx, data.GranteeCapabilities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateEntityShares(ctx, data.EntityGRN.ValueString(), &client.EntityShareRequest{
		SelectedGranteeCapabilities: caps,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create entity share", err.Error())
		return
	}

	mapEntityShareToModel(ctx, result, data.EntityGRN.ValueString(), &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntityShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EntityShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetEntityShares(ctx, data.EntityGRN.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read entity share", err.Error())
		return
	}

	mapEntityShareToModel(ctx, result, data.EntityGRN.ValueString(), &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntityShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EntityShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	caps, diags := mapToStringMap(ctx, data.GranteeCapabilities)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateEntityShares(ctx, data.EntityGRN.ValueString(), &client.EntityShareRequest{
		SelectedGranteeCapabilities: caps,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update entity share", err.Error())
		return
	}

	mapEntityShareToModel(ctx, result, data.EntityGRN.ValueString(), &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntityShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntityShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateEntityShares(ctx, data.EntityGRN.ValueString(), &client.EntityShareRequest{
		SelectedGranteeCapabilities: map[string]string{},
	})
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete entity share", err.Error())
	}
}

func (r *EntityShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("entity_grn"), req, resp)
}

func mapToStringMap(ctx context.Context, input types.Map) (map[string]string, diag.Diagnostics) {
	result := map[string]string{}
	var diags diag.Diagnostics
	if input.IsNull() || input.IsUnknown() {
		return result, diags
	}
	diags = input.ElementsAs(ctx, &result, false)
	return result, diags
}

func mapEntityShareToModel(ctx context.Context, share *client.EntityShareResponse, entityGRN string, data *EntityShareResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(entityGRN)
	data.EntityGRN = types.StringValue(entityGRN)

	caps := share.SelectedGranteeCapabilities
	if len(caps) == 0 {
		caps = map[string]string{}
		for _, active := range share.ActiveShares {
			caps[active.Grantee] = active.Capability
		}
	}

	values, d := types.MapValueFrom(ctx, types.StringType, caps)
	diags.Append(d...)
	data.GranteeCapabilities = values
}
