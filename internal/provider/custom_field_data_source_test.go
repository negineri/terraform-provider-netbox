// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccCustomFieldDataSource は netbox_custom_field データソースの acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccCustomFieldDataSource(t *testing.T) {
	rName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_ds"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  label         = "TF Acc Data Source Test"
  type          = "text"
  content_types = ["dcim.device"]
  description   = "terraform acceptance test data source"
  weight        = 150
}

data "netbox_custom_field" "test" {
  id = netbox_custom_field.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_custom_field.test", "name", rName),
					resource.TestCheckResourceAttr("data.netbox_custom_field.test", "label", "TF Acc Data Source Test"),
					resource.TestCheckResourceAttr("data.netbox_custom_field.test", "type", "text"),
					resource.TestCheckResourceAttr("data.netbox_custom_field.test", "content_types.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_custom_field.test", "content_types.0", "dcim.device"),
					resource.TestCheckResourceAttr("data.netbox_custom_field.test", "description", "terraform acceptance test data source"),
					resource.TestCheckResourceAttr("data.netbox_custom_field.test", "weight", "150"),
					resource.TestCheckResourceAttrSet("data.netbox_custom_field.test", "id"),
				),
			},
		},
	})
}
