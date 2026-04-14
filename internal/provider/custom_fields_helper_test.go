// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestConvertCustomFieldValue は convertCustomFieldValue の型別変換を検証します。
func TestConvertCustomFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldType string
		wantType  string
		wantInt   int64
		wantFloat float64
		wantBool  bool
		wantStr   string
		wantError bool
	}{
		// integer
		{
			name:      "integer field: '42' → int64(42)",
			value:     "42",
			fieldType: "integer",
			wantType:  "int64",
			wantInt:   42,
		},
		{
			name:      "integer field: '0' → int64(0)",
			value:     "0",
			fieldType: "integer",
			wantType:  "int64",
			wantInt:   0,
		},
		{
			name:      "integer field: '-5' → int64(-5)",
			value:     "-5",
			fieldType: "integer",
			wantType:  "int64",
			wantInt:   -5,
		},
		{
			name:      "integer field: 'abc' → error",
			value:     "abc",
			fieldType: "integer",
			wantError: true,
		},
		// decimal
		{
			name:      "decimal field: '3.14' → float64(3.14)",
			value:     "3.14",
			fieldType: "decimal",
			wantType:  "float64",
			wantFloat: 3.14,
		},
		{
			name:      "decimal field: '42' → float64(42)",
			value:     "42",
			fieldType: "decimal",
			wantType:  "float64",
			wantFloat: 42.0,
		},
		// boolean
		{
			name:      "boolean field: 'true' → bool(true)",
			value:     "true",
			fieldType: "boolean",
			wantType:  "bool",
			wantBool:  true,
		},
		{
			name:      "boolean field: 'false' → bool(false)",
			value:     "false",
			fieldType: "boolean",
			wantType:  "bool",
			wantBool:  false,
		},
		{
			name:      "boolean field: '1' → bool(true)",
			value:     "1",
			fieldType: "boolean",
			wantType:  "bool",
			wantBool:  true,
		},
		{
			name:      "boolean field: 'abc' → error",
			value:     "abc",
			fieldType: "boolean",
			wantError: true,
		},
		// text フィールドは常に string
		{
			name:      "text field: '42' remains string '42'",
			value:     "42",
			fieldType: "text",
			wantType:  "string",
			wantStr:   "42",
		},
		{
			name:      "text field: 'true' remains string 'true'",
			value:     "true",
			fieldType: "text",
			wantType:  "string",
			wantStr:   "true",
		},
		{
			name:      "text field: 'hello' remains string 'hello'",
			value:     "hello",
			fieldType: "text",
			wantType:  "string",
			wantStr:   "hello",
		},
		// longtext
		{
			name:      "longtext field: '42' remains string '42'",
			value:     "42",
			fieldType: "longtext",
			wantType:  "string",
			wantStr:   "42",
		},
		// select
		{
			name:      "select field: 'production' remains string",
			value:     "production",
			fieldType: "select",
			wantType:  "string",
			wantStr:   "production",
		},
		// object → int64 (object ID)
		{
			name:      "object field: '42' → int64(42)",
			value:     "42",
			fieldType: "object",
			wantType:  "int64",
			wantInt:   42,
		},
		// date
		{
			name:      "date field: '2026-04-14' remains string",
			value:     "2026-04-14",
			fieldType: "date",
			wantType:  "string",
			wantStr:   "2026-04-14",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertCustomFieldValue(tc.value, tc.fieldType)
			if tc.wantError {
				if err == nil {
					t.Errorf("expected error but got nil, value=%v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			switch tc.wantType {
			case "int64":
				v, ok := got.(int64)
				if !ok {
					t.Errorf("expected int64, got %T (%v)", got, got)
					return
				}
				if v != tc.wantInt {
					t.Errorf("expected %d, got %d", tc.wantInt, v)
				}
			case "float64":
				v, ok := got.(float64)
				if !ok {
					t.Errorf("expected float64, got %T (%v)", got, got)
					return
				}
				if v != tc.wantFloat {
					t.Errorf("expected %f, got %f", tc.wantFloat, v)
				}
			case "bool":
				v, ok := got.(bool)
				if !ok {
					t.Errorf("expected bool, got %T (%v)", got, got)
					return
				}
				if v != tc.wantBool {
					t.Errorf("expected %v, got %v", tc.wantBool, v)
				}
			case "string":
				v, ok := got.(string)
				if !ok {
					t.Errorf("expected string, got %T (%v)", got, got)
					return
				}
				if v != tc.wantStr {
					t.Errorf("expected %q, got %q", tc.wantStr, v)
				}
			}
		})
	}
}

func TestCustomFieldsFromAPI_TypeConversion(t *testing.T) {
	tests := []struct {
		name       string
		input      map[string]interface{}
		wantKey    string
		wantVal    string
		isExcluded bool
	}{
		{
			name:    "integer from API is converted to string without decimal",
			input:   map[string]interface{}{"rack_units": float64(42)},
			wantKey: "rack_units",
			wantVal: "42",
		},
		{
			name:    "zero from API is converted to '0'",
			input:   map[string]interface{}{"count": float64(0)},
			wantKey: "count",
			wantVal: "0",
		},
		{
			name:    "decimal from API is preserved",
			input:   map[string]interface{}{"price": float64(3.14)},
			wantKey: "price",
			wantVal: "3.14",
		},
		{
			name:    "bool true from API is converted to 'true'",
			input:   map[string]interface{}{"is_production": true},
			wantKey: "is_production",
			wantVal: "true",
		},
		{
			name:    "bool false from API is converted to 'false'",
			input:   map[string]interface{}{"is_production": false},
			wantKey: "is_production",
			wantVal: "false",
		},
		{
			name:    "string from API remains unchanged",
			input:   map[string]interface{}{"asset_tag": "ABC-123"},
			wantKey: "asset_tag",
			wantVal: "ABC-123",
		},
		{
			name:       "null value from API is excluded",
			input:      map[string]interface{}{"note": nil},
			wantKey:    "note",
			isExcluded: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, diags := customFieldsFromAPI(tc.input)
			if diags.HasError() {
				t.Fatalf("customFieldsFromAPI returned errors: %v", diags)
			}

			elem, ok := result.Elements()[tc.wantKey]
			if tc.isExcluded {
				if ok {
					t.Errorf("expected key %q to be excluded (null), but it was present", tc.wantKey)
				}
				return
			}
			if !ok {
				t.Fatalf("key %q not found in result", tc.wantKey)
			}

			sv, ok := elem.(types.String)
			if !ok {
				t.Fatalf("expected types.String, got %T", elem)
			}
			if sv.ValueString() != tc.wantVal {
				t.Errorf("expected %q, got %q", tc.wantVal, sv.ValueString())
			}
		})
	}
}
