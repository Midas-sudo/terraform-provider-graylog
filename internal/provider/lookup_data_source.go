package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &LookupDataAdapterDataSource{}

func NewLookupDataAdapterDataSource() datasource.DataSource {
	return &LookupDataAdapterDataSource{}
}

type LookupDataAdapterDataSource struct {
	client *client.Client
}

type LookupDataAdapterDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	PayloadJSON types.String `tfsdk:"payload_json"`
}

func (d *LookupDataAdapterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_data_adapter"
}

func (d *LookupDataAdapterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog lookup data adapter by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"title":       schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *LookupDataAdapterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupDataAdapterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LookupDataAdapterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	adapter, err := d.client.GetLookupDataAdapter(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read lookup data adapter", err.Error())
		return
	}

	mapLookupDataAdapterToDataSourceModel(adapter, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &LookupDataAdaptersDataSource{}

func NewLookupDataAdaptersDataSource() datasource.DataSource {
	return &LookupDataAdaptersDataSource{}
}

type LookupDataAdaptersDataSource struct {
	client *client.Client
}

type LookupDataAdaptersDataSourceModel struct {
	DataAdapters []LookupDataAdapterDataSourceModel `tfsdk:"data_adapters"`
}

func (d *LookupDataAdaptersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_data_adapters"
}

func (d *LookupDataAdaptersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog lookup data adapters.",
		Attributes: map[string]schema.Attribute{
			"data_adapters": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"title":        schema.StringAttribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"description":  schema.StringAttribute{Computed: true},
						"payload_json": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *LookupDataAdaptersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupDataAdaptersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetLookupDataAdapters(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list lookup data adapters", err.Error())
		return
	}

	var data LookupDataAdaptersDataSourceModel
	for _, adapter := range result.DataAdapters {
		row := LookupDataAdapterDataSourceModel{}
		mapLookupDataAdapterToDataSourceModel(&adapter, &row)
		data.DataAdapters = append(data.DataAdapters, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &LookupCacheDataSource{}

func NewLookupCacheDataSource() datasource.DataSource {
	return &LookupCacheDataSource{}
}

type LookupCacheDataSource struct {
	client *client.Client
}

type LookupCacheDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	PayloadJSON types.String `tfsdk:"payload_json"`
}

func (d *LookupCacheDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_cache"
}

func (d *LookupCacheDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog lookup cache by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"title":       schema.StringAttribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *LookupCacheDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupCacheDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LookupCacheDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cache, err := d.client.GetLookupCache(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read lookup cache", err.Error())
		return
	}

	mapLookupCacheToDataSourceModel(cache, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &LookupCachesDataSource{}

func NewLookupCachesDataSource() datasource.DataSource {
	return &LookupCachesDataSource{}
}

type LookupCachesDataSource struct {
	client *client.Client
}

type LookupCachesDataSourceModel struct {
	Caches []LookupCacheDataSourceModel `tfsdk:"caches"`
}

func (d *LookupCachesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_caches"
}

func (d *LookupCachesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog lookup caches.",
		Attributes: map[string]schema.Attribute{
			"caches": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"title":        schema.StringAttribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"description":  schema.StringAttribute{Computed: true},
						"payload_json": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *LookupCachesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupCachesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetLookupCaches(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list lookup caches", err.Error())
		return
	}

	var data LookupCachesDataSourceModel
	for _, cache := range result.Caches {
		row := LookupCacheDataSourceModel{}
		mapLookupCacheToDataSourceModel(&cache, &row)
		data.Caches = append(data.Caches, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &LookupTableDataSource{}

func NewLookupTableDataSource() datasource.DataSource {
	return &LookupTableDataSource{}
}

type LookupTableDataSource struct {
	client *client.Client
}

type LookupTableDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	CacheID       types.String `tfsdk:"cache_id"`
	DataAdapterID types.String `tfsdk:"data_adapter_id"`
	PayloadJSON   types.String `tfsdk:"payload_json"`
}

func (d *LookupTableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_table"
}

func (d *LookupTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog lookup table by ID.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true},
			"title":           schema.StringAttribute{Computed: true},
			"name":            schema.StringAttribute{Computed: true},
			"description":     schema.StringAttribute{Computed: true},
			"cache_id":        schema.StringAttribute{Computed: true},
			"data_adapter_id": schema.StringAttribute{Computed: true},
			"payload_json":    schema.StringAttribute{Computed: true},
		},
	}
}

func (d *LookupTableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LookupTableDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	table, err := d.client.GetLookupTable(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read lookup table", err.Error())
		return
	}

	mapLookupTableToDataSourceModel(table, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &LookupTablesDataSource{}

func NewLookupTablesDataSource() datasource.DataSource {
	return &LookupTablesDataSource{}
}

type LookupTablesDataSource struct {
	client *client.Client
}

type LookupTablesDataSourceModel struct {
	LookupTables []LookupTableDataSourceModel `tfsdk:"lookup_tables"`
}

func (d *LookupTablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lookup_tables"
}

func (d *LookupTablesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog lookup tables.",
		Attributes: map[string]schema.Attribute{
			"lookup_tables": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"title":           schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"description":     schema.StringAttribute{Computed: true},
						"cache_id":        schema.StringAttribute{Computed: true},
						"data_adapter_id": schema.StringAttribute{Computed: true},
						"payload_json":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *LookupTablesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LookupTablesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetLookupTables(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list lookup tables", err.Error())
		return
	}

	var data LookupTablesDataSourceModel
	for _, table := range result.LookupTables {
		row := LookupTableDataSourceModel{}
		mapLookupTableToDataSourceModel(&table, &row)
		data.LookupTables = append(data.LookupTables, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
