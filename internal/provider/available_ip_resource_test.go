// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAvailableIPResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_available_ip" "test" {
  prefix_id = 1
  description = "terraform test IP"
  status      = "active"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_ip.test", "prefix_id", "1"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "description", "terraform test IP"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "status", "active"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "ip_address"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_available_ip" "test" {
  prefix_id = 1
  description = "terraform test IP updated"
  status      = "deprecated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_ip.test", "prefix_id", "1"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "description", "terraform test IP updated"),
					resource.TestCheckResourceAttr("netbox_available_ip.test", "status", "deprecated"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "id"),
					resource.TestCheckResourceAttrSet("netbox_available_ip.test", "ip_address"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
