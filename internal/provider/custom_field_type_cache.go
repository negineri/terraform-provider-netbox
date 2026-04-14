// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"terraform-provider-netbox/internal/client"
)

// customFieldTypeCache はフィールド名 → type ("text", "integer", "boolean" など) のキャッシュ。
// プロセス内グローバルキャッシュ（sync.RWMutex で保護）。
// Terraform の apply 1 回の中でフィールド型が変わることはない前提で TTL は設けない。
type customFieldTypeCache struct {
	mu    sync.RWMutex
	cache map[string]string // field_name → type value
}

var globalCFTypeCache = &customFieldTypeCache{cache: map[string]string{}}

// lookupFieldType はキャッシュからフィールド型を返す。キャッシュミス時は API を呼ぶ。
func (c *customFieldTypeCache) lookupFieldType(
	ctx context.Context,
	cl *client.NetboxClient,
	fieldName string,
) (string, error) {
	// キャッシュヒット確認
	c.mu.RLock()
	if t, ok := c.cache[fieldName]; ok {
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	// API から取得
	path := fmt.Sprintf("api/extras/custom-fields/?name=%s", fieldName)
	bodyStr, err := cl.Get(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to fetch custom field type for %q: %w", fieldName, err)
	}

	var resp struct {
		Results []struct {
			Name string `json:"name"`
			Type struct {
				Value string `json:"value"`
			} `json:"type"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(*bodyStr), &resp); err != nil {
		return "", fmt.Errorf("failed to parse custom field response for %q: %w", fieldName, err)
	}
	if len(resp.Results) == 0 {
		return "", fmt.Errorf("custom field %q not found in Netbox", fieldName)
	}

	fieldType := resp.Results[0].Type.Value

	// キャッシュに保存
	c.mu.Lock()
	c.cache[fieldName] = fieldType
	c.mu.Unlock()

	return fieldType, nil
}
