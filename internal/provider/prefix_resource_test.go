package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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
