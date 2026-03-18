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
	_ resource.Resource                = &ContentPackResource{}
	_ resource.ResourceWithImportState = &ContentPackResource{}
)

func NewContentPackResource() resource.Resource {
	return &ContentPackResource{}
}

type ContentPackResource struct {
	client *client.Client
}

type ContentPackResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ContentPackID types.String `tfsdk:"content_pack_id"`
	Revision      types.Int64  `tfsdk:"revision"`
	Name          types.String `tfsdk:"name"`
	PayloadJSON   types.String `tfsdk:"payload_json"`
}

func (r *ContentPackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_pack"
}

func (r *ContentPackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Graylog content pack revision.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Composite ID formatted as `content_pack_id/revision`.",
			},
			"content_pack_id": schema.StringAttribute{
				Computed: true,
			},
			"revision": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"payload_json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Raw JSON payload for content pack creation.",
			},
		},
	}
}

func (r *ContentPackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContentPackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContentPackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	contentPackReq, diags := contentPackFromPayload(data.PayloadJSON.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CreateContentPack(ctx, contentPackReq); err != nil {
		resp.Diagnostics.AddError("Failed to create content pack", err.Error())
		return
	}

	created, err := r.client.GetContentPack(ctx, contentPackReq.ID, contentPackReq.Rev)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created content pack", err.Error())
		return
	}

	mapContentPackToResourceModel(created, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalContentPackJSON(created))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentPackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContentPackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	contentPack, err := r.client.GetContentPack(ctx, data.ContentPackID.ValueString(), data.Revision.ValueInt64())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read content pack", err.Error())
		return
	}

	mapContentPackToResourceModel(contentPack, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalContentPackJSON(contentPack))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentPackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContentPackResourceModel
	var state ContentPackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	contentPackReq, diags := contentPackFromPayload(data.PayloadJSON.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteContentPack(ctx, state.ContentPackID.ValueString(), state.Revision.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete old content pack revision", err.Error())
		return
	}
	if err := r.client.CreateContentPack(ctx, contentPackReq); err != nil {
		resp.Diagnostics.AddError("Failed to create updated content pack revision", err.Error())
		return
	}

	updated, err := r.client.GetContentPack(ctx, contentPackReq.ID, contentPackReq.Rev)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated content pack", err.Error())
		return
	}

	mapContentPackToResourceModel(updated, &data)
	if data.PayloadJSON.IsNull() || data.PayloadJSON.IsUnknown() || data.PayloadJSON.ValueString() == "" {
		data.PayloadJSON = types.StringValue(marshalContentPackJSON(updated))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentPackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContentPackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteContentPack(ctx, data.ContentPackID.ValueString(), data.Revision.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete content pack", err.Error())
	}
}

func (r *ContentPackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	contentPackID, rev, err := parseContentPackImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("content_pack_id"), contentPackID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("revision"), rev)...)
}
