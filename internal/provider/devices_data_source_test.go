// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDevicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "netbox_devices" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of coffees returned
					resource.TestCheckResourceAttr("data.netbox_devices.test", "devices.#", "12"),
					// Verify the first coffee to ensure all attributes are set
					resource.TestCheckResourceAttr("data.netbox_devices.test", "devices.0.id", "141"),
					resource.TestCheckResourceAttr("data.netbox_devices.test", "devices.0.name", "AUSYD01-SW-1"),
				),
			},
		},
	})
}
