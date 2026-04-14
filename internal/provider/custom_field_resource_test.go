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

// TestAccCustomFieldResource は netbox_custom_field の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccCustomFieldResource(t *testing.T) {
	var capturedID string
	rName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf"), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  label         = "TF Acc Test Field"
  type          = "text"
  content_types = ["dcim.device"]
  description   = "terraform acceptance test custom field"
  required      = false
  weight        = 100
  filter_logic  = "loose"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_field.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "label", "TF Acc Test Field"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "type", "text"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "content_types.#", "1"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "content_types.0", "dcim.device"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "description", "terraform acceptance test custom field"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "required", "false"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "weight", "100"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "filter_logic", "loose"),
					resource.TestCheckResourceAttrSet("netbox_custom_field.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_custom_field.test"]
						if !ok {
							return fmt.Errorf("resource netbox_custom_field.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  label         = "TF Acc Test Field Updated"
  type          = "text"
  content_types = ["dcim.device"]
  description   = "terraform acceptance test custom field updated"
  required      = false
  weight        = 200
  filter_logic  = "exact"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_field.test", "label", "TF Acc Test Field Updated"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "description", "terraform acceptance test custom field updated"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "weight", "200"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "filter_logic", "exact"),
					resource.TestCheckResourceAttrSet("netbox_custom_field.test", "id"),
				),
			},
			// ID が変わっていないことを確認
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  label         = "TF Acc Test Field Updated"
  type          = "text"
  content_types = ["dcim.device"]
  weight        = 200
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_custom_field.test"]
						if !ok {
							return fmt.Errorf("resource netbox_custom_field.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("custom field ID changed: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccCustomFieldResourceByType は各 type の作成・確認を検証するテーブル駆動テストです。
// integer / boolean 型がそれぞれ正しく作成されることを確認します。
func TestAccCustomFieldResourceByType(t *testing.T) {
	tests := []struct {
		name   string
		cfType string
	}{
		{name: "integer", cfType: "integer"},
		{name: "boolean", cfType: "boolean"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_"+tc.name), "-", "_")

			resource.ParallelTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = %q
  content_types = ["dcim.device"]
}
`, rName, tc.cfType),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("netbox_custom_field.test", "name", rName),
							resource.TestCheckResourceAttr("netbox_custom_field.test", "type", tc.cfType),
							resource.TestCheckResourceAttrSet("netbox_custom_field.test", "id"),
						),
					},
					// Delete testing automatically occurs in TestCase
				},
			})
		})
	}
}

// TestAccCustomFieldResourceAdvanced は高度な属性（group_name, ui_visible, ui_editable, is_cloneable）を
// テストします。Netbox 4.x では select 型が choice_set を要求するため longtext 型を使用します。
// required=true + dcim.device を使うため、並列実行すると他のデバイス作成テストが
// required フィールド未指定で失敗する。そのため意図的に逐次実行（resource.Test）にしています。
func TestAccCustomFieldResourceAdvanced(t *testing.T) {
	rName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_adv"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create longtext custom field with advanced options
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "longtext"
  content_types = ["dcim.rack"]
  required      = true
  group_name    = "Advanced"
  ui_visible    = "always"
  ui_editable   = "yes"
  is_cloneable  = true
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_field.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "type", "longtext"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "content_types.#", "1"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "required", "true"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "group_name", "Advanced"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "ui_visible", "always"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "ui_editable", "yes"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "is_cloneable", "true"),
					resource.TestCheckResourceAttrSet("netbox_custom_field.test", "id"),
				),
			},
			// Update group_name
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "longtext"
  content_types = ["dcim.rack"]
  required      = true
  group_name    = "Updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_field.test", "group_name", "Updated"),
					resource.TestCheckResourceAttrSet("netbox_custom_field.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
