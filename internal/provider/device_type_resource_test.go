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

// TestAccDeviceTypeResource は netbox_device_type の acceptance test です。
// slug 未指定時の自動生成・CRUD・rename を一連のステップで検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に manufacturer_id=1 が存在している必要があります。
func TestAccDeviceTypeResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-device-type")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// AutoSlug: slug 未指定時に自動生成されることを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_type" "test" {
  manufacturer_id = 1
  model           = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_type.test", "model", rName),
					resource.TestCheckResourceAttrSet("netbox_device_type.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_device_type.test", "id"),
				),
			},
			// Create and Read testing (with explicit attributes)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_type" "test" {
  manufacturer_id = 1
  model           = %q
  slug            = "tf-acc-test-device-type-slug"
  part_number     = "TEST-PART-001"
  u_height        = 2
  is_full_depth   = true
  description     = "terraform test device type"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_type.test", "model", rName),
					resource.TestCheckResourceAttr("netbox_device_type.test", "slug", "tf-acc-test-device-type-slug"),
					resource.TestCheckResourceAttr("netbox_device_type.test", "part_number", "TEST-PART-001"),
					resource.TestCheckResourceAttr("netbox_device_type.test", "u_height", "2"),
					resource.TestCheckResourceAttr("netbox_device_type.test", "is_full_depth", "true"),
					resource.TestCheckResourceAttr("netbox_device_type.test", "description", "terraform test device type"),
					resource.TestCheckResourceAttrSet("netbox_device_type.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_device_type.test"]
						if !ok {
							return fmt.Errorf("resource netbox_device_type.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_type" "test" {
  manufacturer_id = 1
  model           = %q
  slug            = "tf-acc-test-device-type-slug"
  part_number     = "TEST-PART-002"
  u_height        = 4
  is_full_depth   = false
  description     = "terraform test device type updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_type.test", "model", rName),
					resource.TestCheckResourceAttr("netbox_device_type.test", "part_number", "TEST-PART-002"),
					resource.TestCheckResourceAttr("netbox_device_type.test", "u_height", "4"),
					resource.TestCheckResourceAttr("netbox_device_type.test", "is_full_depth", "false"),
					resource.TestCheckResourceAttr("netbox_device_type.test", "description", "terraform test device type updated"),
					resource.TestCheckResourceAttrSet("netbox_device_type.test", "id"),
				),
			},
			// Rename testing: model 名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_type" "test" {
  manufacturer_id = 1
  model           = %q
  slug            = "tf-acc-test-device-type-slug"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device_type.test", "model", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_device_type.test"]
						if !ok {
							return fmt.Errorf("resource netbox_device_type.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("device type ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
