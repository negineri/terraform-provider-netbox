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

// TestAccDeviceResource は acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1 が存在している必要があります。
func TestAccDeviceResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-device")
	rNameRenamed := rName + "-renamed"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
  description    = "terraform test device"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test", "device_type_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "role_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "site_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_device.test", "description", "terraform test device"),
					resource.TestCheckResourceAttrSet("netbox_device.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_device.test"]
						if !ok {
							return fmt.Errorf("resource netbox_device.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "planned"
  description    = "terraform test device updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test", "device_type_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "role_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "site_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_device.test", "description", "terraform test device updated"),
					resource.TestCheckResourceAttrSet("netbox_device.test", "id"),
				),
			},
			// Rename testing: デバイス名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "planned"
  description    = "terraform test device updated"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_device.test"]
						if !ok {
							return fmt.Errorf("resource netbox_device.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("device ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
