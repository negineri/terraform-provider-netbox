// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMacAddressResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_mac_address" "test" {
  mac_address = "AA:BB:CC:DD:EE:FF"
  description = "terraform test MAC address"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_mac_address.test", "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr("netbox_mac_address.test", "description", "terraform test MAC address"),
					resource.TestCheckResourceAttrSet("netbox_mac_address.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_mac_address" "test" {
  mac_address = "AA:BB:CC:DD:EE:FF"
  description = "terraform test MAC address updated"
  comments    = "updated comment"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_mac_address.test", "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr("netbox_mac_address.test", "description", "terraform test MAC address updated"),
					resource.TestCheckResourceAttr("netbox_mac_address.test", "comments", "updated comment"),
					resource.TestCheckResourceAttrSet("netbox_mac_address.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
