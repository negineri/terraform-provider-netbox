// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-netbox/internal/client"
)

// checkMacAddressAvailable は MAC アドレスが別のオブジェクトに割り当て済みでないことを確認します。
// currentObjectId が 0 の場合は新規割り当て（既存の割り当てがあればエラー）、
// 0 以外の場合は現在のオブジェクトへの割り当てのみ許可します。
func checkMacAddressAvailable(ctx context.Context, c *client.NetboxClient, macId int64, currentObjectId int64) error {
	path := fmt.Sprintf("api/dcim/mac-addresses/%d/", macId)
	bodyStr, err := c.Get(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to get MAC address %d: %w", macId, err)
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		return fmt.Errorf("failed to parse MAC address response: %w", err)
	}

	if objIdRaw := apiResponse["assigned_object_id"]; objIdRaw != nil {
		objIdFloat, ok := objIdRaw.(float64)
		if !ok {
			return fmt.Errorf("unexpected type for assigned_object_id")
		}
		assignedId := int64(objIdFloat)
		if assignedId != currentObjectId {
			return fmt.Errorf(
				"MAC address (id=%d) is already assigned to object id=%d; "+
					"remove primary_mac_address_id from the owning interface first",
				macId, assignedId,
			)
		}
	}

	return nil
}
