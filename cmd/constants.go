/*
Copyright © 2025 Ben Ricker

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import "time"

// Configuration constants for Gitallica analysis
const (
	// DefaultFallbackFileAge is the assumed age for files that exist but weren't
	// modified in the analysis window. This avoids expensive individual git history
	// lookups for every file while providing a reasonable default.
	DefaultFallbackFileAge = 2 * 365 * 24 * time.Hour // 2 years
)

// AuthorMapping represents a mapping from various email/name patterns to a canonical author
type AuthorMapping struct {
	Patterns []string // Email or name patterns to match
	Canonical string  // Canonical email address to use
}

// DefaultAuthorMappings contains the default author normalization mappings.
// These can be overridden in configuration files.
var DefaultAuthorMappings = []AuthorMapping{
	{
		Patterns:  []string{"john", "mayer"},
		Canonical: "john@rockandroll.com",
	},
	{
		Patterns:  []string{"tim", "robinson"},
		Canonical: "tim@ithinkyoushouldleave.com",
	},
	{
		Patterns:  []string{"bo", "jackson"},
		Canonical: "bo@raiders.com",
	},
}
