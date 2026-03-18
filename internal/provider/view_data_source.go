package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &ViewDataSource{}

func NewViewDataSource() datasource.DataSource {
	return &ViewDataSource{}
}

type ViewDataSource struct {
	client *client.Client
}

type ViewDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Title       types.String `tfsdk:"title"`
	Summary     types.String `tfsdk:"summary"`
	Description types.String `tfsdk:"description"`
	SearchID    types.String `tfsdk:"search_id"`
	PayloadJSON types.String `tfsdk:"payload_json"`
}

func (d *ViewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_view"
}

func (d *ViewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog view by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"type":        schema.StringAttribute{Computed: true},
			"title":       schema.StringAttribute{Computed: true},
			"summary":     schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"search_id":   schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Raw JSON payload for the view.",
			},
		},
	}
}

func (d *ViewDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ViewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ViewDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view, err := d.client.GetView(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read view", err.Error())
		return
	}

	mapViewDataSource(view, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &ViewsDataSource{}

func NewViewsDataSource() datasource.DataSource {
	return &ViewsDataSource{}
}

type ViewsDataSource struct {
	client *client.Client
}

type ViewsDataSourceModel struct {
	Views []ViewDataSourceModel `tfsdk:"views"`
}

func (d *ViewsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_views"
}

func (d *ViewsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Graylog views.",
		Attributes: map[string]schema.Attribute{
			"views": schema.ListNestedAttribute{
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

func (d *ViewsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ViewsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetViews(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list views", err.Error())
		return
	}

	var data ViewsDataSourceModel
	for _, v := range result.Views {
		row := ViewDataSourceModel{}
		mapViewDataSource(&v, &row)
		data.Views = append(data.Views, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapViewDataSource(v *client.View, d *ViewDataSourceModel) {
	d.ID = types.StringValue(v.ID)
	d.Type = types.StringValue(v.Type)
	d.Title = types.StringValue(v.Title)
	d.Summary = types.StringValue(v.Summary)
	d.Description = types.StringValue(v.Description)
	d.SearchID = types.StringValue(v.SearchID)
	d.PayloadJSON = types.StringValue(marshalViewJSON(v))
}
