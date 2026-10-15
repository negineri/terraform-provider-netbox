// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDevicesDataSource は netbox_devices データソースの acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1 が存在している必要があります。
// テスト用のタグを付けたデバイスを 2 台作り、そのタグで絞り込んで読みます。
func TestAccDevicesDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-devices")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = %[1]q
}
resource "netbox_device" "a" {
  name           = "%[1]s-a"
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  tags           = [netbox_tag.test.id]
}
resource "netbox_device" "b" {
  name           = "%[1]s-b"
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  tags           = [netbox_tag.test.id]
}
data "netbox_devices" "test" {
  tag        = netbox_tag.test.slug
  depends_on = [netbox_device.a, netbox_device.b]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_devices.test", "devices.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("data.netbox_devices.test", "devices.*", map[string]string{
						"name": rName + "-a",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.netbox_devices.test", "devices.*", map[string]string{
						"name": rName + "-b",
					}),
					resource.TestCheckResourceAttrSet("data.netbox_devices.test", "devices.0.id"),
				),
			},
		},
	})
}
