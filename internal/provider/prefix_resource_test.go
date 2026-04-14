// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPrefixResourceWithCustomFields は custom_fields 属性の acceptance test です。
func TestAccPrefixResourceWithCustomFields(t *testing.T) {
	rCfName := acctest.RandomWithPrefix("tf_acc_cf_prefix")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["ipam.prefix"]
}

resource "netbox_prefix" "test_cf" {
  prefix = "10.99.0.0/24"
  status = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "prefix-cf-value"
  }
}
`, rCfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_prefix.test_cf", "prefix", "10.99.0.0/24"),
					resource.TestCheckResourceAttrSet("netbox_prefix.test_cf", "id"),
					resource.TestCheckResourceAttr("netbox_prefix.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "prefix-cf-value"),
				),
			},
			// カスタムフィールド値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["ipam.prefix"]
}

resource "netbox_prefix" "test_cf" {
  prefix = "10.99.0.0/24"
  status = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "prefix-cf-updated"
  }
}
`, rCfName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_prefix.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "prefix-cf-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPrefixResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_prefix" "test" {
  prefix      = "10.0.0.0/24"
  status      = "active"
  description = "terraform test prefix"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_prefix.test", "prefix", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("netbox_prefix.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_prefix.test", "description", "terraform test prefix"),
					resource.TestCheckResourceAttrSet("netbox_prefix.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_prefix" "test" {
  prefix      = "10.0.0.0/24"
  status      = "reserved"
  description = "terraform test prefix updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_prefix.test", "prefix", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("netbox_prefix.test", "status", "reserved"),
					resource.TestCheckResourceAttr("netbox_prefix.test", "description", "terraform test prefix updated"),
					resource.TestCheckResourceAttrSet("netbox_prefix.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
