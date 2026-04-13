// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVlanDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_vlans" "all" {}
data "netbox_vlan" "test" {
  id = data.netbox_vlans.all.vlans[0].id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_vlan.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_vlan.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_vlan.test", "vid"),
				),
			},
		},
	})
}
