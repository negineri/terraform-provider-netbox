// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccLocationResource は netbox_location の acceptance test です。
// slug 未指定時の自動生成・CRUD・rename を一連のステップで検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccLocationResource(t *testing.T) {
	var capturedID string
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-location")
	rSlug := rName + "-s"
	rNameRenamed := rName + "-renamed"
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_location"), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// AutoSlug: slug 未指定時に自動生成されることを確認する
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
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_location.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_location.test", "id"),
				),
			},
			// Create and Read testing (with explicit slug)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_location" "test" {
  name        = %q
  slug        = %q
  site_id     = netbox_site.test.id
  status      = "active"
  description = "terraform test location"
}
`, rSiteName, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_location.test", "slug", rSlug),
					resource.TestCheckResourceAttr("netbox_location.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_location.test", "description", "terraform test location"),
					resource.TestCheckResourceAttrSet("netbox_location.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_location.test", "site_id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_location.test"]
						if !ok {
							return fmt.Errorf("resource netbox_location.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_location" "test" {
  name        = %q
  slug        = %q
  site_id     = netbox_site.test.id
  status      = "planned"
  description = "terraform test location updated"
}
`, rSiteName, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_location.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_location.test", "description", "terraform test location updated"),
					resource.TestCheckResourceAttrSet("netbox_location.test", "id"),
				),
			},
			// Rename testing: location 名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_location" "test" {
  name    = %q
  slug    = %q
  site_id = netbox_site.test.id
  status  = "planned"
}
`, rSiteName, rNameRenamed, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_location.test"]
						if !ok {
							return fmt.Errorf("resource netbox_location.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("location ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Custom fields: カスタムフィールドを設定する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.location"]
}

resource "netbox_location" "test_cf" {
  name    = %q
  site_id = netbox_site.test.id
  status  = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "location-cf-value"
  }
}
`, rSiteName, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.test_cf", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_location.test_cf", "id"),
					resource.TestCheckResourceAttr("netbox_location.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "location-cf-value"),
				),
			},
			// Custom fields: 値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.location"]
}

resource "netbox_location" "test_cf" {
  name    = %q
  site_id = netbox_site.test.id
  status  = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "location-cf-updated"
  }
}
`, rSiteName, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "location-cf-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccLocationResourceWithParent は parent_id を持つ階層的な location の acceptance test です。
func TestAccLocationResourceWithParent(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rParentName := acctest.RandomWithPrefix("tf-acc-test-location-parent")
	rChildName := acctest.RandomWithPrefix("tf-acc-test-location-child")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_location" "parent" {
  name    = %q
  site_id = netbox_site.test.id
  status  = "active"
}

resource "netbox_location" "child" {
  name      = %q
  site_id   = netbox_site.test.id
  parent_id = netbox_location.parent.id
  status    = "active"
}
`, rSiteName, rParentName, rChildName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.child", "name", rChildName),
					resource.TestCheckResourceAttrSet("netbox_location.child", "parent_id"),
					resource.TestCheckResourceAttrPair("netbox_location.child", "parent_id", "netbox_location.parent", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
