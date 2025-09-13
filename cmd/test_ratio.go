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
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

const (
	// Test ratio thresholds based on findings from the Team Software Process (TSP) study:
	// Watts S. Humphrey, "A Discipline for Software Engineering," Addison-Wesley, 1995.
	testRatioTargetThreshold  = 1.0 // 1:1 ratio is target
	testRatioMinimumThreshold = 0.5 // Below 0.5:1 needs attention

	// Float comparison tolerance to handle precision issues
	floatTolerance = 1e-9
)

const testRatioBenchmarkContext = "Teams often plan to write as much unit test code as production code (TSP study)."

// floatEquals compares two floats with tolerance to handle precision issues
func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < floatTolerance
}

// TestRatioStats represents the test-to-code ratio analysis
type TestRatioStats struct {
	TestLOC        int
	SourceLOC      int
	OtherLOC       int
	TotalLOC       int
	TestRatio      float64
	Status         string
	Recommendation string
	TestFiles      int
	SourceFiles    int
	OtherFiles     int
	TotalFiles     int
}

// isTestFile determines if a file path represents a test file based on common patterns
func isTestFile(filePath string) bool {
	// Convert to lowercase for case-insensitive matching
	lowerPath := strings.ToLower(filePath)
	fileName := filepath.Base(lowerPath)
	dir := filepath.Dir(lowerPath)

	// Strict filename patterns
	if strings.HasSuffix(fileName, "_test.go") { // Go
		return true
	}
	if strings.HasSuffix(fileName, ".test.js") || strings.HasSuffix(fileName, ".test.jsx") ||
		strings.HasSuffix(fileName, ".test.ts") || strings.HasSuffix(fileName, ".test.tsx") {
		return true
	}
	if strings.HasSuffix(fileName, ".spec.js") || strings.HasSuffix(fileName, ".spec.jsx") ||
		strings.HasSuffix(fileName, ".spec.ts") || strings.HasSuffix(fileName, ".spec.tsx") {
		return true
	}
	if strings.HasSuffix(fileName, "_spec.rb") { // Ruby RSpec
		return true
	}
	// Python: test_*.py or *_test.py
	if strings.HasPrefix(fileName, "test_") && strings.HasSuffix(fileName, ".py") {
		return true
	}
	if strings.HasSuffix(fileName, "_test.py") {
		return true
	}
	// Java/C#: *Test.java, *Tests.java, *Test.cs, *Tests.cs
	if strings.HasSuffix(fileName, "test.java") || strings.HasSuffix(fileName, "tests.java") ||
		strings.HasSuffix(fileName, "test.cs") || strings.HasSuffix(fileName, "tests.cs") {
		return true
	}

	// Directory-based: match whole segments
	segments := strings.Split(strings.Trim(dir, string(filepath.Separator)), string(filepath.Separator))
	for i := range segments {
		seg := segments[i]
		if seg == "test" || seg == "tests" || seg == "spec" || seg == "specs" || seg == "__tests__" {
			return true
		}
	}

	return false
}

// classifyFileType classifies a file as source, test, or other
func classifyFileType(filePath string) string {
	// Store result to avoid duplicate computation
	isTest := isTestFile(filePath)
	if isTest {
		return "test"
	}

	// Check if it's a source code file
	ext := strings.ToLower(filepath.Ext(filePath))
	sourceExtensions := []string{
		".go", ".js", ".jsx", ".ts", ".tsx",
		".py", ".rb", ".java", ".cs", ".cpp", ".c", ".h",
		".php", ".swift", ".kt", ".scala", ".rs", ".dart",
	}

	for _, sourceExt := range sourceExtensions {
		if ext == sourceExt {
			return "source"
		}
	}

	return "other"
}

// calculateTestRatio calculates the test-to-code ratio and determines status
func calculateTestRatio(testLOC, sourceLOC int) (float64, string) {
	if sourceLOC == 0 {
		return 0.0, "Unknown"
	}

	ratio := float64(testLOC) / float64(sourceLOC)
	status, _ := classifyTestRatio(ratio)
	return ratio, status
}

// classifyTestRatio classifies the test ratio and provides recommendations
func classifyTestRatio(ratio float64) (string, string) {
	switch {
	case floatEquals(ratio, 0.0):
		return "Critical", "Urgent: add comprehensive test coverage"
	case ratio < 0.25:
		return "Critical", "Urgent: add comprehensive test coverage"
	case ratio < testRatioMinimumThreshold:
		return "Warning", "Increase test coverage significantly"
	case ratio < testRatioTargetThreshold:
		return "Caution", "Consider adding more tests to reach 1:1 ratio"
	case floatEquals(ratio, testRatioTargetThreshold):
		return "Healthy", "Good balance of tests and source code"
	case ratio <= 2.0:
		return "Excellent", "Excellent test coverage"
	default:
		return "Caution", "Consider reviewing test complexity"
	}
}

// analyzeTestRatio analyzes the test-to-code ratio in the repository
func analyzeTestRatio(repo *git.Repository, pathFilters []string) (*TestRatioStats, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD: %v", err)
	}

	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD commit: %v", err)
	}

	tree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD tree: %v", err)
	}

	stats := &TestRatioStats{}

	err = tree.Files().ForEach(func(f *object.File) error {
		// Apply path filter if specified
		if !matchesPathFilter(f.Name, pathFilters) {
			return nil
		}

		// Skip binary files
		isBinary, err := f.IsBinary()
		if err != nil || isBinary {
			return nil
		}

		// Get file content and count lines
		content, err := f.Contents()
		if err != nil {
			return nil
		}

		lineCount := countLines(content)
		fileType := classifyFileType(f.Name)

		switch fileType {
		case "test":
			stats.TestLOC += lineCount
			stats.TestFiles++
		case "source":
			stats.SourceLOC += lineCount
			stats.SourceFiles++
		case "other":
			stats.OtherLOC += lineCount
			stats.OtherFiles++
		}

		stats.TotalFiles++
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error analyzing files: %v", err)
	}

	stats.TotalLOC = stats.TestLOC + stats.SourceLOC + stats.OtherLOC
	stats.TestRatio, stats.Status = calculateTestRatio(stats.TestLOC, stats.SourceLOC)
	_, stats.Recommendation = classifyTestRatio(stats.TestRatio)

	return stats, nil
}

// printTestRatioStats prints the test ratio analysis results
func printTestRatioStats(stats *TestRatioStats, pathFilters []string) {
	fmt.Printf("Code vs Test Ratio Analysis\n")
	if len(pathFilters) > 0 {
		fmt.Printf("Path filters: %s\n", strings.Join(pathFilters, ", "))
	}
	fmt.Printf("Total files analyzed: %d\n", stats.TotalFiles)
	fmt.Printf("Source files: %d (%d LOC)\n", stats.SourceFiles, stats.SourceLOC)
	fmt.Printf("Test files: %d (%d LOC)\n", stats.TestFiles, stats.TestLOC)
	fmt.Printf("Other files: %d (%d LOC)\n", stats.OtherFiles, stats.OtherLOC)
	fmt.Println()

	fmt.Printf("Test-to-Code Ratio: %.2f:1 — %s\n", stats.TestRatio, stats.Status)
	fmt.Printf("Recommendation: %s\n", stats.Recommendation)
	fmt.Println()
	fmt.Println("Context:", testRatioBenchmarkContext)
	fmt.Println()

	// Provide detailed analysis
	fmt.Printf("Detailed Breakdown:\n")
	if stats.SourceLOC > 0 {
		testPercentage := (float64(stats.TestLOC) / float64(stats.SourceLOC)) * 100
		fmt.Printf("  Test coverage: %.1f%% (test LOC / source LOC)\n", testPercentage)
	}

	if stats.TotalLOC > 0 {
		sourcePercentage := (float64(stats.SourceLOC) / float64(stats.TotalLOC)) * 100
		testPercentage := (float64(stats.TestLOC) / float64(stats.TotalLOC)) * 100
		otherPercentage := (float64(stats.OtherLOC) / float64(stats.TotalLOC)) * 100

		fmt.Printf("  Source code: %.1f%% of total codebase\n", sourcePercentage)
		fmt.Printf("  Test code: %.1f%% of total codebase\n", testPercentage)
		fmt.Printf("  Other files: %.1f%% of total codebase\n", otherPercentage)
	}

	// Provide actionable insights
	fmt.Printf("\nHealthy Targets (based on Clean Code principles):\n")
	fmt.Printf("  • Ideal ratio: 1:1 to 2:1 (test:source)\n")
	fmt.Printf("  • Minimum acceptable: 0.75:1\n")
	fmt.Printf("  • Current ratio: %.2f:1\n", stats.TestRatio)

	if stats.TestRatio < testRatioTargetThreshold && stats.SourceLOC > 0 {
		needed := int(float64(stats.SourceLOC)*testRatioTargetThreshold) - stats.TestLOC
		if needed > 0 {
			fmt.Printf("  • Suggested: Add ~%d lines of test code to reach 1:1 ratio\n", needed)
		}
	}
}

// testRatioCmd represents the test-ratio command
var testRatioCmd = &cobra.Command{
	Use:   "test-ratio",
	Short: "Analyze test-to-code ratio",
	Long: `Analyze the ratio of test code to source code in your repository.
Helps ensure comprehensive test coverage following Clean Code principles.

A healthy test-to-code ratio is typically between 1:1 and 2:1, meaning you 
should have roughly equal to double the amount of test code compared to 
source code. This ensures thorough coverage without excessive overhead.

Classifications:
- Excellent: 1:1 to 2:1 ratio
- Healthy: 1:1 ratio  
- Caution: 0.75:1 to 1:1 ratio
- Warning: 0.5:1 to 0.75:1 ratio
- Critical: <0.5:1 ratio or no tests

"Test code is just as important as production code." — Robert C. Martin, Clean Code`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		pathFilters, source := getConfigPaths(cmd, "test-ratio.paths")

		// Print configuration scope
		printCommandScope(cmd, "test-ratio", "", pathFilters, source)

		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Could not open repository: %v", err)
		}

		stats, err := analyzeTestRatio(repo, pathFilters)
		if err != nil {
			log.Fatalf("Error analyzing test ratio: %v", err)
		}

		printTestRatioStats(stats, pathFilters)
	},
}

func init() {
	testRatioCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	rootCmd.AddCommand(testRatioCmd)
}
