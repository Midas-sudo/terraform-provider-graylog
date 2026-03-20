// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import "testing"

func TestExpandCollapseInputType(t *testing.T) {
	t.Parallel()
	if got := expandInputType("SyslogUDPInput"); got != "org.graylog2.inputs.syslog.udp.SyslogUDPInput" {
		t.Fatalf("expand SyslogUDPInput: %q", got)
	}
	if got := expandInputType("org.graylog2.inputs.syslog.udp.SyslogUDPInput"); got != "org.graylog2.inputs.syslog.udp.SyslogUDPInput" {
		t.Fatalf("expand FQCN passthrough: %q", got)
	}
	if got := collapseInputType("org.graylog2.inputs.syslog.udp.SyslogUDPInput"); got != "SyslogUDPInput" {
		t.Fatalf("collapse: %q", got)
	}
	if got := collapseInputType("com.example.CustomInput"); got != "com.example.CustomInput" {
		t.Fatalf("unknown should pass through: %q", got)
	}
}

func TestExpandCollapseOutputType(t *testing.T) {
	t.Parallel()
	if got := expandOutputType("LoggingOutput"); got != "org.graylog2.outputs.LoggingOutput" {
		t.Fatalf("expand LoggingOutput: %q", got)
	}
	if got := collapseOutputType("org.graylog2.outputs.LoggingOutput"); got != "LoggingOutput" {
		t.Fatalf("collapse: %q", got)
	}
}
