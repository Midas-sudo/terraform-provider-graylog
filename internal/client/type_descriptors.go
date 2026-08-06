// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

// ConfigField describes one Graylog plugin configuration key
// (from requested_configuration, json_schema properties, or default_config keys).
type ConfigField struct {
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	HumanName      string      `json:"human_name"`
	Description    string      `json:"description"`
	IsOptional     bool        `json:"is_optional"`
	DefaultValue   interface{} `json:"default_value"`
	Attributes     []string    `json:"attributes"`
	AdditionalInfo interface{} `json:"additional_info,omitempty"`
	Position       int         `json:"position,omitempty"`
}

// TypeDescriptor is a normalized plugin / strategy type description.
type TypeDescriptor struct {
	Type                   string                 `json:"type"`
	Name                   string                 `json:"name"`
	Description            string                 `json:"description,omitempty"`
	LinkToDocs             string                 `json:"link_to_docs,omitempty"`
	RequestedConfiguration []ConfigField          `json:"requested_configuration,omitempty"`
	DefaultConfig          map[string]interface{} `json:"default_config,omitempty"`
}

// --- Raw Graylog response shapes ---

type rawConfigField struct {
	HumanName      string      `json:"human_name"`
	Description    string      `json:"description"`
	DefaultValue   interface{} `json:"default_value"`
	Attributes     []string    `json:"attributes"`
	AdditionalInfo interface{} `json:"additional_info"`
	Position       int         `json:"position"`
	Type           string      `json:"type"`
	IsOptional     bool        `json:"is_optional"`
}

type rawInputTypeInfo struct {
	Type                   string                    `json:"type"`
	Name                   string                    `json:"name"`
	Description            string                    `json:"description"`
	LinkToDocs             string                    `json:"link_to_docs"`
	IsExclusive            bool                      `json:"is_exclusive"`
	RequestedConfiguration map[string]rawConfigField `json:"requested_configuration"`
}

type rawOutputTypeInfo struct {
	Name                   string                    `json:"name"`
	Type                   string                    `json:"type"`
	HumanName              string                    `json:"human_name"`
	LinkToDocs             string                    `json:"link_to_docs"`
	RequestedConfiguration map[string]rawConfigField `json:"requested_configuration"`
}

type AvailableOutputsResponse struct {
	Types map[string]rawOutputTypeInfo `json:"types"`
}

type rawLookupTypeInfo struct {
	Type          string                 `json:"type"`
	ConfigClass   string                 `json:"config_class"`
	DefaultConfig map[string]interface{} `json:"default_config"`
}

type rawJSONSchema struct {
	Type       string                            `json:"type"`
	ID         string                            `json:"id"`
	Required   []string                          `json:"required"`
	Properties map[string]map[string]interface{} `json:"properties"`
}

type rawStrategyInfo struct {
	Type          string                 `json:"type"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	JSONSchema    rawJSONSchema          `json:"json_schema"`
}

type IndexStrategiesResponse struct {
	Total      int               `json:"total"`
	Strategies []rawStrategyInfo `json:"strategies"`
}

type EventEntityTypesResponse struct {
	ProcessorTypes       []string `json:"processor_types"`
	FieldProviderTypes   []string `json:"field_provider_types"`
	StorageHandlerTypes  []string `json:"storage_handler_types"`
	AggregationFunctions []string `json:"aggregation_functions"`
}

type rawLegacyNotificationType struct {
	Name          string                    `json:"name"`
	Configuration map[string]rawConfigField `json:"configuration"`
}

type LegacyNotificationTypesResponse struct {
	Types map[string]rawLegacyNotificationType `json:"types"`
}

// --- API methods ---

func (c *Client) GetInputTypesAll(ctx context.Context) (map[string]TypeDescriptor, error) {
	var raw map[string]rawInputTypeInfo
	if err := c.Get(ctx, "/system/inputs/types/all", &raw); err != nil {
		return nil, err
	}
	out := make(map[string]TypeDescriptor, len(raw))
	for key, info := range raw {
		typeName := info.Type
		if typeName == "" {
			typeName = key
		}
		out[key] = TypeDescriptor{
			Type:                   typeName,
			Name:                   info.Name,
			Description:            info.Description,
			LinkToDocs:             info.LinkToDocs,
			RequestedConfiguration: fieldsFromRequestedConfiguration(info.RequestedConfiguration),
		}
	}
	return out, nil
}

func (c *Client) GetAvailableOutputs(ctx context.Context) (map[string]TypeDescriptor, error) {
	var raw AvailableOutputsResponse
	if err := c.Get(ctx, "/system/outputs/available", &raw); err != nil {
		return nil, err
	}
	out := make(map[string]TypeDescriptor, len(raw.Types))
	for key, info := range raw.Types {
		typeName := info.Type
		if typeName == "" {
			typeName = key
		}
		name := info.HumanName
		if name == "" {
			name = info.Name
		}
		out[key] = TypeDescriptor{
			Type:                   typeName,
			Name:                   name,
			LinkToDocs:             info.LinkToDocs,
			RequestedConfiguration: fieldsFromRequestedConfiguration(info.RequestedConfiguration),
		}
	}
	return out, nil
}

func (c *Client) GetLookupAdapterTypes(ctx context.Context) (map[string]TypeDescriptor, error) {
	var raw map[string]rawLookupTypeInfo
	if err := c.Get(ctx, "/system/lookup/types/adapters", &raw); err != nil {
		return nil, err
	}
	return lookupTypesToDescriptors(raw), nil
}

func (c *Client) GetLookupCacheTypes(ctx context.Context) (map[string]TypeDescriptor, error) {
	var raw map[string]rawLookupTypeInfo
	if err := c.Get(ctx, "/system/lookup/types/caches", &raw); err != nil {
		return nil, err
	}
	return lookupTypesToDescriptors(raw), nil
}

func (c *Client) GetRotationStrategies(ctx context.Context) ([]TypeDescriptor, error) {
	var raw IndexStrategiesResponse
	if err := c.Get(ctx, "/system/indices/rotation/strategies", &raw); err != nil {
		return nil, err
	}
	return strategiesToDescriptors(raw.Strategies), nil
}

func (c *Client) GetRetentionStrategies(ctx context.Context) ([]TypeDescriptor, error) {
	var raw IndexStrategiesResponse
	if err := c.Get(ctx, "/system/indices/retention/strategies", &raw); err != nil {
		return nil, err
	}
	return strategiesToDescriptors(raw.Strategies), nil
}

func (c *Client) GetEventEntityTypes(ctx context.Context) (*EventEntityTypesResponse, error) {
	var result EventEntityTypesResponse
	if err := c.Get(ctx, "/events/entity_types", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetLegacyEventNotificationTypes(ctx context.Context) (map[string]TypeDescriptor, error) {
	var raw LegacyNotificationTypesResponse
	if err := c.Get(ctx, "/events/notifications/legacy/types", &raw); err != nil {
		return nil, err
	}
	out := make(map[string]TypeDescriptor, len(raw.Types))
	for key, info := range raw.Types {
		out[key] = TypeDescriptor{
			Type:                   key,
			Name:                   info.Name,
			RequestedConfiguration: fieldsFromRequestedConfiguration(info.Configuration),
		}
	}
	return out, nil
}

// ModernEventNotificationTypes returns built-in modern notification types.
// Graylog 7 does not expose a discovery API for these (only legacy alarm callbacks).
func ModernEventNotificationTypes() []TypeDescriptor {
	return []TypeDescriptor{
		{
			Type:        "http-notification-v1",
			Name:        "HTTP Notification",
			Description: "Send an HTTP request when an event is triggered.",
			RequestedConfiguration: []ConfigField{
				{Name: "type", Type: "string", HumanName: "Type", Description: "Must be http-notification-v1.", IsOptional: false, DefaultValue: "http-notification-v1"},
				{Name: "url", Type: "string", HumanName: "URL", Description: "Destination URL.", IsOptional: false},
				{Name: "basic_auth", Type: "object", HumanName: "Basic auth", Description: "Optional encrypted basic-auth credentials.", IsOptional: true},
				{Name: "api_key", Type: "string", HumanName: "API key", Description: "Optional API key.", IsOptional: true},
				{Name: "api_secret", Type: "object", HumanName: "API secret", Description: "Optional encrypted API secret (requires api_key).", IsOptional: true},
				{Name: "api_key_as_header", Type: "boolean", HumanName: "API key as header", Description: "Send API key as an HTTP header.", IsOptional: true, DefaultValue: false},
				{Name: "skip_tls_verification", Type: "boolean", HumanName: "Skip TLS verification", Description: "Disable TLS certificate verification.", IsOptional: true, DefaultValue: false},
			},
		},
		{
			Type:        "http-notification-v2",
			Name:        "HTTP Notification V2",
			Description: "HTTP notification with customizable method, headers, and body template.",
			RequestedConfiguration: []ConfigField{
				{Name: "type", Type: "string", HumanName: "Type", Description: "Must be http-notification-v2.", IsOptional: false, DefaultValue: "http-notification-v2"},
				{Name: "url", Type: "string", HumanName: "URL", Description: "Destination URL.", IsOptional: false},
				{Name: "method", Type: "string", HumanName: "HTTP method", Description: "HTTP method (GET, POST, PUT, …).", IsOptional: true},
				{Name: "headers", Type: "string", HumanName: "Headers", Description: "Optional headers body/template.", IsOptional: true},
				{Name: "body_template", Type: "string", HumanName: "Body template", Description: "Optional request body template.", IsOptional: true},
				{Name: "content_type", Type: "string", HumanName: "Content type", Description: "Request content type.", IsOptional: true},
				{Name: "skip_tls_verification", Type: "boolean", HumanName: "Skip TLS verification", IsOptional: true, DefaultValue: false},
			},
		},
		{
			Type:        "email-notification-v1",
			Name:        "Email Notification",
			Description: "Send an email when an event is triggered.",
			RequestedConfiguration: []ConfigField{
				{Name: "type", Type: "string", HumanName: "Type", Description: "Must be email-notification-v1.", IsOptional: false, DefaultValue: "email-notification-v1"},
				{Name: "sender", Type: "string", HumanName: "Sender", Description: "From address.", IsOptional: true},
				{Name: "subject", Type: "string", HumanName: "Subject", Description: "Email subject template.", IsOptional: false},
				{Name: "body_template", Type: "string", HumanName: "Body template", Description: "Plain-text body template.", IsOptional: true},
				{Name: "html_body_template", Type: "string", HumanName: "HTML body template", Description: "HTML body template.", IsOptional: true},
				{Name: "email_recipients", Type: "list", HumanName: "Email recipients", Description: "Recipient email addresses.", IsOptional: true},
				{Name: "user_recipients", Type: "list", HumanName: "User recipients", Description: "Graylog usernames to notify.", IsOptional: true},
			},
		},
	}
}

func lookupTypesToDescriptors(raw map[string]rawLookupTypeInfo) map[string]TypeDescriptor {
	out := make(map[string]TypeDescriptor, len(raw))
	for key, info := range raw {
		typeName := info.Type
		if typeName == "" {
			typeName = key
		}
		out[key] = TypeDescriptor{
			Type:                   typeName,
			Name:                   typeName,
			DefaultConfig:          info.DefaultConfig,
			RequestedConfiguration: fieldsFromDefaultConfig(info.DefaultConfig),
		}
	}
	return out
}

func strategiesToDescriptors(strategies []rawStrategyInfo) []TypeDescriptor {
	out := make([]TypeDescriptor, 0, len(strategies))
	for _, s := range strategies {
		out = append(out, TypeDescriptor{
			Type:                   s.Type,
			Name:                   s.Type,
			DefaultConfig:          s.DefaultConfig,
			RequestedConfiguration: fieldsFromJSONSchema(s.JSONSchema, s.DefaultConfig),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Type < out[j].Type })
	return out
}

func fieldsFromRequestedConfiguration(m map[string]rawConfigField) []ConfigField {
	if len(m) == 0 {
		return nil
	}
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		pi, pj := m[names[i]].Position, m[names[j]].Position
		if pi != pj {
			return pi < pj
		}
		return names[i] < names[j]
	})
	fields := make([]ConfigField, 0, len(names))
	for _, name := range names {
		f := m[name]
		attrs := f.Attributes
		if attrs == nil {
			attrs = []string{}
		}
		fields = append(fields, ConfigField{
			Name:           name,
			Type:           f.Type,
			HumanName:      f.HumanName,
			Description:    f.Description,
			IsOptional:     f.IsOptional,
			DefaultValue:   f.DefaultValue,
			Attributes:     attrs,
			AdditionalInfo: f.AdditionalInfo,
			Position:       f.Position,
		})
	}
	return fields
}

func fieldsFromDefaultConfig(cfg map[string]interface{}) []ConfigField {
	if len(cfg) == 0 {
		return nil
	}
	names := make([]string, 0, len(cfg))
	for name := range cfg {
		names = append(names, name)
	}
	sort.Strings(names)
	fields := make([]ConfigField, 0, len(names))
	for _, name := range names {
		v := cfg[name]
		fields = append(fields, ConfigField{
			Name:         name,
			Type:         inferJSONType(v),
			HumanName:    name,
			IsOptional:   name != "type",
			DefaultValue: v,
			Attributes:   []string{},
		})
	}
	return fields
}

func fieldsFromJSONSchema(schema rawJSONSchema, defaults map[string]interface{}) []ConfigField {
	required := map[string]struct{}{}
	for _, r := range schema.Required {
		required[r] = struct{}{}
	}
	names := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	fields := make([]ConfigField, 0, len(names))
	for _, name := range names {
		prop := schema.Properties[name]
		typeName, _ := prop["type"].(string)
		if typeName == "" {
			typeName = "string"
		}
		_, req := required[name]
		var def interface{}
		if defaults != nil {
			def = defaults[name]
		}
		fields = append(fields, ConfigField{
			Name:         name,
			Type:         typeName,
			HumanName:    name,
			IsOptional:   !req && name != "type",
			DefaultValue: def,
			Attributes:   []string{},
		})
	}
	return fields
}

func inferJSONType(v interface{}) string {
	switch v.(type) {
	case bool:
		return "boolean"
	case float64, json.Number, int, int64:
		return "number"
	case []interface{}:
		return "list"
	case map[string]interface{}:
		return "object"
	case nil:
		return "string"
	default:
		return "string"
	}
}

// DefaultConfigJSON returns the default_config map encoded as JSON, or empty string.
func DefaultConfigJSON(cfg map[string]interface{}) (string, error) {
	if cfg == nil {
		return "", nil
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("encode default_config: %w", err)
	}
	return string(b), nil
}

// DefaultValueJSON returns a single default value encoded as JSON string.
func DefaultValueJSON(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode default_value: %w", err)
	}
	return string(b), nil
}

// AdditionalInfoJSON returns additional_info encoded as JSON, or empty string.
func AdditionalInfoJSON(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	// Skip empty objects.
	if m, ok := v.(map[string]interface{}); ok && len(m) == 0 {
		return "", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode additional_info: %w", err)
	}
	return string(b), nil
}

// SortedTypeDescriptors returns map values sorted by Type.
func SortedTypeDescriptors(m map[string]TypeDescriptor) []TypeDescriptor {
	out := make([]TypeDescriptor, 0, len(m))
	for _, d := range m {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Type < out[j].Type })
	return out
}
