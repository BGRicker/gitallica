package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

// Thresholds based on research: >400 lines or >10-12 files per commit is high risk
const (
	// Risk thresholds based on Nokia Bell Labs research showing correlation 
	// between change size and integration problems
	moderateRiskLinesThreshold = 200  // Moderate complexity
	highRiskLinesThreshold     = 500  // High complexity
	criticalRiskLinesThreshold = 1000 // Critical complexity
	
	// File change thresholds
	moderateRiskFilesThreshold = 5    // Moderate complexity
	highRiskFilesThreshold     = 10   // High complexity
	criticalRiskFilesThreshold = 20   // Critical complexity
)

// HighRiskCommit represents a commit with risk analysis
type HighRiskCommit struct {
	Hash         string
	Author       string
	Date         time.Time
	Message      string
	LinesChanged int // Total additions + deletions
	FilesChanged int
	Risk         string
	Reason       string
}

// HighRiskCommitsStats contains analysis statistics
type HighRiskCommitsStats struct {
	TotalCommits   int
	LowRisk        int
	ModerateRisk   int
	HighRisk       int
	CriticalRisk   int
	AverageLines   float64
	AverageFiles   float64
	LargestCommit  HighRiskCommit
	RiskyCommits   []HighRiskCommit // Only moderate+ risk commits
}

var highRiskCommitsCmd = &cobra.Command{
	Use:   "high-risk-commits",
	Short: "Analyze high-risk commits (monster commits that touch everything)",
	Long: `Identifies commits that pose high risk due to their size or complexity.

Large commits reduce review effectiveness and rollback safety. This command analyzes
commits based on lines changed and number of files touched to identify potential risks.

Research basis:
- "All changes are small. There are only longer and shorter feedback cycles." — Kent Beck
- "Refactoring changes the program in small steps." — Martin Fowler

Thresholds:
- High Risk: >400 lines changed OR ≥12 files touched
- Critical Risk: ≥800 lines changed OR ≥20 files touched

The analysis helps identify commits that may need extra review attention or
architectural consideration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.PlainOpen(".")
		if err != nil {
			return fmt.Errorf("could not open repository: %v", err)
		}

		pathArg, _ := cmd.Flags().GetString("path")
		lastArg, _ := cmd.Flags().GetString("last")
		limitArg, _ := cmd.Flags().GetInt("limit")

		stats, err := analyzeHighRiskCommits(repo, pathArg, lastArg, limitArg)
		if err != nil {
			return err
		}

		printHighRiskCommitsStats(stats, limitArg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(highRiskCommitsCmd)
	highRiskCommitsCmd.Flags().String("last", "", "Specify the time window to analyze (e.g., 30d, 6m, 1y)")
	highRiskCommitsCmd.Flags().String("path", "", "Limit analysis to a specific directory or path")
	highRiskCommitsCmd.Flags().Int("limit", 10, "Number of risky commits to show in detailed output")
}

// classifyCommitRisk determines the risk level and reason for a commit
func classifyCommitRisk(linesChanged, filesChanged int) (string, string) {
	// Critical risk - very large changes
	if linesChanged >= criticalRiskLinesThreshold || filesChanged >= criticalRiskFilesThreshold {
		return "Critical", "Very large changes increase integration risk"
	}
	
	// High risk - large changes that need careful review
	if linesChanged >= highRiskLinesThreshold || filesChanged >= highRiskFilesThreshold {
		return "High", "Large changes increase review complexity"
	}
	
	// Moderate risk - sizeable commits
	if linesChanged >= moderateRiskLinesThreshold || filesChanged >= moderateRiskFilesThreshold {
		return "Moderate", "Moderate complexity requires careful review"
	}
	
	// Low risk - small, focused commits
	return "Low", "Small changes are easier to review and debug"
}

// calculateHighRiskCommitsStats computes statistics for the commit analysis
func calculateHighRiskCommitsStats(commits []HighRiskCommit) *HighRiskCommitsStats {
	stats := &HighRiskCommitsStats{
		TotalCommits: len(commits),
	}
	
	if len(commits) == 0 {
		return stats
	}
	
	totalLines := 0
	totalFiles := 0
	var largestCommit HighRiskCommit
	var riskyCommits []HighRiskCommit
	
	for _, commit := range commits {
		// Count risk categories
		switch commit.Risk {
		case "Low":
			stats.LowRisk++
		case "Moderate":
			stats.ModerateRisk++
			riskyCommits = append(riskyCommits, commit)
		case "High":
			stats.HighRisk++
			riskyCommits = append(riskyCommits, commit)
		case "Critical":
			stats.CriticalRisk++
			riskyCommits = append(riskyCommits, commit)
		}
		
		// Accumulate for averages
		totalLines += commit.LinesChanged
		totalFiles += commit.FilesChanged
		
		// Track largest commit
		if commit.LinesChanged > largestCommit.LinesChanged {
			largestCommit = commit
		}
	}
	
	// Calculate averages
	stats.AverageLines = float64(totalLines) / float64(len(commits))
	stats.AverageFiles = float64(totalFiles) / float64(len(commits))
	stats.LargestCommit = largestCommit
	
	// Sort risky commits by lines changed (descending)
	sort.Slice(riskyCommits, func(i, j int) bool {
		return riskyCommits[i].LinesChanged > riskyCommits[j].LinesChanged
	})
	stats.RiskyCommits = riskyCommits
	
	return stats
}

// analyzeHighRiskCommits performs the main analysis
func analyzeHighRiskCommits(repo *git.Repository, pathArg string, lastArg string, limitArg int) (*HighRiskCommitsStats, error) {
	var since *time.Time
	if lastArg != "" {
		sinceTime, err := parseDurationArg(lastArg)
		if err != nil {
			return nil, fmt.Errorf("invalid time window: %v", err)
		}
		since = &sinceTime
	}
	
	// Get commits within time window
	commitIter, err := repo.Log(&git.LogOptions{
		Since: since,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get commit log: %v", err)
	}
	defer commitIter.Close()
	
	var commits []HighRiskCommit
	
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Skip commits without author information
		if commit.Author.Email == "" {
			return nil
		}
		
		// Skip merge commits for cleaner analysis
		if commit.NumParents() > 1 {
			return nil
		}
		
		// Calculate lines and files changed
		linesChanged, filesChanged, err := calculateCommitChanges(commit, pathArg)
		if err != nil {
			return err
		}
		
		// Skip commits with no changes in the specified path
		if filesChanged == 0 {
			return nil
		}
		
		// Classify risk
		risk, reason := classifyCommitRisk(linesChanged, filesChanged)
		
		commits = append(commits, HighRiskCommit{
			Hash:         commit.Hash.String()[:8],
			Author:       commit.Author.Email,
			Date:         commit.Author.When,
			Message:      commit.Message,
			LinesChanged: linesChanged,
			FilesChanged: filesChanged,
			Risk:         risk,
			Reason:       reason,
		})
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing commits: %v", err)
	}
	
	stats := calculateHighRiskCommitsStats(commits)
	return stats, nil
}

// calculateCommitChanges computes lines and files changed for a commit
func calculateCommitChanges(commit *object.Commit, pathArg string) (int, int, error) {
	var linesChanged int
	var filesChanged int
	
	if commit.NumParents() == 0 {
		// Initial commit - count all files as added
		tree, err := commit.Tree()
		if err != nil {
			return 0, 0, err
		}
		
		err = tree.Files().ForEach(func(file *object.File) error {
			if matchesSinglePathFilter(file.Name, pathArg) {
				filesChanged++
				// Count lines in initial files
				contents, err := file.Contents()
				if err == nil {
					linesChanged += len(splitLines(contents))
				}
			}
			return nil
		})
		return linesChanged, filesChanged, err
	}
	
	// Regular commit - analyze diff with parent
	parent, err := commit.Parent(0)
	if err != nil {
		return 0, 0, err
	}
	
	parentTree, err := parent.Tree()
	if err != nil {
		return 0, 0, err
	}
	
	currentTree, err := commit.Tree()
	if err != nil {
		return 0, 0, err
	}
	
	// Compute single patch between trees for efficiency (avoids O(N) patch computations)
	patch, err := parentTree.Patch(currentTree)
	if err != nil {
		return 0, 0, err
	}
	
	// Get overall stats from the single patch
	stats := patch.Stats()
	fileSet := make(map[string]bool)
	
	// Filter stats to only include files matching the path filter
	for _, fileStat := range stats {
		if matchesSinglePathFilter(fileStat.Name, pathArg) {
			fileSet[fileStat.Name] = true
			linesChanged += fileStat.Addition + fileStat.Deletion
		}
	}
	
	filesChanged = len(fileSet)
	return linesChanged, filesChanged, nil
}

// splitLines splits content into lines for counting
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}
	
	lines := []string{}
	current := ""
	
	for _, char := range content {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	
	// Add the last line if it doesn't end with newline
	if current != "" {
		lines = append(lines, current)
	}
	
	return lines
}

// printHighRiskCommitsStats displays the analysis results
func printHighRiskCommitsStats(stats *HighRiskCommitsStats, limitArg int) {
	fmt.Printf("High-Risk Commits Analysis\n")
	fmt.Printf("Total commits analyzed: %d\n", stats.TotalCommits)
	
	if stats.TotalCommits == 0 {
		fmt.Printf("No commits found in the specified criteria.\n")
		return
	}
	
	fmt.Printf("Average lines changed per commit: %.1f\n", stats.AverageLines)
	fmt.Printf("Average files touched per commit: %.1f\n", stats.AverageFiles)
	fmt.Printf("\n")
	
	// Risk distribution
	fmt.Printf("Risk Distribution:\n")
	if stats.TotalCommits > 0 {
		fmt.Printf("  Low: %d commits (%.1f%%)\n", 
			stats.LowRisk, float64(stats.LowRisk)/float64(stats.TotalCommits)*100)
		fmt.Printf("  Moderate: %d commits (%.1f%%)\n", 
			stats.ModerateRisk, float64(stats.ModerateRisk)/float64(stats.TotalCommits)*100)
		fmt.Printf("  High: %d commits (%.1f%%)\n", 
			stats.HighRisk, float64(stats.HighRisk)/float64(stats.TotalCommits)*100)
		fmt.Printf("  Critical: %d commits (%.1f%%)\n", 
			stats.CriticalRisk, float64(stats.CriticalRisk)/float64(stats.TotalCommits)*100)
	}
	fmt.Printf("\n")
	
	// Context and research
	fmt.Printf("Context: Large commits reduce review effectiveness and rollback safety (Kent Beck, Martin Fowler).\n\n")
	
	// Show largest commit
	if stats.LargestCommit.Hash != "" {
		fmt.Printf("Largest Commit:\n")
		fmt.Printf("  %s by %s (%s)\n", 
			stats.LargestCommit.Hash, stats.LargestCommit.Author, stats.LargestCommit.Date.Format("2006-01-02"))
		fmt.Printf("  %d lines, %d files - %s\n", 
			stats.LargestCommit.LinesChanged, stats.LargestCommit.FilesChanged, stats.LargestCommit.Risk)
		fmt.Printf("  Message: %s\n", trimMessage(stats.LargestCommit.Message))
		fmt.Printf("\n")
	}
	
	// Show risky commits
	if len(stats.RiskyCommits) > 0 {
		fmt.Printf("Risky Commits (showing %d):\n\n", min(len(stats.RiskyCommits), limitArg))
		
		for i, commit := range stats.RiskyCommits {
			if i >= limitArg {
				break
			}
			
			fmt.Printf("%d. %s — %s\n", i+1, commit.Hash, commit.Risk)
			fmt.Printf("   Author: %s\n", commit.Author)
			fmt.Printf("   Date: %s\n", commit.Date.Format("2006-01-02 15:04"))
			fmt.Printf("   Changes: %d lines, %d files\n", commit.LinesChanged, commit.FilesChanged)
			fmt.Printf("   Reason: %s\n", commit.Reason)
			fmt.Printf("   Message: %s\n", trimMessage(commit.Message))
			fmt.Printf("\n")
		}
	}
	
	// Recommendations
	fmt.Printf("Recommendations:\n")
	if stats.CriticalRisk > 0 {
		fmt.Printf("  • %d critical commits need immediate review process improvement\n", stats.CriticalRisk)
	}
	if stats.HighRisk > 0 {
		fmt.Printf("  • %d high-risk commits should be broken into smaller changes\n", stats.HighRisk)
	}
	if stats.ModerateRisk > 0 {
		fmt.Printf("  • %d moderate commits could benefit from more focused scope\n", stats.ModerateRisk)
	}
	
	// Overall assessment
	riskPercentage := float64(stats.ModerateRisk+stats.HighRisk+stats.CriticalRisk) / float64(stats.TotalCommits) * 100
	if riskPercentage > 30 {
		fmt.Printf("  • %.1f%% of commits are risky - consider implementing commit size guidelines\n", riskPercentage)
	} else if riskPercentage > 15 {
		fmt.Printf("  • %.1f%% of commits are risky - monitor commit patterns\n", riskPercentage)
	} else {
		fmt.Printf("  • %.1f%% risky commits - good commit hygiene!\n", riskPercentage)
	}
}

// trimMessage truncates long commit messages for display
func trimMessage(message string) string {
	lines := splitLines(message)
	if len(lines) == 0 {
		return ""
	}
	
	firstLine := lines[0]
	if len(firstLine) > 80 {
		return firstLine[:77] + "..."
	}
	return firstLine
}

