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
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

const (
	// newProjectThreshold defines the age below which a project is considered "new"
	// and dead zone analysis should be skipped to prevent false positives
	newProjectThreshold = 90 * 24 * time.Hour // 90 days
)

// HealthIssue represents a single health issue found in the codebase
type HealthIssue struct {
	Category    string  // e.g., "Code Quality", "Performance", "Risk"
	Metric      string  // e.g., "churn", "bus-factor"
	Severity    string  // "Critical", "High", "Medium", "Low"
	Score       int     // 1-100 severity score
	Description string  // Human-readable description
	Recommendation string // Actionable recommendation
	Details     string  // Additional context or data
}

// HealthReport represents the overall health check results
type HealthReport struct {
	RepositoryPath string
	AnalysisTime   time.Time
	TimeWindow     string
	TotalIssues    int
	CriticalIssues int
	HighIssues     int
	MediumIssues   int
	LowIssues      int
	Issues         []HealthIssue
	Summary        string
}

// getSeverityScore converts severity level to numeric score for ranking
func getSeverityScore(severity string) int {
	switch severity {
	case "Critical":
		return 100
	case "High":
		return 75
	case "Medium":
		return 50
	case "Low":
		return 25
	case "Warning":
		return 60
	case "Caution":
		return 40
	default:
		return 0
	}
}

// categorizeIssue determines the category for an issue based on its metric
func categorizeIssue(metric string) string {
	switch metric {
	case "churn", "churn-files", "survival":
		return "Code Stability"
	case "bus-factor", "ownership-clarity":
		return "Knowledge Management"
	case "test-ratio", "high-risk-commits":
		return "Code Quality"
	case "dead-zones", "directory-entropy":
		return "Technical Debt"
	case "commit-cadence", "commit-size":
		return "Development Practices"
	case "change-lead-time", "long-lived-branches":
		return "DORA Performance"
	default:
		return "General"
	}
}

// analyzeChurnHealth checks churn patterns for issues
func analyzeChurnHealth(repo *git.Repository, since time.Time, pathFilters []string) []HealthIssue {
	var issues []HealthIssue
	
	// Run churn analysis
	ref, err := repo.Head()
	if err != nil {
		return issues
	}
	
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return issues
	}
	
	var additions, deletions int
	err = cIter.ForEach(func(c *object.Commit) error {
		if !since.IsZero() && c.Committer.When.Before(since) {
			return nil
		}
		// Use the existing function from the codebase
		additions_c, deletions_c, _, err := processCommitForSize(c, pathFilters)
		if err != nil {
			return nil // Skip commits with errors
		}
		additions += additions_c
		deletions += deletions_c
		return nil
	})
	if err != nil {
		return issues
	}
	
	// Calculate churn percentage
	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return issues
	}
	tree, err := headCommit.Tree()
	if err != nil {
		return issues
	}
	
	var totalLOC int
	tree.Files().ForEach(func(f *object.File) error {
		if !matchesPathFilter(f.Name, pathFilters) {
			return nil
		}
		isBinary, err := f.IsBinary()
		if err != nil || isBinary {
			return nil
		}
		content, err := f.Contents()
		if err != nil {
			return nil
		}
		totalLOC += countLines(content)
		return nil
	})
	
	if totalLOC > 0 {
		churnPercent := float64(additions+deletions) / float64(totalLOC) * 100
		
		if churnPercent > float64(churnCautionThreshold) {
			severity := "Medium"
			if churnPercent > float64(churnCautionThreshold)*2 {
				severity = "High"
			}
			
			issues = append(issues, HealthIssue{
				Category:      "Code Stability",
				Metric:        "churn",
				Severity:      severity,
				Score:         getSeverityScore(severity),
				Description:   fmt.Sprintf("High code churn detected: %.1f%%", churnPercent),
				Recommendation: "Review recent changes for architectural instability or frequent refactoring needs",
				Details:       fmt.Sprintf("Additions: %d, Deletions: %d, Total LOC: %d", additions, deletions, totalLOC),
			})
		}
	}
	
	return issues
}

// analyzeTestRatioHealth checks test coverage for issues
func analyzeTestRatioHealth(repo *git.Repository, pathFilters []string) []HealthIssue {
	var issues []HealthIssue
	
	stats, err := analyzeTestRatio(repo, pathFilters)
	if err != nil {
		return issues
	}
	
	if stats.TestRatio < testRatioMinimumThreshold {
		severity := "Critical"
		if stats.TestRatio > 0.25 {
			severity = "High"
		}
		
		issues = append(issues, HealthIssue{
			Category:      "Code Quality",
			Metric:        "test-ratio",
			Severity:      severity,
			Score:         getSeverityScore(severity),
			Description:   fmt.Sprintf("Low test coverage: %.2f:1 ratio", stats.TestRatio),
			Recommendation: "Increase test coverage significantly",
			Details:       fmt.Sprintf("Test LOC: %d, Source LOC: %d", stats.TestLOC, stats.SourceLOC),
		})
	}
	
	return issues
}

// analyzeBusFactorHealth checks knowledge concentration for issues
func analyzeBusFactorHealth(repo *git.Repository, since time.Time, pathFilters []string) []HealthIssue {
	var issues []HealthIssue
	
	analysis, err := analyzeBusFactor(repo, since, pathFilters)
	if err != nil {
		return issues
	}
	
	// Check for high-risk directories
	for _, dir := range analysis.OverallRiskDirs {
		if dir.RiskLevel == "Critical" || dir.RiskLevel == "High" {
			// For solo projects, adjust the messaging to be more appropriate
			recommendation := dir.Recommendation
			if len(dir.AuthorLines) == 1 {
				recommendation = "Consider: This is normal for solo projects. Plan for knowledge sharing as team grows."
			}
			
			issues = append(issues, HealthIssue{
				Category:      "Knowledge Management",
				Metric:        "bus-factor",
				Severity:      dir.RiskLevel,
				Score:         getSeverityScore(dir.RiskLevel),
				Description:   fmt.Sprintf("Knowledge concentration risk in %s", dir.Path),
				Recommendation: recommendation,
				Details:       fmt.Sprintf("Bus factor: %d, Contributors: %d", dir.BusFactor, len(dir.AuthorLines)),
			})
		}
	}
	
	return issues
}

// getRealProjectAge determines the actual age of the repository from its first commit
func getRealProjectAge(repo *git.Repository) (time.Duration, error) {
	ref, err := repo.Head()
	if err != nil {
		return 0, err
	}
	
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return 0, err
	}
	defer cIter.Close()
	
	var firstCommitTime time.Time
	err = cIter.ForEach(func(c *object.Commit) error {
		if firstCommitTime.IsZero() || c.Committer.When.Before(firstCommitTime) {
			firstCommitTime = c.Committer.When
		}
		return nil
	})
	
	if err != nil {
		return 0, err
	}
	
	return time.Since(firstCommitTime), nil
}

// analyzeDeadZonesHealth checks for stale code
func analyzeDeadZonesHealth(repo *git.Repository, since time.Time, pathFilters []string) []HealthIssue {
	var issues []HealthIssue
	
	// Skip dead zone analysis for very new projects (less than 3 months old)
	// This prevents false positives for newly created files
	projectAge, err := getRealProjectAge(repo)
	if err == nil && projectAge < newProjectThreshold {
		return issues // Skip dead zone analysis for new projects
	}
	
	analysis, err := analyzeDeadZones(repo, since, pathFilters)
	if err != nil {
		return issues
	}
	
	if analysis.DeadZoneCount > 0 {
		// Check for high-risk dead zones
		highRiskCount := 0
		for _, file := range analysis.DeadZoneFiles {
			if file.RiskLevel == "High Risk" {
				highRiskCount++
			}
		}
		
		if highRiskCount > 0 {
			issues = append(issues, HealthIssue{
				Category:      "Technical Debt",
				Metric:        "dead-zones",
				Severity:      "Medium",
				Score:         getSeverityScore("Medium"),
				Description:   fmt.Sprintf("%d high-risk dead zone files detected", highRiskCount),
				Recommendation: "Review and refactor or remove stale code",
				Details:       fmt.Sprintf("Total dead zones: %d (%.1f%% of codebase)", analysis.DeadZoneCount, analysis.DeadZonePercent),
			})
		}
	}
	
	return issues
}

// analyzeCommitSizeHealth checks for risky commits
func analyzeCommitSizeHealth(repo *git.Repository, since time.Time, pathFilters []string) []HealthIssue {
	var issues []HealthIssue
	
	ref, err := repo.Head()
	if err != nil {
		return issues
	}
	
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return issues
	}
	
	criticalCommits := 0
	highRiskCommits := 0
	
	err = cIter.ForEach(func(c *object.Commit) error {
		if !since.IsZero() && c.Committer.When.Before(since) {
			return nil
		}
		
		additions, deletions, filesChanged, err := processCommitForSize(c, pathFilters)
		if err != nil {
			return nil
		}
		
		riskLevel, _ := calculateCommitRisk(additions, deletions, filesChanged)
		if riskLevel == "Critical" {
			criticalCommits++
		} else if riskLevel == "High" {
			highRiskCommits++
		}
		
		return nil
	})
	
	if criticalCommits > 0 || highRiskCommits > 0 {
		severity := "Medium"
		if criticalCommits > 0 {
			severity = "High"
		}
		
		issues = append(issues, HealthIssue{
			Category:      "Development Practices",
			Metric:        "commit-size",
			Severity:      severity,
			Score:         getSeverityScore(severity),
			Description:   fmt.Sprintf("Large commits detected: %d critical, %d high-risk", criticalCommits, highRiskCommits),
			Recommendation: "Break down large commits into smaller, focused changes",
			Details:       "Large commits reduce review effectiveness and increase rollback risk",
		})
	}
	
	return issues
}

// generateHealthSummary creates a summary based on the issues found
func generateHealthSummary(report *HealthReport) string {
	if report.TotalIssues == 0 {
		return "✅ Excellent! No significant issues detected. Your codebase appears healthy."
	}
	
	var summary strings.Builder
	
	if report.CriticalIssues > 0 {
		issueWord := "issue"
		requireWord := "requires"
		if report.CriticalIssues != 1 {
			issueWord = "issues"
			requireWord = "require"
		}
		summary.WriteString(fmt.Sprintf("🚨 %d critical %s %s immediate attention. ", report.CriticalIssues, issueWord, requireWord))
	}
	if report.HighIssues > 0 {
		issueWord := "issue"
		if report.HighIssues != 1 {
			issueWord = "issues"
		}
		summary.WriteString(fmt.Sprintf("⚠️ %d high-priority %s should be addressed soon. ", report.HighIssues, issueWord))
	}
	if report.MediumIssues > 0 {
		issueWord := "issue"
		needWord := "needs"
		if report.MediumIssues != 1 {
			issueWord = "issues"
			needWord = "need"
		}
		summary.WriteString(fmt.Sprintf("📋 %d medium-priority %s %s attention. ", report.MediumIssues, issueWord, needWord))
	}
	if report.LowIssues > 0 {
		issueWord := "issue"
		if report.LowIssues != 1 {
			issueWord = "issues"
		}
		summary.WriteString(fmt.Sprintf("💡 %d low-priority %s can be addressed when convenient.", report.LowIssues, issueWord))
	}
	
	return summary.String()
}

// printHealthReport prints the health check results in a formatted way
func printHealthReport(report *HealthReport) {
	fmt.Printf("🏥 Gitallica Health Check Report\n")
	fmt.Printf("Repository: %s\n", report.RepositoryPath)
	fmt.Printf("Analysis Time: %s\n", report.AnalysisTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Time Window: %s\n", report.TimeWindow)
	fmt.Printf("Total Issues Found: %d\n", report.TotalIssues)
	fmt.Println()
	
	// Summary
	fmt.Printf("📊 Summary:\n%s\n\n", report.Summary)
	
	if report.TotalIssues == 0 {
		return
	}
	
	// Group issues by category
	categories := make(map[string][]HealthIssue)
	for _, issue := range report.Issues {
		categories[issue.Category] = append(categories[issue.Category], issue)
	}
	
	// Print issues by category
	for category, categoryIssues := range categories {
		fmt.Printf("📁 %s (%d issues)\n", category, len(categoryIssues))
		fmt.Printf("%s\n", strings.Repeat("-", len(category)+20))
		
		for _, issue := range categoryIssues {
			emoji := "🔴"
			switch issue.Severity {
			case "Critical":
				emoji = "🔴"
			case "High":
				emoji = "🟠"
			case "Medium":
				emoji = "🟡"
			case "Low":
				emoji = "🔵"
			}
			
			fmt.Printf("%s [%s] %s\n", emoji, issue.Severity, issue.Description)
			fmt.Printf("   💡 %s\n", issue.Recommendation)
			if issue.Details != "" {
				fmt.Printf("   📋 %s\n", issue.Details)
			}
			fmt.Println()
		}
	}
	
	// Top 3 priorities
	if len(report.Issues) > 0 {
		fmt.Printf("🎯 Top 3 Priorities:\n")
		for i, issue := range report.Issues[:min(3, len(report.Issues))] {
			fmt.Printf("%d. [%s] %s\n", i+1, issue.Severity, issue.Description)
		}
	}
}

// performHealthCheck runs all health checks and returns a comprehensive report
func performHealthCheck(repo *git.Repository, since time.Time, pathFilters []string) (*HealthReport, error) {
	var allIssues []HealthIssue
	
	// Run all health checks
	allIssues = append(allIssues, analyzeChurnHealth(repo, since, pathFilters)...)
	allIssues = append(allIssues, analyzeTestRatioHealth(repo, pathFilters)...)
	allIssues = append(allIssues, analyzeBusFactorHealth(repo, since, pathFilters)...)
	allIssues = append(allIssues, analyzeDeadZonesHealth(repo, since, pathFilters)...)
	allIssues = append(allIssues, analyzeCommitSizeHealth(repo, since, pathFilters)...)
	
	// Sort issues by severity score (highest first)
	sort.Slice(allIssues, func(i, j int) bool {
		return allIssues[i].Score > allIssues[j].Score
	})
	
	// Count issues by severity
	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0
	
	for _, issue := range allIssues {
		switch issue.Severity {
		case "Critical":
			criticalCount++
		case "High":
			highCount++
		case "Medium":
			mediumCount++
		case "Low":
			lowCount++
		}
	}
	
	timeWindow := "all time"
	if !since.IsZero() {
		timeWindow = fmt.Sprintf("since %s", since.Format("2006-01-02"))
	}
	
	report := &HealthReport{
		RepositoryPath: ".",
		AnalysisTime:   time.Now(),
		TimeWindow:     timeWindow,
		TotalIssues:    len(allIssues),
		CriticalIssues: criticalCount,
		HighIssues:     highCount,
		MediumIssues:   mediumCount,
		LowIssues:      lowCount,
		Issues:         allIssues,
	}
	
	report.Summary = generateHealthSummary(report)
	
	return report, nil
}

// healthCheckCmd represents the health-check command
var healthCheckCmd = &cobra.Command{
	Use:   "health-check",
	Short: "Run comprehensive health check and identify top codebase issues",
	Long: `Perform a comprehensive analysis of your codebase health by running all available
metrics and identifying only the issues that exceed healthy thresholds.

This command provides a prioritized list of problems that need attention, omitting
all healthy metrics to focus on actionable insights. Perfect for getting a quick
overview of your codebase's biggest concerns.

The health check covers:
- Code stability (churn patterns)
- Code quality (test coverage)
- Knowledge management (bus factor, ownership)
- Technical debt (dead zones)
- Development practices (commit size)

Issues are ranked by severity and categorized for easy prioritization.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		lastArg, _ := cmd.Flags().GetString("last")
		pathFilters, source := getConfigPaths(cmd, "health-check.paths")
		
		// Print configuration scope
		printCommandScope(cmd, "health-check", lastArg, pathFilters, source)

		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Could not open repository: %v", err)
		}

		// Parse time window
		since := time.Time{}
		if lastArg != "" {
			cutoff, err := parseDurationArg(lastArg)
			if err != nil {
				log.Fatalf("Could not parse --last argument: %v", err)
			}
			since = cutoff
		}

		report, err := performHealthCheck(repo, since, pathFilters)
		if err != nil {
			log.Fatalf("Error performing health check: %v", err)
		}

		printHealthReport(report)
	},
}

func init() {
	healthCheckCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	healthCheckCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	rootCmd.AddCommand(healthCheckCmd)
}
