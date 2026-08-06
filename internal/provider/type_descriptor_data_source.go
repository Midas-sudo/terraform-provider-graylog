// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

type configFieldModel struct {
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	HumanName      types.String `tfsdk:"human_name"`
	Description    types.String `tfsdk:"description"`
	IsOptional     types.Bool   `tfsdk:"is_optional"`
	DefaultValue   types.String `tfsdk:"default_value"`
	Attributes     types.List   `tfsdk:"attributes"`
	AdditionalInfo types.String `tfsdk:"additional_info"`
}

type typeDescriptorModel struct {
	Type                   types.String       `tfsdk:"type"`
	Name                   types.String       `tfsdk:"name"`
	Description            types.String       `tfsdk:"description"`
	LinkToDocs             types.String       `tfsdk:"link_to_docs"`
	DefaultConfig          types.String       `tfsdk:"default_config"`
	RequestedConfiguration []configFieldModel `tfsdk:"requested_configuration"`
}

func typeDescriptorNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Type identifier (short alias when known, otherwise Graylog's native type string).",
		},
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Human-readable name of the type.",
		},
		"description": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Optional description from Graylog.",
		},
		"link_to_docs": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Optional documentation URL from Graylog.",
		},
		"default_config": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "JSON-encoded default configuration object when Graylog provides one.",
		},
		"requested_configuration": schema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "Configuration fields for this type (name, optional/required, defaults, descriptions).",
			NestedObject: schema.NestedAttributeObject{
				Attributes: configFieldNestedAttributes(),
			},
		},
	}
}

func configFieldNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Configuration key to set in the resource Dynamic object.",
		},
		"type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Graylog field type (e.g. text, number, boolean, dropdown, list, string).",
		},
		"human_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "UI label for the field.",
		},
		"description": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Field help text from Graylog.",
		},
		"is_optional": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Whether Graylog treats the field as optional.",
		},
		"default_value": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "JSON-encoded default value, if any.",
		},
		"attributes": schema.ListAttribute{
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Graylog field attributes (e.g. textarea).",
		},
		"additional_info": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "JSON-encoded additional field metadata (e.g. dropdown values).",
		},
	}
}

func typeDescriptorsListSchema(desc string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed:            true,
		MarkdownDescription: desc,
		NestedObject: schema.NestedAttributeObject{
			Attributes: typeDescriptorNestedAttributes(),
		},
	}
}

func mapTypeDescriptors(_ context.Context, descs []client.TypeDescriptor, collapseType func(string) string) ([]typeDescriptorModel, error) {
	out := make([]typeDescriptorModel, 0, len(descs))
	for _, d := range descs {
		typeName := d.Type
		if collapseType != nil {
			typeName = collapseType(typeName)
		}
		defaultCfg, err := client.DefaultConfigJSON(d.DefaultConfig)
		if err != nil {
			return nil, err
		}
		fields, err := mapConfigFields(d.RequestedConfiguration)
		if err != nil {
			return nil, fmt.Errorf("type %s: %w", d.Type, err)
		}
		out = append(out, typeDescriptorModel{
			Type:                   types.StringValue(typeName),
			Name:                   types.StringValue(d.Name),
			Description:            stringOrNull(d.Description),
			LinkToDocs:             stringOrNull(d.LinkToDocs),
			DefaultConfig:          stringOrNull(defaultCfg),
			RequestedConfiguration: fields,
		})
	}
	return out, nil
}

func mapConfigFields(fields []client.ConfigField) ([]configFieldModel, error) {
	if fields == nil {
		return []configFieldModel{}, nil
	}
	out := make([]configFieldModel, 0, len(fields))
	for _, f := range fields {
		defJSON, err := client.DefaultValueJSON(f.DefaultValue)
		if err != nil {
			return nil, err
		}
		addJSON, err := client.AdditionalInfoJSON(f.AdditionalInfo)
		if err != nil {
			return nil, err
		}
		attrs := f.Attributes
		if attrs == nil {
			attrs = []string{}
		}
		attrVals := make([]attr.Value, len(attrs))
		for i, a := range attrs {
			attrVals[i] = types.StringValue(a)
		}
		attrList, diags := types.ListValue(types.StringType, attrVals)
		if diags.HasError() {
			return nil, fmt.Errorf("attributes for %s: %v", f.Name, diags)
		}
		out = append(out, configFieldModel{
			Name:           types.StringValue(f.Name),
			Type:           types.StringValue(f.Type),
			HumanName:      types.StringValue(f.HumanName),
			Description:    types.StringValue(f.Description),
			IsOptional:     types.BoolValue(f.IsOptional),
			DefaultValue:   stringOrNull(defJSON),
			Attributes:     attrList,
			AdditionalInfo: stringOrNull(addJSON),
		})
	}
	return out, nil
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
