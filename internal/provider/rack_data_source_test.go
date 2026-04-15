// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccRackDataSource は netbox_rack data source の acceptance test です。
// netbox_rack リソースで作成したラックを参照して検証します。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に site_id=1 が存在している必要があります。
func TestAccRackDataSource(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-rack")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_rack" "test" {
  name        = %q
  site_id     = netbox_site.test.id
  status      = "active"
  description = "terraform test rack"
  u_height    = 42
}

data "netbox_rack" "test" {
  id = netbox_rack.test.id
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.netbox_rack.test", "id", "netbox_rack.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_rack.test", "name", rName),
					resource.TestCheckResourceAttr("data.netbox_rack.test", "status", "active"),
					resource.TestCheckResourceAttr("data.netbox_rack.test", "description", "terraform test rack"),
					resource.TestCheckResourceAttr("data.netbox_rack.test", "u_height", "42"),
					resource.TestCheckResourceAttrSet("data.netbox_rack.test", "site_id"),
				),
			},
		},
	})
}

// TestAccRacksDataSource は netbox_racks data source の acceptance test です。
func TestAccRacksDataSource(t *testing.T) {
	rSiteName := acctest.RandomWithPrefix("tf-acc-test-site")
	rName := acctest.RandomWithPrefix("tf-acc-test-rack")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_site" "test" {
  name   = %q
  status = "active"
}

resource "netbox_rack" "test" {
  name     = %q
  site_id  = netbox_site.test.id
  status   = "active"
  u_height = 42
}

data "netbox_racks" "test" {
  depends_on = [netbox_rack.test]
}
`, rSiteName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_racks.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_racks.test", "racks.#"),
				),
			},
		},
	})
}
