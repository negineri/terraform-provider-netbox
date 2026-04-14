// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegionDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-region")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_region" "test" {
  name        = %q
  description = "test region for data source"
}

data "netbox_region" "test" {
  id = netbox_region.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_region.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_region.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_region.test", "slug"),
					resource.TestCheckResourceAttr("data.netbox_region.test", "description", "test region for data source"),
				),
			},
		},
	})
}

func TestAccRegionsDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-region")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_region" "test" {
  name = %q
}

data "netbox_regions" "all" {
  depends_on = [netbox_region.test]
}

data "netbox_region" "test" {
  id = data.netbox_regions.all.regions[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_region.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_region.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_region.test", "slug"),
				),
			},
		},
	})
}
