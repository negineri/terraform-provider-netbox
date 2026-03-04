// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceInterfaceResource は acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1 が存在している必要があります。
func TestAccDeviceInterfaceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_device" "test" {
  name           = "tf-acc-test-device-iface"
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
}

resource "netbox_device_interface" "test" {
  device_id   = netbox_device.test.id
  name        = "lo"
  type        = "virtual"
  enabled     = true
  description = "terraform test interface"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_interface.test", "name", "lo"),
					resource.TestCheckResourceAttr("netbox_device_interface.test", "type", "virtual"),
					resource.TestCheckResourceAttr("netbox_device_interface.test", "enabled", "true"),
					resource.TestCheckResourceAttr("netbox_device_interface.test", "description", "terraform test interface"),
					resource.TestCheckResourceAttrSet("netbox_device_interface.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_device_interface.test", "device_id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_device" "test" {
  name           = "tf-acc-test-device-iface"
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
}

resource "netbox_device_interface" "test" {
  device_id   = netbox_device.test.id
  name        = "loopback"
  type        = "virtual"
  enabled     = false
  description = "terraform test interface updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_interface.test", "name", "loopback"),
					resource.TestCheckResourceAttr("netbox_device_interface.test", "type", "virtual"),
					resource.TestCheckResourceAttr("netbox_device_interface.test", "enabled", "false"),
					resource.TestCheckResourceAttr("netbox_device_interface.test", "description", "terraform test interface updated"),
					resource.TestCheckResourceAttrSet("netbox_device_interface.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
