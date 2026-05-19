// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceInterfaceDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-iface-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
}

resource "netbox_mac_address" "eth0_mac" {
  mac_address = "AA:BB:CC:DD:EE:01"
  lifecycle {
    ignore_changes = [assigned_object_type, assigned_object_id]
  }
}

resource "netbox_device_interface" "test" {
  device_id              = netbox_device.test.id
  name                   = "eth0"
  type                   = "virtual"
  primary_mac_address_id = netbox_mac_address.eth0_mac.id
  description            = "terraform test interface data source"
}

data "netbox_device_interface" "test" {
  id = netbox_device_interface.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_device_interface.test", "name", "eth0"),
					resource.TestCheckResourceAttr("data.netbox_device_interface.test", "type", "virtual"),
					resource.TestCheckResourceAttr("data.netbox_device_interface.test", "mac_address", "AA:BB:CC:DD:EE:01"),
					resource.TestCheckResourceAttr("data.netbox_device_interface.test", "description", "terraform test interface data source"),
					resource.TestCheckResourceAttrSet("data.netbox_device_interface.test", "device_id"),
					resource.TestCheckResourceAttrSet("data.netbox_device_interface.test", "id"),
				),
			},
		},
	})
}

func TestAccDeviceInterfacesDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-ifaces-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
}

resource "netbox_mac_address" "eth0_mac" {
  mac_address = "AA:BB:CC:DD:EE:10"
  lifecycle {
    ignore_changes = [assigned_object_type, assigned_object_id]
  }
}

resource "netbox_mac_address" "eth1_mac" {
  mac_address = "AA:BB:CC:DD:EE:11"
  lifecycle {
    ignore_changes = [assigned_object_type, assigned_object_id]
  }
}

resource "netbox_device_interface" "eth0" {
  device_id              = netbox_device.test.id
  name                   = "eth0"
  type                   = "virtual"
  primary_mac_address_id = netbox_mac_address.eth0_mac.id
  description            = "terraform test interface list"
}

resource "netbox_device_interface" "eth1" {
  device_id              = netbox_device.test.id
  name                   = "eth1"
  type                   = "virtual"
  primary_mac_address_id = netbox_mac_address.eth1_mac.id
  description            = "terraform test interface list"
}

data "netbox_device_interfaces" "by_device" {
  device_id  = netbox_device.test.id
  depends_on = [netbox_device_interface.eth0, netbox_device_interface.eth1]
}

data "netbox_device_interfaces" "by_name" {
  device_id  = netbox_device.test.id
  name       = "eth0"
  depends_on = [netbox_device_interface.eth0, netbox_device_interface.eth1]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_device_interfaces.by_device", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_device_interfaces.by_device", "device_interfaces.0.id"),
					resource.TestCheckResourceAttr("data.netbox_device_interfaces.by_name", "device_interfaces.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_device_interfaces.by_name", "device_interfaces.0.name", "eth0"),
					resource.TestCheckResourceAttr("data.netbox_device_interfaces.by_name", "device_interfaces.0.mac_address", "AA:BB:CC:DD:EE:10"),
				),
			},
		},
	})
}
