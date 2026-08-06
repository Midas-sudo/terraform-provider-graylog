// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                 = &ExtractorResource{}
	_ resource.ResourceWithImportState  = &ExtractorResource{}
	_ resource.ResourceWithUpgradeState = &ExtractorResource{}
)

func NewExtractorResource() resource.Resource {
	return &ExtractorResource{}
}

type ExtractorResource struct {
	client *client.Client
}

type ExtractorResourceModel struct {
	ID              types.String  `tfsdk:"id"`
	InputID         types.String  `tfsdk:"input_id"`
	Title           types.String  `tfsdk:"title"`
	ExtractorType   types.String  `tfsdk:"extractor_type"`
	CursorStrategy  types.String  `tfsdk:"cursor_strategy"`
	SourceField     types.String  `tfsdk:"source_field"`
	TargetField     types.String  `tfsdk:"target_field"`
	ConditionType   types.String  `tfsdk:"condition_type"`
	ConditionValue  types.String  `tfsdk:"condition_value"`
	Order           types.Int64   `tfsdk:"order"`
	ExtractorConfig types.Dynamic `tfsdk:"extractor_config"`
	Converters      types.Dynamic `tfsdk:"converters"`
}

type extractorResourceModelV0 struct {
	ID                  types.String `tfsdk:"id"`
	InputID             types.String `tfsdk:"input_id"`
	Title               types.String `tfsdk:"title"`
	ExtractorType       types.String `tfsdk:"extractor_type"`
	CursorStrategy      types.String `tfsdk:"cursor_strategy"`
	SourceField         types.String `tfsdk:"source_field"`
	TargetField         types.String `tfsdk:"target_field"`
	ConditionType       types.String `tfsdk:"condition_type"`
	ConditionValue      types.String `tfsdk:"condition_value"`
	Order               types.Int64  `tfsdk:"order"`
	ExtractorConfigJSON types.String `tfsdk:"extractor_config_json"`
	ConvertersJSON      types.String `tfsdk:"converters_json"`
}

func (r *ExtractorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extractor"
}

func (r *ExtractorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a Graylog extractor for a specific input.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"input_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Graylog input ID that owns this extractor.",
			},
			"title":          schema.StringAttribute{Required: true},
			"extractor_type": schema.StringAttribute{Required: true},
			"cursor_strategy": schema.StringAttribute{
				Required: true,
			},
			"source_field": schema.StringAttribute{
				Required: true,
			},
			"target_field": schema.StringAttribute{
				Required: true,
			},
			"condition_type": schema.StringAttribute{
				Required: true,
			},
			"condition_value": schema.StringAttribute{
				Optional: true,
			},
			"order": schema.Int64Attribute{
				Optional: true,
			},
			"extractor_config": schema.DynamicAttribute{
				Optional: true,
				MarkdownDescription: "Extractor-specific configuration object. Keys depend on `extractor_type` " +
					"(e.g. `regex` uses `regex_value`; `split_and_index` uses `split_by` / `index`; " +
					"`copy_input` often uses `{}`). See the resource docs for a type matrix.",
			},
			"converters": schema.DynamicAttribute{
				Optional: true,
				MarkdownDescription: "List of converter objects. Each entry typically has `type` " +
					"(e.g. `numeric`, `date`, `hash`, `lowercase`, `uppercase`, `tokenizer`) " +
					"and optional `config` object.",
			},
		},
	}
}

func (r *ExtractorResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                    schema.StringAttribute{Computed: true},
					"input_id":              schema.StringAttribute{Required: true},
					"title":                 schema.StringAttribute{Required: true},
					"extractor_type":        schema.StringAttribute{Required: true},
					"cursor_strategy":       schema.StringAttribute{Required: true},
					"source_field":          schema.StringAttribute{Required: true},
					"target_field":          schema.StringAttribute{Required: true},
					"condition_type":        schema.StringAttribute{Required: true},
					"condition_value":       schema.StringAttribute{Optional: true},
					"order":                 schema.Int64Attribute{Optional: true},
					"extractor_config_json": schema.StringAttribute{Optional: true},
					"converters_json":       schema.StringAttribute{Optional: true},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior extractorResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				cfgDyn, err := upgradeJSONStringAttr(prior.ExtractorConfigJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade extractor_config", err.Error())
					return
				}
				if cfgDyn.IsNull() {
					cfgDyn, _ = interfaceToDynamic(ctx, map[string]interface{}{})
				}
				convDyn, err := upgradeJSONStringAttr(prior.ConvertersJSON.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Failed to upgrade converters", err.Error())
					return
				}
				if convDyn.IsNull() {
					convDyn, _ = interfaceToDynamic(ctx, []interface{}{})
				}
				upgraded := ExtractorResourceModel{
					ID:              prior.ID,
					InputID:         prior.InputID,
					Title:           prior.Title,
					ExtractorType:   prior.ExtractorType,
					CursorStrategy:  prior.CursorStrategy,
					SourceField:     prior.SourceField,
					TargetField:     prior.TargetField,
					ConditionType:   prior.ConditionType,
					ConditionValue:  prior.ConditionValue,
					Order:           prior.Order,
					ExtractorConfig: cfgDyn,
					Converters:      convDyn,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *ExtractorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ExtractorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExtractorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	extractorReq, diags := extractorFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateExtractor(ctx, data.InputID.ValueString(), extractorReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create extractor", err.Error())
		return
	}

	resp.Diagnostics.Append(mapExtractorToResourceModel(ctx, created, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExtractorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ExtractorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetExtractor(ctx, data.InputID.ValueString(), data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read extractor", err.Error())
		return
	}

	resp.Diagnostics.Append(mapExtractorToResourceModel(ctx, current, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExtractorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExtractorResourceModel
	var state ExtractorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	extractorReq, diags := extractorFromModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateExtractor(ctx, state.InputID.ValueString(), state.ID.ValueString(), extractorReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update extractor", err.Error())
		return
	}

	data.InputID = state.InputID
	resp.Diagnostics.Append(mapExtractorToResourceModel(ctx, updated, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExtractorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ExtractorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteExtractor(ctx, data.InputID.ValueString(), data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete extractor", err.Error())
	}
}

func (r *ExtractorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Use `input_id/extractor_id` when importing graylog_extractor.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("input_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func extractorFromModel(ctx context.Context, data *ExtractorResourceModel) (*client.Extractor, diag.Diagnostics) {
	var diags diag.Diagnostics

	cfg := map[string]interface{}{}
	if !data.ExtractorConfig.IsNull() && !data.ExtractorConfig.IsUnknown() {
		m, d := dynamicToMap(ctx, data.ExtractorConfig)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		cfg = m
	}

	converters := []map[string]interface{}{}
	if !data.Converters.IsNull() && !data.Converters.IsUnknown() {
		s, d := dynamicToSliceOfMaps(ctx, data.Converters)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		converters = s
	}

	req := &client.Extractor{
		Title:           data.Title.ValueString(),
		ExtractorType:   data.ExtractorType.ValueString(),
		CursorStrategy:  data.CursorStrategy.ValueString(),
		SourceField:     data.SourceField.ValueString(),
		TargetField:     data.TargetField.ValueString(),
		ExtractorConfig: cfg,
		ConditionType:   data.ConditionType.ValueString(),
		Converters:      converters,
	}
	if !data.ConditionValue.IsNull() && !data.ConditionValue.IsUnknown() {
		req.ConditionValue = data.ConditionValue.ValueString()
	}
	if !data.Order.IsNull() && !data.Order.IsUnknown() {
		req.Order = data.Order.ValueInt64()
	}

	return req, diags
}
