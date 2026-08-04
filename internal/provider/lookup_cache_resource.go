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
	_ resource.Resource                 = &LookupCacheResource{}
	_ resource.ResourceWithImportState  = &LookupCacheResource{}
	_ resource.ResourceWithUpgradeState = &LookupCacheResource{}
)

func NewLookupCacheResource() resource.Resource {
	return &LookupCacheResource{}
}

type LookupCacheResource struct {
	client *client.Client
}

type LookupCacheResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Title       types.String  `tfsdk:"title"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	Config      types.Dynamic `tfsdk:"config"`
}

type lookupCacheResourceModelV0 struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ConfigJSON  types.String `tfsdk:"config_json"`
}

func (r *LookupCacheResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_cache"
}

func (r *LookupCacheResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a Graylog lookup cache.",
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
			"config": schema.DynamicAttribute{
				Required:            true,
				MarkdownDescription: "Cache-specific configuration object.",
			},
		},
	}
}

func (r *LookupCacheResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":          schema.StringAttribute{Computed: true},
					"title":       schema.StringAttribute{Required: true},
					"name":        schema.StringAttribute{Required: true},
					"description": schema.StringAttribute{Optional: true},
					"config_json": schema.StringAttribute{Required: true},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior lookupCacheResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				dyn, err := upgradeJSONStringAttr(prior.ConfigJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade lookup cache config", err.Error())
					return
				}
				upgraded := LookupCacheResourceModel{
					ID:          prior.ID,
					Title:       prior.Title,
					Name:        prior.Name,
					Description: prior.Description,
					Config:      dyn,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *LookupCacheResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LookupCacheResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LookupCacheResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cacheReq, diags := lookupCacheFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateLookupCache(ctx, cacheReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create lookup cache", err.Error())
		return
	}

	resp.Diagnostics.Append(mapLookupCacheToResourceModel(ctx, created, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupCacheResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LookupCacheResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetLookupCache(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read lookup cache", err.Error())
		return
	}

	resp.Diagnostics.Append(mapLookupCacheToResourceModel(ctx, current, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupCacheResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LookupCacheResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cacheReq, diags := lookupCacheFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateLookupCache(ctx, data.ID.ValueString(), cacheReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update lookup cache", err.Error())
		return
	}

	resp.Diagnostics.Append(mapLookupCacheToResourceModel(ctx, updated, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupCacheResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LookupCacheResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLookupCache(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete lookup cache", err.Error())
	}
}

func (r *LookupCacheResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func lookupCacheFromModel(ctx context.Context, data *LookupCacheResourceModel) (*client.LookupCache, diag.Diagnostics) {
	cfg, diags := dynamicToMap(ctx, data.Config)
	if diags.HasError() {
		return nil, diags
	}

	cache := &client.LookupCache{
		Title:  data.Title.ValueString(),
		Name:   data.Name.ValueString(),
		Config: cfg,
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		cache.Description = data.Description.ValueString()
	}
	return cache, diags
}
