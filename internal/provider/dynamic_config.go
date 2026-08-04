// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// dynamicToInterface converts a Dynamic attribute into a Go value suitable for
// Graylog JSON APIs (typically map[string]any or []any).
func dynamicToInterface(ctx context.Context, d types.Dynamic) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if d.IsNull() || d.IsUnknown() {
		return nil, diags
	}
	if d.IsUnderlyingValueNull() {
		return nil, diags
	}
	if d.IsUnderlyingValueUnknown() {
		diags.AddError("Invalid dynamic value", "Underlying dynamic value is unknown")
		return nil, diags
	}
	return attrValueToInterface(ctx, d.UnderlyingValue())
}

// dynamicToMap requires the Dynamic value to decode as a JSON object.
func dynamicToMap(ctx context.Context, d types.Dynamic) (map[string]interface{}, diag.Diagnostics) {
	v, diags := dynamicToInterface(ctx, d)
	if diags.HasError() {
		return nil, diags
	}
	if v == nil {
		return map[string]interface{}{}, diags
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		diags.AddError("Invalid dynamic value", fmt.Sprintf("Expected an object, got %T", v))
		return nil, diags
	}
	return m, diags
}

// dynamicToSlice requires the Dynamic value to decode as a JSON array.
func dynamicToSlice(ctx context.Context, d types.Dynamic) ([]interface{}, diag.Diagnostics) {
	v, diags := dynamicToInterface(ctx, d)
	if diags.HasError() {
		return nil, diags
	}
	if v == nil {
		return []interface{}{}, diags
	}
	s, ok := v.([]interface{})
	if !ok {
		diags.AddError("Invalid dynamic value", fmt.Sprintf("Expected a list/tuple, got %T", v))
		return nil, diags
	}
	return s, diags
}

// dynamicToSliceOfMaps requires a Dynamic list of objects.
func dynamicToSliceOfMaps(ctx context.Context, d types.Dynamic) ([]map[string]interface{}, diag.Diagnostics) {
	s, diags := dynamicToSlice(ctx, d)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]map[string]interface{}, 0, len(s))
	for i, item := range s {
		m, ok := item.(map[string]interface{})
		if !ok {
			diags.AddError("Invalid dynamic value", fmt.Sprintf("Expected object at index %d, got %T", i, item))
			return nil, diags
		}
		out = append(out, m)
	}
	return out, diags
}

// interfaceToDynamic converts an arbitrary Go JSON-like value into types.Dynamic.
func interfaceToDynamic(ctx context.Context, v any) (types.Dynamic, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v == nil {
		return types.DynamicNull(), diags
	}
	attrVal, d := interfaceToAttrValue(ctx, v)
	diags.Append(d...)
	if diags.HasError() {
		return types.DynamicNull(), diags
	}
	return types.DynamicValue(attrVal), diags
}

// upgradeJSONStringAttr parses a prior-state JSON string into types.Dynamic.
func upgradeJSONStringAttr(raw string) (types.Dynamic, error) {
	if raw == "" {
		return types.DynamicNull(), nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return types.DynamicNull(), fmt.Errorf("invalid JSON in prior state: %w", err)
	}
	attrVal, diags := interfaceToAttrValue(context.Background(), decoded)
	if diags.HasError() {
		return types.DynamicNull(), fmt.Errorf("%s", diags.Errors()[0].Detail())
	}
	return types.DynamicValue(attrVal), nil
}

func attrValueToInterface(ctx context.Context, v attr.Value) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v == nil || v.IsNull() {
		return nil, diags
	}
	if v.IsUnknown() {
		diags.AddError("Invalid attribute value", "Value is unknown")
		return nil, diags
	}

	switch val := v.(type) {
	case types.Bool:
		return val.ValueBool(), diags
	case types.String:
		return val.ValueString(), diags
	case types.Number:
		f, _ := val.ValueBigFloat().Float64()
		return f, diags
	case types.List:
		elems := val.Elements()
		out := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			iv, d := attrValueToInterface(ctx, e)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			out = append(out, iv)
		}
		return out, diags
	case types.Tuple:
		elems := val.Elements()
		out := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			iv, d := attrValueToInterface(ctx, e)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			out = append(out, iv)
		}
		return out, diags
	case types.Set:
		elems := val.Elements()
		out := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			iv, d := attrValueToInterface(ctx, e)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			out = append(out, iv)
		}
		return out, diags
	case types.Map:
		elems := val.Elements()
		out := make(map[string]interface{}, len(elems))
		for k, e := range elems {
			iv, d := attrValueToInterface(ctx, e)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			out[k] = iv
		}
		return out, diags
	case types.Object:
		attrs := val.Attributes()
		out := make(map[string]interface{}, len(attrs))
		for k, e := range attrs {
			iv, d := attrValueToInterface(ctx, e)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			out[k] = iv
		}
		return out, diags
	case basetypes.DynamicValuable:
		dyn, d := val.ToDynamicValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return dynamicToInterface(ctx, dyn)
	default:
		diags.AddError("Unsupported attribute type", fmt.Sprintf("Cannot convert %T to Go value", v))
		return nil, diags
	}
}

func interfaceToAttrValue(ctx context.Context, v any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch x := v.(type) {
	case nil:
		return types.DynamicNull(), diags
	case bool:
		return types.BoolValue(x), diags
	case string:
		return types.StringValue(x), diags
	case float64:
		return types.NumberValue(big.NewFloat(x)), diags
	case float32:
		return types.NumberValue(big.NewFloat(float64(x))), diags
	case int:
		return types.NumberValue(big.NewFloat(float64(x))), diags
	case int32:
		return types.NumberValue(big.NewFloat(float64(x))), diags
	case int64:
		return types.NumberValue(big.NewFloat(float64(x))), diags
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			diags.AddError("Invalid number", err.Error())
			return nil, diags
		}
		return types.NumberValue(big.NewFloat(f)), diags
	case map[string]interface{}:
		attrTypes := make(map[string]attr.Type, len(x))
		attrVals := make(map[string]attr.Value, len(x))
		for k, child := range x {
			av, d := interfaceToAttrValue(ctx, child)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			attrTypes[k] = av.Type(ctx)
			attrVals[k] = av
		}
		obj, d := types.ObjectValue(attrTypes, attrVals)
		diags.Append(d...)
		return obj, diags
	case []interface{}:
		elemTypes := make([]attr.Type, 0, len(x))
		elemVals := make([]attr.Value, 0, len(x))
		for _, child := range x {
			av, d := interfaceToAttrValue(ctx, child)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			elemTypes = append(elemTypes, av.Type(ctx))
			elemVals = append(elemVals, av)
		}
		tup, d := types.TupleValue(elemTypes, elemVals)
		diags.Append(d...)
		return tup, diags
	default:
		diags.AddError("Unsupported Go value", fmt.Sprintf("Cannot convert %T to Terraform value", v))
		return nil, diags
	}
}
