// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIpAddressDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_ip_addresses" "all" {}
data "netbox_ip_address" "test" {
  id = data.netbox_ip_addresses.all.ip_addresses[0].id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_ip_address.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_ip_address.test", "ip_address"),
					resource.TestCheckResourceAttrSet("data.netbox_ip_address.test", "status"),
				),
			},
		},
	})
}
