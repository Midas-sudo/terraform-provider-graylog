package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &PipelineRuleResource{}
	_ resource.ResourceWithImportState = &PipelineRuleResource{}
)

func NewPipelineRuleResource() resource.Resource {
	return &PipelineRuleResource{}
}

type PipelineRuleResource struct {
	client *client.Client
}

type PipelineRuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	CreatedAt   types.String `tfsdk:"created_at"`
	ModifiedAt  types.String `tfsdk:"modified_at"`
}

func (r *PipelineRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_rule"
}

func (r *PipelineRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog pipeline rule. Rules are written in the " +
			"[Pipeline Rule Language](https://go2docs.graylog.org/current/making_sense_of_your_log_data/pipeline_rules.html) " +
			"and referenced by pipelines.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The pipeline rule ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The rule title. If omitted, Graylog parses it from the `source`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the rule.",
			},
			"source": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The rule source code.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the rule was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the rule was last modified.",
			},
		},
	}
}

func (r *PipelineRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PipelineRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PipelineRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &client.PipelineRule{
		Title:       data.Title.ValueString(),
		Description: data.Description.ValueString(),
		Source:      data.Source.ValueString(),
	}

	result, err := r.client.CreatePipelineRule(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pipeline rule", err.Error())
		return
	}

	mapPipelineRuleToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PipelineRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetPipelineRule(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read pipeline rule", err.Error())
		return
	}

	mapPipelineRuleToModel(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PipelineRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PipelineRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.PipelineRule{
		Title:       data.Title.ValueString(),
		Description: data.Description.ValueString(),
		Source:      data.Source.ValueString(),
	}

	result, err := r.client.UpdatePipelineRule(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update pipeline rule", err.Error())
		return
	}

	mapPipelineRuleToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PipelineRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePipelineRule(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete pipeline rule", err.Error())
	}
}

func (r *PipelineRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapPipelineRuleToModel(rule *client.PipelineRule, data *PipelineRuleResourceModel) {
	data.ID = types.StringValue(rule.ID)
	data.Title = types.StringValue(rule.Title)
	data.Source = types.StringValue(rule.Source)
	data.CreatedAt = types.StringValue(rule.CreatedAt)
	data.ModifiedAt = types.StringValue(rule.ModifiedAt)

	if rule.Description != "" {
		data.Description = types.StringValue(rule.Description)
	}
}
