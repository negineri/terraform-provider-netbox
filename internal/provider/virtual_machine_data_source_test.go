// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVirtualMachineDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_virtual_machines" "all" {}
data "netbox_virtual_machine" "test" {
  id = data.netbox_virtual_machines.all.virtual_machines[0].id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_virtual_machine.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_virtual_machine.test", "name"),
				),
			},
		},
	})
}
