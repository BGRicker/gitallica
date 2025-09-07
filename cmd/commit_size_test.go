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

func TestCalculateCommitRisk(t *testing.T) {
	tests := []struct {
		name           string
		additions      int
		deletions      int
		filesChanged   int
		expectedRisk   string
		expectedScore  int
	}{
		{
			name:           "small safe commit",
			additions:      50,
			deletions:      25,
			filesChanged:   3,
			expectedRisk:   "Medium",
			expectedScore:  105, // 75 + (3 * 10)
		},
		{
			name:           "medium commit",
			additions:      200,
			deletions:      100,
			filesChanged:   8,
			expectedRisk:   "High",
			expectedScore:  380, // 300 + (8 * 10)
		},
		{
			name:           "large risky commit",
			additions:      500,
			deletions:      200,
			filesChanged:   15,
			expectedRisk:   "Critical",
			expectedScore:  850, // 700 + (15 * 10)
		},
		{
			name:           "very large commit",
			additions:      1000,
			deletions:      500,
			filesChanged:   25,
			expectedRisk:   "Critical",
			expectedScore:  1750, // 1500 + (25 * 10)
		},
		{
			name:           "many files but small changes",
			additions:      100,
			deletions:      50,
			filesChanged:   20,
			expectedRisk:   "Critical",
			expectedScore:  350, // 150 + (20 * 10)
		},
		{
			name:           "few files but large changes",
			additions:      800,
			deletions:      200,
			filesChanged:   2,
			expectedRisk:   "Critical",
			expectedScore:  1020, // 1000 + (2 * 10)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk, score := calculateCommitRisk(tt.additions, tt.deletions, tt.filesChanged)
			
			if risk != tt.expectedRisk {
				t.Errorf("calculateCommitRisk() risk = %v, want %v", risk, tt.expectedRisk)
			}
			
			if score != tt.expectedScore {
				t.Errorf("calculateCommitRisk() score = %v, want %v", score, tt.expectedScore)
			}
		})
	}
}

func TestSortCommitsByRisk(t *testing.T) {
	commits := []CommitSizeStats{
		{Hash: "abc123", Message: "small fix", Additions: 10, Deletions: 5, FilesChanged: 1, RiskScore: 15, RiskLevel: "Low"},
		{Hash: "def456", Message: "large feature", Additions: 500, Deletions: 200, FilesChanged: 10, RiskScore: 700, RiskLevel: "High"},
		{Hash: "ghi789", Message: "medium refactor", Additions: 100, Deletions: 50, FilesChanged: 5, RiskScore: 150, RiskLevel: "Medium"},
	}

	sorted := sortCommitsByRisk(commits)

	// Should be sorted by risk score descending
	if sorted[0].Hash != "def456" {
		t.Errorf("Expected first commit to be def456, got %s", sorted[0].Hash)
	}
	if sorted[1].Hash != "ghi789" {
		t.Errorf("Expected second commit to be ghi789, got %s", sorted[1].Hash)
	}
	if sorted[2].Hash != "abc123" {
		t.Errorf("Expected third commit to be abc123, got %s", sorted[2].Hash)
	}
}

func TestFilterCommitsByRisk(t *testing.T) {
	commits := []CommitSizeStats{
		{Hash: "abc123", Message: "small fix", RiskLevel: "Low"},
		{Hash: "def456", Message: "medium change", RiskLevel: "Medium"},
		{Hash: "ghi789", Message: "large feature", RiskLevel: "High"},
		{Hash: "jkl012", Message: "huge refactor", RiskLevel: "Critical"},
	}

	tests := []struct {
		name           string
		minRisk        string
		expectedCount  int
		expectedHashes []string
	}{
		{
			name:           "all commits",
			minRisk:        "Low",
			expectedCount:  4,
			expectedHashes: []string{"abc123", "def456", "ghi789", "jkl012"},
		},
		{
			name:           "medium and above",
			minRisk:        "Medium",
			expectedCount:  3,
			expectedHashes: []string{"def456", "ghi789", "jkl012"},
		},
		{
			name:           "high and above",
			minRisk:        "High",
			expectedCount:  2,
			expectedHashes: []string{"ghi789", "jkl012"},
		},
		{
			name:           "critical only",
			minRisk:        "Critical",
			expectedCount:  1,
			expectedHashes: []string{"jkl012"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterCommitsByRisk(commits, tt.minRisk)
			
			if len(filtered) != tt.expectedCount {
				t.Errorf("Expected %d commits, got %d", tt.expectedCount, len(filtered))
			}
			
			for i, expectedHash := range tt.expectedHashes {
				if i >= len(filtered) {
					t.Errorf("Missing commit at index %d", i)
					continue
				}
				if filtered[i].Hash != expectedHash {
					t.Errorf("Expected commit %d to be %s, got %s", i, expectedHash, filtered[i].Hash)
				}
			}
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		maxLen   int
		expected string
	}{
		{
			name:     "short message",
			message:  "fix bug",
			maxLen:   50,
			expected: "fix bug",
		},
		{
			name:     "exactly max length",
			message:  "This is exactly fifty characters long message here",
			maxLen:   50,
			expected: "This is exactly fifty characters long message here",
		},
		{
			name:     "long message",
			message:  "This is a very long commit message that should be truncated because it exceeds the maximum length",
			maxLen:   50,
			expected: "This is a very long commit message that should ...",
		},
		{
			name:     "very long message",
			message:  "This is an extremely long commit message that goes on and on and on and should definitely be truncated",
			maxLen:   30,
			expected: "This is an extremely long c...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateMessage(tt.message, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}
