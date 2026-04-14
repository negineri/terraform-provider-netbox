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

// TestAccRegionResource は netbox_region の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccRegionResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-region")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_region" "test" {
  name        = %q
  slug        = "tf-acc-test-region-slug"
  description = "terraform test region"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_region.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_region.test", "slug", "tf-acc-test-region-slug"),
					resource.TestCheckResourceAttr("netbox_region.test", "description", "terraform test region"),
					resource.TestCheckResourceAttrSet("netbox_region.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_region.test"]
						if !ok {
							return fmt.Errorf("resource netbox_region.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_region" "test" {
  name        = %q
  slug        = "tf-acc-test-region-slug"
  description = "terraform test region updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_region.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_region.test", "description", "terraform test region updated"),
					resource.TestCheckResourceAttrSet("netbox_region.test", "id"),
				),
			},
			// Rename testing: region 名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_region" "test" {
  name = %q
  slug = "tf-acc-test-region-slug"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_region.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_region.test"]
						if !ok {
							return fmt.Errorf("resource netbox_region.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("region ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccRegionResourceWithParent は parent_id を持つ階層的な region の acceptance test です。
func TestAccRegionResourceWithParent(t *testing.T) {
	rParentName := acctest.RandomWithPrefix("tf-acc-test-region-parent")
	rChildName := acctest.RandomWithPrefix("tf-acc-test-region-child")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_region" "parent" {
  name = %q
}

resource "netbox_region" "child" {
  name      = %q
  parent_id = netbox_region.parent.id
}
`, rParentName, rChildName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_region.child", "name", rChildName),
					resource.TestCheckResourceAttrSet("netbox_region.child", "parent_id"),
					resource.TestCheckResourceAttrPair("netbox_region.child", "parent_id", "netbox_region.parent", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccRegionResourceAutoSlug は slug 未指定時の自動生成を検証する acceptance test です。
func TestAccRegionResourceAutoSlug(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-region")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_region" "test" {
  name = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_region.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_region.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_region.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
