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

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		// Positive cases - should be detected as test files
		{
			name:     "Go test file",
			filePath: "cmd/churn_test.go",
			expected: true,
		},
		{
			name:     "JavaScript test file",
			filePath: "src/components/Button.test.js",
			expected: true,
		},
		{
			name:     "TypeScript test file",
			filePath: "src/utils/helper.spec.ts",
			expected: true,
		},
		{
			name:     "Python test file in tests directory",
			filePath: "tests/test_models.py",
			expected: true,
		},
		{
			name:     "Python test file with test_ prefix",
			filePath: "models/test_user.py",
			expected: true,
		},
		{
			name:     "Python test file with _test suffix",
			filePath: "models/user_test.py",
			expected: true,
		},
		{
			name:     "Ruby test file",
			filePath: "spec/models/user_spec.rb",
			expected: true,
		},
		{
			name:     "Java test file",
			filePath: "src/test/java/UserTest.java",
			expected: true,
		},
		{
			name:     "C# test file",
			filePath: "Tests/UserTests.cs",
			expected: true,
		},
		{
			name:     "Jest test file",
			filePath: "src/__tests__/component.test.js",
			expected: true,
		},
		{
			name:     "TypeScript React test",
			filePath: "src/components/Button.test.tsx",
			expected: true,
		},
		{
			name:     "JavaScript spec file",
			filePath: "src/components/Button.spec.js",
			expected: true,
		},

		// Negative cases - should NOT be detected as test files
		{
			name:     "Regular Go file",
			filePath: "cmd/churn.go",
			expected: false,
		},
		{
			name:     "Regular JavaScript file",
			filePath: "src/components/Button.js",
			expected: false,
		},
		{
			name:     "Regular Python file",
			filePath: "models/user.py",
			expected: false,
		},
		{
			name:     "Config file",
			filePath: "config/database.yml",
			expected: false,
		},
		{
			name:     "Documentation",
			filePath: "README.md",
			expected: false,
		},
		{
			name:     "Package.json",
			filePath: "package.json",
			expected: false,
		},

		// False positive prevention cases
		{
			name:     "testdata.go should not be detected",
			filePath: "pkg/testdata.go",
			expected: false,
		},
		{
			name:     "contest.go should not be detected",
			filePath: "algorithms/contest.go",
			expected: false,
		},
		{
			name:     "latest.go should not be detected",
			filePath: "src/latest.go",
			expected: false,
		},
		{
			name:     "test.go without _test suffix should not be detected",
			filePath: "pkg/test.go",
			expected: false,
		},
		{
			name:     "testing.go should not be detected",
			filePath: "pkg/testing.go",
			expected: false,
		},
		{
			name:     "prototype.go should not be detected",
			filePath: "experiments/prototype.go",
			expected: false,
		},
		{
			name:     "fastest.js should not be detected",
			filePath: "src/fastest.js",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTestFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestCalculateTestRatio(t *testing.T) {
	tests := []struct {
		name           string
		testLOC        int
		sourceLOC      int
		expectedRatio  float64
		expectedStatus string
	}{
		{
			name:           "ideal ratio 1:1",
			testLOC:        1000,
			sourceLOC:      1000,
			expectedRatio:  1.0,
			expectedStatus: "Healthy",
		},
		{
			name:           "excellent ratio 1.5:1",
			testLOC:        1500,
			sourceLOC:      1000,
			expectedRatio:  1.5,
			expectedStatus: "Excellent",
		},
		{
			name:           "good ratio 2:1",
			testLOC:        2000,
			sourceLOC:      1000,
			expectedRatio:  2.0,
			expectedStatus: "Excellent",
		},
		{
			name:           "borderline ratio 0.8:1",
			testLOC:        800,
			sourceLOC:      1000,
			expectedRatio:  0.8,
			expectedStatus: "Caution",
		},
		{
			name:           "poor ratio 0.5:1",
			testLOC:        500,
			sourceLOC:      1000,
			expectedRatio:  0.5,
			expectedStatus: "Warning",
		},
		{
			name:           "very poor ratio 0.2:1",
			testLOC:        200,
			sourceLOC:      1000,
			expectedRatio:  0.2,
			expectedStatus: "Critical",
		},
		{
			name:           "no source code",
			testLOC:        1000,
			sourceLOC:      0,
			expectedRatio:  0.0,
			expectedStatus: "Unknown",
		},
		{
			name:           "no test code",
			testLOC:        0,
			sourceLOC:      1000,
			expectedRatio:  0.0,
			expectedStatus: "Critical",
		},
		{
			name:           "no code at all",
			testLOC:        0,
			sourceLOC:      0,
			expectedRatio:  0.0,
			expectedStatus: "Unknown",
		},
		{
			name:           "excessive tests over 2:1",
			testLOC:        3000,
			sourceLOC:      1000,
			expectedRatio:  3.0,
			expectedStatus: "Caution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio, status := calculateTestRatio(tt.testLOC, tt.sourceLOC)

			if ratio != tt.expectedRatio {
				t.Errorf("calculateTestRatio() ratio = %v, want %v", ratio, tt.expectedRatio)
			}

			if status != tt.expectedStatus {
				t.Errorf("calculateTestRatio() status = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}

func TestClassifyTestRatio(t *testing.T) {
	tests := []struct {
		name           string
		ratio          float64
		expectedStatus string
		expectedRec    string
	}{
		{
			name:           "critical - no tests",
			ratio:          0.0,
			expectedStatus: "Critical",
			expectedRec:    "Urgent: add comprehensive test coverage",
		},
		{
			name:           "warning - low coverage",
			ratio:          0.5,
			expectedStatus: "Warning",
			expectedRec:    "Increase test coverage significantly",
		},
		{
			name:           "caution - below ideal",
			ratio:          0.8,
			expectedStatus: "Caution",
			expectedRec:    "Consider adding more tests",
		},
		{
			name:           "healthy - ideal ratio",
			ratio:          1.0,
			expectedStatus: "Healthy",
			expectedRec:    "Good balance of tests and source code",
		},
		{
			name:           "excellent - strong coverage",
			ratio:          1.5,
			expectedStatus: "Excellent",
			expectedRec:    "Excellent test coverage",
		},
		{
			name:           "excellent - maximum recommended",
			ratio:          2.0,
			expectedStatus: "Excellent",
			expectedRec:    "Excellent test coverage",
		},
		{
			name:           "caution - excessive tests",
			ratio:          3.0,
			expectedStatus: "Caution",
			expectedRec:    "Consider reviewing test complexity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, rec := classifyTestRatio(tt.ratio)

			if status != tt.expectedStatus {
				t.Errorf("classifyTestRatio() status = %v, want %v", status, tt.expectedStatus)
			}

			if rec != tt.expectedRec {
				t.Errorf("classifyTestRatio() recommendation = %v, want %v", rec, tt.expectedRec)
			}
		})
	}
}

func TestFileTypeClassification(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedType string
	}{
		{
			name:         "Go source file",
			filePath:     "cmd/churn.go",
			expectedType: "source",
		},
		{
			name:         "Go test file",
			filePath:     "cmd/churn_test.go",
			expectedType: "test",
		},
		{
			name:         "JavaScript source",
			filePath:     "src/components/Button.js",
			expectedType: "source",
		},
		{
			name:         "JavaScript test",
			filePath:     "src/components/Button.test.js",
			expectedType: "test",
		},
		{
			name:         "Python source",
			filePath:     "models/user.py",
			expectedType: "source",
		},
		{
			name:         "Python test",
			filePath:     "tests/test_user.py",
			expectedType: "test",
		},
		{
			name:         "Documentation",
			filePath:     "README.md",
			expectedType: "other",
		},
		{
			name:         "Config file",
			filePath:     "config.yml",
			expectedType: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyFileType(tt.filePath)
			if result != tt.expectedType {
				t.Errorf("classifyFileType(%q) = %v, want %v", tt.filePath, result, tt.expectedType)
			}
		})
	}
}

func TestTestRatioThresholds(t *testing.T) {
	// Test that our constants match expected research-backed values
	if testRatioTargetThreshold != 1.0 {
		t.Errorf("Expected target test ratio threshold to be 1.0, got %f", testRatioTargetThreshold)
	}

	if testRatioMinimumThreshold != 0.5 {
		t.Errorf("Expected minimum test ratio threshold to be 0.5, got %f", testRatioMinimumThreshold)
	}
}

func TestTestRatioEdgeCases(t *testing.T) {
	t.Run("division by zero protection", func(t *testing.T) {
		ratio, status := calculateTestRatio(100, 0)
		if ratio != 0.0 {
			t.Errorf("Expected ratio 0.0 for zero source LOC, got %f", ratio)
		}
		if status != "Unknown" {
			t.Errorf("Expected 'Unknown' status for zero source LOC, got %s", status)
		}
	})

	t.Run("very large numbers", func(t *testing.T) {
		ratio, status := calculateTestRatio(1000000, 1000000)
		if ratio != 1.0 {
			t.Errorf("Expected ratio 1.0 for large equal numbers, got %f", ratio)
		}
		if status != "Healthy" {
			t.Errorf("Expected 'Healthy' status for 1:1 ratio, got %s", status)
		}
	})

	t.Run("small numbers precision", func(t *testing.T) {
		ratio, status := calculateTestRatio(1, 1)
		if ratio != 1.0 {
			t.Errorf("Expected ratio 1.0 for 1:1, got %f", ratio)
		}
		if status != "Healthy" {
			t.Errorf("Expected 'Healthy' status for 1:1 ratio, got %s", status)
		}
	})
}

func TestFloatEquals(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected bool
	}{
		{
			name:     "exact match",
			a:        1.0,
			b:        1.0,
			expected: true,
		},
		{
			name:     "within tolerance",
			a:        1.0,
			b:        1.0000000001,
			expected: true,
		},
		{
			name:     "outside tolerance",
			a:        1.0,
			b:        1.001,
			expected: false,
		},
		{
			name:     "zero comparison",
			a:        0.0,
			b:        0.0000000001,
			expected: true,
		},
		{
			name:     "negative numbers",
			a:        -1.0,
			b:        -1.0000000001,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := floatEquals(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("floatEquals(%f, %f) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
