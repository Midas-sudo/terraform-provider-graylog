// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ resource.Resource = &StreamRuleResource{}
var _ resource.ResourceWithImportState = &StreamRuleResource{}

func NewStreamRuleResource() resource.Resource {
	return &StreamRuleResource{}
}

type StreamRuleResource struct {
	client *client.Client
}

type StreamRuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	StreamID    types.String `tfsdk:"stream_id"`
	Field       types.String `tfsdk:"field"`
	Value       types.String `tfsdk:"value"`
	Type        types.Int64  `tfsdk:"type"`
	Inverted    types.Bool   `tfsdk:"inverted"`
	Description types.String `tfsdk:"description"`
}

func (r *StreamRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream_rule"
}

func (r *StreamRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a rule on a Graylog stream. Stream rules determine which messages are routed to a stream.\n\n" +
			"Rule types: `1` = EXACT, `2` = REGEX, `3` = GREATER, `4` = SMALLER, `5` = PRESENCE, `6` = CONTAINS, `7` = ALWAYS_MATCH, `8` = MATCH_INPUT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The stream rule ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stream_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the stream this rule belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"field": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The message field to evaluate.",
			},
			"value": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The value to match against.",
			},
			"type": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "The rule type (1=EXACT, 2=REGEX, 3=GREATER, 4=SMALLER, 5=PRESENCE, 6=CONTAINS, 7=ALWAYS_MATCH, 8=MATCH_INPUT). Defaults to `1`.",
			},
			"inverted": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to invert the rule match. Defaults to `false`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the rule.",
			},
		},
	}
}

func (r *StreamRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StreamRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StreamRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := &client.StreamRule{
		Field:       data.Field.ValueString(),
		Value:       data.Value.ValueString(),
		Type:        int(data.Type.ValueInt64()),
		Inverted:    data.Inverted.ValueBool(),
		Description: data.Description.ValueString(),
	}

	result, err := r.client.CreateStreamRule(ctx, data.StreamID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create stream rule", err.Error())
		return
	}

	data.ID = types.StringValue(result.StreamRuleID)

	created, err := r.client.GetStreamRule(ctx, data.StreamID.ValueString(), result.StreamRuleID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created stream rule", err.Error())
		return
	}

	mapStreamRuleToModel(created, data.StreamID.ValueString(), &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StreamRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetStreamRule(ctx, data.StreamID.ValueString(), data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read stream rule", err.Error())
		return
	}

	mapStreamRuleToModel(rule, data.StreamID.ValueString(), &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StreamRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := &client.StreamRule{
		Field:       data.Field.ValueString(),
		Value:       data.Value.ValueString(),
		Type:        int(data.Type.ValueInt64()),
		Inverted:    data.Inverted.ValueBool(),
		Description: data.Description.ValueString(),
	}

	updated, err := r.client.UpdateStreamRule(ctx, data.StreamID.ValueString(), data.ID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update stream rule", err.Error())
		return
	}

	mapStreamRuleToModel(updated, data.StreamID.ValueString(), &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StreamRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStreamRule(ctx, data.StreamID.ValueString(), data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete stream rule", err.Error())
	}
}

// Import format: "stream_id/rule_id"
func (r *StreamRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			"Import ID must be in the format 'stream_id/rule_id'")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stream_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func mapStreamRuleToModel(rule *client.StreamRule, streamID string, data *StreamRuleResourceModel) {
	data.ID = types.StringValue(rule.ID)
	data.StreamID = types.StringValue(streamID)
	data.Field = types.StringValue(rule.Field)
	data.Value = types.StringValue(rule.Value)
	data.Type = types.Int64Value(int64(rule.Type))
	data.Inverted = types.BoolValue(rule.Inverted)
	if rule.Description != "" {
		data.Description = types.StringValue(rule.Description)
	}
}
