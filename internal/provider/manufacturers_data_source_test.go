// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccManufacturersDataSource は netbox_manufacturers data source の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccManufacturersDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_manufacturers" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_manufacturers.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_manufacturers.test", "manufacturers.#"),
				),
			},
		},
	})
}
