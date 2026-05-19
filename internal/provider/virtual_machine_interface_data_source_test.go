// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVirtualMachineInterfaceDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-iface-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name    = %q
  site_id = 1
  status  = "active"
}

resource "netbox_virtual_machine_interface" "test" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0"
  description        = "terraform test vm interface data source"
}

data "netbox_virtual_machine_interface" "test" {
  id = netbox_virtual_machine_interface.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_virtual_machine_interface.test", "name", "eth0"),
					resource.TestCheckResourceAttr("data.netbox_virtual_machine_interface.test", "description", "terraform test vm interface data source"),
					resource.TestCheckResourceAttrSet("data.netbox_virtual_machine_interface.test", "virtual_machine_id"),
					resource.TestCheckResourceAttrSet("data.netbox_virtual_machine_interface.test", "id"),
				),
			},
		},
	})
}

func TestAccVirtualMachineInterfacesDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-ifaces-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name    = %q
  site_id = 1
  status  = "active"
}

resource "netbox_virtual_machine_interface" "eth0" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0"
  description        = "terraform test vm interface list"
}

resource "netbox_virtual_machine_interface" "eth1" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth1"
  description        = "terraform test vm interface list"
}

data "netbox_virtual_machine_interfaces" "by_vm" {
  virtual_machine_id = netbox_virtual_machine.test.id
  depends_on         = [netbox_virtual_machine_interface.eth0, netbox_virtual_machine_interface.eth1]
}

data "netbox_virtual_machine_interfaces" "by_name" {
  virtual_machine_id = netbox_virtual_machine.test.id
  name               = "eth0"
  depends_on         = [netbox_virtual_machine_interface.eth0, netbox_virtual_machine_interface.eth1]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_virtual_machine_interfaces.by_vm", "virtual_machine_interfaces.#", "2"),
					resource.TestCheckResourceAttrSet("data.netbox_virtual_machine_interfaces.by_vm", "id"),
					resource.TestCheckResourceAttr("data.netbox_virtual_machine_interfaces.by_name", "virtual_machine_interfaces.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_virtual_machine_interfaces.by_name", "virtual_machine_interfaces.0.name", "eth0"),
				),
			},
		},
	})
}
