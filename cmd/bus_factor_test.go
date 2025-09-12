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

import (
	"testing"
)

func TestCalculateBusFactor(t *testing.T) {
	tests := []struct {
		name               string
		authorLines         map[string]int
		expectedBusFactor   int
		expectedRiskLevel   string
		description         string
	}{
		{
			name: "single author - high risk",
			authorLines: map[string]int{
				"alice@example.com": 100,
			},
			expectedBusFactor: 1,
			expectedRiskLevel: "Critical",
			description:       "Single author creates critical bus factor risk",
		},
		{
			name: "two authors balanced - medium risk",
			authorLines: map[string]int{
				"alice@example.com": 50,
				"bob@example.com":   50,
			},
			expectedBusFactor: 2,
			expectedRiskLevel: "Medium",
			description:       "Two balanced authors still risky for small teams",
		},
		{
			name: "three authors well distributed - low risk",
			authorLines: map[string]int{
				"alice@example.com":   40,
				"bob@example.com":     35,
				"charlie@example.com": 25,
			},
			expectedBusFactor: 2,
			expectedRiskLevel: "Medium",
			description:       "Well distributed among 3 authors is healthier",
		},
		{
			name: "one dominant author among many - high risk",
			authorLines: map[string]int{
				"alice@example.com":   80,
				"bob@example.com":     10,
				"charlie@example.com": 5,
				"david@example.com":   3,
				"eve@example.com":     2,
			},
			expectedBusFactor: 1,
			expectedRiskLevel: "Critical",
			description:       "Dominant author creates critical risk despite multiple contributors",
		},
		{
			name: "five balanced authors - healthy",
			authorLines: map[string]int{
				"alice@example.com":   25,
				"bob@example.com":     22,
				"charlie@example.com": 20,
				"david@example.com":   18,
				"eve@example.com":     15,
			},
			expectedBusFactor: 3,
			expectedRiskLevel: "Medium",
			description:       "Multiple balanced contributors create reasonable bus factor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			busFactor := calculateBusFactor(tt.authorLines)
			if busFactor != tt.expectedBusFactor {
				t.Errorf("Expected bus factor %d, got %d. %s", tt.expectedBusFactor, busFactor, tt.description)
			}
			
			riskLevel := classifyBusFactorRisk(busFactor, len(tt.authorLines))
			if riskLevel != tt.expectedRiskLevel {
				t.Errorf("Expected risk level %s, got %s. %s", tt.expectedRiskLevel, riskLevel, tt.description)
			}
		})
	}
}

func TestClassifyBusFactorRisk(t *testing.T) {
	tests := []struct {
		name            string
		busFactor       int
		totalContributors int
		expectedRisk    string
		description     string
	}{
		{
			name:            "single person team",
			busFactor:       1,
			totalContributors: 1,
			expectedRisk:    "Critical",
			description:     "Single person creates critical risk",
		},
		{
			name:            "two person team, bus factor 1",
			busFactor:       1,
			totalContributors: 2,
			expectedRisk:    "Critical",
			description:     "One person dominating in small team is critical",
		},
		{
			name:            "small team with good distribution",
			busFactor:       3,
			totalContributors: 4,
			expectedRisk:    "Medium",
			description:     "Good distribution in small team",
		},
		{
			name:            "large team with good distribution",
			busFactor:       6,
			totalContributors: 10,
			expectedRisk:    "Healthy",
			description:     "Large team with good knowledge distribution",
		},
		{
			name:            "medium team with concentration",
			busFactor:       2,
			totalContributors: 8,
			expectedRisk:    "High",
			description:     "Knowledge concentration in medium-sized team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := classifyBusFactorRisk(tt.busFactor, tt.totalContributors)
			if risk != tt.expectedRisk {
				t.Errorf("Expected risk %s, got %s. %s", tt.expectedRisk, risk, tt.description)
			}
		})
	}
}

func TestCalculateAuthorContributionPercentage(t *testing.T) {
	authorLines := map[string]int{
		"alice@example.com": 60,
		"bob@example.com":   30,
		"charlie@example.com": 10,
	}
	
	percentages := calculateAuthorContributionPercentage(authorLines)
	
	expectedPercentages := map[string]float64{
		"alice@example.com": 60.0,
		"bob@example.com":   30.0,
		"charlie@example.com": 10.0,
	}
	
	for author, expected := range expectedPercentages {
		if actual, exists := percentages[author]; !exists {
			t.Errorf("Expected percentage for %s not found", author)
		} else if actual != expected {
			t.Errorf("Expected %s to have %.1f%%, got %.1f%%", author, expected, actual)
		}
	}
}

func TestSortDirectoriesByBusFactorRisk(t *testing.T) {
	dirs := []DirectoryBusFactorStats{
		{Path: "healthy/", BusFactor: 5, RiskLevel: "Healthy"},
		{Path: "critical/", BusFactor: 1, RiskLevel: "Critical"},
		{Path: "medium/", BusFactor: 3, RiskLevel: "Medium"},
		{Path: "high/", BusFactor: 2, RiskLevel: "High"},
	}

	sorted := sortDirectoriesByBusFactorRisk(dirs)

	// Should be sorted by risk (Critical -> High -> Medium -> Healthy)
	expectedOrder := []string{"critical/", "high/", "medium/", "healthy/"}
	
	for i, expected := range expectedOrder {
		if sorted[i].Path != expected {
			t.Errorf("Expected position %d to be %s, got %s", i, expected, sorted[i].Path)
		}
	}
}

func TestBusFactorEdgeCases(t *testing.T) {
	t.Run("empty author lines", func(t *testing.T) {
		busFactor := calculateBusFactor(map[string]int{})
		if busFactor != 0 {
			t.Errorf("Expected bus factor 0 for empty lines, got %d", busFactor)
		}
	})
	
	t.Run("zero total contributors", func(t *testing.T) {
		risk := classifyBusFactorRisk(0, 0)
		if risk != "Unknown" {
			t.Errorf("Expected 'Unknown' risk for zero contributors, got %s", risk)
		}
	})
	
	t.Run("contribution percentage with zero lines", func(t *testing.T) {
		percentages := calculateAuthorContributionPercentage(map[string]int{})
		if len(percentages) != 0 {
			t.Errorf("Expected empty percentages for zero lines, got %d entries", len(percentages))
		}
	})
}

func TestBusFactorThresholds(t *testing.T) {
	// Test that our constants match expected research-backed values
	if criticalBusFactorThreshold != 1 {
		t.Errorf("Expected critical bus factor threshold to be 1, got %d", criticalBusFactorThreshold)
	}
	
	if lowBusFactorThreshold != 2 {
		t.Errorf("Expected low bus factor threshold to be 2, got %d", lowBusFactorThreshold)
	}
}

func TestAuthorNameNormalization(t *testing.T) {
	tests := []struct {
		name     string
		nameInput    string
		emailInput   string
		expected string
	}{
		{
			name:     "valid email and name",
			nameInput:    "Alice Smith",
			emailInput:   "alice@company.com",
			expected: "alice@company.com",
		},
		{
			name:     "generic email with name",
			nameInput:    "Alice Smith",
			emailInput:   "user@localhost",
			expected: "alice smith",
		},
		{
			name:     "example.com email with name",
			nameInput:    "Alice Smith",
			emailInput:   "alice@example.com",
			expected: "alice smith",
		},
		{
			name:     "noreply email with name",
			nameInput:    "Alice Smith",
			emailInput:   "noreply@company.com",
			expected: "alice smith",
		},
		{
			name:     "valid email no name",
			nameInput:    "",
			emailInput:   "alice@company.com",
			expected: "alice@company.com",
		},
		{
			name:     "name only no email",
			nameInput:    "Alice Smith",
			emailInput:   "",
			expected: "alice smith",
		},
		{
			name:     "mixed case email",
			nameInput:    "Alice Smith",
			emailInput:   "Alice.Smith@COMPANY.COM",
			expected: "alice.smith@company.com",
		},
		{
			name:     "empty inputs",
			nameInput:    "",
			emailInput:   "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeAuthorName(tt.nameInput, tt.emailInput)
			if result != tt.expected {
				t.Errorf("normalizeAuthorName(%q, %q) = %q, want %q", tt.nameInput, tt.emailInput, result, tt.expected)
			}
		})
	}
}

func TestIsGenericEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "localhost email",
			email:    "user@localhost",
			expected: true,
		},
		{
			name:     "example.com email",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "noreply email",
			email:    "noreply@company.com",
			expected: true,
		},
		{
			name:     "valid company email",
			email:    "alice@company.com",
			expected: false,
		},
		{
			name:     "gmail email",
			email:    "alice@gmail.com",
			expected: false,
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGenericEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isGenericEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestMatchesPathFilter(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		filter   string
		expected bool
	}{
		{
			name:     "empty filter matches all",
			filePath: "src/main.go",
			filter:   "",
			expected: true,
		},
		{
			name:     "exact directory match",
			filePath: "src/main.go",
			filter:   "src",
			expected: true,
		},
		{
			name:     "subdirectory match",
			filePath: "src/utils/helper.go",
			filter:   "src",
			expected: true,
		},
		{
			name:     "no match",
			filePath: "tests/test.go",
			filter:   "src",
			expected: false,
		},
		{
			name:     "partial name no match",
			filePath: "testing/test.go",
			filter:   "test",
			expected: false,
		},
		{
			name:     "exact file match",
			filePath: "main.go",
			filter:   "main.go",
			expected: true,
		},
		{
			name:     "Windows-style paths",
			filePath: "src\\main.go",
			filter:   "src",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSinglePathFilter(tt.filePath, tt.filter)
			if result != tt.expected {
				t.Errorf("matchesSinglePathFilter(%q, %q) = %v, want %v", tt.filePath, tt.filter, result, tt.expected)
			}
		})
	}
}
