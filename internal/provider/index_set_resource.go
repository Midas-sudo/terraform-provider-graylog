// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

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
	_ resource.Resource                = &IndexSetResource{}
	_ resource.ResourceWithImportState = &IndexSetResource{}
)

func NewIndexSetResource() resource.Resource {
	return &IndexSetResource{}
}

type IndexSetResource struct {
	client *client.Client
}

type RotationStrategyModel struct {
	Type types.String `tfsdk:"type"`
}

type RetentionStrategyModel struct {
	Type               types.String `tfsdk:"type"`
	MaxNumberOfIndices types.Int64  `tfsdk:"max_number_of_indices"`
}

type DataTieringModel struct {
	Type             types.String `tfsdk:"type"`
	IndexLifetimeMin types.String `tfsdk:"index_lifetime_min"`
	IndexLifetimeMax types.String `tfsdk:"index_lifetime_max"`
}

type IndexSetResourceModel struct {
	ID                              types.String            `tfsdk:"id"`
	Title                           types.String            `tfsdk:"title"`
	Description                     types.String            `tfsdk:"description"`
	IndexPrefix                     types.String            `tfsdk:"index_prefix"`
	IndexOptimizationMaxNumSegments types.Int64             `tfsdk:"index_optimization_max_num_segments"`
	IndexOptimizationDisabled       types.Bool              `tfsdk:"index_optimization_disabled"`
	FieldTypeRefreshInterval        types.Int64             `tfsdk:"field_type_refresh_interval"`
	Shards                          types.Int64             `tfsdk:"shards"`
	Replicas                        types.Int64             `tfsdk:"replicas"`
	Writable                        types.Bool              `tfsdk:"writable"`
	IndexAnalyzer                   types.String            `tfsdk:"index_analyzer"`
	UseLegacyRotation               types.Bool              `tfsdk:"use_legacy_rotation"`
	RotationStrategyClass           types.String            `tfsdk:"rotation_strategy_class"`
	RetentionStrategyClass          types.String            `tfsdk:"retention_strategy_class"`
	RotationStrategy                *RotationStrategyModel  `tfsdk:"rotation_strategy"`
	RetentionStrategy               *RetentionStrategyModel `tfsdk:"retention_strategy"`
	DataTiering                     *DataTieringModel       `tfsdk:"data_tiering"`
	SetAsDefault                    types.Bool              `tfsdk:"set_as_default"`
	IsDefault                       types.Bool              `tfsdk:"is_default"`
	SyncTemplate                    types.Bool              `tfsdk:"sync_template"`
}

func (r *IndexSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index_set"
}

func (r *IndexSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				Required: true,
				MarkdownDescription: "Rotation strategy class. Use a short name such as `MessageCountRotationStrategy`, `SizeBasedRotationStrategy`, `TimeBasedRotationStrategy`, or `TimeBasedSizeOptimizingStrategy`, or the full Graylog Java class (e.g. `org.graylog2.indexer.rotation.strategies.MessageCountRotationStrategy`).",
			},
			"retention_strategy_class": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Retention strategy class. Use a short name such as `DeletionRetentionStrategy` or `NoopRetentionStrategy`, or the full Graylog Java class (e.g. `org.graylog2.indexer.retention.strategies.DeletionRetentionStrategy`).",
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
			"rotation_strategy": schema.SingleNestedBlock{
				MarkdownDescription: "Rotation strategy settings.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						MarkdownDescription: "Rotation strategy config type. Use a short name such as `MessageCountRotationStrategyConfig` matching the rotation strategy, or the full Graylog config class name.",
					},
				},
			},
			"retention_strategy": schema.SingleNestedBlock{
				MarkdownDescription: "Retention strategy settings.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						MarkdownDescription: "Retention strategy config type. Use a short name such as `DeletionRetentionStrategyConfig` or `NoopRetentionStrategyConfig`, or the full Graylog config class name.",
					},
					"max_number_of_indices": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(20),
						MarkdownDescription: "Maximum number of indices to retain (strategy dependent).",
					},
				},
			},
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

	createReq := toIndexSetRequest(&data, "")
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
	mapIndexSetToModel(current, &data)
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
	mapIndexSetToModel(current, &data)
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

	updateReq := toIndexSetRequest(&data, state.ID.ValueString())
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

	mapIndexSetToModel(updated, &data)
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

func toIndexSetRequest(data *IndexSetResourceModel, id string) *client.IndexSet {
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
	if data.RotationStrategy != nil {
		req.RotationStrategy = client.RotationStrategyConfig{
			Type: expandRotationStrategyConfigType(data.RotationStrategy.Type.ValueString()),
		}
	}
	if data.RetentionStrategy != nil {
		req.RetentionStrategy = client.RetentionStrategyConfig{
			Type:               expandRetentionStrategyConfigType(data.RetentionStrategy.Type.ValueString()),
			MaxNumberOfIndices: data.RetentionStrategy.MaxNumberOfIndices.ValueInt64(),
		}
	}
	if data.DataTiering != nil {
		req.DataTiering = &client.DataTieringConfig{
			Type:             data.DataTiering.Type.ValueString(),
			IndexLifetimeMin: data.DataTiering.IndexLifetimeMin.ValueString(),
			IndexLifetimeMax: data.DataTiering.IndexLifetimeMax.ValueString(),
		}
	}
	return req
}

func mapIndexSetToModel(src *client.IndexSet, dst *IndexSetResourceModel) {
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
	if src.RotationStrategy.Type != "" {
		dst.RotationStrategy = &RotationStrategyModel{
			Type: types.StringValue(collapseRotationStrategyConfigType(src.RotationStrategy.Type)),
		}
	} else {
		dst.RotationStrategy = nil
	}
	if src.RetentionStrategy.Type != "" || src.RetentionStrategy.MaxNumberOfIndices != 0 {
		dst.RetentionStrategy = &RetentionStrategyModel{
			Type:               types.StringValue(collapseRetentionStrategyConfigType(src.RetentionStrategy.Type)),
			MaxNumberOfIndices: types.Int64Value(src.RetentionStrategy.MaxNumberOfIndices),
		}
	} else {
		dst.RetentionStrategy = nil
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
}
