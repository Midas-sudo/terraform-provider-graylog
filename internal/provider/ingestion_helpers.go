// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

func outputFromPayload(payload string) (*client.Output, diag.Diagnostics) {
	var diags diag.Diagnostics
	var output client.Output
	if err := json.Unmarshal([]byte(payload), &output); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &output, diags
}

func extractorFromPayload(payload string) (*client.Extractor, diag.Diagnostics) {
	var diags diag.Diagnostics
	var extractor client.Extractor
	if err := json.Unmarshal([]byte(payload), &extractor); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &extractor, diags
}

func grokPatternFromPayload(payload string) (*client.GrokPattern, diag.Diagnostics) {
	var diags diag.Diagnostics
	var pattern client.GrokPattern
	if err := json.Unmarshal([]byte(payload), &pattern); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &pattern, diags
}

func marshalOutputJSON(output *client.Output) string {
	clone := *output
	clone.ID = ""
	clone.CreatorUserID = ""
	clone.CreatedAt = ""
	clone.ContentPack = nil
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func marshalExtractorJSON(extractor *client.Extractor) string {
	clone := *extractor
	clone.ID = ""
	clone.Type = ""
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func marshalGrokPatternJSON(pattern *client.GrokPattern) string {
	clone := *pattern
	clone.ID = ""
	clone.ContentPack = nil
	b, err := json.Marshal(clone)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func mapOutputToResourceModel(output *client.Output, data *OutputResourceModel) {
	data.ID = types.StringValue(output.ID)
	data.Title = types.StringValue(output.Title)
	data.Type = types.StringValue(output.Type)
}

func mapExtractorToResourceModel(extractor *client.Extractor, data *ExtractorResourceModel) {
	data.ID = types.StringValue(extractor.ID)
	data.Title = types.StringValue(extractor.Title)
	if extractor.ExtractorType != "" {
		data.ExtractorType = types.StringValue(extractor.ExtractorType)
	} else {
		data.ExtractorType = types.StringValue(extractor.Type)
	}
	data.CursorStrategy = types.StringValue(extractor.CursorStrategy)
	data.SourceField = types.StringValue(extractor.SourceField)
	data.TargetField = types.StringValue(extractor.TargetField)
	data.ConditionType = types.StringValue(extractor.ConditionType)
	data.ConditionValue = types.StringValue(extractor.ConditionValue)
	data.Order = types.Int64Value(extractor.Order)
}

func mapGrokPatternToResourceModel(pattern *client.GrokPattern, data *GrokPatternResourceModel) {
	data.ID = types.StringValue(pattern.ID)
	data.Name = types.StringValue(pattern.Name)
	data.Pattern = types.StringValue(pattern.Pattern)
}

func mapOutputToDataSourceModel(output *client.Output, data *OutputDataSourceModel) {
	data.ID = types.StringValue(output.ID)
	data.Title = types.StringValue(output.Title)
	data.Type = types.StringValue(output.Type)
	if output.Configuration == nil {
		data.ConfigurationJSON = types.StringValue("{}")
		return
	}
	b, err := json.Marshal(output.Configuration)
	if err != nil {
		data.ConfigurationJSON = types.StringValue("{}")
		return
	}
	data.ConfigurationJSON = types.StringValue(string(b))
}

func mapExtractorToDataSourceModel(extractor *client.Extractor, data *ExtractorDataSourceModel) {
	data.ID = types.StringValue(extractor.ID)
	data.Title = types.StringValue(extractor.Title)
	if extractor.ExtractorType != "" {
		data.ExtractorType = types.StringValue(extractor.ExtractorType)
	} else {
		data.ExtractorType = types.StringValue(extractor.Type)
	}
	data.CursorStrategy = types.StringValue(extractor.CursorStrategy)
	data.SourceField = types.StringValue(extractor.SourceField)
	data.TargetField = types.StringValue(extractor.TargetField)
	data.ConditionType = types.StringValue(extractor.ConditionType)
	data.ConditionValue = types.StringValue(extractor.ConditionValue)
	data.Order = types.Int64Value(extractor.Order)

	cfg := extractor.ExtractorConfig
	if cfg == nil {
		data.ExtractorConfigJSON = types.StringValue("{}")
	} else if b, err := json.Marshal(cfg); err == nil {
		data.ExtractorConfigJSON = types.StringValue(string(b))
	} else {
		data.ExtractorConfigJSON = types.StringValue("{}")
	}
	converters := extractor.Converters
	if converters == nil {
		data.ConvertersJSON = types.StringValue("[]")
	} else if b, err := json.Marshal(converters); err == nil {
		data.ConvertersJSON = types.StringValue(string(b))
	} else {
		data.ConvertersJSON = types.StringValue("[]")
	}
}

func mapGrokPatternToDataSourceModel(pattern *client.GrokPattern, data *GrokPatternDataSourceModel) {
	data.ID = types.StringValue(pattern.ID)
	data.Name = types.StringValue(pattern.Name)
	data.Pattern = types.StringValue(pattern.Pattern)
}
