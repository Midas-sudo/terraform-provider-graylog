package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &DashboardDataSource{}

func NewDashboardDataSource() datasource.DataSource {
	return &DashboardDataSource{}
}

type DashboardDataSource struct {
	client *client.Client
}

type DashboardDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Title       types.String `tfsdk:"title"`
	Summary     types.String `tfsdk:"summary"`
	Description types.String `tfsdk:"description"`
	SearchID    types.String `tfsdk:"search_id"`
	PayloadJSON types.String `tfsdk:"payload_json"`
}

func (d *DashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (d *DashboardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog dashboard by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"type":        schema.StringAttribute{Computed: true},
			"title":       schema.StringAttribute{Computed: true},
			"summary":     schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"search_id":   schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *DashboardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *DashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DashboardDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view, err := d.client.GetView(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read dashboard", err.Error())
		return
	}

	mapDashboardDataSource(view, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &DashboardsDataSource{}

func NewDashboardsDataSource() datasource.DataSource {
	return &DashboardsDataSource{}
}

type DashboardsDataSource struct {
	client *client.Client
}

type DashboardsDataSourceModel struct {
	Dashboards []DashboardDataSourceModel `tfsdk:"dashboards"`
}

func (d *DashboardsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboards"
}

func (d *DashboardsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog dashboards.",
		Attributes: map[string]schema.Attribute{
			"dashboards": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"type":         schema.StringAttribute{Computed: true},
						"title":        schema.StringAttribute{Computed: true},
						"summary":      schema.StringAttribute{Computed: true},
						"description":  schema.StringAttribute{Computed: true},
						"search_id":    schema.StringAttribute{Computed: true},
						"payload_json": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *DashboardsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *DashboardsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetDashboards(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list dashboards", err.Error())
		return
	}

	var data DashboardsDataSourceModel
	for _, v := range result.Views {
		row := DashboardDataSourceModel{}
		mapDashboardDataSource(&v, &row)
		data.Dashboards = append(data.Dashboards, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapDashboardDataSource(v *client.View, d *DashboardDataSourceModel) {
	d.ID = types.StringValue(v.ID)
	d.Type = types.StringValue(v.Type)
	d.Title = types.StringValue(v.Title)
	d.Summary = types.StringValue(v.Summary)
	d.Description = types.StringValue(v.Description)
	d.SearchID = types.StringValue(v.SearchID)
	d.PayloadJSON = types.StringValue(marshalViewJSON(v))
}
