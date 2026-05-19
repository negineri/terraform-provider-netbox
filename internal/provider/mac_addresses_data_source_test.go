// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMacAddressesDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "netbox_mac_address" "test1" {
  mac_address = "CA:FE:BA:BE:01:01"
  description = "terraform test MAC list"
}

resource "netbox_mac_address" "test2" {
  mac_address = "CA:FE:BA:BE:01:02"
  description = "terraform test MAC list"
}

data "netbox_mac_addresses" "all" {
  depends_on = [netbox_mac_address.test1, netbox_mac_address.test2]
}

data "netbox_mac_addresses" "filtered" {
  mac_address = "CA:FE:BA:BE:01:01"
  depends_on  = [netbox_mac_address.test1, netbox_mac_address.test2]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_mac_addresses.all", "id"),
					// フィルターなしで 2 件以上存在することを確認
					resource.TestCheckResourceAttrSet("data.netbox_mac_addresses.all", "mac_addresses.0.id"),
					// mac_address フィルターで 1 件に絞れることを確認
					resource.TestCheckResourceAttr("data.netbox_mac_addresses.filtered", "mac_addresses.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_mac_addresses.filtered", "mac_addresses.0.mac_address", "CA:FE:BA:BE:01:01"),
					resource.TestCheckResourceAttr("data.netbox_mac_addresses.filtered", "mac_addresses.0.description", "terraform test MAC list"),
				),
			},
		},
	})
}
