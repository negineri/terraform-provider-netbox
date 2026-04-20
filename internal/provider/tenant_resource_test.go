// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTenantResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-tenant")
	rSlug := rName + "-s"
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// AutoSlug: slug 未指定時に自動生成されることを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_tenant.test", "slug"),
					resource.TestCheckResourceAttrSet("netbox_tenant.test", "id"),
				),
			},
			// Create and Read testing (with explicit slug)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name        = %q
  slug        = %q
  description = "terraform test tenant"
}
`, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_tenant.test", "slug", rSlug),
					resource.TestCheckResourceAttr("netbox_tenant.test", "description", "terraform test tenant"),
					resource.TestCheckResourceAttrSet("netbox_tenant.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_tenant.test"]
						if !ok {
							return fmt.Errorf("resource netbox_tenant.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name        = %q
  slug        = %q
  description = "terraform test tenant updated"
}
`, rName, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_tenant.test", "description", "terraform test tenant updated"),
					resource.TestCheckResourceAttrSet("netbox_tenant.test", "id"),
				),
			},
			// Rename testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_tenant" "test" {
  name = %q
  slug = %q
}
`, rNameRenamed, rSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_tenant.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_tenant.test"]
						if !ok {
							return fmt.Errorf("resource netbox_tenant.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("tenant ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
