package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &StreamResource{}
	_ resource.ResourceWithImportState = &StreamResource{}
)

func NewStreamResource() resource.Resource {
	return &StreamResource{}
}

type StreamResource struct {
	client *client.Client
}

type StreamResourceModel struct {
	ID                             types.String `tfsdk:"id"`
	Title                          types.String `tfsdk:"title"`
	Description                    types.String `tfsdk:"description"`
	IndexSetID                     types.String `tfsdk:"index_set_id"`
	MatchingType                   types.String `tfsdk:"matching_type"`
	RemoveMatchesFromDefaultStream types.Bool   `tfsdk:"remove_matches_from_default_stream"`
	Disabled                       types.Bool   `tfsdk:"disabled"`
}

func (r *StreamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream"
}

func (r *StreamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog stream. Streams route messages into categories based on rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The stream ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The stream title.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "A description of the stream.",
			},
			"index_set_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the index set this stream writes to.",
			},
			"matching_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("AND"),
				MarkdownDescription: "How stream rules are evaluated: `AND` (all must match) or `OR` (any must match). Defaults to `AND`.",
			},
			"remove_matches_from_default_stream": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Remove messages that match this stream from the default stream. Defaults to `false`.",
			},
			"disabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the stream is disabled. Defaults to `false`.",
			},
		},
	}
}

func (r *StreamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StreamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &client.StreamCreateRequest{
		Title:                          data.Title.ValueString(),
		Description:                    data.Description.ValueString(),
		IndexSetID:                     data.IndexSetID.ValueString(),
		MatchingType:                   data.MatchingType.ValueString(),
		RemoveMatchesFromDefaultStream: data.RemoveMatchesFromDefaultStream.ValueBool(),
		Rules:                          []client.StreamRule{},
	}

	result, err := r.client.CreateStream(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create stream", err.Error())
		return
	}

	data.ID = types.StringValue(result.StreamID)

	// Resume the stream if it shouldn't be disabled (streams start paused)
	if !data.Disabled.ValueBool() {
		if err := r.client.ResumeStream(ctx, result.StreamID); err != nil {
			resp.Diagnostics.AddError("Failed to resume stream after creation", err.Error())
			return
		}
	}

	stream, err := r.client.GetStream(ctx, result.StreamID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created stream", err.Error())
		return
	}

	mapStreamToModel(stream, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stream, err := r.client.GetStream(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read stream", err.Error())
		return
	}

	mapStreamToModel(stream, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StreamResourceModel
	var state StreamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.StreamUpdateRequest{
		Title:                          data.Title.ValueString(),
		Description:                    data.Description.ValueString(),
		IndexSetID:                     data.IndexSetID.ValueString(),
		MatchingType:                   data.MatchingType.ValueString(),
		RemoveMatchesFromDefaultStream: data.RemoveMatchesFromDefaultStream.ValueBool(),
	}

	_, err := r.client.UpdateStream(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update stream", err.Error())
		return
	}

	// Handle enabled/disabled transitions
	wasDisabled := state.Disabled.ValueBool()
	wantDisabled := data.Disabled.ValueBool()
	if wasDisabled && !wantDisabled {
		if err := r.client.ResumeStream(ctx, state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to resume stream", err.Error())
			return
		}
	} else if !wasDisabled && wantDisabled {
		if err := r.client.PauseStream(ctx, state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to pause stream", err.Error())
			return
		}
	}

	stream, err := r.client.GetStream(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated stream", err.Error())
		return
	}

	mapStreamToModel(stream, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStream(ctx, data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete stream", err.Error())
	}
}

func (r *StreamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapStreamToModel(stream *client.Stream, data *StreamResourceModel) {
	data.ID = types.StringValue(stream.ID)
	data.Title = types.StringValue(stream.Title)
	data.Description = types.StringValue(stream.Description)
	data.IndexSetID = types.StringValue(stream.IndexSetID)
	data.MatchingType = types.StringValue(stream.MatchingType)
	data.RemoveMatchesFromDefaultStream = types.BoolValue(stream.RemoveMatchesFromDefaultStream)
	data.Disabled = types.BoolValue(stream.Disabled)
}
