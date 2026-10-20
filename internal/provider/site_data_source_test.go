// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSiteDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-site-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name = %q
}

data "netbox_sites" "all" {
  name = netbox_site.test.name
}

data "netbox_site" "test" {
  id = data.netbox_sites.all.sites[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_site.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_site.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_site.test", "slug"),
				),
			},
		},
	})
}
