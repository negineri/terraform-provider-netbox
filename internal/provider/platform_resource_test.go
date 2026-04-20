// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPlatformResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-platform")
	rSlug := rName + "-s"
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// AutoSlug: slug 未指定時に自動生成されることを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_platform" "test" {
  name = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_platform.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_platform.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_platform.test", "id"),
				),
			},
			// Create and Read testing (with explicit slug)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_platform" "test" {
  name        = %q
  slug        = %q
  description = "terraform test platform"
}
`, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_platform.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_platform.test", "slug", rSlug),
					resource.TestCheckResourceAttr("netbox_platform.test", "description", "terraform test platform"),
					resource.TestCheckResourceAttrSet("netbox_platform.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_platform.test"]
						if !ok {
							return fmt.Errorf("resource netbox_platform.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_platform" "test" {
  name        = %q
  slug        = %q
  description = "terraform test platform updated"
}
`, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_platform.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_platform.test", "description", "terraform test platform updated"),
					resource.TestCheckResourceAttrSet("netbox_platform.test", "id"),
				),
			},
			// Rename testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_platform" "test" {
  name = %q
  slug = %q
}
`, rNameRenamed, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_platform.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_platform.test"]
						if !ok {
							return fmt.Errorf("resource netbox_platform.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("platform ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
