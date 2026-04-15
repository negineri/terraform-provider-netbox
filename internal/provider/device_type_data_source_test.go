// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceTypeDataSource は netbox_device_type data source の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に id=1 の device type が存在している必要があります。
func TestAccDeviceTypeDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_device_type" "test" {
  id = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_device_type.test", "id", "1"),
					resource.TestCheckResourceAttrSet("data.netbox_device_type.test", "model"),
					resource.TestCheckResourceAttrSet("data.netbox_device_type.test", "slug"),
					resource.TestCheckResourceAttrSet("data.netbox_device_type.test", "manufacturer_id"),
				),
			},
		},
	})
}
