// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = &ExtractorResource{}
	_ resource.ResourceWithImportState = &ExtractorResource{}
)

func NewExtractorResource() resource.Resource {
	return &ExtractorResource{}
}

type ExtractorResource struct {
	client *client.Client
}

type ExtractorResourceModel struct {
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
			"extractor_config_json": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JSON object with extractor-specific configuration.",
			},
			"converters_json": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JSON array with converter configurations.",
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

	extractorReq, diags := extractorFromModel(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateExtractor(ctx, data.InputID.ValueString(), extractorReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create extractor", err.Error())
		return
	}

	mapExtractorToResourceModel(created, &data)
	populateExtractorJSONFields(created, &data)
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

	mapExtractorToResourceModel(current, &data)
	populateExtractorJSONFields(current, &data)
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

	extractorReq, diags := extractorFromModel(&data)
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
	mapExtractorToResourceModel(updated, &data)
	populateExtractorJSONFields(updated, &data)
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

func extractorFromModel(data *ExtractorResourceModel) (*client.Extractor, diag.Diagnostics) {
	var diags diag.Diagnostics

	cfg := map[string]interface{}{}
	if !data.ExtractorConfigJSON.IsNull() && !data.ExtractorConfigJSON.IsUnknown() && data.ExtractorConfigJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.ExtractorConfigJSON.ValueString()), &cfg); err != nil {
			diags.AddError("Invalid extractor_config_json", fmt.Sprintf("Failed to parse extractor_config_json: %v", err))
			return nil, diags
		}
	}

	converters := []map[string]interface{}{}
	if !data.ConvertersJSON.IsNull() && !data.ConvertersJSON.IsUnknown() && data.ConvertersJSON.ValueString() != "" {
		if err := json.Unmarshal([]byte(data.ConvertersJSON.ValueString()), &converters); err != nil {
			diags.AddError("Invalid converters_json", fmt.Sprintf("Failed to parse converters_json: %v", err))
			return nil, diags
		}
	}

	req := &client.Extractor{
		Title:          data.Title.ValueString(),
		ExtractorType:  data.ExtractorType.ValueString(),
		CursorStrategy: data.CursorStrategy.ValueString(),
		SourceField:    data.SourceField.ValueString(),
		TargetField:    data.TargetField.ValueString(),
		ExtractorConfig: cfg,
		ConditionType:  data.ConditionType.ValueString(),
		Converters:     converters,
	}
	if !data.ConditionValue.IsNull() && !data.ConditionValue.IsUnknown() {
		req.ConditionValue = data.ConditionValue.ValueString()
	}
	if !data.Order.IsNull() && !data.Order.IsUnknown() {
		req.Order = data.Order.ValueInt64()
	}

	return req, diags
}

func populateExtractorJSONFields(extractor *client.Extractor, data *ExtractorResourceModel) {
	cfg := extractor.ExtractorConfig
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	cfgB, err := json.Marshal(cfg)
	if err != nil {
		data.ExtractorConfigJSON = types.StringValue("{}")
	} else {
		data.ExtractorConfigJSON = types.StringValue(string(cfgB))
	}

	converters := extractor.Converters
	if converters == nil {
		converters = []map[string]interface{}{}
	}
	convB, err := json.Marshal(converters)
	if err != nil {
		data.ConvertersJSON = types.StringValue("[]")
	} else {
		data.ConvertersJSON = types.StringValue(string(convB))
	}
}
