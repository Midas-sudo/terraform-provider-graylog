// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
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

func mapEventNotificationToResourceModel(notification *client.EventNotification, data *EventNotificationResourceModel) {
	data.ID = types.StringValue(notification.ID)
	data.Title = types.StringValue(notification.Title)
	data.Description = types.StringValue(notification.Description)
}

func mapEventDefinitionToResourceModel(definition *client.EventDefinition, data *EventDefinitionResourceModel) {
	data.ID = types.StringValue(definition.ID)
	data.Title = types.StringValue(definition.Title)
	data.Description = types.StringValue(definition.Description)
	data.State = types.StringValue(definition.State)
	data.Priority = types.Int64Value(definition.Priority)
	data.Alert = types.BoolValue(definition.Alert)
}

func mapEventNotificationToDataSourceModel(notification *client.EventNotification, data *EventNotificationDataSourceModel) {
	data.ID = types.StringValue(notification.ID)
	data.Title = types.StringValue(notification.Title)
	data.Description = types.StringValue(notification.Description)
	if notification.Config != nil {
		if b, err := json.Marshal(notification.Config); err == nil {
			data.ConfigJSON = types.StringValue(string(b))
		} else {
			data.ConfigJSON = types.StringValue("{}")
		}
	} else {
		data.ConfigJSON = types.StringValue("{}")
	}
}

func mapEventDefinitionToDataSourceModel(definition *client.EventDefinition, data *EventDefinitionDataSourceModel) {
	data.ID = types.StringValue(definition.ID)
	data.Title = types.StringValue(definition.Title)
	data.Description = types.StringValue(definition.Description)
	data.State = types.StringValue(definition.State)
	data.Priority = types.Int64Value(definition.Priority)
	data.Alert = types.BoolValue(definition.Alert)
	if definition.Config != nil {
		if b, err := json.Marshal(definition.Config); err == nil {
			data.ConfigJSON = types.StringValue(string(b))
		} else {
			data.ConfigJSON = types.StringValue("{}")
		}
	} else {
		data.ConfigJSON = types.StringValue("{}")
	}
	if definition.FieldSpec != nil {
		if b, err := json.Marshal(definition.FieldSpec); err == nil {
			data.FieldSpecJSON = types.StringValue(string(b))
		}
	}
	if definition.KeySpec != nil {
		keySpec := make([]types.String, 0, len(definition.KeySpec))
		for _, v := range definition.KeySpec {
			keySpec = append(keySpec, types.StringValue(v))
		}
		data.KeySpec = keySpec
	}
	if definition.NotificationSettings != nil {
		if b, err := json.Marshal(definition.NotificationSettings); err == nil {
			data.NotificationSettingsJSON = types.StringValue(string(b))
		}
	}
	if definition.Notifications != nil {
		if b, err := json.Marshal(definition.Notifications); err == nil {
			data.NotificationsJSON = types.StringValue(string(b))
		}
	}
	if definition.Storage != nil {
		if b, err := json.Marshal(definition.Storage); err == nil {
			data.StorageJSON = types.StringValue(string(b))
		}
	}
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
