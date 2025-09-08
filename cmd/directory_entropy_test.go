package cmd

import (
	"math"
	"strings"
	"testing"
)

func TestCalculateEntropy(t *testing.T) {
	tests := []struct {
		name        string
		fileTypes   map[string]int
		expected    float64
		description string
	}{
		{
			name:        "single file type",
			fileTypes:   map[string]int{"go": 10},
			expected:    0.0,
			description: "Single file type should have zero entropy",
		},
		{
			name:        "two equal file types",
			fileTypes:   map[string]int{"go": 5, "js": 5},
			expected:    1.0,
			description: "Two equal file types should have entropy of 1",
		},
		{
			name:        "three equal file types",
			fileTypes:   map[string]int{"go": 3, "js": 3, "py": 3},
			expected:    1.585, // log2(3) ≈ 1.585
			description: "Three equal file types should have entropy of log2(3)",
		},
		{
			name:        "mixed file types",
			fileTypes:   map[string]int{"go": 8, "js": 2},
			expected:    0.722, // -0.8*log2(0.8) - 0.2*log2(0.2) ≈ 0.722
			description: "Mixed file types should have entropy between 0 and 1",
		},
		{
			name:        "empty file types",
			fileTypes:   map[string]int{},
			expected:    0.0,
			description: "Empty file types should have zero entropy",
		},
		{
			name:        "many file types",
			fileTypes:   map[string]int{"go": 1, "js": 1, "py": 1, "rb": 1, "java": 1},
			expected:    2.322, // log2(5) ≈ 2.322
			description: "Five equal file types should have entropy of log2(5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateEntropy(tt.fileTypes)
			
			// Allow small floating point differences
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("Expected entropy %.3f, got %.3f. %s", 
					tt.expected, result, tt.description)
			}
		})
	}
}

func TestClassifyEntropyLevel(t *testing.T) {
	tests := []struct {
		name        string
		entropy     float64
		avgEntropy  float64
		dirPath     string
		expectedLevel string
		expectedRecContains string
	}{
		{
			name:        "critical entropy subdirectory (80% above avg)",
			entropy:     1.8, // 1.0 * 1.8 = 1.8 threshold
			avgEntropy:  1.0,
			dirPath:     "src",
			expectedLevel: "Critical",
			expectedRecContains: "refactoring",
		},
		{
			name:        "high entropy subdirectory (40% above avg)",
			entropy:     1.4, // 1.0 * 1.4 = 1.4 threshold
			avgEntropy:  1.0,
			dirPath:     "src",
			expectedLevel: "High",
			expectedRecContains: "refactoring",
		},
		{
			name:        "medium entropy subdirectory (10% below avg)",
			entropy:     0.9, // 1.0 * 0.9 = 0.9 threshold
			avgEntropy:  1.0,
			dirPath:     "src",
			expectedLevel: "Medium",
			expectedRecContains: "Monitor",
		},
		{
			name:        "low entropy subdirectory (well below avg)",
			entropy:     0.4,
			avgEntropy:  1.0,
			dirPath:     "src",
			expectedLevel: "Low",
			expectedRecContains: "Good",
		},
		{
			name:        "high entropy root directory (critical + offset)",
			entropy:     2.3, // (1.0 * 1.8) + 0.5 = 2.3 threshold
			avgEntropy:  1.0,
			dirPath:     "root",
			expectedLevel: "High",
			expectedRecContains: "organizing",
		},
		{
			name:        "medium entropy root directory (high + offset)",
			entropy:     1.7, // (1.0 * 1.4) + 0.3 = 1.7 threshold
			avgEntropy:  1.0,
			dirPath:     "root",
			expectedLevel: "Medium",
			expectedRecContains: "Acceptable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a generic project type for testing
			projectType := ProjectType{
				Name: "Test Project",
				RootPatterns: []string{".go", ".md", ".txt"},
				ExpectedDirs: map[string][]string{
					"src": {".go"},
				},
				Description: "Test project",
			}
			
			level, recommendation := classifyEntropyLevelWithContext(tt.entropy, tt.avgEntropy, tt.dirPath, projectType)
			
			if level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s", tt.expectedLevel, level)
			}
			
			if !strings.Contains(recommendation, tt.expectedRecContains) {
				t.Errorf("Expected recommendation to contain '%s', got '%s'", 
					tt.expectedRecContains, recommendation)
			}
		})
	}
}

func TestDirectoryEntropyStats(t *testing.T) {
	tests := []struct {
		name     string
		fileTypes map[string]int
		expectedEntropy float64
		expectedLevel   string
	}{
		{
			name:     "clean directory",
			fileTypes: map[string]int{"go": 10},
			expectedEntropy: 0.0,
			expectedLevel:   "Low",
		},
		{
			name:     "mixed directory",
			fileTypes: map[string]int{"go": 5, "js": 3, "py": 2, "rb": 1},
			expectedEntropy: 1.790, // Calculated entropy: -(5/11)*log2(5/11) - (3/11)*log2(3/11) - (2/11)*log2(2/11) - (1/11)*log2(1/11)
			expectedLevel:   "High", // Assuming avgEntropy = 1.0
		},
		{
			name:     "balanced directory",
			fileTypes: map[string]int{"go": 2, "js": 2},
			expectedEntropy: 1.0,
			expectedLevel:   "Medium", // Assuming avgEntropy = 1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &DirectoryEntropyStats{
				FileTypes: tt.fileTypes,
			}
			
			// Calculate file count
			fileCount := 0
			for _, count := range tt.fileTypes {
				fileCount += count
			}
			stats.FileCount = fileCount
			
			// Calculate entropy
			stats.Entropy = calculateEntropy(tt.fileTypes)
			
			// Test entropy calculation
			if math.Abs(stats.Entropy-tt.expectedEntropy) > 0.001 {
				t.Errorf("Expected entropy %.3f, got %.3f", tt.expectedEntropy, stats.Entropy)
			}
			
			// Test classification (using avgEntropy = 1.0)
			projectType := ProjectType{
				Name: "Test Project",
				RootPatterns: []string{".go", ".md", ".txt"},
				ExpectedDirs: map[string][]string{
					"src": {".go"},
				},
				Description: "Test project",
			}
			level, _ := classifyEntropyLevelWithContext(stats.Entropy, 1.0, "src", projectType)
			if level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s", tt.expectedLevel, level)
			}
		})
	}
}

func TestEntropyEdgeCases(t *testing.T) {
	t.Run("zero entropy with single file type", func(t *testing.T) {
		fileTypes := map[string]int{"go": 1}
		entropy := calculateEntropy(fileTypes)
		if entropy != 0.0 {
			t.Errorf("Expected entropy 0.0 for single file type, got %.3f", entropy)
		}
	})
	
	t.Run("maximum entropy with equal distribution", func(t *testing.T) {
		fileTypes := map[string]int{"go": 1, "js": 1, "py": 1, "rb": 1}
		entropy := calculateEntropy(fileTypes)
		expected := 2.0 // log2(4) = 2.0
		if math.Abs(entropy-expected) > 0.001 {
			t.Errorf("Expected entropy %.3f for 4 equal file types, got %.3f", expected, entropy)
		}
	})
	
	t.Run("classification with zero average entropy", func(t *testing.T) {
		projectType := ProjectType{
			Name: "Test Project",
			RootPatterns: []string{".go", ".md", ".txt"},
			ExpectedDirs: map[string][]string{
				"src": {".go"},
			},
			Description: "Test project",
		}
		level, recommendation := classifyEntropyLevelWithContext(1.0, 0.0, "src", projectType)
		if level != "High" {
			t.Errorf("Expected High level for entropy above zero average (minimum threshold), got %s", level)
		}
		if !strings.Contains(recommendation, "refactoring") {
			t.Errorf("Expected refactoring recommendation, got %s", recommendation)
		}
	})
}
