// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	reNonAlphanumeric = regexp.MustCompile(`[^a-z0-9_]+`)
	reLeadingTrailing = regexp.MustCompile(`^-+|-+$`)
)

// slugify converts a name string into a URL-friendly slug.
// e.g. "Example Site" -> "example-site".
func slugify(name string) string {
	// Normalize unicode characters (decompose accented chars)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, name)

	lower := strings.ToLower(normalized)
	slug := reNonAlphanumeric.ReplaceAllString(lower, "-")
	slug = reLeadingTrailing.ReplaceAllString(slug, "")
	return slug
}
