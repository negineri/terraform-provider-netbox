// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDeviceResource は acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1 が存在している必要があります。
func TestAccDeviceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_device" "test" {
  name           = "tf-acc-test-device"
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
  description    = "terraform test device"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "name", "tf-acc-test-device"),
					resource.TestCheckResourceAttr("netbox_device.test", "device_type_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "role_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "site_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_device.test", "description", "terraform test device"),
					resource.TestCheckResourceAttrSet("netbox_device.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_device" "test" {
  name           = "tf-acc-test-device"
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "planned"
  description    = "terraform test device updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "name", "tf-acc-test-device"),
					resource.TestCheckResourceAttr("netbox_device.test", "device_type_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "role_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "site_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_device.test", "description", "terraform test device updated"),
					resource.TestCheckResourceAttrSet("netbox_device.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
