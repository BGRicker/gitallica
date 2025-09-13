package cmd

import (
	"testing"
)

func TestHighRiskCommitsThresholds(t *testing.T) {
	tests := []struct {
		name           string
		linesChanged   int
		filesChanged   int
		expectedRisk   string
		expectedReason string
	}{
		{
			name:           "low risk - small commit",
			linesChanged:   50,
			filesChanged:   3,
			expectedRisk:   "Low",
			expectedReason: "Small, focused commit",
		},
		{
			name:           "moderate risk - many files",
			linesChanged:   200,
			filesChanged:   8,
			expectedRisk:   "Moderate",
			expectedReason: "Moderately large commit",
		},
		{
			name:           "high risk - many lines",
			linesChanged:   450,
			filesChanged:   5,
			expectedRisk:   "High",
			expectedReason: "Large line changes - review carefully",
		},
		{
			name:           "high risk - many files",
			linesChanged:   100,
			filesChanged:   15,
			expectedRisk:   "High",
			expectedReason: "Touches many files - high complexity",
		},
		{
			name:           "critical risk - monster commit",
			linesChanged:   800,
			filesChanged:   20,
			expectedRisk:   "Critical",
			expectedReason: "Monster commit - extremely high risk",
		},
		{
			name:           "boundary case - exactly 400 lines",
			linesChanged:   400,
			filesChanged:   5,
			expectedRisk:   "Moderate",
			expectedReason: "Moderately large commit",
		},
		{
			name:           "boundary case - exactly 12 files",
			linesChanged:   100,
			filesChanged:   12,
			expectedRisk:   "High",
			expectedReason: "Touches many files - high complexity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk, reason := classifyCommitRisk(tt.linesChanged, tt.filesChanged)

			if risk != tt.expectedRisk {
				t.Errorf("classifyCommitRisk() risk = %v, want %v", risk, tt.expectedRisk)
			}

			if reason != tt.expectedReason {
				t.Errorf("classifyCommitRisk() reason = %v, want %v", reason, tt.expectedReason)
			}
		})
	}
}

func TestCalculateHighRiskCommitsStats(t *testing.T) {
	commits := []HighRiskCommit{
		{Hash: "abc123", LinesChanged: 50, FilesChanged: 2, Risk: "Low"},
		{Hash: "def456", LinesChanged: 300, FilesChanged: 8, Risk: "Moderate"},
		{Hash: "ghi789", LinesChanged: 500, FilesChanged: 15, Risk: "High"},
		{Hash: "jkl012", LinesChanged: 1000, FilesChanged: 25, Risk: "Critical"},
		{Hash: "mno345", LinesChanged: 450, FilesChanged: 6, Risk: "High"},
	}

	stats := calculateHighRiskCommitsStats(commits)

	// Test total counts
	if stats.TotalCommits != 5 {
		t.Errorf("Expected TotalCommits = 5, got %d", stats.TotalCommits)
	}

	// Test risk distribution
	if stats.LowRisk != 1 {
		t.Errorf("Expected LowRisk = 1, got %d", stats.LowRisk)
	}
	if stats.ModerateRisk != 1 {
		t.Errorf("Expected ModerateRisk = 1, got %d", stats.ModerateRisk)
	}
	if stats.HighRisk != 2 {
		t.Errorf("Expected HighRisk = 2, got %d", stats.HighRisk)
	}
	if stats.CriticalRisk != 1 {
		t.Errorf("Expected CriticalRisk = 1, got %d", stats.CriticalRisk)
	}

	// Test averages - use tolerance for float comparison
	tolerance := 0.01
	expectedAvgLines := 460.0 // (50+300+500+1000+450)/5
	if abs(stats.AverageLines-expectedAvgLines) > tolerance {
		t.Errorf("Expected AverageLines ≈ %.2f, got %.2f", expectedAvgLines, stats.AverageLines)
	}

	expectedAvgFiles := 11.2 // (2+8+15+25+6)/5
	if abs(stats.AverageFiles-expectedAvgFiles) > tolerance {
		t.Errorf("Expected AverageFiles ≈ %.2f, got %.2f", expectedAvgFiles, stats.AverageFiles)
	}

	// Test largest commit identification
	if stats.LargestCommit.Hash != "jkl012" {
		t.Errorf("Expected LargestCommit.Hash = jkl012, got %s", stats.LargestCommit.Hash)
	}
	if stats.LargestCommit.LinesChanged != 1000 {
		t.Errorf("Expected LargestCommit.LinesChanged = 1000, got %d", stats.LargestCommit.LinesChanged)
	}
}

func TestHighRiskCommitsEdgeCases(t *testing.T) {
	// Test empty commits
	emptyCommits := []HighRiskCommit{}
	stats := calculateHighRiskCommitsStats(emptyCommits)

	if stats.TotalCommits != 0 {
		t.Errorf("Expected TotalCommits = 0 for empty list, got %d", stats.TotalCommits)
	}
	if stats.AverageLines != 0.0 {
		t.Errorf("Expected AverageLines = 0.0 for empty list, got %.2f", stats.AverageLines)
	}

	// Test single commit
	singleCommit := []HighRiskCommit{
		{Hash: "single", LinesChanged: 100, FilesChanged: 5, Risk: "Low"},
	}
	singleStats := calculateHighRiskCommitsStats(singleCommit)

	if singleStats.TotalCommits != 1 {
		t.Errorf("Expected TotalCommits = 1 for single commit, got %d", singleStats.TotalCommits)
	}
	if singleStats.AverageLines != 100.0 {
		t.Errorf("Expected AverageLines = 100.0 for single commit, got %.2f", singleStats.AverageLines)
	}
	if singleStats.LargestCommit.Hash != "single" {
		t.Errorf("Expected LargestCommit.Hash = single, got %s", singleStats.LargestCommit.Hash)
	}
}
