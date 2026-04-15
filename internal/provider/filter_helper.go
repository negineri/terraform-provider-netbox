// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// buildFilterQuery はフィルタパラメータの map から URL クエリ文字列を構築します。
// 値が空文字のエントリは除外されます。
func buildFilterQuery(params map[string]string) string {
	parts := make([]string, 0, len(params))
	for k, v := range params {
		if v == "" {
			continue
		}
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "&")
}

// combineQueryStrings は複数のクエリ文字列を & で結合します。空文字は除外されます。
func combineQueryStrings(queries ...string) string {
	parts := make([]string, 0, len(queries))
	for _, q := range queries {
		if q != "" {
			parts = append(parts, q)
		}
	}
	return strings.Join(parts, "&")
}

// stringFilterParam は types.String フィルタを params map に追加します。Null/Unknown は無視します。
func stringFilterParam(params map[string]string, key string, val types.String) {
	if !val.IsNull() && !val.IsUnknown() && val.ValueString() != "" {
		params[key] = val.ValueString()
	}
}

// int64FilterParam は types.Int64 フィルタを params map に追加します。Null/Unknown は無視します。
func int64FilterParam(params map[string]string, key string, val types.Int64) {
	if !val.IsNull() && !val.IsUnknown() {
		params[key] = fmt.Sprintf("%d", val.ValueInt64())
	}
}
