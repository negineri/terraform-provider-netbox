package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIpAddressResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_ip_address" "test" {
  ip_address  = "192.168.100.1/24"
  status      = "active"
  description = "terraform test IP address"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", "192.168.100.1/24"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", "terraform test IP address"),
					resource.TestCheckResourceAttrSet("netbox_ip_address.test", "id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "netbox_ip_address" "test" {
  ip_address  = "192.168.100.1/24"
  status      = "deprecated"
  description = "terraform test IP address updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", "192.168.100.1/24"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "deprecated"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", "terraform test IP address updated"),
					resource.TestCheckResourceAttrSet("netbox_ip_address.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
