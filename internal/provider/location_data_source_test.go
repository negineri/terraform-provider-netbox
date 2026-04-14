// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLocationDataSource(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-location")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_location" "test" {
  name        = %q
  site_id     = netbox_site.test.id
  status      = "active"
  description = "test location for data source"
}

data "netbox_location" "test" {
  id = netbox_location.test.id
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_location.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_location.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_location.test", "slug"),
					resource.TestCheckResourceAttrSet("data.netbox_location.test", "site_id"),
					resource.TestCheckResourceAttr("data.netbox_location.test", "status", "active"),
					resource.TestCheckResourceAttr("data.netbox_location.test", "description", "test location for data source"),
				),
			},
		},
	})
}

func TestAccLocationsDataSource(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-location")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_location" "test" {
  name    = %q
  site_id = netbox_site.test.id
  status  = "active"
}

data "netbox_locations" "all" {
  depends_on = [netbox_location.test]
}

data "netbox_location" "test" {
  id = data.netbox_locations.all.locations[0].id
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_location.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_location.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_location.test", "slug"),
				),
			},
		},
	})
}
