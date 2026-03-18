package provider

import (
	"context"
	"fmt"
	"strings"

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
	ID            types.String `tfsdk:"id"`
	InputID       types.String `tfsdk:"input_id"`
	Title         types.String `tfsdk:"title"`
	ExtractorType types.String `tfsdk:"extractor_type"`
	PayloadJSON   types.String `tfsdk:"payload_json"`
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
			"title":          schema.StringAttribute{Computed: true},
			"extractor_type": schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Raw JSON payload for the extractor request body.",
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

	extractorReq, diags := extractorFromPayload(data.PayloadJSON.ValueString())
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
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalExtractorJSON(created))
	}
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
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalExtractorJSON(current))
	}
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

	extractorReq, diags := extractorFromPayload(data.PayloadJSON.ValueString())
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
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalExtractorJSON(updated))
	}
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
