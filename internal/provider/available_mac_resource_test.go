// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAvailableMacResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_available_mac" "test" {
  prefix      = "AA:BB:CC"
  description = "terraform test available MAC"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_available_mac.test", "mac_address", regexp.MustCompile(`(?i)^AA:BB:CC:[0-9A-F]{2}:[0-9A-F]{2}:[0-9A-F]{2}$`)),
					resource.TestCheckResourceAttr("netbox_available_mac.test", "description", "terraform test available MAC"),
					resource.TestCheckResourceAttrSet("netbox_available_mac.test", "id"),
				),
			},
			// Update description
			{
				Config: providerConfig + `
resource "netbox_available_mac" "test" {
  prefix      = "AA:BB:CC"
  description = "terraform test available MAC updated"
  comments    = "updated comment"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_available_mac.test", "description", "terraform test available MAC updated"),
					resource.TestCheckResourceAttr("netbox_available_mac.test", "comments", "updated comment"),
					resource.TestCheckResourceAttrSet("netbox_available_mac.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAvailableMacResourceCustomAttempts(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "netbox_available_mac" "test" {
  prefix       = "11:22:33"
  max_attempts = 20
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("netbox_available_mac.test", "mac_address", regexp.MustCompile(`(?i)^11:22:33:[0-9A-F]{2}:[0-9A-F]{2}:[0-9A-F]{2}$`)),
					resource.TestCheckResourceAttr("netbox_available_mac.test", "max_attempts", "20"),
					resource.TestCheckResourceAttrSet("netbox_available_mac.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// testCheckMacCount は NetBox API に直接クエリして、指定した MAC アドレスの登録件数を確認する
// カスタム TestCheckFunc を返します。
func testCheckMacCount(mac string, wantCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		serverURL := os.Getenv("NETBOX_SERVER_URL")
		token := os.Getenv("NETBOX_TOKEN")
		keyV2 := os.Getenv("NETBOX_KEY_V2")
		tokenV2 := os.Getenv("NETBOX_TOKEN_V2")

		var c *client.NetboxClient
		if token != "" {
			c = client.NewNetboxClientV1(serverURL, token)
		} else {
			c = client.NewNetboxClient(serverURL, keyV2, tokenV2)
		}

		r := &availableMacResource{client: c}
		count, err := r.macAddressCount(context.Background(), mac)
		if err != nil {
			return fmt.Errorf("macAddressCount(%q) failed: %w", mac, err)
		}
		if count != wantCount {
			return fmt.Errorf("MAC address %q: expected count %d, got %d", mac, wantCount, count)
		}
		return nil
	}
}

// TestAccMacAddressCountCheck は同じ MAC アドレスを 2 件登録した場合に
// macAddressCount が 2 を返すことを確認します。
func TestAccMacAddressCountCheck(t *testing.T) {
	const testMac = "DE:AD:BE:EF:00:01"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_mac_address" "dup1" {
  mac_address = %q
}
resource "netbox_mac_address" "dup2" {
  mac_address = %q
}
`, testMac, testMac),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_mac_address.dup1", "mac_address", testMac),
					resource.TestCheckResourceAttr("netbox_mac_address.dup2", "mac_address", testMac),
					testCheckMacCount(testMac, 2),
				),
			},
		},
	})
}

// TestAccAvailableMacSkipsDuplicate は事前に重複した MAC アドレスが登録されている状態で
// available_mac がリトライして別の MAC アドレスを取得することを確認します。
// プレフィックスを 5 オクテット（残り 1 バイト = 256 通り）にして、そのうち 1 個 (AA) を
// 重複状態（2 件）にした後、available_mac が AA 以外の MAC を取得することを確認します。
func TestAccAvailableMacSkipsDuplicate(t *testing.T) {
	const prefix = "FB:00:00:00:11"
	const dupMac = prefix + ":AA"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// dupMac を 2 件登録して重複状態を作り、available_mac がリトライすることを確認する
				Config: providerConfig + fmt.Sprintf(`
resource "netbox_mac_address" "dup1" {
  mac_address = %q
}
resource "netbox_mac_address" "dup2" {
  mac_address = %q
}
resource "netbox_available_mac" "test" {
  prefix       = %q
  max_attempts = 50
  depends_on   = [netbox_mac_address.dup1, netbox_mac_address.dup2]
}
`, dupMac, dupMac, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					// dupMac が 2 件登録されていることを確認
					testCheckMacCount(dupMac, 2),
					// available_mac が取得した MAC はプレフィックスに合致する
					resource.TestMatchResourceAttr("netbox_available_mac.test", "mac_address",
						regexp.MustCompile(`(?i)^FB:00:00:00:11:[0-9A-F]{2}$`)),
					// available_mac が取得した MAC は count == 1 （重複なし）
					testCheckMacAttrCount("netbox_available_mac.test", "mac_address", 1),
					resource.TestCheckResourceAttrSet("netbox_available_mac.test", "id"),
				),
			},
		},
	})
}

// testCheckMacAttrCount は Terraform state の指定属性から MAC アドレスを読み取り、
// NetBox 上の登録件数が wantCount であることを確認します。
func testCheckMacAttrCount(resourceName, attr string, wantCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}
		mac, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("attribute %q not found in resource %q", attr, resourceName)
		}
		return testCheckMacCount(mac, wantCount)(s)
	}
}

func TestGenerateMacAddress(t *testing.T) {
	tests := []struct {
		prefix       string
		expectedLen  int
		prefixOctets int
	}{
		{"AA:BB:CC", 17, 3},
		{"11:22", 17, 2},
		{"FF", 17, 1},
		{"AA:BB:CC:DD:EE", 17, 5},
	}

	for _, tt := range tests {
		mac, err := generateMacAddress(tt.prefix)
		if err != nil {
			t.Errorf("generateMacAddress(%q) returned error: %v", tt.prefix, err)
			continue
		}
		if len(mac) != tt.expectedLen {
			t.Errorf("generateMacAddress(%q) = %q, want length %d, got %d", tt.prefix, mac, tt.expectedLen, len(mac))
		}
		if !regexp.MustCompile(`^[0-9A-F]{2}(:[0-9A-F]{2}){5}$`).MatchString(mac) {
			t.Errorf("generateMacAddress(%q) = %q, not a valid MAC address", tt.prefix, mac)
		}
		prefixUpper := tt.prefix
		if mac[:len(prefixUpper)] != prefixUpper {
			t.Errorf("generateMacAddress(%q) = %q, does not start with prefix", tt.prefix, mac)
		}
	}
}
