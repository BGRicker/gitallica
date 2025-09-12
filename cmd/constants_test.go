package cmd

import (
	"testing"
)

func TestAuthorMappings(t *testing.T) {
	tests := []struct {
		testName string
		email    string
		name     string
		expected string
	}{
		{
			testName: "john email variation",
			email:    "john@example.com",
			name:     "John Mayer",
			expected: "john@rockandroll.com",
		},
		{
			testName: "mayer email variation",
			email:    "mayer@company.com",
			name:     "John Mayer",
			expected: "john@rockandroll.com",
		},
		{
			testName: "tim email variation",
			email:    "tim@example.com",
			name:     "Tim Robinson",
			expected: "tim@ithinkyoushouldleave.com",
		},
		{
			testName: "robinson email variation",
			email:    "robinson@example.com",
			name:     "Tim Robinson",
			expected: "tim@ithinkyoushouldleave.com",
		},
		{
			testName: "bo email variation",
			email:    "bo@example.com",
			name:     "Bo Jackson",
			expected: "bo@raiders.com",
		},
		{
			testName: "jackson email variation",
			email:    "jackson@example.com",
			name:     "Bo Jackson",
			expected: "bo@raiders.com",
		},
		{
			testName: "unknown author",
			email:    "unknown@example.com",
			name:     "Unknown Author",
			expected: "unknown author",
		},
		{
			testName: "generic email with name",
			email:    "noreply@example.com",
			name:     "Jane Smith",
			expected: "jane smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := normalizeAuthorName(tt.name, tt.email)
			if result != tt.expected {
				t.Errorf("normalizeAuthorName(%q, %q) = %q, want %q", tt.name, tt.email, result, tt.expected)
			}
		})
	}
}

func TestDefaultAuthorMappings(t *testing.T) {
	// Test that the default mappings are properly configured
	if len(DefaultAuthorMappings) == 0 {
		t.Error("DefaultAuthorMappings should not be empty")
	}

	// Test that each mapping has patterns and canonical email
	for i, mapping := range DefaultAuthorMappings {
		if len(mapping.Patterns) == 0 {
			t.Errorf("AuthorMapping[%d] should have patterns", i)
		}
		if mapping.Canonical == "" {
			t.Errorf("AuthorMapping[%d] should have canonical email", i)
		}
	}
}
