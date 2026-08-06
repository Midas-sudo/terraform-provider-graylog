// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestFieldsFromRequestedConfiguration_InputsFixture(t *testing.T) {
	var raw map[string]rawInputTypeInfo
	if err := json.Unmarshal(loadFixture(t, "inputs_types_all.json"), &raw); err != nil {
		t.Fatal(err)
	}
	info, ok := raw["org.graylog2.inputs.syslog.udp.SyslogUDPInput"]
	if !ok {
		t.Fatal("missing SyslogUDPInput in fixture")
	}
	fields := fieldsFromRequestedConfiguration(info.RequestedConfiguration)
	if len(fields) == 0 {
		t.Fatal("expected configuration fields")
	}
	found := false
	for _, f := range fields {
		if f.Name == "bind_address" {
			found = true
			if f.IsOptional {
				t.Fatalf("bind_address should be required, got %+v", f)
			}
			if f.Type != "text" {
				t.Fatalf("bind_address type: got %q", f.Type)
			}
		}
	}
	if !found {
		t.Fatal("bind_address not found")
	}
}

func TestFieldsFromDefaultConfig_LookupFixture(t *testing.T) {
	var raw map[string]rawLookupTypeInfo
	if err := json.Unmarshal(loadFixture(t, "lookup_adapter_types.json"), &raw); err != nil {
		t.Fatal(err)
	}
	desc := lookupTypesToDescriptors(raw)["csvfile"]
	if desc.Type != "csvfile" {
		t.Fatalf("type: %q", desc.Type)
	}
	if len(desc.RequestedConfiguration) < 3 {
		t.Fatalf("expected synthesized fields, got %d", len(desc.RequestedConfiguration))
	}
	if desc.DefaultConfig["path"] == nil {
		t.Fatal("expected default path")
	}
}

func TestStrategiesToDescriptors_RotationFixture(t *testing.T) {
	var raw IndexStrategiesResponse
	if err := json.Unmarshal(loadFixture(t, "rotation_strategies.json"), &raw); err != nil {
		t.Fatal(err)
	}
	descs := strategiesToDescriptors(raw.Strategies)
	if len(descs) == 0 {
		t.Fatal("expected strategies")
	}
	var timeBased *TypeDescriptor
	for i := range descs {
		if descs[i].Type == "org.graylog2.indexer.rotation.strategies.TimeBasedRotationStrategy" {
			timeBased = &descs[i]
			break
		}
	}
	if timeBased == nil {
		t.Fatal("TimeBasedRotationStrategy missing")
	}
	foundPeriod := false
	for _, f := range timeBased.RequestedConfiguration {
		if f.Name == "rotation_period" {
			foundPeriod = true
			if f.Type != "string" {
				t.Fatalf("rotation_period type: %q", f.Type)
			}
		}
	}
	if !foundPeriod {
		t.Fatal("rotation_period missing from json_schema fields")
	}
}

func TestOutputsAvailableFixture(t *testing.T) {
	var raw AvailableOutputsResponse
	if err := json.Unmarshal(loadFixture(t, "outputs_available.json"), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw.Types["org.graylog2.outputs.LoggingOutput"]; !ok {
		t.Fatal("LoggingOutput missing")
	}
}

func TestModernEventNotificationTypes(t *testing.T) {
	types := ModernEventNotificationTypes()
	if len(types) < 3 {
		t.Fatalf("expected modern types, got %d", len(types))
	}
	if types[0].Type != "http-notification-v1" {
		t.Fatalf("first type: %q", types[0].Type)
	}
}
