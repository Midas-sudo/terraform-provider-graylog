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

var _ datasource.DataSource = &ContentPackDataSource{}

func NewContentPackDataSource() datasource.DataSource {
	return &ContentPackDataSource{}
}

type ContentPackDataSource struct {
	client *client.Client
}

type ContentPackDataSourceModel struct {
	ContentPackID  types.String `tfsdk:"content_pack_id"`
	Revision       types.Int64  `tfsdk:"revision"`
	V              types.String `tfsdk:"v"`
	Name           types.String `tfsdk:"name"`
	Summary        types.String `tfsdk:"summary"`
	Description    types.String `tfsdk:"description"`
	Vendor         types.String `tfsdk:"vendor"`
	URL            types.String `tfsdk:"url"`
	ParametersJSON types.String `tfsdk:"parameters_json"`
	EntitiesJSON   types.String `tfsdk:"entities_json"`
}

func (d *ContentPackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_pack"
}

func (d *ContentPackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog content pack by ID and revision.",
		Attributes: map[string]schema.Attribute{
			"content_pack_id": schema.StringAttribute{
				Required: true,
			},
			"revision": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"v":               schema.StringAttribute{Computed: true},
			"summary":         schema.StringAttribute{Computed: true},
			"description":     schema.StringAttribute{Computed: true},
			"vendor":          schema.StringAttribute{Computed: true},
			"url":             schema.StringAttribute{Computed: true},
			"parameters_json": schema.StringAttribute{Computed: true},
			"entities_json":   schema.StringAttribute{Computed: true},
		},
	}
}

func (d *ContentPackDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *ContentPackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContentPackDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	contentPack, err := d.client.GetContentPack(ctx, data.ContentPackID.ValueString(), data.Revision.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read content pack", err.Error())
		return
	}

	data.ContentPackID = types.StringValue(contentPack.ID)
	data.Revision = types.Int64Value(contentPack.Rev)
	data.V = types.StringValue(contentPack.V)
	data.Name = types.StringValue(contentPack.Name)
	data.Summary = types.StringValue(contentPack.Summary)
	data.Description = types.StringValue(contentPack.Description)
	data.Vendor = types.StringValue(contentPack.Vendor)
	data.URL = types.StringValue(contentPack.URL)
	if b, err := json.Marshal(contentPack.Parameters); err == nil {
		data.ParametersJSON = types.StringValue(string(b))
	}
	if b, err := json.Marshal(contentPack.Entities); err == nil {
		data.EntitiesJSON = types.StringValue(string(b))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &ContentPacksDataSource{}

func NewContentPacksDataSource() datasource.DataSource {
	return &ContentPacksDataSource{}
}

type ContentPacksDataSource struct {
	client *client.Client
}

type ContentPacksDataSourceModel struct {
	ContentPacks []ContentPackDataSourceModel `tfsdk:"content_packs"`
}

func (d *ContentPacksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_packs"
}

func (d *ContentPacksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists latest Graylog content pack revisions.",
		Attributes: map[string]schema.Attribute{
			"content_packs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"content_pack_id": schema.StringAttribute{Computed: true},
						"revision":        schema.Int64Attribute{Computed: true},
						"v":               schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"summary":         schema.StringAttribute{Computed: true},
						"description":     schema.StringAttribute{Computed: true},
						"vendor":          schema.StringAttribute{Computed: true},
						"url":             schema.StringAttribute{Computed: true},
						"parameters_json": schema.StringAttribute{Computed: true},
						"entities_json":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ContentPacksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *ContentPacksDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetContentPacks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list content packs", err.Error())
		return
	}

	var data ContentPacksDataSourceModel
	for _, cp := range result.ContentPacks {
		row := ContentPackDataSourceModel{
			ContentPackID: types.StringValue(cp.ID),
			Revision:      types.Int64Value(cp.Rev),
			V:             types.StringValue(cp.V),
			Name:          types.StringValue(cp.Name),
			Summary:       types.StringValue(cp.Summary),
			Description:   types.StringValue(cp.Description),
			Vendor:        types.StringValue(cp.Vendor),
			URL:           types.StringValue(cp.URL),
		}
		if b, err := json.Marshal(cp.Parameters); err == nil {
			row.ParametersJSON = types.StringValue(string(b))
		}
		if b, err := json.Marshal(cp.Entities); err == nil {
			row.EntitiesJSON = types.StringValue(string(b))
		}
		data.ContentPacks = append(data.ContentPacks, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
