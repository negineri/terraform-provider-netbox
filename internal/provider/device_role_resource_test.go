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

// TestAccDeviceRoleResource は netbox_device_role の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccDeviceRoleResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-device-role")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name        = %q
  slug        = "tf-acc-test-device-role-slug"
  color       = "aa1409"
  vm_role     = false
  description = "terraform test device role"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_role.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_device_role.test", "slug", "tf-acc-test-device-role-slug"),
					resource.TestCheckResourceAttr("netbox_device_role.test", "color", "aa1409"),
					resource.TestCheckResourceAttr("netbox_device_role.test", "vm_role", "false"),
					resource.TestCheckResourceAttr("netbox_device_role.test", "description", "terraform test device role"),
					resource.TestCheckResourceAttrSet("netbox_device_role.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_device_role.test"]
						if !ok {
							return fmt.Errorf("resource netbox_device_role.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name        = %q
  slug        = "tf-acc-test-device-role-slug"
  color       = "4caf50"
  vm_role     = true
  description = "terraform test device role updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_role.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_device_role.test", "color", "4caf50"),
					resource.TestCheckResourceAttr("netbox_device_role.test", "vm_role", "true"),
					resource.TestCheckResourceAttr("netbox_device_role.test", "description", "terraform test device role updated"),
					resource.TestCheckResourceAttrSet("netbox_device_role.test", "id"),
				),
			},
			// Rename testing: device_role 名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name  = %q
  slug  = "tf-acc-test-device-role-slug"
  color = "4caf50"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_role.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_device_role.test"]
						if !ok {
							return fmt.Errorf("resource netbox_device_role.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("device role ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccDeviceRoleResourceAutoSlug は slug 未指定時の自動生成を検証する acceptance test です。
func TestAccDeviceRoleResourceAutoSlug(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-role")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name  = %q
  color = "aa1409"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_role.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_device_role.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_device_role.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
