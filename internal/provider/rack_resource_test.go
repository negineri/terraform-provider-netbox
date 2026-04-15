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

// TestAccRackResource は netbox_rack の acceptance test です。
// CRUD・rename・各フィールド更新を一連のステップで検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に site_id=1 が存在している必要があります。
func TestAccRackResource(t *testing.T) {
	var capturedID string
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-rack")
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

resource "netbox_rack" "test" {
  name        = %q
  site_id     = netbox_site.test.id
  status      = "active"
  description = "terraform test rack"
  u_height    = 42
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_rack.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_rack.test", "description", "terraform test rack"),
					resource.TestCheckResourceAttr("netbox_rack.test", "u_height", "42"),
					resource.TestCheckResourceAttrSet("netbox_rack.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_rack.test", "site_id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_rack.test"]
						if !ok {
							return fmt.Errorf("resource netbox_rack.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update: status・description を変更する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_rack" "test" {
  name        = %q
  site_id     = netbox_site.test.id
  status      = "planned"
  description = "terraform test rack updated"
  u_height    = 42
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_rack.test", "description", "terraform test rack updated"),
					resource.TestCheckResourceAttrSet("netbox_rack.test", "id"),
				),
			},
			// Rename: ラック名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_rack" "test" {
  name     = %q
  site_id  = netbox_site.test.id
  status   = "planned"
  u_height = 42
}
`, rSiteName, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_rack.test"]
						if !ok {
							return fmt.Errorf("resource netbox_rack.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("rack ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccRackResourceWithFacilityId は facility_id フィールドの acceptance test です。
func TestAccRackResourceWithFacilityId(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-rack")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_rack" "test" {
  name        = %q
  site_id     = netbox_site.test.id
  status      = "active"
  facility_id = "RACK-001"
  u_height    = 42
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_rack.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_rack.test", "facility_id", "RACK-001"),
					resource.TestCheckResourceAttrSet("netbox_rack.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
