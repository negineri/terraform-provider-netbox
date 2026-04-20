// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPlatformDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-platform")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_platform" "test" {
  name        = %q
  description = "test platform for data source"
}

data "netbox_platform" "test" {
  id = netbox_platform.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_platform.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_platform.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_platform.test", "slug"),
					resource.TestCheckResourceAttr("data.netbox_platform.test", "description", "test platform for data source"),
				),
			},
		},
	})
}

func TestAccPlatformsDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-platform")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_platform" "test" {
  name = %q
}

data "netbox_platforms" "all" {
  depends_on = [netbox_platform.test]
}

data "netbox_platform" "test" {
  id = data.netbox_platforms.all.platforms[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_platform.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_platform.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_platform.test", "slug"),
				),
			},
		},
	})
}
