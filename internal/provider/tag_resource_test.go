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

// TestAccTagResource は acceptance test です。
// auto-slug・CRUD・explicit slug・rename を一連のステップで検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccTagResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-tag")
	rSlug := "custom-slug-" + acctest.RandString(8)
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create (auto-slug): slug 未指定時に自動生成されることを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tag" "test" {
  name        = %q
  color       = "aa1409"
  description = "terraform test tag"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tag.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_tag.test", "color", "aa1409"),
					resource.TestCheckResourceAttr("netbox_tag.test", "description", "terraform test tag"),
					resource.TestCheckResourceAttrSet("netbox_tag.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_tag.test", "slug"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_tag.test"]
						if !ok {
							return fmt.Errorf("resource netbox_tag.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update (explicit slug): slug を明示的に指定する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tag" "test" {
  name  = %q
  color = "aa1409"
  slug  = %q
}
`, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tag.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_tag.test", "slug", rSlug),
				),
			},
			// Update: color と description を変更する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tag" "test" {
  name        = %q
  color       = "4caf50"
  slug        = %q
  description = "terraform test tag updated"
}
`, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tag.test", "color", "4caf50"),
					resource.TestCheckResourceAttr("netbox_tag.test", "description", "terraform test tag updated"),
					resource.TestCheckResourceAttrSet("netbox_tag.test", "id"),
				),
			},
			// Rename testing: タグ名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tag" "test" {
  name  = %q
  color = "4caf50"
  slug  = %q
}
`, rNameRenamed, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tag.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_tag.test"]
						if !ok {
							return fmt.Errorf("resource netbox_tag.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("tag ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
