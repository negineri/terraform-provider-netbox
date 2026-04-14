// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIpAddressResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_ip_address" "test" {
  ip_address  = "192.168.100.1/24"
  status      = "active"
  description = "terraform test IP address"
  dns_name    = "test.example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", "192.168.100.1/24"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", "terraform test IP address"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "dns_name", "test.example.com"),
					resource.TestCheckResourceAttrSet("netbox_ip_address.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_ip_address" "test" {
  ip_address  = "192.168.100.1/24"
  status      = "deprecated"
  description = "terraform test IP address updated"
  dns_name    = "updated.example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", "192.168.100.1/24"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "deprecated"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", "terraform test IP address updated"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "dns_name", "updated.example.com"),
					resource.TestCheckResourceAttrSet("netbox_ip_address.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccIpAddressResourceWithCustomFields は custom_fields 属性の acceptance test です。
func TestAccIpAddressResourceWithCustomFields(t *testing.T) {
	rCfName := acctest.RandomWithPrefix("tf_acc_cf_ipaddr")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["ipam.ipaddress"]
}

resource "netbox_ip_address" "test_cf" {
  ip_address = "192.168.199.1/24"
  status     = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "ipaddr-cf-value"
  }
}
`, rCfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test_cf", "ip_address", "192.168.199.1/24"),
					resource.TestCheckResourceAttrSet("netbox_ip_address.test_cf", "id"),
					resource.TestCheckResourceAttr("netbox_ip_address.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "ipaddr-cf-value"),
				),
			},
			// カスタムフィールド値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["ipam.ipaddress"]
}

resource "netbox_ip_address" "test_cf" {
  ip_address = "192.168.199.1/24"
  status     = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "ipaddr-cf-updated"
  }
}
`, rCfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "ipaddr-cf-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccIpAddressResourceWithInterface は IP アドレスをデバイスインターフェースに割り当てる acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1 が存在している必要があります。
func TestAccIpAddressResourceWithInterface(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-ipaddr")

	resource.Test(t, resource.TestCase{
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

resource "netbox_ip_address" "test" {
  ip_address     = "192.168.200.1/24"
  status         = "active"
  description    = "terraform test IP with interface"
  interface_id   = netbox_device_interface.test.id
  interface_type = "dcim.interface"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", "192.168.200.1/24"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", "terraform test IP with interface"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "interface_type", "dcim.interface"),
					resource.TestCheckResourceAttrSet("netbox_ip_address.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_ip_address.test", "interface_id"),
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

resource "netbox_ip_address" "test" {
  ip_address  = "192.168.200.1/24"
  status      = "active"
  description = "terraform test IP detached"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", "192.168.200.1/24"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", "terraform test IP detached"),
					resource.TestCheckNoResourceAttr("netbox_ip_address.test", "interface_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
