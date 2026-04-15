// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccManufacturerDataSource は netbox_manufacturer data source の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に id=1 の manufacturer が存在している必要があります。
func TestAccManufacturerDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_manufacturer" "test" {
  id = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_manufacturer.test", "id", "1"),
					resource.TestCheckResourceAttrSet("data.netbox_manufacturer.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_manufacturer.test", "slug"),
				),
			},
		},
	})
}
