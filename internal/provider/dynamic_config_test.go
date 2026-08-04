// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"
)

func TestDynamicConfigRoundTripObject(t *testing.T) {
	ctx := context.Background()
	in := map[string]interface{}{
		"bind_address":     "0.0.0.0",
		"port":             float64(1514),
		"recv_buffer_size": float64(262144),
		"enabled":          true,
		"nested": map[string]interface{}{
			"a": "b",
		},
		"tags": []interface{}{"x", "y"},
	}

	dyn, diags := interfaceToDynamic(ctx, in)
	if diags.HasError() {
		t.Fatalf("interfaceToDynamic diagnostics: %v", diags)
	}
	if dyn.IsNull() || dyn.IsUnknown() {
		t.Fatalf("expected known dynamic value")
	}

	out, diags := dynamicToMap(ctx, dyn)
	if diags.HasError() {
		t.Fatalf("dynamicToMap diagnostics: %v", diags)
	}
	if out["bind_address"] != "0.0.0.0" {
		t.Fatalf("bind_address=%v", out["bind_address"])
	}
	if out["port"] != float64(1514) {
		t.Fatalf("port=%v", out["port"])
	}
	if out["enabled"] != true {
		t.Fatalf("enabled=%v", out["enabled"])
	}
	nested, ok := out["nested"].(map[string]interface{})
	if !ok || nested["a"] != "b" {
		t.Fatalf("nested=%v", out["nested"])
	}
	tags, ok := out["tags"].([]interface{})
	if !ok || len(tags) != 2 || tags[0] != "x" {
		t.Fatalf("tags=%v", out["tags"])
	}
}

func TestUpgradeJSONStringAttr(t *testing.T) {
	dyn, err := upgradeJSONStringAttr(`{"type":"none","n":1}`)
	if err != nil {
		t.Fatalf("upgradeJSONStringAttr: %v", err)
	}
	out, diags := dynamicToMap(context.Background(), dyn)
	if diags.HasError() {
		t.Fatalf("dynamicToMap: %v", diags)
	}
	if out["type"] != "none" {
		t.Fatalf("type=%v", out["type"])
	}
	if out["n"] != float64(1) {
		t.Fatalf("n=%v", out["n"])
	}
}

func TestDynamicToSlice(t *testing.T) {
	ctx := context.Background()
	dyn, diags := interfaceToDynamic(ctx, []interface{}{
		map[string]interface{}{"id": "a"},
		map[string]interface{}{"id": "b"},
	})
	if diags.HasError() {
		t.Fatalf("interfaceToDynamic: %v", diags)
	}
	out, diags := dynamicToSlice(ctx, dyn)
	if diags.HasError() {
		t.Fatalf("dynamicToSlice: %v", diags)
	}
	if len(out) != 2 {
		t.Fatalf("len=%d", len(out))
	}
}
