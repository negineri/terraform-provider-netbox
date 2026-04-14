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
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccLocationResource(t *testing.T) {
	var capturedID string
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-location")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_location" "test" {
  name        = %q
  slug        = "tf-acc-test-location-slug"
  site_id     = netbox_site.test.id
  status      = "active"
  description = "terraform test location"
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_location.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_location.test", "slug", "tf-acc-test-location-slug"),
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
  slug        = "tf-acc-test-location-slug"
  site_id     = netbox_site.test.id
  status      = "planned"
  description = "terraform test location updated"
}
`, rSiteName, rName),
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
  slug    = "tf-acc-test-location-slug"
  site_id = netbox_site.test.id
  status  = "planned"
}
`, rSiteName, rNameRenamed),
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

// TestAccLocationResourceWithCustomFields は custom_fields 属性の acceptance test です。
func TestAccLocationResourceWithCustomFields(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-location-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_location"), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// カスタムフィールドを作成してロケーションに設定する
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
			// カスタムフィールド値を更新する
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

// TestAccLocationResourceAutoSlug は slug 未指定時の自動生成を検証する acceptance test です。
func TestAccLocationResourceAutoSlug(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-location")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// slug を明示せずに作成し、Computed として返ってくることを確認
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
			// Delete testing automatically occurs in TestCase
		},
	})
}
