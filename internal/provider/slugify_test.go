// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple lowercase",
			input: "example",
			want:  "example",
		},
		{
			name:  "space separated words",
			input: "Example Site",
			want:  "example-site",
		},
		{
			name:  "multiple spaces",
			input: "Example  VLAN  Group",
			want:  "example-vlan-group",
		},
		{
			name:  "leading and trailing spaces",
			input: "  test  ",
			want:  "test",
		},
		{
			name:  "special characters",
			input: "My Site #1",
			want:  "my-site-1",
		},
		{
			name:  "slash",
			input: "Rack A/B",
			want:  "rack-a-b",
		},
		{
			name:  "already slug-like",
			input: "example-site",
			want:  "example-site",
		},
		{
			name:  "unicode accented characters",
			input: "Café",
			want:  "cafe",
		},
		{
			name:  "japanese characters become empty segments",
			input: "東京 Site",
			want:  "site",
		},
		{
			name:  "numbers",
			input: "VLAN 100",
			want:  "vlan-100",
		},
		{
			name:  "consecutive hyphens from special chars",
			input: "test---value",
			want:  "test-value",
		},
		{
			name:  "mixed case",
			input: "Tokyo-DC01",
			want:  "tokyo-dc01",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := slugify(tc.input)
			if got != tc.want {
				t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
