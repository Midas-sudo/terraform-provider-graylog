// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

func eventNotificationFromPayload(payload string) (*client.EventNotification, diag.Diagnostics) {
	var diags diag.Diagnostics
	var notification client.EventNotification
	if err := json.Unmarshal([]byte(payload), &notification); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &notification, diags
}

func eventDefinitionFromPayload(payload string) (*client.EventDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	var definition client.EventDefinition
	if err := json.Unmarshal([]byte(payload), &definition); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &definition, diags
}

func marshalEventNotificationJSON(notification *client.EventNotification) string {
	clone := *notification
	clone.ID = ""
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func marshalEventDefinitionJSON(definition *client.EventDefinition) string {
	clone := *definition
	clone.ID = ""
	clone.Scope = ""
	clone.EntitySource = nil
	clone.UpdatedAt = ""
	clone.MatchedAt = ""
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func mapEventNotificationToResourceModel(ctx context.Context, notification *client.EventNotification, data *EventNotificationResourceModel) diag.Diagnostics {
	data.ID = types.StringValue(notification.ID)
	data.Title = types.StringValue(notification.Title)
	data.Description = types.StringValue(notification.Description)
	// Preserve planned/state Dynamic config: API responses often add attributes and
	// change number encoding, which breaks Framework Dynamic object type equality.
	if !data.Config.IsNull() && !data.Config.IsUnknown() {
		return nil
	}
	cfg := notification.Config
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	dyn, diags := interfaceToDynamic(ctx, cfg)
	data.Config = dyn
	return diags
}

func mapEventDefinitionToResourceModel(ctx context.Context, definition *client.EventDefinition, data *EventDefinitionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ID = types.StringValue(definition.ID)
	data.Title = types.StringValue(definition.Title)
	data.Description = types.StringValue(definition.Description)
	data.State = types.StringValue(definition.State)
	data.Priority = types.Int64Value(definition.Priority)
	data.Alert = types.BoolValue(definition.Alert)

	if data.Config.IsNull() || data.Config.IsUnknown() {
		cfg := definition.Config
		if cfg == nil {
			cfg = map[string]interface{}{}
		}
		cfgDyn, d := interfaceToDynamic(ctx, cfg)
		diags.Append(d...)
		data.Config = cfgDyn
	}

	if data.FieldSpec.IsNull() || data.FieldSpec.IsUnknown() {
		if definition.FieldSpec != nil {
			fsDyn, d := interfaceToDynamic(ctx, definition.FieldSpec)
			diags.Append(d...)
			data.FieldSpec = fsDyn
		} else {
			data.FieldSpec = types.DynamicNull()
		}
	}

	if definition.KeySpec != nil {
		keys := make([]types.String, 0, len(definition.KeySpec))
		for _, k := range definition.KeySpec {
			keys = append(keys, types.StringValue(k))
		}
		data.KeySpec = keys
	}

	if definition.NotificationSettings != nil {
		data.NotificationSettings = &eventDefinitionNotificationSettingsModel{
			GracePeriodMs: types.Int64Value(jsonNumberAsInt64(definition.NotificationSettings["grace_period_ms"])),
			BacklogSize:   types.Int64Value(jsonNumberAsInt64(definition.NotificationSettings["backlog_size"])),
		}
	}

	if data.Notifications.IsNull() || data.Notifications.IsUnknown() {
		if definition.Notifications != nil {
			raw := make([]interface{}, 0, len(definition.Notifications))
			for _, n := range definition.Notifications {
				item := map[string]interface{}{
					"notification_id": n.NotificationID,
				}
				if n.NotificationParameters != nil {
					item["notification_parameters"] = n.NotificationParameters
				}
				raw = append(raw, item)
			}
			nDyn, d := interfaceToDynamic(ctx, raw)
			diags.Append(d...)
			data.Notifications = nDyn
		} else {
			empty, d := interfaceToDynamic(ctx, []interface{}{})
			diags.Append(d...)
			data.Notifications = empty
		}
	}

	if data.Storage.IsNull() || data.Storage.IsUnknown() {
		if definition.Storage != nil {
			raw := make([]interface{}, 0, len(definition.Storage))
			for _, s := range definition.Storage {
				raw = append(raw, s)
			}
			sDyn, d := interfaceToDynamic(ctx, raw)
			diags.Append(d...)
			data.Storage = sDyn
		} else {
			data.Storage = types.DynamicNull()
		}
	}

	return diags
}

func mapEventNotificationToDataSourceModel(ctx context.Context, notification *client.EventNotification, data *EventNotificationDataSourceModel) diag.Diagnostics {
	data.ID = types.StringValue(notification.ID)
	data.Title = types.StringValue(notification.Title)
	data.Description = types.StringValue(notification.Description)
	cfg := notification.Config
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	dyn, diags := interfaceToDynamic(ctx, cfg)
	data.Config = dyn
	return diags
}

func mapEventDefinitionToDataSourceModel(ctx context.Context, definition *client.EventDefinition, data *EventDefinitionDataSourceModel) diag.Diagnostics {
	// Reuse resource mapping into a temporary model, then copy shared fields.
	tmp := EventDefinitionResourceModel{}
	diags := mapEventDefinitionToResourceModel(ctx, definition, &tmp)
	data.ID = tmp.ID
	data.Title = tmp.Title
	data.Description = tmp.Description
	data.State = tmp.State
	data.Priority = tmp.Priority
	data.Alert = tmp.Alert
	data.Config = tmp.Config
	data.FieldSpec = tmp.FieldSpec
	data.KeySpec = tmp.KeySpec
	data.NotificationSettings = tmp.NotificationSettings
	data.Notifications = tmp.Notifications
	data.Storage = tmp.Storage
	return diags
}

func mapEventDefinitionToListItemModel(definition *client.EventDefinition) eventDefinitionListItemModel {
	item := eventDefinitionListItemModel{
		ID:          types.StringValue(definition.ID),
		Title:       types.StringValue(definition.Title),
		Description: types.StringValue(definition.Description),
		State:       types.StringValue(definition.State),
		Priority:    types.Int64Value(definition.Priority),
		Alert:       types.BoolValue(definition.Alert),
		Config:      lookupConfigJSONString(definition.Config),
	}
	if definition.FieldSpec != nil {
		item.FieldSpec = lookupConfigJSONString(definition.FieldSpec)
	} else {
		item.FieldSpec = types.StringNull()
	}
	if definition.KeySpec != nil {
		keys := make([]types.String, 0, len(definition.KeySpec))
		for _, v := range definition.KeySpec {
			keys = append(keys, types.StringValue(v))
		}
		item.KeySpec = keys
	}
	if definition.NotificationSettings != nil {
		item.NotificationSettings = &eventDefinitionNotificationSettingsModel{
			GracePeriodMs: types.Int64Value(jsonNumberAsInt64(definition.NotificationSettings["grace_period_ms"])),
			BacklogSize:   types.Int64Value(jsonNumberAsInt64(definition.NotificationSettings["backlog_size"])),
		}
	}
	if definition.Notifications != nil {
		if b, err := json.Marshal(definition.Notifications); err == nil {
			item.Notifications = types.StringValue(string(b))
		} else {
			item.Notifications = types.StringValue("[]")
		}
	} else {
		item.Notifications = types.StringValue("[]")
	}
	if definition.Storage != nil {
		if b, err := json.Marshal(definition.Storage); err == nil {
			item.Storage = types.StringValue(string(b))
		} else {
			item.Storage = types.StringNull()
		}
	} else {
		item.Storage = types.StringNull()
	}
	return item
}

func eventDefinitionNotificationIDs(definition *client.EventDefinition) []string {
	ids := make([]string, 0, len(definition.Notifications))
	for _, n := range definition.Notifications {
		if n.NotificationID != "" {
			ids = append(ids, n.NotificationID)
		}
	}
	sort.Strings(ids)
	return ids
}

func eventDefinitionNotificationsFromIDs(ids []string) []client.EventDefinitionNotification {
	sortedIDs := append([]string(nil), ids...)
	sort.Strings(sortedIDs)

	notifications := make([]client.EventDefinitionNotification, 0, len(sortedIDs))
	for _, id := range sortedIDs {
		if id == "" {
			continue
		}
		notifications = append(notifications, client.EventDefinitionNotification{
			NotificationID: id,
		})
	}
	return notifications
}
