// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildFilterQuery(t *testing.T) {
	t.Run("empty map returns empty string", func(t *testing.T) {
		got := buildFilterQuery(map[string]string{})
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("nil map returns empty string", func(t *testing.T) {
		got := buildFilterQuery(nil)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("single param is included", func(t *testing.T) {
		got := buildFilterQuery(map[string]string{"status": "active"})
		if got != "status=active" {
			t.Errorf("expected %q, got %q", "status=active", got)
		}
	})

	t.Run("empty value is excluded", func(t *testing.T) {
		got := buildFilterQuery(map[string]string{"name": "", "status": "active"})
		if got != "status=active" {
			t.Errorf("expected %q, got %q", "status=active", got)
		}
	})

	t.Run("multiple params are all included", func(t *testing.T) {
		got := buildFilterQuery(map[string]string{"status": "active", "tag": "prod"})
		if !strings.Contains(got, "status=active") {
			t.Errorf("expected status=active in %q", got)
		}
		if !strings.Contains(got, "tag=prod") {
			t.Errorf("expected tag=prod in %q", got)
		}
	})

	t.Run("special characters are percent-encoded", func(t *testing.T) {
		got := buildFilterQuery(map[string]string{"name": "my rack"})
		if got != "name=my+rack" && got != "name=my%20rack" {
			t.Errorf("expected encoded space in %q", got)
		}
	})
}

func TestCombineQueryStrings(t *testing.T) {
	t.Run("both empty returns empty string", func(t *testing.T) {
		got := combineQueryStrings("", "")
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("one empty returns the other", func(t *testing.T) {
		got := combineQueryStrings("status=active", "")
		if got != "status=active" {
			t.Errorf("expected %q, got %q", "status=active", got)
		}
	})

	t.Run("both non-empty are joined with &", func(t *testing.T) {
		got := combineQueryStrings("status=active", "cf_tier=core")
		if got != "status=active&cf_tier=core" {
			t.Errorf("expected %q, got %q", "status=active&cf_tier=core", got)
		}
	})
}

func TestStringFilterParam(t *testing.T) {
	t.Run("non-empty value is added", func(t *testing.T) {
		params := map[string]string{}
		stringFilterParam(params, "name", types.StringValue("tokyo"))
		if params["name"] != "tokyo" {
			t.Errorf("expected tokyo, got %q", params["name"])
		}
	})

	t.Run("null value is not added", func(t *testing.T) {
		params := map[string]string{}
		stringFilterParam(params, "name", types.StringNull())
		if _, ok := params["name"]; ok {
			t.Errorf("expected key to be absent for null value")
		}
	})

	t.Run("empty string value is not added", func(t *testing.T) {
		params := map[string]string{}
		stringFilterParam(params, "name", types.StringValue(""))
		if _, ok := params["name"]; ok {
			t.Errorf("expected key to be absent for empty string value")
		}
	})
}

func TestInt64FilterParam(t *testing.T) {
	t.Run("non-zero value is added as string", func(t *testing.T) {
		params := map[string]string{}
		int64FilterParam(params, "site_id", types.Int64Value(42))
		if params["site_id"] != "42" {
			t.Errorf("expected 42, got %q", params["site_id"])
		}
	})

	t.Run("zero value is added", func(t *testing.T) {
		params := map[string]string{}
		int64FilterParam(params, "vid", types.Int64Value(0))
		if params["vid"] != "0" {
			t.Errorf("expected 0, got %q", params["vid"])
		}
	})

	t.Run("null value is not added", func(t *testing.T) {
		params := map[string]string{}
		int64FilterParam(params, "site_id", types.Int64Null())
		if _, ok := params["site_id"]; ok {
			t.Errorf("expected key to be absent for null value")
		}
	})
}
