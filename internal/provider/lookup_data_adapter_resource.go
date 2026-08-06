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
	_ resource.Resource                 = &LookupDataAdapterResource{}
	_ resource.ResourceWithImportState  = &LookupDataAdapterResource{}
	_ resource.ResourceWithUpgradeState = &LookupDataAdapterResource{}
)

func NewLookupDataAdapterResource() resource.Resource {
	return &LookupDataAdapterResource{}
}

type LookupDataAdapterResource struct {
	client *client.Client
}

type LookupDataAdapterResourceModel struct {
	ID                    types.String  `tfsdk:"id"`
	Title                 types.String  `tfsdk:"title"`
	Name                  types.String  `tfsdk:"name"`
	Description           types.String  `tfsdk:"description"`
	Config                types.Dynamic `tfsdk:"config"`
	CustomErrorTTLEnabled types.Bool    `tfsdk:"custom_error_ttl_enabled"`
	CustomErrorTTL        types.Int64   `tfsdk:"custom_error_ttl"`
	CustomErrorTTLUnit    types.String  `tfsdk:"custom_error_ttl_unit"`
}

type lookupDataAdapterResourceModelV0 struct {
	ID                    types.String `tfsdk:"id"`
	Title                 types.String `tfsdk:"title"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	ConfigJSON            types.String `tfsdk:"config_json"`
	CustomErrorTTLEnabled types.Bool   `tfsdk:"custom_error_ttl_enabled"`
	CustomErrorTTL        types.Int64  `tfsdk:"custom_error_ttl"`
	CustomErrorTTLUnit    types.String `tfsdk:"custom_error_ttl_unit"`
}

func (r *LookupDataAdapterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_data_adapter"
}

func (r *LookupDataAdapterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a Graylog lookup data adapter.",
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
				Required: true,
				MarkdownDescription: "HCL object passed through to Graylog. Must include `type` (e.g. `csvfile`, `httpjsonpath`, `dnslookup`). " +
					"Discover keys via [`graylog_lookup_adapter_types`](../data-sources/lookup_adapter_types.md) " +
					"(`default_config` / synthesized `requested_configuration`).",
			},
			"custom_error_ttl_enabled": schema.BoolAttribute{
				Optional: true,
			},
			"custom_error_ttl": schema.Int64Attribute{
				Optional: true,
			},
			"custom_error_ttl_unit": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *LookupDataAdapterResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                       schema.StringAttribute{Computed: true},
					"title":                    schema.StringAttribute{Required: true},
					"name":                     schema.StringAttribute{Required: true},
					"description":              schema.StringAttribute{Optional: true},
					"config_json":              schema.StringAttribute{Required: true},
					"custom_error_ttl_enabled": schema.BoolAttribute{Optional: true},
					"custom_error_ttl":         schema.Int64Attribute{Optional: true},
					"custom_error_ttl_unit":    schema.StringAttribute{Optional: true},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior lookupDataAdapterResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				dyn, err := upgradeJSONStringAttr(prior.ConfigJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade lookup data adapter config", err.Error())
					return
				}
				upgraded := LookupDataAdapterResourceModel{
					ID:                    prior.ID,
					Title:                 prior.Title,
					Name:                  prior.Name,
					Description:           prior.Description,
					Config:                dyn,
					CustomErrorTTLEnabled: prior.CustomErrorTTLEnabled,
					CustomErrorTTL:        prior.CustomErrorTTL,
					CustomErrorTTLUnit:    prior.CustomErrorTTLUnit,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *LookupDataAdapterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LookupDataAdapterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LookupDataAdapterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adapterReq, diags := lookupDataAdapterFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateLookupDataAdapter(ctx, adapterReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create lookup data adapter", err.Error())
		return
	}

	resp.Diagnostics.Append(mapLookupDataAdapterToResourceModel(ctx, created, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupDataAdapterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LookupDataAdapterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetLookupDataAdapter(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read lookup data adapter", err.Error())
		return
	}

	resp.Diagnostics.Append(mapLookupDataAdapterToResourceModel(ctx, current, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupDataAdapterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LookupDataAdapterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adapterReq, diags := lookupDataAdapterFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateLookupDataAdapter(ctx, data.ID.ValueString(), adapterReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update lookup data adapter", err.Error())
		return
	}

	resp.Diagnostics.Append(mapLookupDataAdapterToResourceModel(ctx, updated, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LookupDataAdapterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LookupDataAdapterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLookupDataAdapter(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete lookup data adapter", err.Error())
	}
}

func (r *LookupDataAdapterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func lookupDataAdapterFromModel(ctx context.Context, data *LookupDataAdapterResourceModel) (*client.LookupDataAdapter, diag.Diagnostics) {
	cfg, diags := dynamicToMap(ctx, data.Config)
	if diags.HasError() {
		return nil, diags
	}

	adapter := &client.LookupDataAdapter{
		Title:  data.Title.ValueString(),
		Name:   data.Name.ValueString(),
		Config: cfg,
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		adapter.Description = data.Description.ValueString()
	}
	if !data.CustomErrorTTLEnabled.IsNull() && !data.CustomErrorTTLEnabled.IsUnknown() {
		v := data.CustomErrorTTLEnabled.ValueBool()
		adapter.CustomErrorTTLEnabled = &v
	}
	if !data.CustomErrorTTL.IsNull() && !data.CustomErrorTTL.IsUnknown() {
		v := data.CustomErrorTTL.ValueInt64()
		adapter.CustomErrorTTL = &v
	}
	if !data.CustomErrorTTLUnit.IsNull() && !data.CustomErrorTTLUnit.IsUnknown() {
		v := data.CustomErrorTTLUnit.ValueString()
		adapter.CustomErrorTTLUnit = &v
	}
	return adapter, diags
}
