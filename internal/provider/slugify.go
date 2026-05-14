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
	reNonSlugChar       = regexp.MustCompile(`[^a-z0-9_-]+`)
	reConsecutiveHyphen = regexp.MustCompile(`-{2,}`)
	reLeadingHyphen     = regexp.MustCompile(`^-+`)
	reTrailingHyphen    = regexp.MustCompile(`-+$`)
)

// slugify converts a name string into a URL-friendly slug.
// e.g. "Example Site" -> "example-site".
func slugify(name string) string {
	name = strings.TrimSpace(name)

	hasLeadingHyphen := strings.HasPrefix(name, "-")
	hasTrailingHyphen := strings.HasSuffix(name, "-")

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, name)

	lower := strings.ToLower(normalized)
	slug := reNonSlugChar.ReplaceAllString(lower, "-")
	slug = reConsecutiveHyphen.ReplaceAllString(slug, "-")
	slug = reLeadingHyphen.ReplaceAllString(slug, "")
	slug = reTrailingHyphen.ReplaceAllString(slug, "")

	if hasLeadingHyphen {
		slug = "-" + slug
	}
	if hasTrailingHyphen {
		slug = slug + "-"
	}

	return slug
}
