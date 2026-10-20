// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVirtualMachineDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name       = %q
  cluster_id = 1
  status     = "active"
}

data "netbox_virtual_machines" "all" {
  name = netbox_virtual_machine.test.name
}

data "netbox_virtual_machine" "test" {
  id = data.netbox_virtual_machines.all.virtual_machines[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_virtual_machine.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_virtual_machine.test", "name", rName),
				),
			},
		},
	})
}
