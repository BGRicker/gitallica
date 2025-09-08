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

// TestCalculateOwnershipClarity tests ownership clarity calculation logic
func TestCalculateOwnershipClarity(t *testing.T) {
	tests := []struct {
		name                   string
		commitsByContributor   map[string]int
		totalCommits          int
		expectedTopOwnership  float64
		expectedStatus        string
		expectedContributors  int
	}{
		{
			name: "clear_ownership_60_percent",
			commitsByContributor: map[string]int{
				"alice@example.com": 12,
				"bob@example.com":   5,
				"carol@example.com": 3,
			},
			totalCommits:         20,
			expectedTopOwnership: 0.60,
			expectedStatus:       "Healthy",
			expectedContributors: 3,
		},
		{
			name: "borderline_ownership_50_percent",
			commitsByContributor: map[string]int{
				"alice@example.com": 10,
				"bob@example.com":   6,
				"carol@example.com": 4,
			},
			totalCommits:         20,
			expectedTopOwnership: 0.50,
			expectedStatus:       "Healthy",
			expectedContributors: 3,
		},
		{
			name: "diffuse_ownership_no_clear_owner",
			commitsByContributor: map[string]int{
				"alice@example.com":   4,
				"bob@example.com":     4,
				"carol@example.com":   4,
				"david@example.com":   4,
				"eve@example.com":     4,
			},
			totalCommits:         20,
			expectedTopOwnership: 0.20,
			expectedStatus:       "Warning",
			expectedContributors: 5,
		},
		{
			name: "extremely_diffuse_ownership_many_contributors",
			commitsByContributor: map[string]int{
				"contributor1@example.com":  2,
				"contributor2@example.com":  2,
				"contributor3@example.com":  2,
				"contributor4@example.com":  2,
				"contributor5@example.com":  2,
				"contributor6@example.com":  2,
				"contributor7@example.com":  2,
				"contributor8@example.com":  2,
				"contributor9@example.com":  2,
				"contributor10@example.com": 2,
				"contributor11@example.com": 2,
			},
			totalCommits:         22,
			expectedTopOwnership: 0.091, // 2/22 ≈ 0.091
			expectedStatus:       "Critical",
			expectedContributors: 11,
		},
		{
			name: "concentrated_ownership_bottleneck",
			commitsByContributor: map[string]int{
				"alice@example.com": 18,
				"bob@example.com":   1,
				"carol@example.com": 1,
			},
			totalCommits:         20,
			expectedTopOwnership: 0.90,
			expectedStatus:       "Caution",
			expectedContributors: 3,
		},
		{
			name: "single_contributor_no_problem",
			commitsByContributor: map[string]int{
				"alice@example.com": 20,
			},
			totalCommits:         20,
			expectedTopOwnership: 1.00,
			expectedStatus:       "Healthy",
			expectedContributors: 1,
		},
		{
			name: "two_contributors_shared_ownership",
			commitsByContributor: map[string]int{
				"alice@example.com": 12,
				"bob@example.com":   8,
			},
			totalCommits:         20,
			expectedTopOwnership: 0.60,
			expectedStatus:       "Healthy",
			expectedContributors: 2,
		},
		{
			name: "minimal_contributors_for_diffuse_analysis",
			commitsByContributor: map[string]int{
				"alice@example.com":   2,
				"bob@example.com":     2,
				"carol@example.com":   2,
				"david@example.com":   2,
				"eve@example.com":     2,
				"frank@example.com":   2,
				"grace@example.com":   2,
				"henry@example.com":   2,
				"iris@example.com":    2,
				"jack@example.com":    2,
				"kate@example.com":    2,
			},
			totalCommits:         22,
			expectedTopOwnership: 0.091,
			expectedStatus:       "Critical",
			expectedContributors: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topOwnership, status, contributors := calculateOwnershipClarity(tt.commitsByContributor)
			
			if floatDifference(topOwnership, tt.expectedTopOwnership) > testToleranceOwnership { // Allow small floating point differences
				t.Errorf("calculateOwnershipClarity() topOwnership = %v, want %v", topOwnership, tt.expectedTopOwnership)
			}
			if status != tt.expectedStatus {
				t.Errorf("calculateOwnershipClarity() status = %v, want %v", status, tt.expectedStatus)
			}
			if contributors != tt.expectedContributors {
				t.Errorf("calculateOwnershipClarity() contributors = %v, want %v", contributors, tt.expectedContributors)
			}
		})
	}
}

// floatDifference returns the absolute difference between two floats
func floatDifference(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

// TestClassifyOwnershipClarity tests ownership classification logic
func TestClassifyOwnershipClarity(t *testing.T) {
	tests := []struct {
		name                 string
		topOwnership        float64
		totalContributors   int
		expectedStatus      string
		expectedRecommendation string
	}{
		{
			name:                   "healthy_clear_ownership",
			topOwnership:          0.60,
			totalContributors:     3,
			expectedStatus:        "Healthy",
			expectedRecommendation: "Good ownership balance",
		},
		{
			name:                   "healthy_borderline_ownership",
			topOwnership:          0.50,
			totalContributors:     4,
			expectedStatus:        "Healthy",
			expectedRecommendation: "Good ownership balance",
		},
		{
			name:                   "caution_too_concentrated",
			topOwnership:          0.85,
			totalContributors:     5,
			expectedStatus:        "Caution",
			expectedRecommendation: "Consider encouraging more contributors to avoid bottlenecks",
		},
		{
			name:                   "warning_diffuse_ownership",
			topOwnership:          0.35,
			totalContributors:     8,
			expectedStatus:        "Warning",
			expectedRecommendation: "Consider assigning clearer ownership or primary maintainers",
		},
		{
			name:                   "critical_extremely_diffuse",
			topOwnership:          0.15,
			totalContributors:     12,
			expectedStatus:        "Critical",
			expectedRecommendation: "Urgent: assign primary maintainers for clear ownership",
		},
		{
			name:                   "healthy_single_owner",
			topOwnership:          1.00,
			totalContributors:     1,
			expectedStatus:        "Healthy",
			expectedRecommendation: "Good ownership balance",
		},
		{
			name:                   "healthy_small_team",
			topOwnership:          0.30,
			totalContributors:     3, // Small team, diffuse ownership is okay
			expectedStatus:        "Healthy",
			expectedRecommendation: "Good ownership balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, recommendation := classifyOwnershipClarity(tt.topOwnership, tt.totalContributors)
			
			if status != tt.expectedStatus {
				t.Errorf("classifyOwnershipClarity() status = %v, want %v", status, tt.expectedStatus)
			}
			if recommendation != tt.expectedRecommendation {
				t.Errorf("classifyOwnershipClarity() recommendation = %v, want %v", recommendation, tt.expectedRecommendation)
			}
		})
	}
}

// TestOwnershipClarityThresholds tests the threshold constants
func TestOwnershipClarityThresholds(t *testing.T) {
	// Test that our constants are reasonable
	if ownershipHealthyMinThreshold < 0.4 || ownershipHealthyMinThreshold > 0.6 {
		t.Errorf("ownershipHealthyMinThreshold should be between 0.4 and 0.6, got %v", ownershipHealthyMinThreshold)
	}
	
	if ownershipConcentratedThreshold < 0.7 || ownershipConcentratedThreshold > 0.9 {
		t.Errorf("ownershipConcentratedThreshold should be between 0.7 and 0.9, got %v", ownershipConcentratedThreshold)
	}
	
	if ownershipCriticalThreshold < 0.1 || ownershipCriticalThreshold > 0.3 {
		t.Errorf("ownershipCriticalThreshold should be between 0.1 and 0.3, got %v", ownershipCriticalThreshold)
	}
	
	if ownershipMinContributorsForAnalysis < 3 || ownershipMinContributorsForAnalysis > 15 {
		t.Errorf("ownershipMinContributorsForAnalysis should be between 3 and 15, got %v", ownershipMinContributorsForAnalysis)
	}
}

// TestOwnershipClarityEdgeCases tests edge cases and error conditions
func TestOwnershipClarityEdgeCases(t *testing.T) {
	tests := []struct {
		name                   string
		commitsByContributor   map[string]int
		expectedTopOwnership  float64
		expectedStatus        string
	}{
		{
			name:                   "empty_commits_map",
			commitsByContributor:   map[string]int{},
			expectedTopOwnership:  0.0,
			expectedStatus:        "Unknown",
		},
		{
			name: "zero_commits_contributors",
			commitsByContributor: map[string]int{
				"alice@example.com": 0,
				"bob@example.com":   0,
			},
			expectedTopOwnership: 0.0,
			expectedStatus:       "Unknown",
		},
		{
			name: "negative_commits_handled",
			commitsByContributor: map[string]int{
				"alice@example.com": 10,
				"bob@example.com":   -2, // Should be ignored or handled gracefully
			},
			expectedTopOwnership: 1.0, // Only alice's commits count
			expectedStatus:       "Healthy",
		},
		{
			name: "very_large_numbers",
			commitsByContributor: map[string]int{
				"alice@example.com": 1000000,
				"bob@example.com":   500000,
				"carol@example.com": 100000,
			},
			expectedTopOwnership: 0.625, // 1000000 / 1600000
			expectedStatus:       "Healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topOwnership, status, _ := calculateOwnershipClarity(tt.commitsByContributor)
			
			if floatDifference(topOwnership, tt.expectedTopOwnership) > testToleranceOwnership {
				t.Errorf("calculateOwnershipClarity() topOwnership = %v, want %v", topOwnership, tt.expectedTopOwnership)
			}
			if status != tt.expectedStatus {
				t.Errorf("calculateOwnershipClarity() status = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}
