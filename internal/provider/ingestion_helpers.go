// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
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

func mapOutputToResourceModel(ctx context.Context, output *client.Output, data *OutputResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ID = types.StringValue(output.ID)
	data.Title = types.StringValue(output.Title)
	data.Type = types.StringValue(collapseOutputType(output.Type))
	if !data.Configuration.IsNull() && !data.Configuration.IsUnknown() {
		return diags
	}
	cfg := output.Configuration
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	dyn, d := interfaceToDynamic(ctx, cfg)
	diags.Append(d...)
	data.Configuration = dyn
	return diags
}

func mapExtractorToResourceModel(ctx context.Context, extractor *client.Extractor, data *ExtractorResourceModel) diag.Diagnostics {
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

	var diags diag.Diagnostics
	if data.ExtractorConfig.IsNull() || data.ExtractorConfig.IsUnknown() {
		cfg := extractor.ExtractorConfig
		if cfg == nil {
			cfg = map[string]interface{}{}
		}
		cfgDyn, d := interfaceToDynamic(ctx, cfg)
		diags.Append(d...)
		data.ExtractorConfig = cfgDyn
	}

	if data.Converters.IsNull() || data.Converters.IsUnknown() {
		converters := make([]interface{}, 0, len(extractor.Converters))
		for _, c := range extractor.Converters {
			converters = append(converters, c)
		}
		convDyn, d := interfaceToDynamic(ctx, converters)
		diags.Append(d...)
		data.Converters = convDyn
	}
	return diags
}

func mapGrokPatternToResourceModel(pattern *client.GrokPattern, data *GrokPatternResourceModel) {
	data.ID = types.StringValue(pattern.ID)
	data.Name = types.StringValue(pattern.Name)
	data.Pattern = types.StringValue(pattern.Pattern)
}

func mapOutputToDataSourceModel(ctx context.Context, output *client.Output, data *OutputDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ID = types.StringValue(output.ID)
	data.Title = types.StringValue(output.Title)
	data.Type = types.StringValue(collapseOutputType(output.Type))
	cfg := output.Configuration
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	dyn, d := interfaceToDynamic(ctx, cfg)
	diags.Append(d...)
	data.Configuration = dyn
	return diags
}

func mapExtractorToDataSourceModel(ctx context.Context, extractor *client.Extractor, data *ExtractorDataSourceModel) diag.Diagnostics {
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
		cfg = map[string]interface{}{}
	}
	cfgDyn, diags := interfaceToDynamic(ctx, cfg)
	data.ExtractorConfig = cfgDyn
	if diags.HasError() {
		return diags
	}

	converters := make([]interface{}, 0, len(extractor.Converters))
	for _, c := range extractor.Converters {
		converters = append(converters, c)
	}
	convDyn, d := interfaceToDynamic(ctx, converters)
	diags.Append(d...)
	data.Converters = convDyn
	return diags
}

func extractorConfigJSONString(cfg map[string]interface{}) types.String {
	return lookupConfigJSONString(cfg)
}

func extractorConvertersJSONString(converters []map[string]interface{}) types.String {
	if converters == nil {
		return types.StringValue("[]")
	}
	b, err := json.Marshal(converters)
	if err != nil {
		return types.StringValue("[]")
	}
	return types.StringValue(string(b))
}

func mapGrokPatternToDataSourceModel(pattern *client.GrokPattern, data *GrokPatternDataSourceModel) {
	data.ID = types.StringValue(pattern.ID)
	data.Name = types.StringValue(pattern.Name)
	data.Pattern = types.StringValue(pattern.Pattern)
}
