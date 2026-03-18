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
	_ resource.Resource                = &GrokPatternResource{}
	_ resource.ResourceWithImportState = &GrokPatternResource{}
)

func NewGrokPatternResource() resource.Resource {
	return &GrokPatternResource{}
}

type GrokPatternResource struct {
	client *client.Client
}

type GrokPatternResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Pattern     types.String `tfsdk:"pattern"`
	PayloadJSON types.String `tfsdk:"payload_json"`
}

func (r *GrokPatternResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grok_pattern"
}

func (r *GrokPatternResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog grok pattern.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name":    schema.StringAttribute{Computed: true},
			"pattern": schema.StringAttribute{Computed: true},
			"payload_json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Raw JSON payload for the grok pattern object.",
			},
		},
	}
}

func (r *GrokPatternResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GrokPatternResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GrokPatternResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patternReq, diags := grokPatternFromPayload(data.PayloadJSON.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateGrokPattern(ctx, patternReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create grok pattern", err.Error())
		return
	}

	mapGrokPatternToResourceModel(created, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalGrokPatternJSON(created))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrokPatternResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GrokPatternResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetGrokPattern(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read grok pattern", err.Error())
		return
	}

	mapGrokPatternToResourceModel(current, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalGrokPatternJSON(current))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrokPatternResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GrokPatternResourceModel
	var state GrokPatternResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patternReq, diags := grokPatternFromPayload(data.PayloadJSON.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateGrokPattern(ctx, state.ID.ValueString(), patternReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update grok pattern", err.Error())
		return
	}

	mapGrokPatternToResourceModel(updated, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalGrokPatternJSON(updated))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GrokPatternResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GrokPatternResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGrokPattern(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete grok pattern", err.Error())
	}
}

func (r *GrokPatternResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
