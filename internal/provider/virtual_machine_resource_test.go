// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccVirtualMachineResource は acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に cluster_id=1 が存在している必要があります。
func TestAccVirtualMachineResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-vm")
	rNameRenamed := rName + "-renamed"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name        = %q
  cluster_id  = 1
  status      = "active"
  vcpus       = 2
  memory      = 4096
  disk        = 50
  description = "terraform test virtual machine"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "cluster_id", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "vcpus", "2"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "memory", "4096"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "disk", "50"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "description", "terraform test virtual machine"),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_virtual_machine.test"]
						if !ok {
							return fmt.Errorf("resource netbox_virtual_machine.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name        = %q
  cluster_id  = 1
  status      = "planned"
  vcpus       = 4
  memory      = 8192
  disk        = 100
  description = "terraform test virtual machine updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "cluster_id", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "status", "planned"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "vcpus", "4"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "memory", "8192"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "disk", "100"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "description", "terraform test virtual machine updated"),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine.test", "id"),
				),
			},
			// Rename testing: VM 名を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test" {
  name        = %q
  cluster_id  = 1
  status      = "planned"
  vcpus       = 4
  memory      = 8192
  disk        = 100
  description = "terraform test virtual machine updated"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_virtual_machine.test"]
						if !ok {
							return fmt.Errorf("resource netbox_virtual_machine.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("virtual machine ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccVirtualMachineResourceWithCustomFields は custom_fields 属性の acceptance test です。
// 実行前に NetBox 上に cluster_id=1 が存在している必要があります。
func TestAccVirtualMachineResourceWithCustomFields(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_vm"), "-", "_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["virtualization.virtualmachine"]
}

resource "netbox_virtual_machine" "test_cf" {
  name       = %q
  cluster_id = 1
  status     = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "vm-cf-value"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_cf", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine.test_cf", "id"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "vm-cf-value"),
				),
			},
			// カスタムフィールド値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["virtualization.virtualmachine"]
}

resource "netbox_virtual_machine" "test_cf" {
  name       = %q
  cluster_id = 1
  status     = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "vm-cf-updated"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "vm-cf-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccVirtualMachineResourceWithoutCluster は cluster_id を指定せず site_id のみで作成する acceptance test です。
// 実行前に NetBox 上に site_id=1 が存在している必要があります。
func TestAccVirtualMachineResourceWithoutCluster(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-nocluster")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with site_id only (no cluster_id)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test_nocluster" {
  name    = %q
  site_id = 1
  status  = "active"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_nocluster", "name", rName),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_nocluster", "site_id", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_nocluster", "status", "active"),
					resource.TestCheckNoResourceAttr("netbox_virtual_machine.test_nocluster", "cluster_id"),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine.test_nocluster", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccVirtualMachineResourceWithTags は tags 属性の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数と、
// NetBox 上に cluster_id=1, tag_id=37 が存在している必要があります。
func TestAccVirtualMachineResourceWithTags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vm-tags")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with tags
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test_tags" {
  name       = %q
  cluster_id = 1
  status     = "active"
  tags       = [37]
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_tags", "name", rName),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_tags", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_tags", "tags.0", "37"),
					resource.TestCheckResourceAttrSet("netbox_virtual_machine.test_tags", "id"),
				),
			},
			// Update tags
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_virtual_machine" "test_tags" {
  name       = %q
  cluster_id = 1
  status     = "active"
  tags       = []
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_tags", "name", rName),
					resource.TestCheckResourceAttr("netbox_virtual_machine.test_tags", "tags.#", "0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
