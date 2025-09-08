package cmd

import (
	"math"
	"testing"
	"time"
)

func TestChangeLeadTimeAnalysis(t *testing.T) {
	// Test data: commits with different lead times
	commits := []CommitLeadTime{
		{Hash: "abc123", CommitTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), DeployTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), LeadTimeHours: 4.0},   // Elite: 4 hours
		{Hash: "def456", CommitTime: time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC), DeployTime: time.Date(2024, 1, 4, 9, 0, 0, 0, time.UTC), LeadTimeHours: 48.0},     // High: 2 days
		{Hash: "ghi789", CommitTime: time.Date(2024, 1, 5, 8, 0, 0, 0, time.UTC), DeployTime: time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC), LeadTimeHours: 240.0},   // Medium: 10 days
		{Hash: "jkl012", CommitTime: time.Date(2024, 1, 10, 7, 0, 0, 0, time.UTC), DeployTime: time.Date(2024, 2, 20, 7, 0, 0, 0, time.UTC), LeadTimeHours: 984.0}, // Low: 41 days
		{Hash: "mno345", CommitTime: time.Date(2024, 1, 15, 6, 0, 0, 0, time.UTC), DeployTime: time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC), LeadTimeHours: 12.0}, // Elite: 12 hours
	}

	stats := calculateChangeLeadTimeStats(commits)

	// Test basic statistics
	if stats.TotalCommits != 5 {
		t.Errorf("Expected TotalCommits = 5, got %d", stats.TotalCommits)
	}

	// Calculate expected average programmatically for maintainability
	var totalHours float64
	for _, c := range commits {
		totalHours += c.LeadTimeHours
	}
	expectedAverage := totalHours / float64(len(commits)) // (4+48+240+984+12)/5 = 257.6 hours
	tolerance := 0.1
	if math.Abs(stats.AverageLeadTimeHours-expectedAverage) > tolerance {
		t.Errorf("Expected AverageLeadTimeHours ≈ %.1f, got %.1f", expectedAverage, stats.AverageLeadTimeHours)
	}

	// Test DORA classification distribution
	if stats.EliteCommits != 2 { // 4 and 12 hours
		t.Errorf("Expected EliteCommits = 2, got %d", stats.EliteCommits)
	}

	if stats.HighCommits != 1 { // 48 hours (2 days)
		t.Errorf("Expected HighCommits = 1, got %d", stats.HighCommits)
	}

	if stats.MediumCommits != 1 { // 240 hours (10 days)
		t.Errorf("Expected MediumCommits = 1, got %d", stats.MediumCommits)
	}

	if stats.LowCommits != 1 { // 984 hours (41 days)
		t.Errorf("Expected LowCommits = 1, got %d", stats.LowCommits)
	}

	// Test DORA performance level  
	// Elite: 2/5 = 40%, Elite+High: 3/5 = 60% → qualifies as High performance
	expectedPerformance := "High" // 60% elite+high commits meets high threshold
	if stats.DORAPerformanceLevel != expectedPerformance {
		t.Errorf("Expected DORAPerformanceLevel = %s, got %s", expectedPerformance, stats.DORAPerformanceLevel)
	}
}

func TestClassifyDORALeadTime(t *testing.T) {
	tests := []struct {
		name                string
		leadTimeHours       float64
		expectedClassification string
	}{
		{
			name:                "elite same day",
			leadTimeHours:       4.0,
			expectedClassification: "Elite",
		},
		{
			name:                "elite boundary",
			leadTimeHours:       23.9, // Just under 24 hours
			expectedClassification: "Elite",
		},
		{
			name:                "high one day",
			leadTimeHours:       24.0,
			expectedClassification: "High",
		},
		{
			name:                "high few days",
			leadTimeHours:       72.0, // 3 days
			expectedClassification: "High",
		},
		{
			name:                "high boundary",
			leadTimeHours:       167.9, // Just under 7 days
			expectedClassification: "High",
		},
		{
			name:                "medium one week",
			leadTimeHours:       168.0, // 7 days
			expectedClassification: "Medium",
		},
		{
			name:                "medium few weeks",
			leadTimeHours:       480.0, // 20 days
			expectedClassification: "Medium",
		},
		{
			name:                "medium boundary",
			leadTimeHours:       719.9, // Just under 30 days
			expectedClassification: "Medium",
		},
		{
			name:                "low one month",
			leadTimeHours:       720.0, // 30 days
			expectedClassification: "Low",
		},
		{
			name:                "low multiple months",
			leadTimeHours:       2160.0, // 90 days
			expectedClassification: "Low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := classifyDORALeadTime(tt.leadTimeHours)
			
			if classification != tt.expectedClassification {
				t.Errorf("Expected classification = %s, got %s", tt.expectedClassification, classification)
			}
		})
	}
}

func TestClassifyDORAPerformanceLevel(t *testing.T) {
	tests := []struct {
		name                string
		eliteCommits        int
		highCommits         int
		mediumCommits       int
		lowCommits          int
		totalCommits        int
		expectedPerformance string
	}{
		{
			name:                "elite performance",
			eliteCommits:        8,
			highCommits:         1,
			mediumCommits:       1,
			lowCommits:          0,
			totalCommits:        10,
			expectedPerformance: "Elite",
		},
		{
			name:                "high performance",
			eliteCommits:        4,
			highCommits:         4,
			mediumCommits:       2,
			lowCommits:          0,
			totalCommits:        10,
			expectedPerformance: "High",
		},
		{
			name:                "medium performance",
			eliteCommits:        2,
			highCommits:         2,
			mediumCommits:       4,
			lowCommits:          2,
			totalCommits:        10,
			expectedPerformance: "Medium",
		},
		{
			name:                "low performance",
			eliteCommits:        1,
			highCommits:         1,
			mediumCommits:       2,
			lowCommits:          6,
			totalCommits:        10,
			expectedPerformance: "Low",
		},
		{
			name:                "no commits edge case",
			eliteCommits:        0,
			highCommits:         0,
			mediumCommits:       0,
			lowCommits:          0,
			totalCommits:        0,
			expectedPerformance: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			performance := classifyDORAPerformanceLevel(tt.eliteCommits, tt.highCommits, tt.mediumCommits, tt.lowCommits, tt.totalCommits)
			
			if performance != tt.expectedPerformance {
				t.Errorf("Expected performance = %s, got %s", tt.expectedPerformance, performance)
			}
		})
	}
}

func TestCalculateLeadTime(t *testing.T) {
	commitTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	deployTime := time.Date(2024, 1, 3, 14, 0, 0, 0, time.UTC)

	expectedHours := 52.0 // 2 days and 4 hours
	leadTime := calculateLeadTime(commitTime, deployTime)

	tolerance := 0.1
	if math.Abs(leadTime-expectedHours) > tolerance {
		t.Errorf("Expected leadTime ≈ %.1f hours, got %.1f", expectedHours, leadTime)
	}
}

func TestCalculatePercentiles(t *testing.T) {
	// Test data with known percentiles
	leadTimes := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}

	p50 := calculatePercentile(leadTimes, 50)
	p95 := calculatePercentile(leadTimes, 95)

	// For sorted data, 50th percentile should be median (5.5)
	expectedP50 := 5.5
	tolerance := 0.1
	if math.Abs(p50-expectedP50) > tolerance {
		t.Errorf("Expected P50 ≈ %.1f, got %.1f", expectedP50, p50)
	}

	// 95th percentile should be near the high end
	expectedP95 := 9.5
	if math.Abs(p95-expectedP95) > tolerance {
		t.Errorf("Expected P95 ≈ %.1f, got %.1f", expectedP95, p95)
	}
}

func TestChangeLeadTimeEdgeCases(t *testing.T) {
	// Test empty commits
	emptyCommits := []CommitLeadTime{}
	stats := calculateChangeLeadTimeStats(emptyCommits)
	
	if stats.TotalCommits != 0 {
		t.Errorf("Expected TotalCommits = 0 for empty commits, got %d", stats.TotalCommits)
	}
	if stats.DORAPerformanceLevel != "Unknown" {
		t.Errorf("Expected DORAPerformanceLevel = Unknown for empty commits, got %s", stats.DORAPerformanceLevel)
	}

	// Test single commit
	singleCommit := []CommitLeadTime{
		{Hash: "abc123", CommitTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), DeployTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), LeadTimeHours: 4.0},
	}
	singleStats := calculateChangeLeadTimeStats(singleCommit)
	
	if singleStats.TotalCommits != 1 {
		t.Errorf("Expected TotalCommits = 1 for single commit, got %d", singleStats.TotalCommits)
	}
	if singleStats.DORAPerformanceLevel != "Elite" {
		t.Errorf("Expected DORAPerformanceLevel = Elite for single elite commit, got %s", singleStats.DORAPerformanceLevel)
	}
}

func TestLeadTimeFromCommitToMerge(t *testing.T) {
	// Test calculating lead time based on commit to merge time
	commitTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	mergeTime := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)

	expectedHours := 24.0 // 1 day
	leadTime := calculateLeadTime(commitTime, mergeTime)

	tolerance := 0.1
	if math.Abs(leadTime-expectedHours) > tolerance {
		t.Errorf("Expected leadTime ≈ %.1f hours, got %.1f", expectedHours, leadTime)
	}

	// Should be classified as High (1 day-1 week)
	classification := classifyDORALeadTime(leadTime)
	if classification != "High" {
		t.Errorf("Expected classification = High, got %s", classification)
	}
}
