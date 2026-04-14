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

// TestAccSiteResource は netbox_site の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccSiteResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-site")
	rNameRenamed := rName + "-renamed"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name             = %q
  slug             = "tf-acc-test-site-slug"
  status           = "active"
  description      = "terraform test site"
  facility         = "Test Datacenter"
  time_zone        = "Asia/Tokyo"
  physical_address = "1-1-1 Test Street, Tokyo"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_site.test", "slug", "tf-acc-test-site-slug"),
					resource.TestCheckResourceAttr("netbox_site.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_site.test", "description", "terraform test site"),
					resource.TestCheckResourceAttr("netbox_site.test", "facility", "Test Datacenter"),
					resource.TestCheckResourceAttr("netbox_site.test", "time_zone", "Asia/Tokyo"),
					resource.TestCheckResourceAttr("netbox_site.test", "physical_address", "1-1-1 Test Street, Tokyo"),
					resource.TestCheckResourceAttrSet("netbox_site.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_site.test"]
						if !ok {
							return fmt.Errorf("resource netbox_site.test not found")
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
  name             = %q
  slug             = "tf-acc-test-site-slug"
  status           = "planned"
  description      = "terraform test site updated"
  facility         = "Updated Datacenter"
  time_zone        = "America/New_York"
  physical_address = "2-2-2 Updated Street, Osaka"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_site.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_site.test", "description", "terraform test site updated"),
					resource.TestCheckResourceAttr("netbox_site.test", "facility", "Updated Datacenter"),
					resource.TestCheckResourceAttr("netbox_site.test", "time_zone", "America/New_York"),
					resource.TestCheckResourceAttr("netbox_site.test", "physical_address", "2-2-2 Updated Street, Osaka"),
					resource.TestCheckResourceAttrSet("netbox_site.test", "id"),
				),
			},
			// Rename testing: site名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name      = %q
  slug      = "tf-acc-test-site-slug"
  status    = "planned"
  time_zone = "America/New_York"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_site.test"]
						if !ok {
							return fmt.Errorf("resource netbox_site.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("site ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccSiteResourceWithCustomFields は custom_fields 属性の acceptance test です。
func TestAccSiteResourceWithCustomFields(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-site-cf")
	rCfName := acctest.RandomWithPrefix("tf_acc_cf_site")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// カスタムフィールドを作成してサイトに設定する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.site"]
}

resource "netbox_site" "test_cf" {
  name   = %q
  status = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "site-cf-value"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test_cf", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_site.test_cf", "id"),
					resource.TestCheckResourceAttr("netbox_site.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "site-cf-value"),
				),
			},
			// カスタムフィールド値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.site"]
}

resource "netbox_site" "test_cf" {
  name   = %q
  status = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "site-cf-updated"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "site-cf-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccSiteResourceAutoSlug は slug 未指定時の自動生成を検証する acceptance test です。
func TestAccSiteResourceAutoSlug(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-site")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// slug を明示せずに作成し、Computed として返ってくることを確認
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_site.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_site.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
