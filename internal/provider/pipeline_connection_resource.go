package provider

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &PipelineConnectionResource{}
	_ resource.ResourceWithImportState = &PipelineConnectionResource{}
)

func NewPipelineConnectionResource() resource.Resource {
	return &PipelineConnectionResource{}
}

type PipelineConnectionResource struct {
	client *client.Client
}

type PipelineConnectionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	StreamID    types.String `tfsdk:"stream_id"`
	PipelineIDs types.Set    `tfsdk:"pipeline_ids"`
}

func (r *PipelineConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_connection"
}

func (r *PipelineConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Connects one or more processing pipelines to a stream. " +
			"Each stream can have multiple pipelines connected to it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The connection ID (same as stream_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stream_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The stream ID to connect pipelines to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pipeline_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "Set of pipeline IDs to connect to the stream.",
			},
		},
	}
}

func (r *PipelineConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PipelineConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PipelineConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipelineIDs, diags := extractStringSet(ctx, data.PipelineIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connReq := &client.PipelineConnectionRequest{
		StreamID:    data.StreamID.ValueString(),
		PipelineIDs: pipelineIDs,
	}

	result, err := r.client.ConnectPipelinesToStream(ctx, connReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pipeline connection", err.Error())
		return
	}

	data.ID = types.StringValue(result.StreamID)
	data.StreamID = types.StringValue(result.StreamID)
	pipeSet, diags := types.SetValueFrom(ctx, types.StringType, result.PipelineIDs)
	resp.Diagnostics.Append(diags...)
	data.PipelineIDs = pipeSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PipelineConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := r.client.GetPipelineConnectionsForStream(ctx, data.StreamID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read pipeline connection", err.Error())
		return
	}

	data.ID = types.StringValue(conn.StreamID)
	data.StreamID = types.StringValue(conn.StreamID)
	pipeSet, diags := types.SetValueFrom(ctx, types.StringType, conn.PipelineIDs)
	resp.Diagnostics.Append(diags...)
	data.PipelineIDs = pipeSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PipelineConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipelineIDs, diags := extractStringSet(ctx, data.PipelineIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connReq := &client.PipelineConnectionRequest{
		StreamID:    data.StreamID.ValueString(),
		PipelineIDs: pipelineIDs,
	}

	result, err := r.client.ConnectPipelinesToStream(ctx, connReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update pipeline connection", err.Error())
		return
	}

	data.ID = types.StringValue(result.StreamID)
	pipeSet, setDiags := types.SetValueFrom(ctx, types.StringType, result.PipelineIDs)
	resp.Diagnostics.Append(setDiags...)
	data.PipelineIDs = pipeSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PipelineConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Disconnect all pipelines by setting an empty list
	connReq := &client.PipelineConnectionRequest{
		StreamID:    data.StreamID.ValueString(),
		PipelineIDs: []string{},
	}

	_, err := r.client.ConnectPipelinesToStream(ctx, connReq)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete pipeline connection", err.Error())
	}
}

func (r *PipelineConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("stream_id"), req, resp)
}

func extractStringSet(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var result []string
	diags := set.ElementsAs(ctx, &result, false)
	return result, diags
}
