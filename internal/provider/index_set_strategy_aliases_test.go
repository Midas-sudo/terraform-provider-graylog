// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import "testing"

func TestExpandCollapseRotationStrategyClass(t *testing.T) {
	full := rotationStrategyPackage + "MessageCountRotationStrategy"
	if got := expandRotationStrategyClass("MessageCountRotationStrategy"); got != full {
		t.Fatalf("expand simple: got %q want %q", got, full)
	}
	if got := expandRotationStrategyClass(full); got != full {
		t.Fatalf("expand fqcn passthrough: got %q", got)
	}
	if got := collapseRotationStrategyClass(full); got != "MessageCountRotationStrategy" {
		t.Fatalf("collapse: got %q", got)
	}
}

func TestExpandCollapseRetentionNestedTypes(t *testing.T) {
	cfgFull := retentionStrategyPackage + "DeletionRetentionStrategyConfig"
	if got := expandRetentionStrategyConfigType("DeletionRetentionStrategyConfig"); got != cfgFull {
		t.Fatalf("expand config: got %q want %q", got, cfgFull)
	}
	if got := collapseRetentionStrategyConfigType(cfgFull); got != "DeletionRetentionStrategyConfig" {
		t.Fatalf("collapse config: got %q", got)
	}
}
