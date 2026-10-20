// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-tag-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = %q
}

data "netbox_tags" "all" {
  name = netbox_tag.test.name
}

data "netbox_tag" "test" {
  id = data.netbox_tags.all.tags[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_tag.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_tag.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_tag.test", "slug"),
				),
			},
		},
	})
}
