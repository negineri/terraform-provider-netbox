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

// TestAccManufacturerResource は netbox_manufacturer の acceptance test です。
// slug 未指定時の自動生成・CRUD・rename を一連のステップで検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccManufacturerResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-manufacturer")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// AutoSlug: slug 未指定時に自動生成されることを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_manufacturer.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_manufacturer.test", "id"),
				),
			},
			// Create and Read testing (with explicit attributes)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name        = %q
  slug        = "tf-acc-test-manufacturer-slug"
  description = "terraform test manufacturer"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "slug", "tf-acc-test-manufacturer-slug"),
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "description", "terraform test manufacturer"),
					resource.TestCheckResourceAttrSet("netbox_manufacturer.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_manufacturer.test"]
						if !ok {
							return fmt.Errorf("resource netbox_manufacturer.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name        = %q
  slug        = "tf-acc-test-manufacturer-slug"
  description = "terraform test manufacturer updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "description", "terraform test manufacturer updated"),
					resource.TestCheckResourceAttrSet("netbox_manufacturer.test", "id"),
				),
			},
			// Rename testing: name を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_manufacturer" "test" {
  name = %q
  slug = "tf-acc-test-manufacturer-slug"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_manufacturer.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_manufacturer.test"]
						if !ok {
							return fmt.Errorf("resource netbox_manufacturer.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("manufacturer ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
