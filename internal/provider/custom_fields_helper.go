// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

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
func customFieldsToPayload(ctx context.Context, m types.Map, diags *diag.Diagnostics) map[string]interface{} {
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
		result[k] = v
	}
	return result
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
			elems[k] = types.StringValue(fmt.Sprintf("%v", val))
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
