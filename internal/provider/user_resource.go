// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *client.Client
}

type UserResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
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

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Username.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password used only at create time.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "User email.",
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Full name.",
			},
			"first_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "First name.",
			},
			"last_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Last name.",
			},
			"roles": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "Role names assigned to user.",
			},
			"permissions": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Effective permissions reported by Graylog.",
			},
			"timezone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Timezone.",
			},
			"session_timeout_ms": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Session timeout in milliseconds.",
			},
			"service_account": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether user is a service account.",
			},
			"account_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current account status.",
			},
			"read_only": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether user is read only.",
			},
			"external": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether user is externally managed.",
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Password.IsNull() || data.Password.ValueString() == "" {
		resp.Diagnostics.AddError("Missing password", "password must be provided when creating a user.")
		return
	}

	roles, rDiags := extractStringSet(ctx, data.Roles)
	resp.Diagnostics.Append(rDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	firstName := data.FirstName.ValueString()
	lastName := data.LastName.ValueString()
	if firstName == "" {
		firstName = data.Username.ValueString()
	}
	if lastName == "" {
		lastName = "User"
	}

	err := r.client.CreateUser(ctx, &client.CreateUserRequest{
		Username:         data.Username.ValueString(),
		Password:         data.Password.ValueString(),
		Email:            data.Email.ValueString(),
		FullName:         data.FullName.ValueString(),
		FirstName:        firstName,
		LastName:         lastName,
		Roles:            nonNilStrings(roles),
		Permissions:      []string{},
		Timezone:         data.Timezone.ValueString(),
		SessionTimeoutMS: data.SessionTimeoutMS.ValueInt64(),
		ServiceAccount:   data.ServiceAccount.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create user", err.Error())
		return
	}

	user, err := r.client.GetUserByUsername(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created user", err.Error())
		return
	}

	data.ID = types.StringValue(user.ID)
	mapUserToModel(ctx, user, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read user", err.Error())
		return
	}

	// Keep password from prior state (write-only).
	statePassword := data.Password
	mapUserToModel(ctx, user, &data, &resp.Diagnostics)
	data.Password = statePassword
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel
	var state UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, rDiags := extractStringSet(ctx, data.Roles)
	resp.Diagnostics.Append(rDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	firstName := data.FirstName.ValueString()
	lastName := data.LastName.ValueString()
	if firstName == "" {
		firstName = state.FirstName.ValueString()
	}
	if lastName == "" {
		lastName = state.LastName.ValueString()
	}

	err := r.client.UpdateUser(ctx, state.ID.ValueString(), &client.UpdateUserRequest{
		Email:            data.Email.ValueString(),
		FullName:         data.FullName.ValueString(),
		FirstName:        firstName,
		LastName:         lastName,
		Roles:            nonNilStrings(roles),
		Permissions:      []string{},
		Timezone:         data.Timezone.ValueString(),
		SessionTimeoutMS: data.SessionTimeoutMS.ValueInt64(),
		ServiceAccount:   data.ServiceAccount.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update user", err.Error())
		return
	}

	current, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated user", err.Error())
		return
	}

	mapUserToModel(ctx, current, &data, &resp.Diagnostics)
	// Keep prior password for state continuity.
	data.Password = state.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete user", err.Error())
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapUserToModel(ctx context.Context, user *client.User, data *UserResourceModel, diags *diag.Diagnostics) {
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

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
