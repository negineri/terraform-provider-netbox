// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAvailablePrefixResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_prefix" "parent" {
  prefix      = "10.0.0.0/24"
  status      = "active"
  description = "terraform test parent prefix"
}

resource "netbox_available_prefix" "test" {
  parent_prefix_id = netbox_prefix.parent.id
  prefix_length    = 25
  status           = "active"
  description      = "terraform test available prefix"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_prefix.test", "prefix_length", "25"),
					resource.TestCheckResourceAttr("netbox_available_prefix.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_available_prefix.test", "description", "terraform test available prefix"),
					resource.TestCheckResourceAttrSet("netbox_available_prefix.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_prefix.test", "prefix"),
					resource.TestCheckResourceAttrSet("netbox_available_prefix.test", "parent_prefix_id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_prefix" "parent" {
  prefix      = "10.0.0.0/24"
  status      = "active"
  description = "terraform test parent prefix"
}

resource "netbox_available_prefix" "test" {
  parent_prefix_id = netbox_prefix.parent.id
  prefix_length    = 25
  status           = "reserved"
  description      = "terraform test available prefix updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_prefix.test", "prefix_length", "25"),
					resource.TestCheckResourceAttr("netbox_available_prefix.test", "status", "reserved"),
					resource.TestCheckResourceAttr("netbox_available_prefix.test", "description", "terraform test available prefix updated"),
					resource.TestCheckResourceAttrSet("netbox_available_prefix.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_prefix.test", "prefix"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
