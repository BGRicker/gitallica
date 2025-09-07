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

func TestCalculateFileChurn(t *testing.T) {
	tests := []struct {
		name           string
		additions      int
		deletions      int
		totalLOC       int
		expectedChurn  float64
		expectedStatus string
	}{
		{
			name:           "healthy churn",
			additions:      20,
			deletions:      10,
			totalLOC:       1000,
			expectedChurn:  3.0,
			expectedStatus: "Healthy",
		},
		{
			name:           "caution churn",
			additions:      100,
			deletions:      50,
			totalLOC:       2000,
			expectedChurn:  7.5,
			expectedStatus: "Caution",
		},
		{
			name:           "warning churn",
			additions:      200,
			deletions:      100,
			totalLOC:       1000,
			expectedChurn:  30.0,
			expectedStatus: "Warning",
		},
		{
			name:           "zero total LOC",
			additions:      100,
			deletions:      50,
			totalLOC:       0,
			expectedChurn:  0.0,
			expectedStatus: "Healthy",
		},
		{
			name:           "exactly at threshold",
			additions:      100,
			deletions:      100,
			totalLOC:       1000,
			expectedChurn:  20.0,
			expectedStatus: "Caution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			churn, status := calculateFileChurn(tt.additions, tt.deletions, tt.totalLOC)
			
			if churn != tt.expectedChurn {
				t.Errorf("calculateFileChurn() churn = %v, want %v", churn, tt.expectedChurn)
			}
			
			if status != tt.expectedStatus {
				t.Errorf("calculateFileChurn() status = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}

func TestSortFilesByChurn(t *testing.T) {
	files := []FileChurnStats{
		{Path: "low.go", Additions: 10, Deletions: 5, TotalLOC: 1000, ChurnPercent: 1.5},
		{Path: "high.go", Additions: 200, Deletions: 100, TotalLOC: 500, ChurnPercent: 60.0},
		{Path: "medium.go", Additions: 50, Deletions: 25, TotalLOC: 300, ChurnPercent: 25.0},
	}

	sorted := sortFilesByChurn(files)

	// Should be sorted by churn percentage descending
	if sorted[0].Path != "high.go" {
		t.Errorf("Expected first file to be high.go, got %s", sorted[0].Path)
	}
	if sorted[1].Path != "medium.go" {
		t.Errorf("Expected second file to be medium.go, got %s", sorted[1].Path)
	}
	if sorted[2].Path != "low.go" {
		t.Errorf("Expected third file to be low.go, got %s", sorted[2].Path)
	}
}

func TestAggregateDirectoryChurn(t *testing.T) {
	files := []FileChurnStats{
		{Path: "src/main.go", Additions: 100, Deletions: 50, TotalLOC: 500, ChurnPercent: 30.0},
		{Path: "src/utils.go", Additions: 20, Deletions: 10, TotalLOC: 200, ChurnPercent: 15.0},
		{Path: "tests/test.go", Additions: 30, Deletions: 15, TotalLOC: 100, ChurnPercent: 45.0},
	}

	dirs := aggregateDirectoryChurn(files)

	// Should have src/ and tests/ directories
	if len(dirs) != 2 {
		t.Errorf("Expected 2 directories, got %d", len(dirs))
	}

	// Find src/ directory
	var srcDir *DirectoryChurnStats
	for _, dir := range dirs {
		if dir.Path == "src/" {
			srcDir = &dir
			break
		}
	}
	if srcDir == nil {
		t.Fatal("Expected to find src/ directory")
	}

	// src/ should have 120 additions, 60 deletions, 700 total LOC
	if srcDir.Additions != 120 {
		t.Errorf("Expected src/ additions to be 120, got %d", srcDir.Additions)
	}
	if srcDir.Deletions != 60 {
		t.Errorf("Expected src/ deletions to be 60, got %d", srcDir.Deletions)
	}
	if srcDir.TotalLOC != 700 {
		t.Errorf("Expected src/ total LOC to be 700, got %d", srcDir.TotalLOC)
	}
}

func TestChurnCalculationEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		additions      int
		deletions      int
		totalLOC       int
		expectedChurn  float64
		expectedStatus string
	}{
		{
			name:           "single line file with high churn",
			additions:      5,
			deletions:      5,
			totalLOC:       1,
			expectedChurn:  1000.0, // 10/1 * 100
			expectedStatus: "Warning",
		},
		{
			name:           "large file with low churn",
			additions:      1000,
			deletions:      1000,
			totalLOC:       100000,
			expectedChurn:  2.0,
			expectedStatus: "Healthy",
		},
		{
			name:           "only additions",
			additions:      100,
			deletions:      0,
			totalLOC:       1000,
			expectedChurn:  10.0,
			expectedStatus: "Caution",
		},
		{
			name:           "only deletions",
			additions:      0,
			deletions:      100,
			totalLOC:       1000,
			expectedChurn:  10.0,
			expectedStatus: "Caution",
		},
		{
			name:           "boundary case: exactly 5%",
			additions:      25,
			deletions:      25,
			totalLOC:       1000,
			expectedChurn:  5.0,
			expectedStatus: "Healthy",
		},
		{
			name:           "boundary case: just under 5%",
			additions:      24,
			deletions:      24,
			totalLOC:       1000,
			expectedChurn:  4.8,
			expectedStatus: "Healthy",
		},
		{
			name:           "boundary case: just over 20%",
			additions:      100,
			deletions:      100,
			totalLOC:       1000,
			expectedChurn:  20.0,
			expectedStatus: "Caution",
		},
		{
			name:           "boundary case: just under 20%",
			additions:      99,
			deletions:      99,
			totalLOC:       1000,
			expectedChurn:  19.8,
			expectedStatus: "Caution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			churn, status := calculateFileChurn(tt.additions, tt.deletions, tt.totalLOC)
			
			if churn != tt.expectedChurn {
				t.Errorf("calculateFileChurn() churn = %v, want %v", churn, tt.expectedChurn)
			}
			
			if status != tt.expectedStatus {
				t.Errorf("calculateFileChurn() status = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}

func TestDirectoryAggregationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		files    []FileChurnStats
		expected []DirectoryChurnStats
	}{
		{
			name: "empty file list",
			files: []FileChurnStats{},
			expected: []DirectoryChurnStats{},
		},
		{
			name: "single file",
			files: []FileChurnStats{
				{Path: "src/main.go", Additions: 100, Deletions: 50, TotalLOC: 500, ChurnPercent: 30.0},
			},
			expected: []DirectoryChurnStats{
				{Path: "src/", Additions: 100, Deletions: 50, TotalLOC: 500, ChurnPercent: 30.0, FileCount: 1},
			},
		},
		{
			name: "root directory files",
			files: []FileChurnStats{
				{Path: "main.go", Additions: 100, Deletions: 50, TotalLOC: 500, ChurnPercent: 30.0},
				{Path: "config.go", Additions: 20, Deletions: 10, TotalLOC: 200, ChurnPercent: 15.0},
			},
			expected: []DirectoryChurnStats{
				{Path: "root/", Additions: 120, Deletions: 60, TotalLOC: 700, ChurnPercent: 25.71, FileCount: 2},
			},
		},
		{
			name: "nested directories",
			files: []FileChurnStats{
				{Path: "src/main.go", Additions: 100, Deletions: 50, TotalLOC: 500, ChurnPercent: 30.0},
				{Path: "src/utils.go", Additions: 20, Deletions: 10, TotalLOC: 200, ChurnPercent: 15.0},
				{Path: "tests/test.go", Additions: 30, Deletions: 15, TotalLOC: 100, ChurnPercent: 45.0},
				{Path: "tests/integration/test.go", Additions: 10, Deletions: 5, TotalLOC: 50, ChurnPercent: 30.0},
			},
			expected: []DirectoryChurnStats{
				{Path: "tests/", Additions: 30, Deletions: 15, TotalLOC: 100, ChurnPercent: 45.0, FileCount: 1},
				{Path: "tests/integration/", Additions: 10, Deletions: 5, TotalLOC: 50, ChurnPercent: 30.0, FileCount: 1},
				{Path: "src/", Additions: 120, Deletions: 60, TotalLOC: 700, ChurnPercent: 25.71, FileCount: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aggregateDirectoryChurn(tt.files)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d directories, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing directory at index %d", i)
					continue
				}
				
				actual := result[i]
				if actual.Path != expected.Path {
					t.Errorf("Directory %d: expected path %s, got %s", i, expected.Path, actual.Path)
				}
				if actual.Additions != expected.Additions {
					t.Errorf("Directory %d: expected additions %d, got %d", i, expected.Additions, actual.Additions)
				}
				if actual.Deletions != expected.Deletions {
					t.Errorf("Directory %d: expected deletions %d, got %d", i, expected.Deletions, actual.Deletions)
				}
				if actual.TotalLOC != expected.TotalLOC {
					t.Errorf("Directory %d: expected total LOC %d, got %d", i, expected.TotalLOC, actual.TotalLOC)
				}
				if actual.FileCount != expected.FileCount {
					t.Errorf("Directory %d: expected file count %d, got %d", i, expected.FileCount, actual.FileCount)
				}
				// Allow small floating point differences
				if abs(actual.ChurnPercent - expected.ChurnPercent) > 0.1 {
					t.Errorf("Directory %d: expected churn %f, got %f", i, expected.ChurnPercent, actual.ChurnPercent)
				}
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
