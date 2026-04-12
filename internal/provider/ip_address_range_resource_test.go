// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestParseIPRange(t *testing.T) {
	tests := []struct {
		input   string
		wantIPs []string
		wantErr bool
	}{
		{
			input:   "10.18.48.[224-226]/24",
			wantIPs: []string{"10.18.48.224/24", "10.18.48.225/24", "10.18.48.226/24"},
		},
		{
			input:   "192.168.1.[1-1]/24",
			wantIPs: []string{"192.168.1.1/24"},
		},
		{
			input:   "10.0.0.[0-255]/8",
			wantIPs: nil,
		},
		{
			input:   "invalid",
			wantErr: true,
		},
		{
			input:   "10.18.48.[239-224]/24",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		ips, err := parseIPRange(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseIPRange(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseIPRange(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if tt.wantIPs != nil {
			if len(ips) != len(tt.wantIPs) {
				t.Errorf("parseIPRange(%q) got %d IPs, want %d", tt.input, len(ips), len(tt.wantIPs))
				continue
			}
			for i, ip := range ips {
				if ip != tt.wantIPs[i] {
					t.Errorf("parseIPRange(%q)[%d] = %q, want %q", tt.input, i, ip, tt.wantIPs[i])
				}
			}
		}
	}
}

func TestAccIpAddressRangeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "netbox_ip_address_range" "test" {
  ip_range    = "192.168.150.[10-12]/24"
  status      = "dhcp"
  description = "terraform test IP range"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "ip_range", "192.168.150.[10-12]/24"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "status", "dhcp"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "description", "terraform test IP range"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "allocated_ips.#", "3"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "allocated_ips.0.ip_address", "192.168.150.10/24"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "allocated_ips.1.ip_address", "192.168.150.11/24"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "allocated_ips.2.ip_address", "192.168.150.12/24"),
					resource.TestCheckResourceAttrSet("netbox_ip_address_range.test", "allocated_ips.0.id"),
					resource.TestCheckResourceAttrSet("netbox_ip_address_range.test", "allocated_ips.1.id"),
					resource.TestCheckResourceAttrSet("netbox_ip_address_range.test", "allocated_ips.2.id"),
				),
			},
			// Update status and description
			{
				Config: providerConfig + `
resource "netbox_ip_address_range" "test" {
  ip_range    = "192.168.150.[10-12]/24"
  status      = "active"
  description = "terraform test IP range updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "description", "terraform test IP range updated"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "allocated_ips.#", "3"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
