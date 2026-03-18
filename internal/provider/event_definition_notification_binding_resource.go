package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var (
	_ resource.Resource                = &EventDefinitionNotificationBindingResource{}
	_ resource.ResourceWithImportState = &EventDefinitionNotificationBindingResource{}
)

func NewEventDefinitionNotificationBindingResource() resource.Resource {
	return &EventDefinitionNotificationBindingResource{}
}

type EventDefinitionNotificationBindingResource struct {
	client *client.Client
}

type EventDefinitionNotificationBindingResourceModel struct {
	ID                types.String `tfsdk:"id"`
	EventDefinitionID types.String `tfsdk:"event_definition_id"`
	NotificationIDs   types.Set    `tfsdk:"notification_ids"`
}

func (r *EventDefinitionNotificationBindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_definition_notification_binding"
}

func (r *EventDefinitionNotificationBindingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages notification bindings for a Graylog event definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Binding ID (same as `event_definition_id`).",
			},
			"event_definition_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Event definition ID to bind notifications to.",
			},
			"notification_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Set of event notification IDs attached to this event definition.",
			},
		},
	}
}

func (r *EventDefinitionNotificationBindingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EventDefinitionNotificationBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EventDefinitionNotificationBindingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyBinding(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventDefinitionNotificationBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EventDefinitionNotificationBindingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	definition, err := r.client.GetEventDefinition(ctx, data.EventDefinitionID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read event definition binding", err.Error())
		return
	}

	ids := eventDefinitionNotificationIDs(definition)
	notificationSet, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(data.EventDefinitionID.ValueString())
	data.NotificationIDs = notificationSet
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventDefinitionNotificationBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EventDefinitionNotificationBindingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyBinding(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EventDefinitionNotificationBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EventDefinitionNotificationBindingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	definition, err := r.client.GetEventDefinition(ctx, data.EventDefinitionID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to read event definition for delete", err.Error())
		return
	}

	definition.Notifications = []client.EventDefinitionNotification{}
	if _, err := r.client.UpdateEventDefinition(ctx, definition.ID, definition); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to clear event definition notifications", err.Error())
	}
}

func (r *EventDefinitionNotificationBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root("event_definition_id"), req, resp)
}

func (r *EventDefinitionNotificationBindingResource) applyBinding(
	ctx context.Context,
	data *EventDefinitionNotificationBindingResourceModel,
	diags *diag.Diagnostics,
) {
	definition, err := r.client.GetEventDefinition(ctx, data.EventDefinitionID.ValueString())
	if err != nil {
		diags.AddError("Failed to read event definition", err.Error())
		return
	}

	var ids []string
	diags.Append(data.NotificationIDs.ElementsAs(ctx, &ids, false)...)
	if diags.HasError() {
		return
	}

	definition.Notifications = eventDefinitionNotificationsFromIDs(ids)
	updated, err := r.client.UpdateEventDefinition(ctx, definition.ID, definition)
	if err != nil {
		diags.AddError("Failed to update event definition notifications", err.Error())
		return
	}

	boundIDs := eventDefinitionNotificationIDs(updated)
	notificationSet, setDiags := types.SetValueFrom(ctx, types.StringType, boundIDs)
	diags.Append(setDiags...)
	if diags.HasError() {
		return
	}

	data.ID = types.StringValue(data.EventDefinitionID.ValueString())
	data.NotificationIDs = notificationSet
}
