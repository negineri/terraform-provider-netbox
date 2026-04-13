// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVirtualMachinesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "netbox_virtual_machines" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_virtual_machines.test", "id", "netbox_virtual_machines"),
					resource.TestCheckResourceAttrSet("data.netbox_virtual_machines.test", "virtual_machines.#"),
				),
			},
		},
	})
}
