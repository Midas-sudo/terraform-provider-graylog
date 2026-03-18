package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &RoleDataSource{}

func NewRoleDataSource() datasource.DataSource {
	return &RoleDataSource{}
}

type RoleDataSource struct {
	client *client.Client
}

type RoleDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
	ReadOnly    types.Bool   `tfsdk:"read_only"`
}

func (d *RoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *RoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog role by name.",
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{Required: true, MarkdownDescription: "Role name."},
			"description": schema.StringAttribute{Computed: true},
			"permissions": schema.SetAttribute{Computed: true, ElementType: types.StringType},
			"read_only":   schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *RoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.client.GetRole(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read role", err.Error())
		return
	}

	data.Name = types.StringValue(role.Name)
	data.Description = types.StringValue(role.Description)
	data.ReadOnly = types.BoolValue(role.ReadOnly)
	perms, diags := types.SetValueFrom(ctx, types.StringType, role.Permissions)
	resp.Diagnostics.Append(diags...)
	data.Permissions = perms

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &RolesDataSource{}

func NewRolesDataSource() datasource.DataSource {
	return &RolesDataSource{}
}

type RolesDataSource struct {
	client *client.Client
}

type RolesDataSourceModel struct {
	Roles []RoleDataSourceModel `tfsdk:"roles"`
}

func (d *RolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *RolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog roles.",
		Attributes: map[string]schema.Attribute{
			"roles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"permissions": schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"read_only":   schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *RolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RolesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetRoles(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list roles", err.Error())
		return
	}

	var data RolesDataSourceModel
	for _, role := range result.Roles {
		row := RoleDataSourceModel{
			Name:        types.StringValue(role.Name),
			Description: types.StringValue(role.Description),
			ReadOnly:    types.BoolValue(role.ReadOnly),
		}
		perms, diags := types.SetValueFrom(ctx, types.StringType, role.Permissions)
		resp.Diagnostics.Append(diags...)
		row.Permissions = perms
		data.Roles = append(data.Roles, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
