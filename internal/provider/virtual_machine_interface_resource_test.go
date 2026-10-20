// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccVirtualMachineInterfaceResource は acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に cluster_id=1 が存在している必要があります。
func TestAccVirtualMachineInterfaceResource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-iface")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
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
  enabled            = true
  description        = "terraform test vm interface"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "name", "eth0"),
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "enabled", "true"),
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "description", "terraform test vm interface"),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine_interface.test", "id"),
					resource.TestCheckResourceAttrPair(
						"netbox_virtual_machine_interface.test", "virtual_machine_id",
						"netbox_virtual_machine.test", "id",
					),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name       = %q
  cluster_id = 1
  status     = "active"
}

resource "netbox_virtual_machine_interface" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0-updated"
  enabled            = false
  description        = "terraform test vm interface updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "name", "eth0-updated"),
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "enabled", "false"),
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "description", "terraform test vm interface updated"),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine_interface.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccVirtualMachineInterfaceResource_macAddress は mac_address で primary MAC を作る場合の acceptance test です。
// mac_address の書き込みは NetBox 4.7 以降でのみ有効です。
func TestAccVirtualMachineInterfaceResource_macAddress(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-iface-mac")

	config := func(iface string) string {
		return providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name       = %q
  cluster_id = 1
  status     = "active"
}

resource "netbox_virtual_machine_interface" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0"
%s
}
`, rName, iface)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccSkipIfNetboxOlderThan(t, 4, 7) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// mac_address と primary_mac_address_id は同時に指定できないこと
			{
				Config: config(`  mac_address            = "BC:24:11:34:E1:75"
  primary_mac_address_id = 1`),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
			// 小文字の MAC で作成しても apply 後と次の plan で差分が出ないこと
			{
				Config: config(`  mac_address = "bc:24:11:34:e1:73"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "mac_address", "bc:24:11:34:e1:73"),
					resource.TestCheckNoResourceAttr("netbox_virtual_machine_interface.test", "primary_mac_address_id"),
				),
			},
			// MAC の変更
			{
				Config: config(`  mac_address = "BC:24:11:34:E1:74"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine_interface.test", "mac_address", "BC:24:11:34:E1:74"),
					resource.TestCheckNoResourceAttr("netbox_virtual_machine_interface.test", "primary_mac_address_id"),
				),
			},
			// MAC の削除
			{
				Config: config(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("netbox_virtual_machine_interface.test", "mac_address"),
					resource.TestCheckNoResourceAttr("netbox_virtual_machine_interface.test", "primary_mac_address_id"),
				),
			},
		},
	})
}
