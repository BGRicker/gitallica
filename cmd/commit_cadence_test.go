package cmd

import (
	"testing"
	"time"
)

func TestCommitCadenceTrends(t *testing.T) {
	// Test data: commit counts over time periods
	timePeriods := []TimePeriod{
		{Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC), CommitCount: 10},
		{Start: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 14, 23, 59, 59, 0, time.UTC), CommitCount: 12},
		{Start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 21, 23, 59, 59, 0, time.UTC), CommitCount: 15},
		{Start: time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 28, 23, 59, 59, 0, time.UTC), CommitCount: 18},
		{Start: time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 2, 4, 23, 59, 59, 0, time.UTC), CommitCount: 20},
	}

	stats := calculateCommitCadenceStats(timePeriods)

	// Test basic statistics
	if stats.TotalCommits != 75 {
		t.Errorf("Expected TotalCommits = 75, got %d", stats.TotalCommits)
	}

	expectedAverage := 15.0 // 75/5
	tolerance := 0.01
	if abs(stats.AverageCommitsPerPeriod-expectedAverage) > tolerance {
		t.Errorf("Expected AverageCommitsPerPeriod ≈ %.2f, got %.2f", expectedAverage, stats.AverageCommitsPerPeriod)
	}

	// Test trend detection (increasing trend)
	expectedTrend := "Increasing"
	if stats.TrendDirection != expectedTrend {
		t.Errorf("Expected TrendDirection = %s, got %s", expectedTrend, stats.TrendDirection)
	}

	// Test that trend strength is calculated
	if stats.TrendStrength <= 0 {
		t.Errorf("Expected positive TrendStrength, got %.3f", stats.TrendStrength)
	}
}

func TestDetectCommitSpikes(t *testing.T) {
	timePeriods := []TimePeriod{
		{Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC), CommitCount: 10},
		{Start: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 14, 23, 59, 59, 0, time.UTC), CommitCount: 12},
		{Start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 21, 23, 59, 59, 0, time.UTC), CommitCount: 50}, // Spike
		{Start: time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 28, 23, 59, 59, 0, time.UTC), CommitCount: 11},
		{Start: time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 2, 4, 23, 59, 59, 0, time.UTC), CommitCount: 13},
	}

	spikes := detectCommitSpikes(timePeriods)

	if len(spikes) != 1 {
		t.Errorf("Expected 1 spike, got %d", len(spikes))
	}

	if len(spikes) > 0 {
		spike := spikes[0]
		if spike.CommitCount != 50 {
			t.Errorf("Expected spike CommitCount = 50, got %d", spike.CommitCount)
		}
		if spike.Severity != "High" {
			t.Errorf("Expected spike Severity = High, got %s", spike.Severity)
		}
	}
}

func TestDetectCommitDips(t *testing.T) {
	timePeriods := []TimePeriod{
		{Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC), CommitCount: 20},
		{Start: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 14, 23, 59, 59, 0, time.UTC), CommitCount: 18},
		{Start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 21, 23, 59, 59, 0, time.UTC), CommitCount: 2}, // Dip
		{Start: time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 28, 23, 59, 59, 0, time.UTC), CommitCount: 19},
		{Start: time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 2, 4, 23, 59, 59, 0, time.UTC), CommitCount: 21},
	}

	dips := detectCommitDips(timePeriods)

	if len(dips) != 1 {
		t.Errorf("Expected 1 dip, got %d", len(dips))
	}

	if len(dips) > 0 {
		dip := dips[0]
		if dip.CommitCount != 2 {
			t.Errorf("Expected dip CommitCount = 2, got %d", dip.CommitCount)
		}
		if dip.Severity != "High" {
			t.Errorf("Expected dip Severity = High, got %s", dip.Severity)
		}
	}
}

func TestClassifyTrendDirection(t *testing.T) {
	tests := []struct {
		name             string
		slope            float64
		expectedTrend    string
		expectedStrength float64
	}{
		{
			name:             "strong increasing trend",
			slope:            2.5,
			expectedTrend:    "Increasing",
			expectedStrength: 2.5,
		},
		{
			name:             "mild increasing trend",
			slope:            0.3,
			expectedTrend:    "Increasing",
			expectedStrength: 0.3,
		},
		{
			name:             "stable trend",
			slope:            0.05,
			expectedTrend:    "Stable",
			expectedStrength: 0.05,
		},
		{
			name:             "mild decreasing trend",
			slope:            -0.4,
			expectedTrend:    "Decreasing",
			expectedStrength: 0.4, // Absolute value
		},
		{
			name:             "strong decreasing trend",
			slope:            -1.8,
			expectedTrend:    "Decreasing",
			expectedStrength: 1.8, // Absolute value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend, strength := classifyTrendDirection(tt.slope)

			if trend != tt.expectedTrend {
				t.Errorf("Expected trend = %s, got %s", tt.expectedTrend, trend)
			}

			tolerance := 0.01
			if abs(strength-tt.expectedStrength) > tolerance {
				t.Errorf("Expected strength ≈ %.2f, got %.2f", tt.expectedStrength, strength)
			}
		})
	}
}

func TestClassifySustainabilityLevel(t *testing.T) {
	tests := []struct {
		name           string
		avgCommits     float64
		spikeCount     int
		dipCount       int
		trendDirection string
		expected       string
	}{
		{
			name:           "healthy sustainable pace",
			avgCommits:     15.0,
			spikeCount:     0,
			dipCount:       0,
			trendDirection: "Stable",
			expected:       "Healthy",
		},
		{
			name:           "warning with spikes",
			avgCommits:     20.0,
			spikeCount:     2,
			dipCount:       0,
			trendDirection: "Increasing",
			expected:       "Warning",
		},
		{
			name:           "critical with many spikes and dips",
			avgCommits:     25.0,
			spikeCount:     4,
			dipCount:       3,
			trendDirection: "Increasing",
			expected:       "Critical",
		},
		{
			name:           "caution with decreasing trend",
			avgCommits:     8.0,
			spikeCount:     0,
			dipCount:       1,
			trendDirection: "Decreasing",
			expected:       "Caution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := classifySustainabilityLevel(tt.avgCommits, tt.spikeCount, tt.dipCount, tt.trendDirection)

			if level != tt.expected {
				t.Errorf("Expected sustainability level = %s, got %s", tt.expected, level)
			}
		})
	}
}

func TestGroupCommitsByTimePeriod(t *testing.T) {
	// Test commits grouped by week
	commits := []CommitInfo{
		{Time: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), Hash: "abc1"},
		{Time: time.Date(2024, 1, 2, 14, 0, 0, 0, time.UTC), Hash: "abc2"},
		{Time: time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC), Hash: "abc3"},
		{Time: time.Date(2024, 1, 15, 16, 0, 0, 0, time.UTC), Hash: "abc4"},
		{Time: time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC), Hash: "abc5"},
	}

	periods := groupCommitsByTimePeriod(commits, "week")

	if len(periods) != 3 {
		t.Errorf("Expected 3 periods, got %d", len(periods))
	}

	// First week should have 2 commits
	if periods[0].CommitCount != 2 {
		t.Errorf("Expected first period to have 2 commits, got %d", periods[0].CommitCount)
	}

	// Second week should have 1 commit
	if periods[1].CommitCount != 1 {
		t.Errorf("Expected second period to have 1 commit, got %d", periods[1].CommitCount)
	}

	// Third week should have 2 commits
	if periods[2].CommitCount != 2 {
		t.Errorf("Expected third period to have 2 commits, got %d", periods[2].CommitCount)
	}
}

func TestCommitCadenceEdgeCases(t *testing.T) {
	// Test empty periods
	emptyPeriods := []TimePeriod{}
	stats := calculateCommitCadenceStats(emptyPeriods)

	if stats.TotalCommits != 0 {
		t.Errorf("Expected TotalCommits = 0 for empty periods, got %d", stats.TotalCommits)
	}
	if stats.TrendDirection != "Unknown" {
		t.Errorf("Expected TrendDirection = Unknown for empty periods, got %s", stats.TrendDirection)
	}

	// Test single period
	singlePeriod := []TimePeriod{
		{Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC), CommitCount: 10},
	}
	singleStats := calculateCommitCadenceStats(singlePeriod)

	if singleStats.TotalCommits != 10 {
		t.Errorf("Expected TotalCommits = 10 for single period, got %d", singleStats.TotalCommits)
	}
	if singleStats.TrendDirection != "Unknown" {
		t.Errorf("Expected TrendDirection = Unknown for single period, got %s", singleStats.TrendDirection)
	}
}
