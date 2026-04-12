// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestParseIPRange(t *testing.T) {
	tests := []struct {
		input     string
		wantStart string
		wantEnd   string
		wantErr   bool
	}{
		{
			input:     "10.18.48.[224-239]/24",
			wantStart: "10.18.48.224/24",
			wantEnd:   "10.18.48.239/24",
		},
		{
			input:     "192.168.1.[1-1]/24",
			wantStart: "192.168.1.1/24",
			wantEnd:   "192.168.1.1/24",
		},
		{
			input:     "10.0.0.[0-255]/8",
			wantStart: "10.0.0.0/8",
			wantEnd:   "10.0.0.255/8",
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
		start, end, err := parseIPRange(tt.input)
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
		if start != tt.wantStart {
			t.Errorf("parseIPRange(%q) start = %q, want %q", tt.input, start, tt.wantStart)
		}
		if end != tt.wantEnd {
			t.Errorf("parseIPRange(%q) end = %q, want %q", tt.input, end, tt.wantEnd)
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
  ip_range    = "192.168.150.[10-20]/24"
  status      = "active"
  description = "terraform test IP range"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "ip_range", "192.168.150.[10-20]/24"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "description", "terraform test IP range"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "start_address", "192.168.150.10/24"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "end_address", "192.168.150.20/24"),
					resource.TestCheckResourceAttrSet("netbox_ip_address_range.test", "id"),
				),
			},
			// Update status and description
			{
				Config: providerConfig + `
resource "netbox_ip_address_range" "test" {
  ip_range    = "192.168.150.[10-20]/24"
  status      = "reserved"
  description = "terraform test IP range updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "status", "reserved"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "description", "terraform test IP range updated"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "start_address", "192.168.150.10/24"),
					resource.TestCheckResourceAttr("netbox_ip_address_range.test", "end_address", "192.168.150.20/24"),
					resource.TestCheckResourceAttrSet("netbox_ip_address_range.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
