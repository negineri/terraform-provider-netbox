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

// customFieldDeviceConfig は netbox_custom_field と netbox_device を組み合わせた設定を生成するヘルパーです。
func customFieldDeviceConfig(cfName, devName, cfType, cfValue string) string {
	return providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = %q
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = %q
  }
}
`, cfName, cfType, devName, cfValue)
}

// TestAccDeviceResource は acceptance test です。
// CRUD・rename・serial・description・status の更新を一連のステップで検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1 が存在している必要があります。
func TestAccDeviceResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-device")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
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
  serial         = "SN-001"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test", "device_type_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "role_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "site_id", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_device.test", "description", "terraform test device"),
					resource.TestCheckResourceAttr("netbox_device.test", "serial", "SN-001"),
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
			// Update: status・description・serial を変更する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "planned"
  description    = "terraform test device updated"
  serial         = "SN-002"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_device.test", "description", "terraform test device updated"),
					resource.TestCheckResourceAttr("netbox_device.test", "serial", "SN-002"),
					resource.TestCheckResourceAttrSet("netbox_device.test", "id"),
				),
			},
			// Rename: デバイス名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "planned"
  description    = "terraform test device updated"
  serial         = "SN-002"
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

// TestAccDeviceResourceWithTypedCustomFields はカスタムフィールド型ごとの値の設定・読み返しを検証します。
// - integer 型: "42" が送信・読み返しともに "42"（ドリフトなし）
// - boolean 型: "true"/"false" が送信・読み返しともに同じ値
// - text 型に数値文字列: integer に変換されない（回帰テスト）
// - text 型に "true": boolean に変換されない（回帰テスト）

func TestAccDeviceResourceWithTypedCustomFields(t *testing.T) {
	tests := []struct {
		name         string
		cfPrefix     string // Netbox CF 名は 50 文字以内のため短い prefix を使う
		cfType       string
		initialValue string
		updatedValue string
	}{
		{
			name:         "integer",
			cfPrefix:     "tf_acc_cf_int",
			cfType:       "integer",
			initialValue: "42",
			updatedValue: "100",
		},
		{
			name:         "boolean",
			cfPrefix:     "tf_acc_cf_bool",
			cfType:       "boolean",
			initialValue: "true",
			updatedValue: "false",
		},
		{
			name:         "text_with_numeric_string",
			cfPrefix:     "tf_acc_cf_tnum",
			cfType:       "text",
			initialValue: "42",
			updatedValue: "99",
		},
		{
			name:         "text_with_true_string",
			cfPrefix:     "tf_acc_cf_ttrue",
			cfType:       "text",
			initialValue: "true",
			updatedValue: "false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rName := acctest.RandomWithPrefix("tf-acc-test-dev-" + tc.name)
			rCfName := strings.ReplaceAll(acctest.RandomWithPrefix(tc.cfPrefix), "-", "_")

			resource.ParallelTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					// 初期値を設定する
					{
						Config: customFieldDeviceConfig(rCfName, rName, tc.cfType, tc.initialValue),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("netbox_device.test_cf", "name", rName),
							resource.TestCheckResourceAttr("netbox_device.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), tc.initialValue),
						),
					},
					// 値を更新する
					{
						Config: customFieldDeviceConfig(rCfName, rName, tc.cfType, tc.updatedValue),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("netbox_device.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), tc.updatedValue),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

// TestAccDeviceResourceWithTags は tags 属性の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1, tag_id=37 が存在している必要があります。
func TestAccDeviceResourceWithTags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-tags")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with tags
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
  tags           = [37]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_device.test", "tags.0", "37"),
					resource.TestCheckResourceAttrSet("netbox_device.test", "id"),
				),
			},
			// Update: タグを空にする
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
  tags           = []
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test", "tags.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
