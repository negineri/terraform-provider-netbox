// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrefixDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-prefix-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_prefix" "test" {
  prefix      = "198.51.100.0/24"
  status      = "active"
  description = %q
}

data "netbox_prefixes" "all" {
  prefix = netbox_prefix.test.prefix
}

data "netbox_prefix" "test" {
  id = data.netbox_prefixes.all.prefixes[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_prefix.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_prefix.test", "prefix", "198.51.100.0/24"),
					resource.TestCheckResourceAttr("data.netbox_prefix.test", "status", "active"),
					resource.TestCheckResourceAttr("data.netbox_prefix.test", "description", rName),
				),
			},
		},
	})
}
