package cmd

import (
	"testing"
	"time"
)

func TestLongLivedBranchesAnalysis(t *testing.T) {
	// Test data: branches with different ages
	branches := []BranchInfo{
		{Name: "feature/quick-fix", AgeInDays: 1, Status: "active"},
		{Name: "feature/new-feature", AgeInDays: 5, Status: "active"},
		{Name: "feature/old-work", AgeInDays: 15, Status: "active"},
		{Name: "hotfix/urgent", AgeInDays: 0, Status: "active"},
		{Name: "experiment/research", AgeInDays: 30, Status: "active"},
	}

	stats := calculateLongLivedBranchesStats(branches)

	// Test basic statistics
	if stats.TotalBranches != 5 {
		t.Errorf("Expected TotalBranches = 5, got %d", stats.TotalBranches)
	}

	// Calculate expected average programmatically for maintainability
	var totalAge int
	for _, b := range branches {
		totalAge += int(b.AgeInDays)
	}
	expectedAverage := float64(totalAge) / float64(len(branches)) // (1+5+15+0+30)/5 = 10.2
	tolerance := 0.1
	if abs(stats.AverageBranchAge-expectedAverage) > tolerance {
		t.Errorf("Expected AverageBranchAge ≈ %.1f, got %.1f", expectedAverage, stats.AverageBranchAge)
	}

	// Test risk distribution
	if stats.HealthyBranches != 2 { // 0 and 1 days
		t.Errorf("Expected HealthyBranches = 2, got %d", stats.HealthyBranches)
	}

	if stats.WarningBranches != 1 { // 5 days
		t.Errorf("Expected WarningBranches = 1, got %d", stats.WarningBranches)
	}

	if stats.RiskyBranches != 1 { // 15 days
		t.Errorf("Expected RiskyBranches = 1, got %d", stats.RiskyBranches)
	}

	if stats.CriticalBranches != 1 { // 30 days
		t.Errorf("Expected CriticalBranches = 1, got %d", stats.CriticalBranches)
	}

	// Test trunk-based compliance
	expectedCompliance := "Moderate" // 40% (2/5) are healthy
	if stats.TrunkBasedCompliance != expectedCompliance {
		t.Errorf("Expected TrunkBasedCompliance = %s, got %s", expectedCompliance, stats.TrunkBasedCompliance)
	}
}

func TestClassifyBranchRisk(t *testing.T) {
	tests := []struct {
		name         string
		ageInDays    float64
		expectedRisk string
	}{
		{
			name:         "same day branch",
			ageInDays:    0.5,
			expectedRisk: "Healthy",
		},
		{
			name:         "one day old",
			ageInDays:    1.0,
			expectedRisk: "Healthy",
		},
		{
			name:         "two day boundary",
			ageInDays:    2.0,
			expectedRisk: "Healthy",
		},
		{
			name:         "three day boundary",
			ageInDays:    3.0,
			expectedRisk: "Warning",
		},
		{
			name:         "five day old",
			ageInDays:    5.0,
			expectedRisk: "Warning",
		},
		{
			name:         "seven day boundary",
			ageInDays:    7.0,
			expectedRisk: "Warning",
		},
		{
			name:         "eight day old",
			ageInDays:    8.0,
			expectedRisk: "Risky",
		},
		{
			name:         "fifteen day old",
			ageInDays:    15.0,
			expectedRisk: "Risky",
		},
		{
			name:         "thirty day old",
			ageInDays:    30.0,
			expectedRisk: "Critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := classifyBranchRisk(tt.ageInDays)

			if risk != tt.expectedRisk {
				t.Errorf("Expected risk = %s, got %s", tt.expectedRisk, risk)
			}
		})
	}
}

func TestClassifyTrunkBasedCompliance(t *testing.T) {
	tests := []struct {
		name               string
		healthyBranches    int
		totalBranches      int
		expectedCompliance string
	}{
		{
			name:               "excellent compliance",
			healthyBranches:    9,
			totalBranches:      10,
			expectedCompliance: "Excellent",
		},
		{
			name:               "good compliance",
			healthyBranches:    7,
			totalBranches:      10,
			expectedCompliance: "Good",
		},
		{
			name:               "moderate compliance",
			healthyBranches:    5,
			totalBranches:      10,
			expectedCompliance: "Moderate",
		},
		{
			name:               "poor compliance",
			healthyBranches:    2,
			totalBranches:      10,
			expectedCompliance: "Poor",
		},
		{
			name:               "critical compliance",
			healthyBranches:    1,
			totalBranches:      10,
			expectedCompliance: "Critical",
		},
		{
			name:               "no branches edge case",
			healthyBranches:    0,
			totalBranches:      0,
			expectedCompliance: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compliance := classifyTrunkBasedCompliance(tt.healthyBranches, tt.totalBranches)

			if compliance != tt.expectedCompliance {
				t.Errorf("Expected compliance = %s, got %s", tt.expectedCompliance, compliance)
			}
		})
	}
}

func TestCalculateBranchAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		createdTime  time.Time
		expectedDays float64
	}{
		{
			name:         "same day",
			createdTime:  now.Add(-12 * time.Hour),
			expectedDays: 0.5,
		},
		{
			name:         "one day ago",
			createdTime:  now.Add(-24 * time.Hour),
			expectedDays: 1.0,
		},
		{
			name:         "three days ago",
			createdTime:  now.Add(-72 * time.Hour),
			expectedDays: 3.0,
		},
		{
			name:         "one week ago",
			createdTime:  now.Add(-7 * 24 * time.Hour),
			expectedDays: 7.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			age := calculateBranchAge(tt.createdTime, now)

			tolerance := 0.1
			if abs(age-tt.expectedDays) > tolerance {
				t.Errorf("Expected age ≈ %.1f days, got %.1f", tt.expectedDays, age)
			}
		})
	}
}

func TestLongLivedBranchesEdgeCases(t *testing.T) {
	// Test empty branches
	emptyBranches := []BranchInfo{}
	stats := calculateLongLivedBranchesStats(emptyBranches)

	if stats.TotalBranches != 0 {
		t.Errorf("Expected TotalBranches = 0 for empty branches, got %d", stats.TotalBranches)
	}
	if stats.TrunkBasedCompliance != "Unknown" {
		t.Errorf("Expected TrunkBasedCompliance = Unknown for empty branches, got %s", stats.TrunkBasedCompliance)
	}

	// Test single healthy branch
	singleBranch := []BranchInfo{
		{Name: "feature/quick", AgeInDays: 1, Status: "active"},
	}
	singleStats := calculateLongLivedBranchesStats(singleBranch)

	if singleStats.TotalBranches != 1 {
		t.Errorf("Expected TotalBranches = 1 for single branch, got %d", singleStats.TotalBranches)
	}
	if singleStats.TrunkBasedCompliance != "Excellent" {
		t.Errorf("Expected TrunkBasedCompliance = Excellent for single healthy branch, got %s", singleStats.TrunkBasedCompliance)
	}
}
