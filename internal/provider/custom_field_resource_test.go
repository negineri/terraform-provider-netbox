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

// TestAccCustomFieldResource は netbox_custom_field の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccCustomFieldResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf_acc_cf")

	resource.Test(t, resource.TestCase{
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

// TestAccCustomFieldResourceSelect は select 型カスタムフィールドの acceptance test です。
func TestAccCustomFieldResourceSelect(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf_acc_cf_sel")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create select type custom field
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "select"
  content_types = ["dcim.device", "virtualization.virtualmachine"]
  choices       = ["production", "staging", "development"]
  default       = "development"
  required      = true
  group_name    = "Deployment"
  ui_visible    = "always"
  ui_editable   = "yes"
  is_cloneable  = true
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_field.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "type", "select"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "content_types.#", "2"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "choices.#", "3"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "required", "true"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "group_name", "Deployment"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "ui_visible", "always"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "ui_editable", "yes"),
					resource.TestCheckResourceAttr("netbox_custom_field.test", "is_cloneable", "true"),
					resource.TestCheckResourceAttrSet("netbox_custom_field.test", "id"),
				),
			},
			// Update choices
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "select"
  content_types = ["dcim.device", "virtualization.virtualmachine"]
  choices       = ["production", "staging", "development", "testing"]
  required      = true
  group_name    = "Deployment"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_custom_field.test", "choices.#", "4"),
					resource.TestCheckResourceAttrSet("netbox_custom_field.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
