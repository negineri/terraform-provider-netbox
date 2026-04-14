// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// customFieldsSchema は custom_fields 属性のスキーマ定義を返します。
func customFieldsSchema() schema.MapAttribute {
	return schema.MapAttribute{
		MarkdownDescription: "Custom field values as a map of field name to string value.",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
	}
}

// customFieldsToPayload は types.Map から API ペイロード用の map[string]interface{} に変換します。
// Custom Field の型情報を Netbox API から取得し、型に応じた変換を行います。
//
// 型別変換ルール:
//   - text, longtext, url, json, date, datetime, select, multiselect → string
//   - integer → int64
//   - decimal → float64
//   - boolean → bool
//   - object, multiobject → int64 (object ID)
func customFieldsToPayload(
	ctx context.Context,
	cl *client.NetboxClient,
	m types.Map,
	diags *diag.Diagnostics,
) map[string]interface{} {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	var vals map[string]string
	diags.Append(m.ElementsAs(ctx, &vals, false)...)
	if diags.HasError() {
		return nil
	}

	result := make(map[string]interface{}, len(vals))
	for k, v := range vals {
		fieldType, err := globalCFTypeCache.lookupFieldType(ctx, cl, k)
		if err != nil {
			diags.AddError("Error looking up custom field type", err.Error())
			return nil
		}

		converted, err := convertCustomFieldValue(v, fieldType)
		if err != nil {
			diags.AddError(
				fmt.Sprintf("Error converting custom field %q", k),
				err.Error(),
			)
			return nil
		}
		result[k] = converted
	}
	return result
}

// convertCustomFieldValue は文字列値を Netbox フィールド型に応じた Go 値に変換します。
func convertCustomFieldValue(v, fieldType string) (interface{}, error) {
	switch fieldType {
	case "integer", "object", "multiobject":
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to integer for field type %q: %w", v, fieldType, err)
		}
		return i, nil
	case "decimal":
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to decimal for field type %q: %w", v, fieldType, err)
		}
		return f, nil
	case "boolean":
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to boolean for field type %q: %w", v, fieldType, err)
		}
		return b, nil
	default:
		// text, longtext, url, json, date, datetime, select, multiselect → string
		return v, nil
	}
}

// customFieldsFromAPI は API レスポンスの custom_fields を types.Map に変換します。
// null 値のフィールドは除外します。
func customFieldsFromAPI(raw interface{}) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	cfMap, ok := raw.(map[string]interface{})
	if !ok || cfMap == nil {
		return types.MapValueMust(types.StringType, map[string]attr.Value{}), diags
	}

	elems := make(map[string]attr.Value, len(cfMap))
	for k, v := range cfMap {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case string:
			elems[k] = types.StringValue(val)
		case float64:
			// 整数値は小数点なし ("42")、小数値はそのまま ("3.14") として変換する
			if val == float64(int64(val)) {
				elems[k] = types.StringValue(strconv.FormatInt(int64(val), 10))
			} else {
				elems[k] = types.StringValue(strconv.FormatFloat(val, 'f', -1, 64))
			}
		case bool:
			if val {
				elems[k] = types.StringValue("true")
			} else {
				elems[k] = types.StringValue("false")
			}
		default:
			elems[k] = types.StringValue(fmt.Sprintf("%v", val))
		}
	}

	result, d := types.MapValue(types.StringType, elems)
	diags.Append(d...)
	return result, diags
}
