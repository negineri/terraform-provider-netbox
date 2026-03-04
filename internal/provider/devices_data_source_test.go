// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strconv"
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
					// Verify number of devices returned is at least 12
					resource.TestCheckResourceAttrWith("data.netbox_devices.test", "devices.#", func(value string) error {
						if valueInt, err := strconv.Atoi(value); err != nil {
							return err
						} else if valueInt < 12 {
							return fmt.Errorf("expected at least 12 devices, got %d", valueInt)
						}
						return nil
					}),
					// Verify the first device to ensure all attributes are set
					resource.TestCheckResourceAttr("data.netbox_devices.test", "devices.0.id", "141"),
					resource.TestCheckResourceAttr("data.netbox_devices.test", "devices.0.name", "AUSYD01-SW-1"),
				),
			},
		},
	})
}
