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

// TestAccVlanResource は netbox_vlan の acceptance test です。
// 実行前に NETBOX_SERVER_URL / NETBOX_KEY_V2 / NETBOX_TOKEN_V2 環境変数が必要です。
func TestAccVlanResource(t *testing.T) {
	var capturedID string
	rName := acctest.RandomWithPrefix("tf-acc-test-vlan")
	rNameRenamed := rName + "-renamed"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan" "test" {
  name        = %q
  vid         = 100
  status      = "active"
  description = "terraform test VLAN"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_vlan.test", "vid", "100"),
					resource.TestCheckResourceAttr("netbox_vlan.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_vlan.test", "description", "terraform test VLAN"),
					resource.TestCheckResourceAttrSet("netbox_vlan.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_vlan.test"]
						if !ok {
							return fmt.Errorf("resource netbox_vlan.test not found")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan" "test" {
  name        = %q
  vid         = 100
  status      = "reserved"
  description = "terraform test VLAN updated"
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan.test", "name", rName),
					resource.TestCheckResourceAttr("netbox_vlan.test", "vid", "100"),
					resource.TestCheckResourceAttr("netbox_vlan.test", "status", "reserved"),
					resource.TestCheckResourceAttr("netbox_vlan.test", "description", "terraform test VLAN updated"),
					resource.TestCheckResourceAttrSet("netbox_vlan.test", "id"),
				),
			},
			// Rename testing: 名前を変更してもIDが変化しないことを確認する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan" "test" {
  name   = %q
  vid    = 100
  status = "reserved"
}
`, rNameRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan.test", "name", rNameRenamed),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["netbox_vlan.test"]
						if !ok {
							return fmt.Errorf("resource netbox_vlan.test not found")
						}
						if rs.Primary.ID != capturedID {
							return fmt.Errorf("vlan ID changed after rename: was %s, now %s", capturedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccVlanResourceWithCustomFields は custom_fields 属性の acceptance test です。
func TestAccVlanResourceWithCustomFields(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-vlan-cf")
	rCfName := strings.ReplaceAll(acctest.RandomWithPrefix("tf_acc_cf_vlan"), "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["ipam.vlan"]
}

resource "netbox_vlan" "test_cf" {
  name   = %q
  vid    = 300
  status = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "vlan-cf-value"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan.test_cf", "name", rName),
					resource.TestCheckResourceAttrSet("netbox_vlan.test_cf", "id"),
					resource.TestCheckResourceAttr("netbox_vlan.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "vlan-cf-value"),
				),
			},
			// カスタムフィールド値を更新する
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_custom_field" "test" {
  name          = %q
  type          = "text"
  content_types = ["ipam.vlan"]
}

resource "netbox_vlan" "test_cf" {
  name   = %q
  vid    = 300
  status = "active"

  custom_fields = {
    (netbox_custom_field.test.name) = "vlan-cf-updated"
  }
}
`, rCfName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan.test_cf", fmt.Sprintf("custom_fields.%s", rCfName), "vlan-cf-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccVlanResourceWithGroup は group_id を使った acceptance test です。
// VLAN group を作成してその配下に VLAN を所属させ、その後 group から切り離す。
func TestAccVlanResourceWithGroup(t *testing.T) {
	rGroupName := acctest.RandomWithPrefix("tf-acc-test-vlan-group")
	rVlanName := acctest.RandomWithPrefix("tf-acc-test-vlan")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// VLAN group 配下に VLAN を作成
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan_group" "test" {
  name = %q
}

resource "netbox_vlan" "test" {
  name     = %q
  vid      = 200
  status   = "active"
  group_id = netbox_vlan_group.test.id
}
`, rGroupName, rVlanName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan.test", "name", rVlanName),
					resource.TestCheckResourceAttr("netbox_vlan.test", "vid", "200"),
					resource.TestCheckResourceAttr("netbox_vlan.test", "status", "active"),
					resource.TestCheckResourceAttrSet("netbox_vlan.test", "group_id"),
					resource.TestCheckResourceAttrSet("netbox_vlan.test", "id"),
				),
			},
			// VLAN group から切り離す
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_vlan_group" "test" {
  name = %q
}

resource "netbox_vlan" "test" {
  name   = %q
  vid    = 200
  status = "active"
}
`, rGroupName, rVlanName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_vlan.test", "name", rVlanName),
					resource.TestCheckNoResourceAttr("netbox_vlan.test", "group_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
