// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                 = &IndexSetResource{}
	_ resource.ResourceWithImportState  = &IndexSetResource{}
	_ resource.ResourceWithUpgradeState = &IndexSetResource{}
)

func NewIndexSetResource() resource.Resource {
	return &IndexSetResource{}
}

type IndexSetResource struct {
	client *client.Client
}

type DataTieringModel struct {
	Type             types.String `tfsdk:"type"`
	IndexLifetimeMin types.String `tfsdk:"index_lifetime_min"`
	IndexLifetimeMax types.String `tfsdk:"index_lifetime_max"`
}

type IndexSetResourceModel struct {
	ID                              types.String      `tfsdk:"id"`
	Title                           types.String      `tfsdk:"title"`
	Description                     types.String      `tfsdk:"description"`
	IndexPrefix                     types.String      `tfsdk:"index_prefix"`
	IndexOptimizationMaxNumSegments types.Int64       `tfsdk:"index_optimization_max_num_segments"`
	IndexOptimizationDisabled       types.Bool        `tfsdk:"index_optimization_disabled"`
	FieldTypeRefreshInterval        types.Int64       `tfsdk:"field_type_refresh_interval"`
	Shards                          types.Int64       `tfsdk:"shards"`
	Replicas                        types.Int64       `tfsdk:"replicas"`
	Writable                        types.Bool        `tfsdk:"writable"`
	IndexAnalyzer                   types.String      `tfsdk:"index_analyzer"`
	UseLegacyRotation               types.Bool        `tfsdk:"use_legacy_rotation"`
	RotationStrategyClass           types.String      `tfsdk:"rotation_strategy_class"`
	RetentionStrategyClass          types.String      `tfsdk:"retention_strategy_class"`
	RotationStrategy                types.Dynamic     `tfsdk:"rotation_strategy"`
	RetentionStrategy               types.Dynamic     `tfsdk:"retention_strategy"`
	DataTiering                     *DataTieringModel `tfsdk:"data_tiering"`
	SetAsDefault                    types.Bool        `tfsdk:"set_as_default"`
	IsDefault                       types.Bool        `tfsdk:"is_default"`
	SyncTemplate                    types.Bool        `tfsdk:"sync_template"`
}

type indexSetResourceModelV0 struct {
	ID                              types.String `tfsdk:"id"`
	Title                           types.String `tfsdk:"title"`
	Description                     types.String `tfsdk:"description"`
	IndexPrefix                     types.String `tfsdk:"index_prefix"`
	IndexOptimizationMaxNumSegments types.Int64  `tfsdk:"index_optimization_max_num_segments"`
	IndexOptimizationDisabled       types.Bool   `tfsdk:"index_optimization_disabled"`
	FieldTypeRefreshInterval        types.Int64  `tfsdk:"field_type_refresh_interval"`
	Shards                          types.Int64  `tfsdk:"shards"`
	Replicas                        types.Int64  `tfsdk:"replicas"`
	Writable                        types.Bool   `tfsdk:"writable"`
	IndexAnalyzer                   types.String `tfsdk:"index_analyzer"`
	UseLegacyRotation               types.Bool   `tfsdk:"use_legacy_rotation"`
	RotationStrategyClass           types.String `tfsdk:"rotation_strategy_class"`
	RetentionStrategyClass          types.String `tfsdk:"retention_strategy_class"`
	RotationStrategy                *struct {
		Type types.String `tfsdk:"type"`
	} `tfsdk:"rotation_strategy"`
	RetentionStrategy *struct {
		Type               types.String `tfsdk:"type"`
		MaxNumberOfIndices types.Int64  `tfsdk:"max_number_of_indices"`
	} `tfsdk:"retention_strategy"`
	DataTiering  *DataTieringModel `tfsdk:"data_tiering"`
	SetAsDefault types.Bool        `tfsdk:"set_as_default"`
	IsDefault    types.Bool        `tfsdk:"is_default"`
	SyncTemplate types.Bool        `tfsdk:"sync_template"`
}

func (r *IndexSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index_set"
}

func (r *IndexSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a Graylog index set, including retention and rotation strategies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Index set ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Index set display name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Index set description.",
			},
			"index_prefix": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Index prefix used for created indices.",
			},
			"index_optimization_max_num_segments": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Maximum segments after index optimization. Defaults to 1.",
			},
			"index_optimization_disabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Disables index optimization when true.",
			},
			"field_type_refresh_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(5000),
				MarkdownDescription: "Field type refresh interval in milliseconds. Defaults to 5000.",
			},
			"shards": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Number of shards per index. Defaults to 1.",
			},
			"replicas": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Number of replicas per index. Defaults to 0.",
			},
			"writable": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this index set is writable.",
			},
			"index_analyzer": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("standard"),
				MarkdownDescription: "Index analyzer name. Defaults to `standard`.",
			},
			"use_legacy_rotation": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether legacy rotation mode is enabled.",
			},
			"rotation_strategy_class": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Rotation strategy class. Use a short name such as `MessageCountRotationStrategy`, `SizeBasedRotationStrategy`, `TimeBasedRotationStrategy`, or `TimeBasedSizeOptimizingStrategy`, or the full Graylog Java class (e.g. `org.graylog2.indexer.rotation.strategies.MessageCountRotationStrategy`).",
			},
			"retention_strategy_class": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Retention strategy class. Use a short name such as `DeletionRetentionStrategy` or `NoopRetentionStrategy`, or the full Graylog Java class (e.g. `org.graylog2.indexer.retention.strategies.DeletionRetentionStrategy`).",
			},
			"rotation_strategy": schema.DynamicAttribute{
				Required: true,
				MarkdownDescription: "HCL object passed through to Graylog. Must include `type` (short config name or FQCN). " +
					"Discover fields with [`graylog_index_set_strategy_types`](../data-sources/index_set_strategy_types.md) " +
					"(`rotation`). Examples: TimeBased uses `rotation_period`; MessageCount uses `max_docs_per_index`; " +
					"SizeBased uses `max_size`.",
			},
			"retention_strategy": schema.DynamicAttribute{
				Required: true,
				MarkdownDescription: "HCL object passed through to Graylog. Must include `type` (short config name or FQCN). " +
					"Discover fields with [`graylog_index_set_strategy_types`](../data-sources/index_set_strategy_types.md) " +
					"(`retention`). Deletion uses `max_number_of_indices`.",
			},
			"set_as_default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "If true, sets this index set as Graylog default after create/update.",
			},
			"is_default": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this index set is currently the Graylog default.",
			},
			"sync_template": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "If true, syncs index template after create/update.",
			},
		},
		Blocks: map[string]schema.Block{
			"data_tiering": schema.SingleNestedBlock{
				MarkdownDescription: "Data tiering settings for tier-aware rotation strategies.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Data tiering mode type (for example `hot_only`).",
					},
					"index_lifetime_min": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Minimum index lifetime as ISO-8601 duration (for example `P30D`).",
					},
					"index_lifetime_max": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						MarkdownDescription: "Maximum index lifetime as ISO-8601 duration (for example `P40D`).",
					},
				},
			},
		},
	}
}

func (r *IndexSetResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                                  schema.StringAttribute{Computed: true},
					"title":                               schema.StringAttribute{Required: true},
					"description":                         schema.StringAttribute{Optional: true},
					"index_prefix":                        schema.StringAttribute{Required: true},
					"index_optimization_max_num_segments": schema.Int64Attribute{Optional: true},
					"index_optimization_disabled":         schema.BoolAttribute{Optional: true},
					"field_type_refresh_interval":         schema.Int64Attribute{Optional: true},
					"shards":                              schema.Int64Attribute{Optional: true},
					"replicas":                            schema.Int64Attribute{Optional: true},
					"writable":                            schema.BoolAttribute{Optional: true},
					"index_analyzer":                      schema.StringAttribute{Optional: true},
					"use_legacy_rotation":                 schema.BoolAttribute{Optional: true},
					"rotation_strategy_class":             schema.StringAttribute{Required: true},
					"retention_strategy_class":            schema.StringAttribute{Required: true},
					"set_as_default":                      schema.BoolAttribute{Optional: true},
					"is_default":                          schema.BoolAttribute{Computed: true},
					"sync_template":                       schema.BoolAttribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					"rotation_strategy": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{Required: true},
						},
					},
					"retention_strategy": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"type":                  schema.StringAttribute{Required: true},
							"max_number_of_indices": schema.Int64Attribute{Optional: true},
						},
					},
					"data_tiering": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"type":               schema.StringAttribute{Optional: true},
							"index_lifetime_min": schema.StringAttribute{Optional: true},
							"index_lifetime_max": schema.StringAttribute{Optional: true},
						},
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior indexSetResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				rot := map[string]interface{}{}
				if prior.RotationStrategy != nil && !prior.RotationStrategy.Type.IsNull() {
					rot["type"] = prior.RotationStrategy.Type.ValueString()
				}
				ret := map[string]interface{}{}
				if prior.RetentionStrategy != nil {
					if !prior.RetentionStrategy.Type.IsNull() {
						ret["type"] = prior.RetentionStrategy.Type.ValueString()
					}
					if !prior.RetentionStrategy.MaxNumberOfIndices.IsNull() {
						ret["max_number_of_indices"] = prior.RetentionStrategy.MaxNumberOfIndices.ValueInt64()
					}
				}
				rotDyn, d := interfaceToDynamic(ctx, rot)
				resp.Diagnostics.Append(d...)
				retDyn, d := interfaceToDynamic(ctx, ret)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := IndexSetResourceModel{
					ID:                              prior.ID,
					Title:                           prior.Title,
					Description:                     prior.Description,
					IndexPrefix:                     prior.IndexPrefix,
					IndexOptimizationMaxNumSegments: prior.IndexOptimizationMaxNumSegments,
					IndexOptimizationDisabled:       prior.IndexOptimizationDisabled,
					FieldTypeRefreshInterval:        prior.FieldTypeRefreshInterval,
					Shards:                          prior.Shards,
					Replicas:                        prior.Replicas,
					Writable:                        prior.Writable,
					IndexAnalyzer:                   prior.IndexAnalyzer,
					UseLegacyRotation:               prior.UseLegacyRotation,
					RotationStrategyClass:           prior.RotationStrategyClass,
					RetentionStrategyClass:          prior.RetentionStrategyClass,
					RotationStrategy:                rotDyn,
					RetentionStrategy:               retDyn,
					DataTiering:                     prior.DataTiering,
					SetAsDefault:                    prior.SetAsDefault,
					IsDefault:                       prior.IsDefault,
					SyncTemplate:                    prior.SyncTemplate,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *IndexSetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IndexSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IndexSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plannedRotation := data.RotationStrategy
	plannedRetention := data.RetentionStrategy

	createReq, diags := toIndexSetRequest(ctx, &data, "")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateIndexSet(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create index set", err.Error())
		return
	}

	if data.SetAsDefault.ValueBool() {
		if err := r.client.SetDefaultIndexSet(ctx, created.ID); err != nil {
			resp.Diagnostics.AddError("Failed to set default index set", err.Error())
			return
		}
	}
	if data.SyncTemplate.ValueBool() {
		if _, err := r.client.SyncIndexTemplate(ctx, created.ID); err != nil && !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Failed to sync index template", err.Error())
			return
		}
	}

	var current *client.IndexSet
	var errRead error
	for i := 0; i < 5; i++ {
		current, errRead = r.client.GetIndexSet(ctx, created.ID)
		if errRead == nil {
			break
		}
		if !client.IsNotFound(errRead) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if errRead != nil {
		resp.Diagnostics.AddError("Failed to read created index set", errRead.Error())
		return
	}
	resp.Diagnostics.Append(mapIndexSetToModel(ctx, current, &data)...)
	// Preserve planned Dynamic strategy objects to avoid Framework type-drift on apply.
	data.RotationStrategy = plannedRotation
	data.RetentionStrategy = plannedRetention
	data.SetAsDefault = types.BoolValue(data.SetAsDefault.ValueBool())
	data.SyncTemplate = types.BoolValue(data.SyncTemplate.ValueBool())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IndexSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetIndexSet(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read index set", err.Error())
		return
	}
	resp.Diagnostics.Append(mapIndexSetToModel(ctx, current, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IndexSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state IndexSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plannedRotation := data.RotationStrategy
	plannedRetention := data.RetentionStrategy

	updateReq, diags := toIndexSetRequest(ctx, &data, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateIndexSet(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update index set", err.Error())
		return
	}

	if data.SetAsDefault.ValueBool() {
		if err := r.client.SetDefaultIndexSet(ctx, state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to set default index set", err.Error())
			return
		}
	}
	if data.SyncTemplate.ValueBool() {
		if _, err := r.client.SyncIndexTemplate(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Failed to sync index template", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(mapIndexSetToModel(ctx, updated, &data)...)
	data.RotationStrategy = plannedRotation
	data.RetentionStrategy = plannedRetention
	data.SetAsDefault = types.BoolValue(data.SetAsDefault.ValueBool())
	data.SyncTemplate = types.BoolValue(data.SyncTemplate.ValueBool())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IndexSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteIndexSet(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete index set", err.Error())
	}
}

func (r *IndexSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func toIndexSetRequest(ctx context.Context, data *IndexSetResourceModel, id string) (*client.IndexSet, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.IndexSet{
		ID:                              id,
		Title:                           data.Title.ValueString(),
		Description:                     data.Description.ValueString(),
		IndexPrefix:                     data.IndexPrefix.ValueString(),
		IndexOptimizationMaxNumSegments: data.IndexOptimizationMaxNumSegments.ValueInt64(),
		IndexOptimizationDisabled:       data.IndexOptimizationDisabled.ValueBool(),
		FieldTypeRefreshInterval:        data.FieldTypeRefreshInterval.ValueInt64(),
		Shards:                          data.Shards.ValueInt64(),
		Replicas:                        data.Replicas.ValueInt64(),
		Writable:                        data.Writable.ValueBool(),
		IndexAnalyzer:                   data.IndexAnalyzer.ValueString(),
		UseLegacyRotation:               data.UseLegacyRotation.ValueBool(),
		RotationStrategyClass:           expandRotationStrategyClass(data.RotationStrategyClass.ValueString()),
		RetentionStrategyClass:          expandRetentionStrategyClass(data.RetentionStrategyClass.ValueString()),
	}

	rot, d := dynamicToMap(ctx, data.RotationStrategy)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	if strategyMapType(rot) == "" {
		diags.AddAttributeError(path.Root("rotation_strategy"), "Invalid rotation_strategy", "`type` is required")
		return nil, diags
	}
	req.RotationStrategy = expandRotationStrategyMap(rot)

	ret, d := dynamicToMap(ctx, data.RetentionStrategy)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	if strategyMapType(ret) == "" {
		diags.AddAttributeError(path.Root("retention_strategy"), "Invalid retention_strategy", "`type` is required")
		return nil, diags
	}
	req.RetentionStrategy = expandRetentionStrategyMap(ret)

	if data.DataTiering != nil {
		req.DataTiering = &client.DataTieringConfig{
			Type:             data.DataTiering.Type.ValueString(),
			IndexLifetimeMin: data.DataTiering.IndexLifetimeMin.ValueString(),
			IndexLifetimeMax: data.DataTiering.IndexLifetimeMax.ValueString(),
		}
	}
	return req, diags
}

func mapIndexSetToModel(ctx context.Context, src *client.IndexSet, dst *IndexSetResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	dst.ID = types.StringValue(src.ID)
	dst.Title = types.StringValue(src.Title)
	dst.Description = types.StringValue(src.Description)
	dst.IndexPrefix = types.StringValue(src.IndexPrefix)
	dst.IndexOptimizationMaxNumSegments = types.Int64Value(src.IndexOptimizationMaxNumSegments)
	dst.IndexOptimizationDisabled = types.BoolValue(src.IndexOptimizationDisabled)
	dst.FieldTypeRefreshInterval = types.Int64Value(src.FieldTypeRefreshInterval)
	dst.Shards = types.Int64Value(src.Shards)
	dst.Replicas = types.Int64Value(src.Replicas)
	dst.Writable = types.BoolValue(src.Writable)
	dst.IndexAnalyzer = types.StringValue(src.IndexAnalyzer)
	dst.UseLegacyRotation = types.BoolValue(src.UseLegacyRotation)
	dst.RotationStrategyClass = types.StringValue(collapseRotationStrategyClass(src.RotationStrategyClass))
	dst.RetentionStrategyClass = types.StringValue(collapseRetentionStrategyClass(src.RetentionStrategyClass))
	dst.IsDefault = types.BoolValue(src.Default)

	// Only hydrate Dynamic strategy objects when unset (e.g. import). Graylog often
	// echoes nulls or strategy-specific extras that would otherwise cause perpetual drift.
	if dst.RotationStrategy.IsNull() || dst.RotationStrategy.IsUnknown() {
		rotDyn, d := interfaceToDynamic(ctx, collapseRotationStrategyMap(omitNilStrategyValues(src.RotationStrategy)))
		diags.Append(d...)
		dst.RotationStrategy = rotDyn
	}
	if dst.RetentionStrategy.IsNull() || dst.RetentionStrategy.IsUnknown() {
		retDyn, d := interfaceToDynamic(ctx, collapseRetentionStrategyMap(omitNilStrategyValues(src.RetentionStrategy)))
		diags.Append(d...)
		dst.RetentionStrategy = retDyn
	}

	if src.DataTiering != nil && src.DataTiering.Type != "" {
		dst.DataTiering = &DataTieringModel{
			Type:             types.StringValue(src.DataTiering.Type),
			IndexLifetimeMin: types.StringValue(src.DataTiering.IndexLifetimeMin),
			IndexLifetimeMax: types.StringValue(src.DataTiering.IndexLifetimeMax),
		}
	} else {
		dst.DataTiering = nil
	}
	return diags
}
