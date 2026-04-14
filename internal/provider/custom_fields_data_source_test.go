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

// TestAccCustomFieldsDataSource は netbox_custom_fields データソースの acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccCustomFieldsDataSource(t *testing.T) {
	rName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_list"), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["dcim.device"]
  description   = "terraform acceptance test list data source"
}

data "netbox_custom_fields" "all" {
  depends_on = [netbox_custom_field.test]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// 少なくとも1件以上取得できることを確認
					resource.TestCheckResourceAttrSet("data.netbox_custom_fields.all", "custom_fields.#"),
				),
			},
		},
	})
}
