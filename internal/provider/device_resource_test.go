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

// TestAccDeviceResourceWithCustomFields は custom_fields 属性の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1 が存在している必要があります。
func TestAccDeviceResourceWithCustomFields(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_device"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// カスタムフィールドを作成してデバイスに設定する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "initial-value"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_cf", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_device.test_cf", "id"),
					resource.TestCheckResourceAttrSet("netbox_device.test_cf", fmt.Sprintf("custom_fields.%s", rCfName)),
				),
			},
			// カスタムフィールド値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "updated-value"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_cf", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "updated-value"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccDeviceResourceWithIntegerCustomField は integer 型カスタムフィールドの設定・読み返しを検証します。
// 数値文字列 "42" が API に integer として送信され、読み返し時も "42" として取得できることを確認します。
func TestAccDeviceResourceWithIntegerCustomField(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-int-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_int_dev"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// integer カスタムフィールドに数値を設定する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "integer"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_int_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "42"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_int_cf", "name", rName),
					// API に integer として送信され、読み返しても "42" のまま（ドリフトなし）
					resource.TestCheckResourceAttr("netbox_device.test_int_cf", fmt.Sprintf("custom_fields.%s", rCfName), "42"),
				),
			},
			// 値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "integer"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_int_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "100"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_int_cf", fmt.Sprintf("custom_fields.%s", rCfName), "100"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccDeviceResourceWithBooleanCustomField は boolean 型カスタムフィールドの設定・読み返しを検証します。
// "true"/"false" が API に boolean として送信され、読み返し時も "true"/"false" として取得できることを確認します。
func TestAccDeviceResourceWithBooleanCustomField(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-bool-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_bool_dev"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// boolean カスタムフィールドに true を設定する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "boolean"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_bool_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "true"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_bool_cf", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test_bool_cf", fmt.Sprintf("custom_fields.%s", rCfName), "true"),
				),
			},
			// false に更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "boolean"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_bool_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "false"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_bool_cf", fmt.Sprintf("custom_fields.%s", rCfName), "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccDeviceResourceWithTextCustomField は text 型カスタムフィールドに数値文字列を設定しても
// integer に変換されないことを検証します（型認識対応の回帰テスト）。
func TestAccDeviceResourceWithTextCustomField(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-text-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_text_dev"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// text フィールドに数値文字列 "42" を設定 → "42" のまま（integer 扱いされない）
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_text_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "42"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_text_cf", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test_text_cf", fmt.Sprintf("custom_fields.%s", rCfName), "42"),
				),
			},
		},
	})
}

// TestAccDeviceResourceWithTextTrueCustomField は text 型カスタムフィールドに "true" を設定しても
// boolean に変換されないことを検証します（型認識対応の回帰テスト）。
func TestAccDeviceResourceWithTextTrueCustomField(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-texttrue-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_texttrue_dev"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// text フィールドに "true" を設定 → "true" のまま（bool 扱いされない）
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.device"]
}

resource "netbox_device" "test_texttrue_cf" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "true"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_texttrue_cf", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test_texttrue_cf", fmt.Sprintf("custom_fields.%s", rCfName), "true"),
				),
			},
		},
	})
}

// TestAccDeviceResourceWithTags は tags 属性の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に device_type_id=1, role_id=1, site_id=1, tag_id=37 が存在している必要があります。
func TestAccDeviceResourceWithTags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-tags")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with tags
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test_tags" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
  tags           = [37]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_tags", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test_tags", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_device.test_tags", "tags.0", "37"),
					resource.TestCheckResourceAttrSet("netbox_device.test_tags", "id"),
				),
			},
			// Update tags
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device" "test_tags" {
  name           = %q
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
  tags           = []
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_device.test_tags", "name", rName),
					resource.TestCheckResourceAttr("netbox_device.test_tags", "tags.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
