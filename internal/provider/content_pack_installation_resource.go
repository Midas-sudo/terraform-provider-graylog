// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &ContentPackInstallationResource{}
	_ resource.ResourceWithImportState = &ContentPackInstallationResource{}
)

func NewContentPackInstallationResource() resource.Resource {
	return &ContentPackInstallationResource{}
}

type ContentPackInstallationResource struct {
	client *client.Client
}

type ContentPackInstallationResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ContentPackID  types.String `tfsdk:"content_pack_id"`
	Revision       types.Int64  `tfsdk:"revision"`
	Comment        types.String `tfsdk:"comment"`
	ParametersJSON types.String `tfsdk:"parameters_json"`
}

func (r *ContentPackInstallationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_pack_installation"
}

func (r *ContentPackInstallationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog content pack installation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Installation ID returned by Graylog.",
			},
			"content_pack_id": schema.StringAttribute{
				Required: true,
			},
			"revision": schema.Int64Attribute{
				Required: true,
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional installation comment.",
			},
			"parameters_json": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JSON object with content pack installation parameters.",
			},
		},
	}
}

func (r *ContentPackInstallationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContentPackInstallationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContentPackInstallationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installReq, diags := contentPackInstallationRequestFromModel(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.InstallContentPack(ctx, data.ContentPackID.ValueString(), data.Revision.ValueInt64(), installReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to install content pack", err.Error())
		return
	}

	mapContentPackInstallationToResourceModel(created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentPackInstallationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContentPackInstallationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installations, err := r.client.GetContentPackInstallations(ctx, data.ContentPackID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list content pack installations", err.Error())
		return
	}

	for _, inst := range installations.Installations {
		if inst.ID == data.ID.ValueString() {
			mapContentPackInstallationToResourceModel(&inst, &data)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *ContentPackInstallationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContentPackInstallationResourceModel
	var state ContentPackInstallationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteContentPackInstallation(ctx, state.ContentPackID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete old content pack installation", err.Error())
		return
	}

	installReq, diags := contentPackInstallationRequestFromModel(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.InstallContentPack(ctx, data.ContentPackID.ValueString(), data.Revision.ValueInt64(), installReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to reinstall content pack", err.Error())
		return
	}

	mapContentPackInstallationToResourceModel(created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentPackInstallationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContentPackInstallationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteContentPackInstallation(ctx, data.ContentPackID.ValueString(), data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete content pack installation", err.Error())
	}
}

func (r *ContentPackInstallationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import identifier", "Use `content_pack_id/installation_id` for import.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("content_pack_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func contentPackInstallationRequestFromModel(data *ContentPackInstallationResourceModel) (*client.ContentPackInstallationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.ContentPackInstallationRequest{}

	if !data.Comment.IsNull() && !data.Comment.IsUnknown() {
		req.Comment = data.Comment.ValueString()
	}

	if !data.ParametersJSON.IsNull() && !data.ParametersJSON.IsUnknown() && data.ParametersJSON.ValueString() != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(data.ParametersJSON.ValueString()), &params); err != nil {
			diags.AddError("Invalid parameters_json", fmt.Sprintf("Failed to parse parameters_json: %v", err))
			return req, diags
		}
		req.Parameters = params
	}

	return req, diags
}
