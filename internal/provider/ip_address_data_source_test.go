// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIpAddressDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-ipaddr-ds")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "203.0.113.10/24"
  status     = "active"
  dns_name   = "%s.example.com"
}

data "netbox_ip_addresses" "all" {
  dns_name = netbox_ip_address.test.dns_name
}

data "netbox_ip_address" "test" {
  id = data.netbox_ip_addresses.all.ip_addresses[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_ip_address.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_ip_address.test", "ip_address", "203.0.113.10/24"),
					resource.TestCheckResourceAttr("data.netbox_ip_address.test", "status", "active"),
				),
			},
		},
	})
}
