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

func TestCalculateFileAge(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name           string
		lastModified   time.Time
		referenceTime  time.Time
		expectedMonths int
	}{
		{
			name:           "file modified 6 months ago",
			lastModified:   now.AddDate(0, -6, 0),
			referenceTime:  now,
			expectedMonths: 6,
		},
		{
			name:           "file modified 12 months ago",
			lastModified:   now.AddDate(0, -12, 0),
			referenceTime:  now,
			expectedMonths: 12,
		},
		{
			name:           "file modified 18 months ago",
			lastModified:   now.AddDate(0, -18, 0),
			referenceTime:  now,
			expectedMonths: 18,
		},
		{
			name:           "file modified yesterday",
			lastModified:   now.AddDate(0, 0, -1),
			referenceTime:  now,
			expectedMonths: 0,
		},
		{
			name:           "file modified exactly 1 year ago",
			lastModified:   now.AddDate(-1, 0, 0),
			referenceTime:  now,
			expectedMonths: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			months := calculateFileAge(tt.lastModified, tt.referenceTime)
			if months != tt.expectedMonths {
				t.Errorf("Expected %d months, got %d", tt.expectedMonths, months)
			}
		})
	}
}

func TestIsDeadZone(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name         string
		lastModified time.Time
		referenceTime time.Time
		expected     bool
		description  string
	}{
		{
			name:         "fresh file (1 month old)",
			lastModified: now.AddDate(0, -1, 0),
			referenceTime: now,
			expected:     false,
			description:  "Recent files should not be dead zones",
		},
		{
			name:         "borderline file (11 months old)",
			lastModified: now.AddDate(0, -11, 0),
			referenceTime: now,
			expected:     false,
			description:  "Files just under 12 months should not be dead zones",
		},
		{
			name:         "exactly 12 months old",
			lastModified: now.AddDate(0, -12, 0),
			referenceTime: now,
			expected:     true,
			description:  "Files exactly 12 months old should be dead zones",
		},
		{
			name:         "old file (18 months old)",
			lastModified: now.AddDate(0, -18, 0),
			referenceTime: now,
			expected:     true,
			description:  "Files older than 12 months should be dead zones",
		},
		{
			name:         "very old file (2 years old)",
			lastModified: now.AddDate(-2, 0, 0),
			referenceTime: now,
			expected:     true,
			description:  "Very old files should definitely be dead zones",
		},
		{
			name:         "future file (edge case)",
			lastModified: now.AddDate(0, 1, 0),
			referenceTime: now,
			expected:     false,
			description:  "Future dates should not be dead zones",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDeadZone(tt.lastModified, tt.referenceTime)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v. %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestClassifyDeadZoneRisk(t *testing.T) {
	tests := []struct {
		name          string
		ageInMonths   int
		expectedLevel string
		expectedRec   string
	}{
		{
			name:          "not a dead zone",
			ageInMonths:   6,
			expectedLevel: "Active",
			expectedRec:   "regularly maintained",
		},
		{
			name:          "low risk dead zone",
			ageInMonths:   14,
			expectedLevel: "Low Risk",
			expectedRec:   "Consider reviewing",
		},
		{
			name:          "medium risk dead zone",
			ageInMonths:   20,
			expectedLevel: "Medium Risk",
			expectedRec:   "Needs attention",
		},
		{
			name:          "high risk dead zone",
			ageInMonths:   30,
			expectedLevel: "High Risk",
			expectedRec:   "Refactor or remove",
		},
		{
			name:          "critical dead zone",
			ageInMonths:   40,
			expectedLevel: "Critical",
			expectedRec:   "Urgent: refactor or delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, rec := classifyDeadZoneRisk(tt.ageInMonths)
			if level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s", tt.expectedLevel, level)
			}
			if rec != tt.expectedRec {
				t.Errorf("Expected recommendation %s, got %s", tt.expectedRec, rec)
			}
		})
	}
}

func TestDeadZoneFileStats(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name         string
		path         string
		lastModified time.Time
		size         int64
		referenceTime time.Time
		expectedDead bool
		expectedAge  int
	}{
		{
			name:         "small active file",
			path:         "src/active.go",
			lastModified: now.AddDate(0, -3, 0),
			size:         1000,
			referenceTime: now,
			expectedDead: false,
			expectedAge:  3,
		},
		{
			name:         "large dead file",
			path:         "legacy/old.go",
			lastModified: now.AddDate(0, -15, 0),
			size:         5000,
			referenceTime: now,
			expectedDead: true,
			expectedAge:  15,
		},
		{
			name:         "config file dead zone",
			path:         "config/deprecated.yaml",
			lastModified: now.AddDate(0, -24, 0),
			size:         500,
			referenceTime: now,
			expectedDead: true,
			expectedAge:  24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &DeadZoneFileStats{
				Path:         tt.path,
				LastModified: tt.lastModified,
				Size:         tt.size,
			}
			
			// Calculate age and dead zone status
			age := calculateFileAge(tt.lastModified, tt.referenceTime)
			isDead := isDeadZone(tt.lastModified, tt.referenceTime)
			
			if age != tt.expectedAge {
				t.Errorf("Expected age %d months, got %d", tt.expectedAge, age)
			}
			
			if isDead != tt.expectedDead {
				t.Errorf("Expected dead zone status %v, got %v", tt.expectedDead, isDead)
			}
			
			if stats.Path != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, stats.Path)
			}
		})
	}
}

func TestSortDeadZonesByAge(t *testing.T) {
	now := time.Now()
	
	files := []DeadZoneFileStats{
		{Path: "new.go", LastModified: now.AddDate(0, -13, 0), AgeInMonths: 13},
		{Path: "very_old.go", LastModified: now.AddDate(0, -30, 0), AgeInMonths: 30},
		{Path: "old.go", LastModified: now.AddDate(0, -20, 0), AgeInMonths: 20},
	}

	sorted := sortDeadZonesByAge(files)

	// Should be sorted by age descending (oldest first)
	if sorted[0].Path != "very_old.go" {
		t.Errorf("Expected first file to be very_old.go, got %s", sorted[0].Path)
	}
	if sorted[1].Path != "old.go" {
		t.Errorf("Expected second file to be old.go, got %s", sorted[1].Path)
	}
	if sorted[2].Path != "new.go" {
		t.Errorf("Expected third file to be new.go, got %s", sorted[2].Path)
	}
}

func TestDeadZoneThresholds(t *testing.T) {
	// Test that our constants match expected research-backed values
	if deadZoneThresholdMonths != 12 {
		t.Errorf("Expected dead zone threshold to be 12 months, got %d", deadZoneThresholdMonths)
	}
	
	if deadZoneLowRiskThresholdMonths != 24 {
		t.Errorf("Expected low risk threshold to be 24 months, got %d", deadZoneLowRiskThresholdMonths)
	}
	
	if deadZoneHighRiskThresholdMonths != 36 {
		t.Errorf("Expected high risk threshold to be 36 months, got %d", deadZoneHighRiskThresholdMonths)
	}
}

func TestDeadZoneEdgeCases(t *testing.T) {
	now := time.Now()
	
	t.Run("zero time", func(t *testing.T) {
		age := calculateFileAge(time.Time{}, now)
		// Should handle zero time gracefully
		if age < 0 {
			t.Errorf("Expected non-negative age for zero time, got %d", age)
		}
	})
	
	t.Run("same time", func(t *testing.T) {
		age := calculateFileAge(now, now)
		if age != 0 {
			t.Errorf("Expected 0 months for same time, got %d", age)
		}
	})
	
	t.Run("future time", func(t *testing.T) {
		future := now.AddDate(0, 6, 0)
		age := calculateFileAge(future, now)
		if age < 0 {
			t.Errorf("Expected non-negative age for future time, got %d", age)
		}
	})
}
