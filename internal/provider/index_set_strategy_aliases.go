// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import "strings"

const (
	rotationStrategyPackage  = "org.graylog2.indexer.rotation.strategies."
	retentionStrategyPackage = "org.graylog2.indexer.retention.strategies."
)

// rotationStrategyAliases maps simple Java class names to full Graylog rotation strategy class names.
var rotationStrategyAliases = map[string]string{
	"MessageCountRotationStrategy":         rotationStrategyPackage + "MessageCountRotationStrategy",
	"SizeBasedRotationStrategy":            rotationStrategyPackage + "SizeBasedRotationStrategy",
	"TimeBasedRotationStrategy":            rotationStrategyPackage + "TimeBasedRotationStrategy",
	"TimeBasedSizeOptimizingStrategy":      rotationStrategyPackage + "TimeBasedSizeOptimizingStrategy",
}

// rotationStrategyConfigAliases maps simple config type names to full Graylog rotation strategy config types.
var rotationStrategyConfigAliases = map[string]string{
	"MessageCountRotationStrategyConfig":    rotationStrategyPackage + "MessageCountRotationStrategyConfig",
	"SizeBasedRotationStrategyConfig":       rotationStrategyPackage + "SizeBasedRotationStrategyConfig",
	"TimeBasedRotationStrategyConfig":       rotationStrategyPackage + "TimeBasedRotationStrategyConfig",
	"TimeBasedSizeOptimizingStrategyConfig": rotationStrategyPackage + "TimeBasedSizeOptimizingStrategyConfig",
}

// retentionStrategyAliases maps simple Java class names to full Graylog retention strategy class names.
var retentionStrategyAliases = map[string]string{
	"DeletionRetentionStrategy": retentionStrategyPackage + "DeletionRetentionStrategy",
	"NoopRetentionStrategy":     retentionStrategyPackage + "NoopRetentionStrategy",
}

// retentionStrategyConfigAliases maps simple config type names to full Graylog retention strategy config types.
var retentionStrategyConfigAliases = map[string]string{
	"DeletionRetentionStrategyConfig": retentionStrategyPackage + "DeletionRetentionStrategyConfig",
	"NoopRetentionStrategyConfig":     retentionStrategyPackage + "NoopRetentionStrategyConfig",
}

func isFullyQualifiedGraylogStrategy(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "org.graylog2.indexer.")
}

// expandRotationStrategyClass accepts a simple class name (e.g. MessageCountRotationStrategy) or a full FQCN.
func expandRotationStrategyClass(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || isFullyQualifiedGraylogStrategy(s) {
		return s
	}
	if full, ok := rotationStrategyAliases[s]; ok {
		return full
	}
	return s
}

// expandRetentionStrategyClass accepts a simple class name (e.g. DeletionRetentionStrategy) or a full FQCN.
func expandRetentionStrategyClass(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || isFullyQualifiedGraylogStrategy(s) {
		return s
	}
	if full, ok := retentionStrategyAliases[s]; ok {
		return full
	}
	return s
}

// expandRotationStrategyConfigType expands nested rotation_strategy.type values.
func expandRotationStrategyConfigType(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || isFullyQualifiedGraylogStrategy(s) {
		return s
	}
	if full, ok := rotationStrategyConfigAliases[s]; ok {
		return full
	}
	return s
}

// expandRetentionStrategyConfigType expands nested retention_strategy.type values.
func expandRetentionStrategyConfigType(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || isFullyQualifiedGraylogStrategy(s) {
		return s
	}
	if full, ok := retentionStrategyConfigAliases[s]; ok {
		return full
	}
	return s
}

func collapseRotationStrategyClass(full string) string {
	full = strings.TrimSpace(full)
	for short, fq := range rotationStrategyAliases {
		if full == fq {
			return short
		}
	}
	return full
}

func collapseRetentionStrategyClass(full string) string {
	full = strings.TrimSpace(full)
	for short, fq := range retentionStrategyAliases {
		if full == fq {
			return short
		}
	}
	return full
}

func collapseRotationStrategyConfigType(full string) string {
	full = strings.TrimSpace(full)
	for short, fq := range rotationStrategyConfigAliases {
		if full == fq {
			return short
		}
	}
	return full
}

func collapseRetentionStrategyConfigType(full string) string {
	full = strings.TrimSpace(full)
	for short, fq := range retentionStrategyConfigAliases {
		if full == fq {
			return short
		}
	}
	return full
}

func cloneStrategyMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func omitNilStrategyValues(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		if v == nil {
			continue
		}
		out[k] = v
	}
	return out
}

// expandRotationStrategyMap expands short strategy config type names for API calls.
func expandRotationStrategyMap(in map[string]interface{}) map[string]interface{} {
	out := cloneStrategyMap(in)
	if t, ok := out["type"].(string); ok {
		out["type"] = expandRotationStrategyConfigType(t)
	}
	return out
}

// collapseRotationStrategyMap collapses known FQCNs for Terraform state.
func collapseRotationStrategyMap(in map[string]interface{}) map[string]interface{} {
	out := cloneStrategyMap(in)
	if t, ok := out["type"].(string); ok {
		out["type"] = collapseRotationStrategyConfigType(t)
	}
	return out
}

func expandRetentionStrategyMap(in map[string]interface{}) map[string]interface{} {
	out := cloneStrategyMap(in)
	if t, ok := out["type"].(string); ok {
		out["type"] = expandRetentionStrategyConfigType(t)
	}
	return out
}

func collapseRetentionStrategyMap(in map[string]interface{}) map[string]interface{} {
	out := cloneStrategyMap(in)
	if t, ok := out["type"].(string); ok {
		out["type"] = collapseRetentionStrategyConfigType(t)
	}
	return out
}

func strategyMapType(m map[string]interface{}) string {
	if m == nil {
		return ""
	}
	t, _ := m["type"].(string)
	return t
}
