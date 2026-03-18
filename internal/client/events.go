// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type EventNotification struct {
	ID          string                 `json:"id,omitempty"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config"`
}

type EventDefinitionNotification struct {
	NotificationID         string      `json:"notification_id"`
	NotificationParameters interface{} `json:"notification_parameters,omitempty"`
}

type EventDefinition struct {
	ID                   string                        `json:"id,omitempty"`
	Title                string                        `json:"title"`
	Description          string                        `json:"description,omitempty"`
	Priority             int64                         `json:"priority"`
	Alert                bool                          `json:"alert"`
	Config               map[string]interface{}        `json:"config"`
	FieldSpec            map[string]interface{}        `json:"field_spec,omitempty"`
	KeySpec              []string                      `json:"key_spec"`
	NotificationSettings map[string]interface{}        `json:"notification_settings,omitempty"`
	Notifications        []EventDefinitionNotification `json:"notifications,omitempty"`
	Storage              []map[string]interface{}      `json:"storage,omitempty"`
	Scheduler            interface{}                   `json:"scheduler,omitempty"`
	State                string                        `json:"state,omitempty"`
	Scope                string                        `json:"_scope,omitempty"`
	EntitySource         map[string]interface{}        `json:"_entity_source,omitempty"`
	UpdatedAt            string                        `json:"updated_at,omitempty"`
	MatchedAt            string                        `json:"matched_at,omitempty"`
}

type eventNotificationEntityRequest struct {
	Entity *EventNotification `json:"entity"`
}

type eventDefinitionEntityRequest struct {
	Entity *EventDefinition `json:"entity"`
}

type EventNotificationsResponse struct {
	Total         int                 `json:"total"`
	Page          int                 `json:"page"`
	PerPage       int                 `json:"per_page"`
	Count         int                 `json:"count"`
	Notifications []EventNotification `json:"notifications"`
}

type EventDefinitionsResponse struct {
	Total            int               `json:"total"`
	Page             int               `json:"page"`
	PerPage          int               `json:"per_page"`
	Count            int               `json:"count"`
	EventDefinitions []EventDefinition `json:"event_definitions"`
}

func (c *Client) GetEventNotifications(ctx context.Context) (*EventNotificationsResponse, error) {
	var result EventNotificationsResponse
	if err := c.Get(ctx, "/events/notifications", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetEventNotification(ctx context.Context, id string) (*EventNotification, error) {
	var result EventNotification
	if err := c.Get(ctx, fmt.Sprintf("/events/notifications/%s", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateEventNotification(ctx context.Context, notification *EventNotification) (*EventNotification, error) {
	var result EventNotification
	if err := c.Post(ctx, "/events/notifications", &eventNotificationEntityRequest{Entity: notification}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateEventNotification(ctx context.Context, id string, notification *EventNotification) (*EventNotification, error) {
	if notification.ID == "" {
		notification.ID = id
	}
	var result EventNotification
	if err := c.Put(ctx, fmt.Sprintf("/events/notifications/%s", id), notification, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteEventNotification(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/events/notifications/%s", id))
}

func (c *Client) GetEventDefinitions(ctx context.Context) (*EventDefinitionsResponse, error) {
	var result EventDefinitionsResponse
	if err := c.Get(ctx, "/events/definitions", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetEventDefinition(ctx context.Context, id string) (*EventDefinition, error) {
	var result EventDefinition
	if err := c.Get(ctx, fmt.Sprintf("/events/definitions/%s", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateEventDefinition(ctx context.Context, definition *EventDefinition) (*EventDefinition, error) {
	normalizeEventDefinition(definition)
	var result EventDefinition
	if err := c.Post(ctx, "/events/definitions?schedule=true", &eventDefinitionEntityRequest{Entity: definition}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateEventDefinition(ctx context.Context, id string, definition *EventDefinition) (*EventDefinition, error) {
	definition.ID = id
	normalizeEventDefinition(definition)
	var result EventDefinition
	if err := c.Put(ctx, fmt.Sprintf("/events/definitions/%s?schedule=true", id), definition, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteEventDefinition(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/events/definitions/%s", id))
}

func normalizeEventDefinition(definition *EventDefinition) {
	if definition.Config == nil {
		definition.Config = map[string]interface{}{}
	}
	if definition.FieldSpec == nil {
		definition.FieldSpec = map[string]interface{}{}
	}
	if definition.KeySpec == nil {
		definition.KeySpec = []string{}
	}
	if definition.NotificationSettings == nil {
		definition.NotificationSettings = map[string]interface{}{}
	}
	if definition.Notifications == nil {
		definition.Notifications = []EventDefinitionNotification{}
	}
	if definition.Storage == nil {
		definition.Storage = []map[string]interface{}{}
	}
}
