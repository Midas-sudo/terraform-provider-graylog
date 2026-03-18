// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ datasource.DataSource = &StreamDataSource{}

func NewStreamDataSource() datasource.DataSource {
	return &StreamDataSource{}
}

type StreamDataSource struct {
	client *client.Client
}

type StreamDataSourceModel struct {
	ID                             types.String `tfsdk:"id"`
	Title                          types.String `tfsdk:"title"`
	Description                    types.String `tfsdk:"description"`
	IndexSetID                     types.String `tfsdk:"index_set_id"`
	MatchingType                   types.String `tfsdk:"matching_type"`
	RemoveMatchesFromDefaultStream types.Bool   `tfsdk:"remove_matches_from_default_stream"`
	Disabled                       types.Bool   `tfsdk:"disabled"`
	IsDefault                      types.Bool   `tfsdk:"is_default"`
}

func (d *StreamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream"
}

func (d *StreamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single Graylog stream by ID.",
		Attributes: map[string]schema.Attribute{
			"id":                                  schema.StringAttribute{Required: true, MarkdownDescription: "The stream ID."},
			"title":                               schema.StringAttribute{Computed: true},
			"description":                         schema.StringAttribute{Computed: true},
			"index_set_id":                        schema.StringAttribute{Computed: true},
			"matching_type":                       schema.StringAttribute{Computed: true},
			"remove_matches_from_default_stream":  schema.BoolAttribute{Computed: true},
			"disabled":                            schema.BoolAttribute{Computed: true},
			"is_default":                          schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *StreamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StreamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StreamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stream, err := d.client.GetStream(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read stream", err.Error())
		return
	}

	data.Title = types.StringValue(stream.Title)
	data.Description = types.StringValue(stream.Description)
	data.IndexSetID = types.StringValue(stream.IndexSetID)
	data.MatchingType = types.StringValue(stream.MatchingType)
	data.RemoveMatchesFromDefaultStream = types.BoolValue(stream.RemoveMatchesFromDefaultStream)
	data.Disabled = types.BoolValue(stream.Disabled)
	data.IsDefault = types.BoolValue(stream.IsDefault)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// StreamsDataSource lists all streams.
var _ datasource.DataSource = &StreamsDataSource{}

func NewStreamsDataSource() datasource.DataSource {
	return &StreamsDataSource{}
}

type StreamsDataSource struct {
	client *client.Client
}

type StreamsDataSourceModel struct {
	Streams []StreamDataSourceModel `tfsdk:"streams"`
}

func (d *StreamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_streams"
}

func (d *StreamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all Graylog streams.",
		Attributes: map[string]schema.Attribute{
			"streams": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                                  schema.StringAttribute{Computed: true},
						"title":                               schema.StringAttribute{Computed: true},
						"description":                         schema.StringAttribute{Computed: true},
						"index_set_id":                        schema.StringAttribute{Computed: true},
						"matching_type":                       schema.StringAttribute{Computed: true},
						"remove_matches_from_default_stream":  schema.BoolAttribute{Computed: true},
						"disabled":                            schema.BoolAttribute{Computed: true},
						"is_default":                          schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *StreamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StreamsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.GetStreams(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list streams", err.Error())
		return
	}

	var data StreamsDataSourceModel
	for _, s := range result.Streams {
		data.Streams = append(data.Streams, StreamDataSourceModel{
			ID:                             types.StringValue(s.ID),
			Title:                          types.StringValue(s.Title),
			Description:                    types.StringValue(s.Description),
			IndexSetID:                     types.StringValue(s.IndexSetID),
			MatchingType:                   types.StringValue(s.MatchingType),
			RemoveMatchesFromDefaultStream: types.BoolValue(s.RemoveMatchesFromDefaultStream),
			Disabled:                       types.BoolValue(s.Disabled),
			IsDefault:                      types.BoolValue(s.IsDefault),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
