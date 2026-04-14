// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccVirtualMachinePrimaryIPResource は acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に cluster_id=1 が存在している必要があります。
func TestAccVirtualMachinePrimaryIPResource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-primary-ip")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create: IPv4 と IPv6 の両方を primary IP として設定する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name       = %q
  cluster_id = 1
  status     = "active"
}

resource "netbox_virtual_machine_interface" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0"
}

resource "netbox_ip_address" "ipv4" {
  ip_address     = "192.168.60.1/24"
  status         = "active"
  interface_id   = netbox_virtual_machine_interface.test.id
  interface_type = "virtualization.vminterface"
}

resource "netbox_ip_address" "ipv6" {
  ip_address     = "2001:db8:1::1/64"
  status         = "active"
  interface_id   = netbox_virtual_machine_interface.test.id
  interface_type = "virtualization.vminterface"
}

resource "netbox_virtual_machine_primary_ip" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  primary_ipv4_id    = netbox_ip_address.ipv4.id
  primary_ipv6_id    = netbox_ip_address.ipv6.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"netbox_virtual_machine_primary_ip.test", "virtual_machine_id",
						"netbox_virtual_machine.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"netbox_virtual_machine_primary_ip.test", "primary_ipv4_id",
						"netbox_ip_address.ipv4", "id",
					),
					resource.TestCheckResourceAttrPair(
						"netbox_virtual_machine_primary_ip.test", "primary_ipv6_id",
						"netbox_ip_address.ipv6", "id",
					),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine_primary_ip.test", "id"),
				),
			},
			// Update: primary IPv6 を削除し IPv4 のみにする
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name       = %q
  cluster_id = 1
  status     = "active"
}

resource "netbox_virtual_machine_interface" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0"
}

resource "netbox_ip_address" "ipv4" {
  ip_address     = "192.168.60.1/24"
  status         = "active"
  interface_id   = netbox_virtual_machine_interface.test.id
  interface_type = "virtualization.vminterface"
}

resource "netbox_ip_address" "ipv6" {
  ip_address     = "2001:db8:1::1/64"
  status         = "active"
  interface_id   = netbox_virtual_machine_interface.test.id
  interface_type = "virtualization.vminterface"
}

resource "netbox_virtual_machine_primary_ip" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  primary_ipv4_id    = netbox_ip_address.ipv4.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"netbox_virtual_machine_primary_ip.test", "primary_ipv4_id",
						"netbox_ip_address.ipv4", "id",
					),
					resource.TestCheckNoResourceAttr("netbox_virtual_machine_primary_ip.test", "primary_ipv6_id"),
				),
			},
			// Update: primary IPv4 を別のアドレスに変更する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name       = %q
  cluster_id = 1
  status     = "active"
}

resource "netbox_virtual_machine_interface" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0"
}

resource "netbox_ip_address" "ipv4" {
  ip_address     = "192.168.60.1/24"
  status         = "active"
  interface_id   = netbox_virtual_machine_interface.test.id
  interface_type = "virtualization.vminterface"
}

resource "netbox_ip_address" "ipv4_new" {
  ip_address     = "192.168.60.2/24"
  status         = "active"
  interface_id   = netbox_virtual_machine_interface.test.id
  interface_type = "virtualization.vminterface"
}

resource "netbox_ip_address" "ipv6" {
  ip_address     = "2001:db8:1::1/64"
  status         = "active"
  interface_id   = netbox_virtual_machine_interface.test.id
  interface_type = "virtualization.vminterface"
}

resource "netbox_virtual_machine_primary_ip" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  primary_ipv4_id    = netbox_ip_address.ipv4_new.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"netbox_virtual_machine_primary_ip.test", "primary_ipv4_id",
						"netbox_ip_address.ipv4_new", "id",
					),
					resource.TestCheckNoResourceAttr("netbox_virtual_machine_primary_ip.test", "primary_ipv6_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
