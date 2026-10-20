// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVlanDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vlan-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan" "test" {
  name   = %q
  vid    = 3999
  status = "active"
}

data "netbox_vlans" "all" {
  name = netbox_vlan.test.name
}

data "netbox_vlan" "test" {
  id = data.netbox_vlans.all.vlans[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_vlan.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_vlan.test", "name", rName),
					resource.TestCheckResourceAttr("data.netbox_vlan.test", "vid", "3999"),
				),
			},
		},
	})
}
