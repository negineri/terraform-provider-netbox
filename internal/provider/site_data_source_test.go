// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSiteDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_sites" "all" {}
data "netbox_site" "test" {
  id = data.netbox_sites.all.sites[0].id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_site.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_site.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_site.test", "slug"),
				),
			},
		},
	})
}
