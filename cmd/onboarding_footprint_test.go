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
	"time"
)

// TestClassifyOnboardingComplexity tests the classification of onboarding complexity
func TestClassifyOnboardingComplexity(t *testing.T) {
	tests := []struct {
		name                 string
		filesCount          int
		commitsCount        int
		expectedStatus      string
		expectedRecommendation string
	}{
		{
			name:                   "simple_onboarding_single_file",
			filesCount:            1,
			commitsCount:          3,
			expectedStatus:        "Simple",
			expectedRecommendation: "Excellent focused onboarding",
		},
		{
			name:                   "simple_onboarding_few_files",
			filesCount:            4,
			commitsCount:          5,
			expectedStatus:        "Simple",
			expectedRecommendation: "Excellent focused onboarding",
		},
		{
			name:                   "moderate_onboarding_reasonable_scope",
			filesCount:            8,
			commitsCount:          5,
			expectedStatus:        "Moderate",
			expectedRecommendation: "Reasonable onboarding complexity",
		},
		{
			name:                   "complex_onboarding_many_files",
			filesCount:            15,
			commitsCount:          5,
			expectedStatus:        "Complex",
			expectedRecommendation: "Consider simplifying initial tasks",
		},
		{
			name:                   "overwhelming_onboarding_too_many_files",
			filesCount:            25,
			commitsCount:          5,
			expectedStatus:        "Overwhelming",
			expectedRecommendation: "Urgent: simplify onboarding process",
		},
		{
			name:                   "borderline_simple_to_moderate",
			filesCount:            5,
			commitsCount:          5,
			expectedStatus:        "Simple",
			expectedRecommendation: "Excellent focused onboarding",
		},
		{
			name:                   "borderline_moderate_to_complex",
			filesCount:            10,
			commitsCount:          5,
			expectedStatus:        "Moderate",
			expectedRecommendation: "Reasonable onboarding complexity",
		},
		{
			name:                   "borderline_complex_to_overwhelming",
			filesCount:            20,
			commitsCount:          5,
			expectedStatus:        "Complex",
			expectedRecommendation: "Consider simplifying initial tasks",
		},
		{
			name:                   "early_contributor_few_commits",
			filesCount:            3,
			commitsCount:          2,
			expectedStatus:        "Simple",
			expectedRecommendation: "Excellent focused onboarding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, recommendation := classifyOnboardingComplexity(tt.filesCount)
			
			if status != tt.expectedStatus {
				t.Errorf("classifyOnboardingComplexity() status = %v, want %v", status, tt.expectedStatus)
			}
			if recommendation != tt.expectedRecommendation {
				t.Errorf("classifyOnboardingComplexity() recommendation = %v, want %v", recommendation, tt.expectedRecommendation)
			}
		})
	}
}

// TestCalculateOnboardingStats tests onboarding statistics calculation
func TestCalculateOnboardingStats(t *testing.T) {
	tests := []struct {
		name                    string
		contributors           []MockNewContributor
		expectedAvgFiles       float64
		expectedSimpleCount    int
		expectedModerateCount  int
		expectedComplexCount   int
		expectedOverwhelmingCount int
	}{
		{
			name: "mixed_onboarding_complexity",
			contributors: []MockNewContributor{
				{Email: "alice@example.com", FilesCount: 3, CommitsCount: 4},
				{Email: "bob@example.com", FilesCount: 8, CommitsCount: 5},
				{Email: "carol@example.com", FilesCount: 15, CommitsCount: 5},
				{Email: "david@example.com", FilesCount: 25, CommitsCount: 5},
			},
			expectedAvgFiles:         12.75, // (3+8+15+25)/4
			expectedSimpleCount:      1,     // alice
			expectedModerateCount:    1,     // bob
			expectedComplexCount:     1,     // carol
			expectedOverwhelmingCount: 1,    // david
		},
		{
			name: "all_simple_onboarding",
			contributors: []MockNewContributor{
				{Email: "alice@example.com", FilesCount: 2, CommitsCount: 3},
				{Email: "bob@example.com", FilesCount: 4, CommitsCount: 4},
				{Email: "carol@example.com", FilesCount: 1, CommitsCount: 2},
			},
			expectedAvgFiles:         2.33, // (2+4+1)/3 ≈ 2.33
			expectedSimpleCount:      3,
			expectedModerateCount:    0,
			expectedComplexCount:     0,
			expectedOverwhelmingCount: 0,
		},
		{
			name: "all_complex_onboarding",
			contributors: []MockNewContributor{
				{Email: "alice@example.com", FilesCount: 18, CommitsCount: 5},
				{Email: "bob@example.com", FilesCount: 22, CommitsCount: 5},
			},
			expectedAvgFiles:         20.0, // (18+22)/2
			expectedSimpleCount:      0,
			expectedModerateCount:    0,
			expectedComplexCount:     1,    // alice (18 files)
			expectedOverwhelmingCount: 1,   // bob (22 files)
		},
		{
			name:                    "no_contributors",
			contributors:           []MockNewContributor{},
			expectedAvgFiles:       0.0,
			expectedSimpleCount:    0,
			expectedModerateCount:  0,
			expectedComplexCount:   0,
			expectedOverwhelmingCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avgFiles, simpleCnt, moderateCnt, complexCnt, overwhelmingCnt := calculateOnboardingStats(tt.contributors)
			
			tolerance := 0.01
			if floatDiff(avgFiles, tt.expectedAvgFiles) > tolerance {
				t.Errorf("calculateOnboardingStats() avgFiles = %v, want %v", avgFiles, tt.expectedAvgFiles)
			}
			if simpleCnt != tt.expectedSimpleCount {
				t.Errorf("calculateOnboardingStats() simpleCount = %v, want %v", simpleCnt, tt.expectedSimpleCount)
			}
			if moderateCnt != tt.expectedModerateCount {
				t.Errorf("calculateOnboardingStats() moderateCount = %v, want %v", moderateCnt, tt.expectedModerateCount)
			}
			if complexCnt != tt.expectedComplexCount {
				t.Errorf("calculateOnboardingStats() complexCount = %v, want %v", complexCnt, tt.expectedComplexCount)
			}
			if overwhelmingCnt != tt.expectedOverwhelmingCount {
				t.Errorf("calculateOnboardingStats() overwhelmingCount = %v, want %v", overwhelmingCnt, tt.expectedOverwhelmingCount)
			}
		})
	}
}

// TestOnboardingThresholds tests the threshold constants
func TestOnboardingThresholds(t *testing.T) {
	// Test that our constants are reasonable
	if onboardingSimpleThreshold < 3 || onboardingSimpleThreshold > 7 {
		t.Errorf("onboardingSimpleThreshold should be between 3 and 7, got %v", onboardingSimpleThreshold)
	}
	
	if onboardingModerateThreshold < 8 || onboardingModerateThreshold > 12 {
		t.Errorf("onboardingModerateThreshold should be between 8 and 12, got %v", onboardingModerateThreshold)
	}
	
	if onboardingComplexThreshold < 15 || onboardingComplexThreshold > 25 {
		t.Errorf("onboardingComplexThreshold should be between 15 and 25, got %v", onboardingComplexThreshold)
	}
	
	if onboardingDefaultCommitLimit < 3 || onboardingDefaultCommitLimit > 10 {
		t.Errorf("onboardingDefaultCommitLimit should be between 3 and 10, got %v", onboardingDefaultCommitLimit)
	}
}

// TestIdentifyNewContributors tests new contributor identification logic
func TestIdentifyNewContributors(t *testing.T) {
	tests := []struct {
		name                  string
		commits              []MockCommitInfo
		expectedContributors []string
		expectedFirstCommits map[string]time.Time
	}{
		{
			name: "multiple_new_contributors",
			commits: []MockCommitInfo{
				{Author: "alice@example.com", Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Files: []string{"main.go"}},
				{Author: "bob@example.com", Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Files: []string{"utils.go"}},
				{Author: "alice@example.com", Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Files: []string{"main.go", "config.go"}},
				{Author: "carol@example.com", Time: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), Files: []string{"test.go"}},
			},
			expectedContributors: []string{"alice@example.com", "bob@example.com", "carol@example.com"},
			expectedFirstCommits: map[string]time.Time{
				"alice@example.com": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				"bob@example.com":   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				"carol@example.com": time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "single_contributor_multiple_commits",
			commits: []MockCommitInfo{
				{Author: "alice@example.com", Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Files: []string{"main.go"}},
				{Author: "alice@example.com", Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Files: []string{"utils.go"}},
				{Author: "alice@example.com", Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Files: []string{"config.go"}},
			},
			expectedContributors: []string{"alice@example.com"},
			expectedFirstCommits: map[string]time.Time{
				"alice@example.com": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:                 "no_commits",
			commits:              []MockCommitInfo{},
			expectedContributors: []string{},
			expectedFirstCommits: map[string]time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contributors, firstCommits := identifyNewContributors(tt.commits)
			
			if len(contributors) != len(tt.expectedContributors) {
				t.Errorf("identifyNewContributors() contributors count = %v, want %v", len(contributors), len(tt.expectedContributors))
			}
			
			for _, expected := range tt.expectedContributors {
				found := false
				for _, actual := range contributors {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("identifyNewContributors() missing contributor %v", expected)
				}
			}
			
			for email, expectedTime := range tt.expectedFirstCommits {
				if actualTime, exists := firstCommits[email]; !exists || !actualTime.Equal(expectedTime) {
					t.Errorf("identifyNewContributors() first commit time for %v = %v, want %v", email, actualTime, expectedTime)
				}
			}
		})
	}
}

// TestAnalyzeOnboardingFootprintMock tests the analysis logic using mock data
func TestAnalyzeOnboardingFootprintMock(t *testing.T) {
	tests := []struct {
		name                 string
		commits             []MockCommitInfo
		commitLimit         int
		expectedContributors int
		expectedAvgFiles    float64
		expectedStatus      map[string]string
	}{
		{
			name: "diverse_onboarding_patterns",
			commits: []MockCommitInfo{
				// Alice - simple onboarding (3 files in first 5 commits)
				{Author: "alice@example.com", Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Files: []string{"main.go"}},
				{Author: "alice@example.com", Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Files: []string{"utils.go"}},
				{Author: "alice@example.com", Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Files: []string{"config.go"}},
				// Bob - complex onboarding (15 files in first 5 commits)
				{Author: "bob@example.com", Time: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), Files: []string{"a.go", "b.go", "c.go", "d.go", "e.go"}},
				{Author: "bob@example.com", Time: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), Files: []string{"f.go", "g.go", "h.go", "i.go", "j.go"}},
				{Author: "bob@example.com", Time: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), Files: []string{"k.go", "l.go", "m.go", "n.go", "o.go"}},
			},
			commitLimit:         5,
			expectedContributors: 2,
			expectedAvgFiles:    9.0, // (3+15)/2
			expectedStatus: map[string]string{
				"alice@example.com": "Simple",
				"bob@example.com":   "Complex",
			},
		},
		{
			name: "commit_limit_enforced",
			commits: []MockCommitInfo{
				// Alice - only count first 2 commits when limit=2
				{Author: "alice@example.com", Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Files: []string{"main.go"}},
				{Author: "alice@example.com", Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Files: []string{"utils.go"}}, // Should count
				{Author: "alice@example.com", Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Files: []string{"ignored.go"}}, // Should be ignored due to limit
			},
			commitLimit:         2,
			expectedContributors: 1,
			expectedAvgFiles:    2.0, // Only main.go and utils.go counted
			expectedStatus: map[string]string{
				"alice@example.com": "Simple",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avgFiles, contributors := mockAnalyzeOnboardingFootprint(tt.commits, tt.commitLimit)
			
			if len(contributors) != tt.expectedContributors {
				t.Errorf("mockAnalyzeOnboardingFootprint() contributors count = %v, want %v", len(contributors), tt.expectedContributors)
			}
			
			tolerance := 0.01
			if floatDiff(avgFiles, tt.expectedAvgFiles) > tolerance {
				t.Errorf("mockAnalyzeOnboardingFootprint() avgFiles = %v, want %v", avgFiles, tt.expectedAvgFiles)
			}
			
			for email, expectedStatus := range tt.expectedStatus {
				found := false
				for _, contributor := range contributors {
					if contributor.Email == email {
						found = true
						if contributor.Status != expectedStatus {
							t.Errorf("mockAnalyzeOnboardingFootprint() status for %v = %v, want %v", email, contributor.Status, expectedStatus)
						}
						break
					}
				}
				if !found {
					t.Errorf("mockAnalyzeOnboardingFootprint() missing contributor %v", email)
				}
			}
		})
	}
}

// MockNewContributor represents a simplified new contributor for testing
type MockNewContributor struct {
	Email        string
	FilesCount   int
	CommitsCount int
}

// MockCommitInfo represents a simplified commit for testing
type MockCommitInfo struct {
	Author string
	Time   time.Time
	Files  []string
}

// floatDiff returns the absolute difference between two floats
func floatDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

// Mock functions for testing

// calculateOnboardingStats analyzes onboarding statistics from contributors
func calculateOnboardingStats(contributors []MockNewContributor) (float64, int, int, int, int) {
	if len(contributors) == 0 {
		return 0.0, 0, 0, 0, 0
	}
	
	var totalFiles int
	var simpleCnt, moderateCnt, complexCnt, overwhelmingCnt int
	
	for _, contributor := range contributors {
		totalFiles += contributor.FilesCount
		
		status, _ := classifyOnboardingComplexity(contributor.FilesCount)
		switch status {
		case "Simple":
			simpleCnt++
		case "Moderate":
			moderateCnt++
		case "Complex":
			complexCnt++
		case "Overwhelming":
			overwhelmingCnt++
		}
	}
	
	avgFiles := float64(totalFiles) / float64(len(contributors))
	return avgFiles, simpleCnt, moderateCnt, complexCnt, overwhelmingCnt
}

// identifyNewContributors identifies new contributors from commit history
func identifyNewContributors(commits []MockCommitInfo) ([]string, map[string]time.Time) {
	firstCommits := make(map[string]time.Time)
	
	for _, commit := range commits {
		if _, exists := firstCommits[commit.Author]; !exists {
			firstCommits[commit.Author] = commit.Time
		}
	}
	
	var contributors []string
	for email := range firstCommits {
		contributors = append(contributors, email)
	}
	
	return contributors, firstCommits
}

// mockAnalyzeOnboardingFootprint analyzes onboarding patterns for testing
func mockAnalyzeOnboardingFootprint(commits []MockCommitInfo, commitLimit int) (float64, []MockAnalyzedContributor) {
	contributors, firstCommits := identifyNewContributors(commits)
	
	var analyzedContributors []MockAnalyzedContributor
	var totalFiles int
	
	for _, email := range contributors {
		filesSet := make(map[string]bool)
		commitCount := 0
		
		for _, commit := range commits {
			if commit.Author == email && commitCount < commitLimit {
				for _, file := range commit.Files {
					filesSet[file] = true
				}
				commitCount++
			}
		}
		
		fileCount := len(filesSet)
		totalFiles += fileCount
		
		status, recommendation := classifyOnboardingComplexity(fileCount)
		
		analyzedContributors = append(analyzedContributors, MockAnalyzedContributor{
			Email:          email,
			FilesCount:     fileCount,
			CommitsCount:   commitCount,
			Status:         status,
			Recommendation: recommendation,
			FirstCommit:    firstCommits[email],
		})
	}
	
	var avgFiles float64
	if len(contributors) > 0 {
		avgFiles = float64(totalFiles) / float64(len(contributors))
	}
	
	return avgFiles, analyzedContributors
}

// MockAnalyzedContributor represents an analyzed contributor for testing
type MockAnalyzedContributor struct {
	Email          string
	FilesCount     int
	CommitsCount   int
	Status         string
	Recommendation string
	FirstCommit    time.Time
}
