package provider

import (
	"context"
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
	ContentPackID types.String `tfsdk:"content_pack_id"`
	Revision      types.Int64  `tfsdk:"revision"`
	Name          types.String `tfsdk:"name"`
	PayloadJSON   types.String `tfsdk:"payload_json"`
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
			"payload_json": schema.StringAttribute{
				Computed: true,
			},
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
	data.Name = types.StringValue(contentPack.Name)
	data.PayloadJSON = types.StringValue(marshalContentPackJSON(contentPack))
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
						"name":            schema.StringAttribute{Computed: true},
						"payload_json":    schema.StringAttribute{Computed: true},
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
			Name:          types.StringValue(cp.Name),
			PayloadJSON:   types.StringValue(marshalContentPackJSON(&cp)),
		}
		data.ContentPacks = append(data.ContentPacks, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
