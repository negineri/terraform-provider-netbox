// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMacAddressDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "netbox_mac_address" "test" {
  mac_address = "CA:FE:BA:BE:00:01"
  description = "terraform test MAC data source"
  comments    = "test comment"
}

data "netbox_mac_address" "test" {
  id = netbox_mac_address.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_mac_address.test", "mac_address", "CA:FE:BA:BE:00:01"),
					resource.TestCheckResourceAttr("data.netbox_mac_address.test", "description", "terraform test MAC data source"),
					resource.TestCheckResourceAttr("data.netbox_mac_address.test", "comments", "test comment"),
					resource.TestCheckResourceAttrSet("data.netbox_mac_address.test", "id"),
				),
			},
		},
	})
}
