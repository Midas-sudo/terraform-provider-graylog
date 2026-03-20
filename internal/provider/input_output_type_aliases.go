// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import "strings"

// inputTypeAliases maps simple Graylog input class names (e.g. SyslogUDPInput) to full Java type strings.
// Extend this map when Graylog adds new built-in inputs; unknown types can still use the full FQCN.
var inputTypeAliases = map[string]string{
	"SyslogUDPInput":  "org.graylog2.inputs.syslog.udp.SyslogUDPInput",
	"SyslogTCPInput":  "org.graylog2.inputs.syslog.tcp.SyslogTCPInput",
	"GELFUDPInput":    "org.graylog2.inputs.gelf.udp.GELFUDPInput",
	"GELFTCPInput":    "org.graylog2.inputs.gelf.tcp.GELFTCPInput",
	"GELFHttpInput":   "org.graylog2.inputs.gelf.http.GELFHttpInput",
	"RawUDPInput":     "org.graylog2.inputs.raw.udp.RawUDPInput",
	"RawTCPInput":     "org.graylog2.inputs.raw.tcp.RawTCPInput",
	"JsonPathInput":   "org.graylog2.inputs.misc.jsonpath.JsonPathInput",
	"BeatsInput":      "org.graylog2.inputs.beats.BeatsInput",
	"KafkaInput":      "org.graylog2.inputs.kafka.KafkaInput",
	"CEFUDPInput":     "org.graylog2.inputs.cefd.udp.CEFUDPInput",
	"SyslogAMQPInput": "org.graylog2.inputs.syslog.amqp.SyslogAMQPInput",
	"GELFAMQPInput":   "org.graylog2.inputs.gelf.amqp.GELFAMQPInput",
}

// outputTypeAliases maps simple Graylog output class names to full Java type strings.
var outputTypeAliases = map[string]string{
	"LoggingOutput":              "org.graylog2.outputs.LoggingOutput",
	"GelfOutput":                 "org.graylog2.outputs.GelfOutput",
	"ElasticSearchOutput":        "org.graylog2.outputs.ElasticSearchOutput",
	"DiscardMessageOutput":       "org.graylog2.outputs.DiscardMessageOutput",
	"BatchedMessageFilterOutput": "org.graylog2.outputs.BatchedMessageFilterOutput",
}

func isFullyQualifiedInputType(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "org.graylog2.inputs.")
}

func isFullyQualifiedOutputType(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "org.graylog2.outputs.")
}

// expandInputType resolves a short name to Graylog's input type string, or returns s if already qualified / unknown.
func expandInputType(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || isFullyQualifiedInputType(s) {
		return s
	}
	if full, ok := inputTypeAliases[s]; ok {
		return full
	}
	return s
}

// collapseInputType returns the short alias when s matches a known built-in input type.
func collapseInputType(s string) string {
	s = strings.TrimSpace(s)
	for short, full := range inputTypeAliases {
		if s == full {
			return short
		}
	}
	return s
}

func expandOutputType(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || isFullyQualifiedOutputType(s) {
		return s
	}
	if full, ok := outputTypeAliases[s]; ok {
		return full
	}
	return s
}

func collapseOutputType(s string) string {
	s = strings.TrimSpace(s)
	for short, full := range outputTypeAliases {
		if s == full {
			return short
		}
	}
	return s
}
