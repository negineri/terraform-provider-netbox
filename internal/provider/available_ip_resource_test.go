// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAvailableIPResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_available_ip" "test" {
  prefix_id   = 1
  description = "terraform test IP"
  status      = "active"
  dns_name    = "test.example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_ip.test", "prefix_id", "1"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "description", "terraform test IP"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "dns_name", "test.example.com"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "ip_address"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_available_ip" "test" {
  prefix_id   = 1
  description = "terraform test IP updated"
  status      = "deprecated"
  dns_name    = "updated.example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_ip.test", "prefix_id", "1"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "description", "terraform test IP updated"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "status", "deprecated"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "dns_name", "updated.example.com"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "ip_address"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAvailableIPResourceWithInterface は available IP をデバイスインターフェースに割り当てる acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1, prefix_id=1 が存在している必要があります。
func TestAccAvailableIPResourceWithInterface(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-availip")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with interface and dns_name
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

resource "netbox_available_ip" "test" {
  prefix_id      = 1
  status         = "active"
  description    = "terraform test available IP with interface"
  dns_name       = "avail.example.com"
  interface_id   = netbox_device_interface.test.id
  interface_type = "dcim.interface"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_ip.test", "prefix_id", "1"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "description", "terraform test available IP with interface"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "dns_name", "avail.example.com"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "interface_type", "dcim.interface"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "ip_address"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "interface_id"),
				),
			},
			// Update: change dns_name and detach interface
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

resource "netbox_available_ip" "test" {
  prefix_id   = 1
  status      = "active"
  description = "terraform test available IP detached"
  dns_name    = "detached.example.com"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_ip.test", "prefix_id", "1"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "description", "terraform test available IP detached"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "dns_name", "detached.example.com"),
					resource.TestCheckNoResourceAttr("netbox_available_ip.test", "interface_id"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "ip_address"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
