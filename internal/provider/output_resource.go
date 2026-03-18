package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &OutputResource{}
	_ resource.ResourceWithImportState = &OutputResource{}
)

func NewOutputResource() resource.Resource {
	return &OutputResource{}
}

type OutputResource struct {
	client *client.Client
}

type OutputResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Type        types.String `tfsdk:"type"`
	PayloadJSON types.String `tfsdk:"payload_json"`
}

func (r *OutputResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_output"
}

func (r *OutputResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog output.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"title": schema.StringAttribute{Computed: true},
			"type":  schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Raw JSON payload for the output object.",
			},
		},
	}
}

func (r *OutputResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OutputResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OutputResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outputReq, diags := outputFromPayload(data.PayloadJSON.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateOutput(ctx, outputReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create output", err.Error())
		return
	}

	mapOutputToResourceModel(created, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalOutputJSON(created))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OutputResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OutputResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetOutput(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read output", err.Error())
		return
	}

	mapOutputToResourceModel(current, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalOutputJSON(current))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OutputResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OutputResourceModel
	var state OutputResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Graylog output PUT is unreliable in this target environment; replace on update.
	outputReq, diags := outputFromPayload(data.PayloadJSON.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateOutput(ctx, outputReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to recreate output during update", err.Error())
		return
	}

	if err := r.client.DeleteOutput(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete old output during update", err.Error())
		return
	}

	mapOutputToResourceModel(created, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalOutputJSON(created))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OutputResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OutputResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOutput(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete output", err.Error())
	}
}

func (r *OutputResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
