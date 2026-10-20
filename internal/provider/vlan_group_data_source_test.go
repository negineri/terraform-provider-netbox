// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVlanGroupDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vlan-group-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan_group" "test" {
  name = %q
}

# netbox_vlan_groups には絞り込み条件がないため、作成した VLAN group を直接読む
data "netbox_vlan_group" "test" {
  id = netbox_vlan_group.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_vlan_group.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_vlan_group.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_vlan_group.test", "slug"),
				),
			},
		},
	})
}
