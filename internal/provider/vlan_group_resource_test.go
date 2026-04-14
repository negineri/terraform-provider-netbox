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

// TestAccVlanGroupResource は netbox_vlan_group の acceptance test です。
// slug 未指定時の自動生成・CRUD・rename を一連のステップで検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccVlanGroupResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-vlan-group")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// AutoSlug: slug 未指定時に自動生成されることを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan_group" "test" {
  name = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_vlan_group.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_vlan_group.test", "id"),
				),
			},
			// Create and Read testing (with explicit slug)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan_group" "test" {
  name        = %q
  slug        = "tf-acc-test-vlan-group-slug"
  description = "terraform test VLAN group"
  min_vid     = 1
  max_vid     = 4094
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "slug", "tf-acc-test-vlan-group-slug"),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "description", "terraform test VLAN group"),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "min_vid", "1"),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "max_vid", "4094"),
					resource.TestCheckResourceAttrSet("netbox_vlan_group.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_vlan_group.test"]
						if !ok {
							return fmt.Errorf("resource netbox_vlan_group.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan_group" "test" {
  name        = %q
  slug        = "tf-acc-test-vlan-group-slug"
  description = "terraform test VLAN group updated"
  min_vid     = 100
  max_vid     = 200
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "description", "terraform test VLAN group updated"),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "min_vid", "100"),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "max_vid", "200"),
					resource.TestCheckResourceAttrSet("netbox_vlan_group.test", "id"),
				),
			},
			// Rename testing: 名前を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan_group" "test" {
  name = %q
  slug = "tf-acc-test-vlan-group-slug"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_vlan_group.test"]
						if !ok {
							return fmt.Errorf("resource netbox_vlan_group.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("vlan_group ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccVlanGroupResourceWithScope は scope_type / scope_id を使った acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccVlanGroupResourceWithScope(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site-for-vg")
	rGroupName := acctest.RandomWithPrefix("tf-acc-test-vlan-group")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create VLAN group scoped to a site
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_vlan_group" "test" {
  name       = %q
  scope_type = "dcim.site"
  scope_id   = netbox_site.test.id
}
`, rSiteName, rGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "name", rGroupName),
					resource.TestCheckResourceAttr("netbox_vlan_group.test", "scope_type", "dcim.site"),
					resource.TestCheckResourceAttrSet("netbox_vlan_group.test", "scope_id"),
					resource.TestCheckResourceAttrSet("netbox_vlan_group.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_vlan_group.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
