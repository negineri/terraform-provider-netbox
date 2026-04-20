// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTenantDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-tenant")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name        = %q
  description = "test tenant for data source"
}

data "netbox_tenant" "test" {
  id = netbox_tenant.test.id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_tenant.test", "id"),
					resource.TestCheckResourceAttr("data.netbox_tenant.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.netbox_tenant.test", "slug"),
					resource.TestCheckResourceAttr("data.netbox_tenant.test", "description", "test tenant for data source"),
				),
			},
		},
	})
}

func TestAccTenantsDataSource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-tenant")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = %q
}

data "netbox_tenants" "all" {
  depends_on = [netbox_tenant.test]
}

data "netbox_tenant" "test" {
  id = data.netbox_tenants.all.tenants[0].id
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netbox_tenant.test", "id"),
					resource.TestCheckResourceAttrSet("data.netbox_tenant.test", "name"),
					resource.TestCheckResourceAttrSet("data.netbox_tenant.test", "slug"),
				),
			},
		},
	})
}
