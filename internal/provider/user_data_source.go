// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *client.Client
}

type UserDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Username         types.String `tfsdk:"username"`
	Email            types.String `tfsdk:"email"`
	FullName         types.String `tfsdk:"full_name"`
	FirstName        types.String `tfsdk:"first_name"`
	LastName         types.String `tfsdk:"last_name"`
	Roles            types.Set    `tfsdk:"roles"`
	Permissions      types.Set    `tfsdk:"permissions"`
	Timezone         types.String `tfsdk:"timezone"`
	SessionTimeoutMS types.Int64  `tfsdk:"session_timeout_ms"`
	ServiceAccount   types.Bool   `tfsdk:"service_account"`
	AccountStatus    types.String `tfsdk:"account_status"`
	ReadOnly         types.Bool   `tfsdk:"read_only"`
	External         types.Bool   `tfsdk:"external"`
}

func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Graylog user by ID.",
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Required: true},
			"username":           schema.StringAttribute{Computed: true},
			"email":              schema.StringAttribute{Computed: true},
			"full_name":          schema.StringAttribute{Computed: true},
			"first_name":         schema.StringAttribute{Computed: true},
			"last_name":          schema.StringAttribute{Computed: true},
			"roles":              schema.SetAttribute{Computed: true, ElementType: types.StringType},
			"permissions":        schema.SetAttribute{Computed: true, ElementType: types.StringType},
			"timezone":           schema.StringAttribute{Computed: true},
			"session_timeout_ms": schema.Int64Attribute{Computed: true},
			"service_account":    schema.BoolAttribute{Computed: true},
			"account_status":     schema.StringAttribute{Computed: true},
			"read_only":          schema.BoolAttribute{Computed: true},
			"external":           schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := d.client.GetUser(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read user", err.Error())
		return
	}

	mapUserDataSource(ctx, user, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *client.Client
}

type UsersDataSourceModel struct {
	Users []UserDataSourceModel `tfsdk:"users"`
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog users.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                 schema.StringAttribute{Computed: true},
						"username":           schema.StringAttribute{Computed: true},
						"email":              schema.StringAttribute{Computed: true},
						"full_name":          schema.StringAttribute{Computed: true},
						"first_name":         schema.StringAttribute{Computed: true},
						"last_name":          schema.StringAttribute{Computed: true},
						"roles":              schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"permissions":        schema.SetAttribute{Computed: true, ElementType: types.StringType},
						"timezone":           schema.StringAttribute{Computed: true},
						"session_timeout_ms": schema.Int64Attribute{Computed: true},
						"service_account":    schema.BoolAttribute{Computed: true},
						"account_status":     schema.StringAttribute{Computed: true},
						"read_only":          schema.BoolAttribute{Computed: true},
						"external":           schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list users", err.Error())
		return
	}

	var data UsersDataSourceModel
	for _, user := range result.Users {
		row := UserDataSourceModel{}
		mapUserDataSource(ctx, &user, &row, &resp.Diagnostics)
		data.Users = append(data.Users, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapUserDataSource(ctx context.Context, user *client.User, data *UserDataSourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(user.ID)
	data.Username = types.StringValue(user.Username)
	data.Email = types.StringValue(user.Email)
	data.FullName = types.StringValue(user.FullName)
	data.FirstName = types.StringValue(user.FirstName)
	data.LastName = types.StringValue(user.LastName)
	data.Timezone = types.StringValue(user.Timezone)
	data.SessionTimeoutMS = types.Int64Value(user.SessionTimeoutMS)
	data.ServiceAccount = types.BoolValue(user.ServiceAccount)
	data.AccountStatus = types.StringValue(user.AccountStatus)
	data.ReadOnly = types.BoolValue(user.ReadOnly)
	data.External = types.BoolValue(user.External)

	roleSet, rDiags := types.SetValueFrom(ctx, types.StringType, nonNilStrings(user.Roles))
	diags.Append(rDiags...)
	data.Roles = roleSet

	permSet, pDiags := types.SetValueFrom(ctx, types.StringType, nonNilStrings(user.Permissions))
	diags.Append(pDiags...)
	data.Permissions = permSet
}
