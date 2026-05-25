// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrefixesDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "netbox_prefixes" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_prefixes.test", "id", "netbox_prefixes"),
					resource.TestCheckResourceAttrSet("data.netbox_prefixes.test", "prefixes.#"),
				),
			},
		},
	})
}

func TestAccPrefixesDataSourceWithPrefixFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "netbox_prefix" "test_filter" {
  prefix      = "192.168.100.0/24"
  status      = "active"
  description = "terraform test prefix for filter"
}

data "netbox_prefixes" "by_prefix" {
  prefix = "192.168.100.0/24"
  depends_on = [netbox_prefix.test_filter]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_prefixes.by_prefix", "id", "netbox_prefixes"),
					resource.TestCheckResourceAttr("data.netbox_prefixes.by_prefix", "prefixes.#", "1"),
					resource.TestCheckResourceAttr("data.netbox_prefixes.by_prefix", "prefixes.0.prefix", "192.168.100.0/24"),
				),
			},
		},
	})
}

func TestAccPrefixesDataSourceWithFamilyFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "netbox_prefixes" "ipv4" {
  family = 4
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_prefixes.ipv4", "id", "netbox_prefixes"),
					resource.TestCheckResourceAttrSet("data.netbox_prefixes.ipv4", "prefixes.#"),
				),
			},
		},
	})
}

func TestAccPrefixesDataSourceWithWithinFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "netbox_prefix" "test_within" {
  prefix      = "10.20.1.0/24"
  status      = "active"
  description = "terraform test prefix for within filter"
}

data "netbox_prefixes" "within_10_20" {
  within = "10.20.0.0/16"
  depends_on = [netbox_prefix.test_within]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_prefixes.within_10_20", "id", "netbox_prefixes"),
					resource.TestCheckResourceAttrSet("data.netbox_prefixes.within_10_20", "prefixes.#"),
				),
			},
		},
	})
}
