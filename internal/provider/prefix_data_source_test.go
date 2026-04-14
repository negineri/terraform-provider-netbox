// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrefixDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_prefixes" "all" {}
data "netbox_prefix" "test" {
  id = data.netbox_prefixes.all.prefixes[0].id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_prefix.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_prefix.test", "prefix"),
					resource.TestCheckResourceAttrSet("data.netbox_prefix.test", "status"),
				),
			},
		},
	})
}
