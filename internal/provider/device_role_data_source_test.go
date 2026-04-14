// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceRoleDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-role")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name        = %q
  color       = "aa1409"
  description = "test device role for data source"
}

data "netbox_device_role" "test" {
  id = netbox_device_role.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_device_role.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_device_role.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_device_role.test", "slug"),
					resource.TestCheckResourceAttr("data.netbox_device_role.test", "color", "aa1409"),
					resource.TestCheckResourceAttr("data.netbox_device_role.test", "description", "test device role for data source"),
				),
			},
		},
	})
}

func TestAccDeviceRolesDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-device-role")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_device_role" "test" {
  name  = %q
  color = "aa1409"
}

data "netbox_device_roles" "all" {
  depends_on = [netbox_device_role.test]
}

data "netbox_device_role" "test" {
  id = data.netbox_device_roles.all.device_roles[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_device_role.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_device_role.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_device_role.test", "slug"),
					resource.TestCheckResourceAttrSet("data.netbox_device_role.test", "color"),
				),
			},
		},
	})
}
