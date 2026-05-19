// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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

// TestAccMacAddressResourceWithInterface は MAC アドレスをデバイスインターフェースに割り当てる acceptance test です。
func TestAccMacAddressResourceWithInterface(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-mac")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with interface assignment
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
}

resource "netbox_device_interface" "test" {
  device_id = netbox_device.test.id
  name      = "eth0"
  type      = "virtual"
}

resource "netbox_mac_address" "test" {
  mac_address          = "11:22:33:44:55:66"
  assigned_object_type = "dcim.interface"
  assigned_object_id   = netbox_device_interface.test.id
  description          = "terraform test MAC with interface"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_mac_address.test", "mac_address", "11:22:33:44:55:66"),
					resource.TestCheckResourceAttr("netbox_mac_address.test", "assigned_object_type", "dcim.interface"),
					resource.TestCheckResourceAttr("netbox_mac_address.test", "description", "terraform test MAC with interface"),
					resource.TestCheckResourceAttrSet("netbox_mac_address.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_mac_address.test", "assigned_object_id"),
				),
			},
			// Update: detach interface
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
}

resource "netbox_device_interface" "test" {
  device_id = netbox_device.test.id
  name      = "eth0"
  type      = "virtual"
}

resource "netbox_mac_address" "test" {
  mac_address = "11:22:33:44:55:66"
  description = "terraform test MAC detached"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_mac_address.test", "mac_address", "11:22:33:44:55:66"),
					resource.TestCheckResourceAttr("netbox_mac_address.test", "description", "terraform test MAC detached"),
					resource.TestCheckNoResourceAttr("netbox_mac_address.test", "assigned_object_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
