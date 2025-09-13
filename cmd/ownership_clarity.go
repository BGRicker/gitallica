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
	// Ownership thresholds based on Microsoft research
	// >9 contributors increases vulnerability likelihood 16x
	// Strong ownership (one developer ≥80%) tends to improve quality
	ownershipStrongThreshold            = 0.80 // One developer owns ≥80% of changes
	ownershipRiskThreshold              = 9    // >9 contributors increases vulnerability risk
	ownershipMinContributorsForAnalysis = 3    // Only analyze files with ≥3 contributors
)

const ownershipBenchmarkContext = "Microsoft Research: Strong code ownership (one developer ≥80%) improves quality, while files with >9 contributors are 16x more likely to have vulnerabilities (Bird et al., MSR 2011)."

// OwnershipClarityStats represents the ownership clarity analysis for files
type OwnershipClarityStats struct {
	TotalFiles    int
	FilesAnalyzed int
	HealthyFiles  int
	CautionFiles  int
	WarningFiles  int
	CriticalFiles int
	UnknownFiles  int
	FileOwnership []FileOwnership
}

// FileOwnership represents ownership information for a single file
type FileOwnership struct {
	FilePath          string
	TopContributor    string
	TopOwnership      float64
	TotalContributors int
	Status            string
	Recommendation    string
	CommitsByAuthor   map[string]int
}

// calculateOwnershipClarity calculates ownership clarity metrics
func calculateOwnershipClarity(commitsByContributor map[string]int) (float64, string, int) {
	if len(commitsByContributor) == 0 {
		return 0.0, "Unknown", 0
	}

	total := 0
	maxCommits := 0
	validContributors := 0

	// Count total commits and find max, filtering out negative values
	for _, commits := range commitsByContributor {
		if commits > 0 {
			total += commits
			validContributors++
			if commits > maxCommits {
				maxCommits = commits
			}
		}
	}

	if total == 0 {
		return 0.0, "Unknown", validContributors
	}

	topOwnership := float64(maxCommits) / float64(total)
	status, _ := classifyOwnershipClarity(topOwnership, validContributors)

	return topOwnership, status, validContributors
}

// classifyOwnershipClarity classifies ownership clarity and provides recommendations
func classifyOwnershipClarity(topOwnership float64, totalContributors int) (string, string) {
	// Single contributor is always healthy
	if totalContributors <= 1 {
		return "Healthy", "Good ownership balance"
	}

	// Based on Microsoft research findings
	// Strong ownership tends to improve quality regardless of contributor count
	if topOwnership >= ownershipStrongThreshold {
		return "Healthy", "Strong ownership tends to improve quality"
	}

	// Excessive number of contributors without clear owner increases risk
	if totalContributors > ownershipRiskThreshold {
		return "Critical", "Too many contributors (>9) increases vulnerability risk 16x"
	}

	// Small teams benefit from knowledge sharing
	if totalContributors <= 3 {
		return "Caution", "Small team – consider knowledge sharing"
	}

	return "Warning", "Multiple contributors without clear ownership"
}

// analyzeOwnershipClarity analyzes ownership clarity across repository files
func analyzeOwnershipClarity(repo *git.Repository, pathFilters []string, lastArg string) (*OwnershipClarityStats, error) {
	var since *time.Time
	// Set a sensible default time window to cap resource usage
	if lastArg == "" {
		lastArg = "1y" // Default to last year instead of full history
	}
	sinceTime, err := parseDurationArg(lastArg)
	if err != nil {
		return nil, fmt.Errorf("invalid time window: %v", err)
	}
	since = &sinceTime

	// Get file ownership data with efficient analysis
	fileOwnership, err := analyzeFileOwnership(repo, pathFilters, since)
	if err != nil {
		return nil, fmt.Errorf("error analyzing file ownership: %v", err)
	}

	stats := &OwnershipClarityStats{
		TotalFiles:    len(fileOwnership),
		FilesAnalyzed: len(fileOwnership),
		FileOwnership: fileOwnership,
	}

	// Count files by status
	for _, file := range fileOwnership {
		switch file.Status {
		case "Healthy":
			stats.HealthyFiles++
		case "Caution":
			stats.CautionFiles++
		case "Warning":
			stats.WarningFiles++
		case "Critical":
			stats.CriticalFiles++
		default:
			stats.UnknownFiles++
		}
	}

	return stats, nil
}

// analyzeFileOwnership analyzes ownership for individual files using efficient log options
func analyzeFileOwnership(repo *git.Repository, pathFilters []string, since *time.Time) ([]FileOwnership, error) {
	// Use efficient log options with early path filtering
	commitIter, err := repo.Log(&git.LogOptions{
		Since: since,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get commit log: %v", err)
	}
	defer commitIter.Close()

	// Map of file -> author -> commit count
	fileCommits := make(map[string]map[string]int)

	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Add check for commit author to prevent runtime crashes
		if commit.Author.Email == "" {
			return nil // Skip commits without author information
		}
		author := commit.Author.Email

		// Skip merge commits for performance (they often don't represent meaningful ownership)
		if commit.NumParents() > 1 {
			return nil
		}

		// Get files changed in this commit using efficient approach
		if commit.NumParents() == 0 {
			// Initial commit - treat as adding all files
			tree, err := commit.Tree()
			if err != nil {
				return err
			}

			return tree.Files().ForEach(func(file *object.File) error {
				if !matchesPathFilter(file.Name, pathFilters) {
					return nil
				}

				if fileCommits[file.Name] == nil {
					fileCommits[file.Name] = make(map[string]int)
				}
				fileCommits[file.Name][author]++
				return nil
			})
		}

		// Regular commit - use more efficient diff approach
		parent, err := commit.Parent(0)
		if err != nil {
			return err
		}

		parentTree, err := parent.Tree()
		if err != nil {
			return err
		}

		currentTree, err := commit.Tree()
		if err != nil {
			return err
		}

		// Use name-only changes to reduce memory usage
		changes, err := parentTree.Diff(currentTree)
		if err != nil {
			return err
		}

		// Process changes with early filtering to bound memory
		for _, change := range changes {
			var filePath string
			if change.To.Name != "" {
				filePath = change.To.Name
			} else if change.From.Name != "" {
				filePath = change.From.Name
			}

			// Early path filtering to avoid processing irrelevant files
			if filePath != "" && matchesPathFilter(filePath, pathFilters) {
				if fileCommits[filePath] == nil {
					fileCommits[filePath] = make(map[string]int)
				}
				fileCommits[filePath][author]++

				// Handle rename detection by tracking both old and new names
				if change.From.Name != "" && change.To.Name != "" && change.From.Name != change.To.Name {
					// File was renamed - credit both paths to maintain history
					if matchesPathFilter(change.From.Name, pathFilters) {
						if fileCommits[change.From.Name] == nil {
							fileCommits[change.From.Name] = make(map[string]int)
						}
						fileCommits[change.From.Name][author]++
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error iterating commits: %v", err)
	}

	// Convert to FileOwnership slice
	var ownership []FileOwnership
	for filePath, commits := range fileCommits {
		topOwnership, status, contributors := calculateOwnershipClarity(commits)
		_, recommendation := classifyOwnershipClarity(topOwnership, contributors)

		// Find top contributor
		topContributor := ""
		maxCommits := 0
		for author, commitCount := range commits {
			if commitCount > maxCommits {
				maxCommits = commitCount
				topContributor = author
			}
		}

		ownership = append(ownership, FileOwnership{
			FilePath:          filePath,
			TopContributor:    topContributor,
			TopOwnership:      topOwnership,
			TotalContributors: contributors,
			Status:            status,
			Recommendation:    recommendation,
			CommitsByAuthor:   commits,
		})
	}

	// Sort by ownership clarity (most concerning first)
	sort.Slice(ownership, func(i, j int) bool {
		statusPriority := map[string]int{
			"Critical": 0,
			"Warning":  1,
			"Caution":  2,
			"Healthy":  3,
			"Unknown":  4,
		}

		if statusPriority[ownership[i].Status] != statusPriority[ownership[j].Status] {
			return statusPriority[ownership[i].Status] < statusPriority[ownership[j].Status]
		}

		// Within same status, sort by ownership percentage (lower first for problematic statuses)
		if ownership[i].Status == "Critical" || ownership[i].Status == "Warning" {
			return ownership[i].TopOwnership < ownership[j].TopOwnership
		}
		return ownership[i].TopOwnership > ownership[j].TopOwnership
	})

	return ownership, nil
}

// printOwnershipClarityStats prints the ownership clarity analysis results
func printOwnershipClarityStats(stats *OwnershipClarityStats, pathFilters []string, limit int) {
	fmt.Printf("Ownership Clarity Analysis\n")
	if len(pathFilters) > 0 {
		fmt.Printf("Path filters: %s\n", strings.Join(pathFilters, ", "))
	}
	fmt.Printf("Files analyzed: %d\n", stats.FilesAnalyzed)
	fmt.Println()

	// Summary by status
	fmt.Printf("Ownership Distribution:\n")
	if stats.FilesAnalyzed == 0 {
		fmt.Printf("  Healthy: %d files (0.0%%)\n", stats.HealthyFiles)
		fmt.Printf("  Caution: %d files (0.0%%)\n", stats.CautionFiles)
		fmt.Printf("  Warning: %d files (0.0%%)\n", stats.WarningFiles)
		fmt.Printf("  Critical: %d files (0.0%%)\n", stats.CriticalFiles)
		if stats.UnknownFiles > 0 {
			fmt.Printf("  Unknown: %d files (0.0%%)\n", stats.UnknownFiles)
		}
	} else {
		fmt.Printf("  Healthy: %d files (%.1f%%)\n", stats.HealthyFiles,
			float64(stats.HealthyFiles)/float64(stats.FilesAnalyzed)*100)
		fmt.Printf("  Caution: %d files (%.1f%%)\n", stats.CautionFiles,
			float64(stats.CautionFiles)/float64(stats.FilesAnalyzed)*100)
		fmt.Printf("  Warning: %d files (%.1f%%)\n", stats.WarningFiles,
			float64(stats.WarningFiles)/float64(stats.FilesAnalyzed)*100)
		fmt.Printf("  Critical: %d files (%.1f%%)\n", stats.CriticalFiles,
			float64(stats.CriticalFiles)/float64(stats.FilesAnalyzed)*100)
		if stats.UnknownFiles > 0 {
			fmt.Printf("  Unknown: %d files (%.1f%%)\n", stats.UnknownFiles,
				float64(stats.UnknownFiles)/float64(stats.FilesAnalyzed)*100)
		}
	}
	fmt.Println()

	fmt.Println("Context:", ownershipBenchmarkContext)
	fmt.Println()

	// Show detailed file analysis
	displayCount := limit
	if displayCount > len(stats.FileOwnership) {
		displayCount = len(stats.FileOwnership)
	}

	if displayCount > 0 {
		fmt.Printf("Top %d files by ownership risk:\n", displayCount)
		for i := 0; i < displayCount; i++ {
			file := stats.FileOwnership[i]
			fmt.Printf("\n%d. %s — %s\n", i+1, file.FilePath, file.Status)
			fmt.Printf("   Top owner: %s (%.1f%% of commits)\n",
				file.TopContributor, file.TopOwnership*100)
			fmt.Printf("   Contributors: %d\n", file.TotalContributors)
			fmt.Printf("   Recommendation: %s\n", file.Recommendation)

			// Show top contributors
			if len(file.CommitsByAuthor) > 1 {
				type authorCommit struct {
					author  string
					commits int
				}
				var authors []authorCommit
				total := 0
				for author, commits := range file.CommitsByAuthor {
					authors = append(authors, authorCommit{author, commits})
					total += commits
				}
				sort.Slice(authors, func(a, b int) bool {
					return authors[a].commits > authors[b].commits
				})

				fmt.Printf("   Top contributors: ")
				for j, ac := range authors {
					if j >= 3 { // Show top 3
						break
					}
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s (%.1f%%)", ac.author, float64(ac.commits)/float64(total)*100)
				}
				fmt.Println()
			}
		}
	}

	// Provide actionable insights
	fmt.Printf("\nRecommendations:\n")
	if stats.CriticalFiles > 0 {
		fmt.Printf("  • %d files need urgent attention - assign primary maintainers\n", stats.CriticalFiles)
	}
	if stats.WarningFiles > 0 {
		fmt.Printf("  • %d files need clearer ownership assignments\n", stats.WarningFiles)
	}
	if stats.CautionFiles > 0 {
		fmt.Printf("  • %d files may have ownership bottlenecks - encourage broader contribution\n", stats.CautionFiles)
	}
	if stats.HealthyFiles > 0 {
		fmt.Printf("  • %d files have good ownership balance\n", stats.HealthyFiles)
	}
}

// ownershipClarityCmd represents the ownership-clarity command
var ownershipClarityCmd = &cobra.Command{
	Use:   "ownership-clarity",
	Short: "Analyze ownership clarity across repository files",
	Long: `Analyze ownership clarity to identify files with unclear ownership patterns.
Helps ensure balanced ownership that avoids both bottlenecks and diffuse responsibility.

Performance optimized for large repositories:
- Defaults to analyzing last 1 year (use --last to override)
- Skips merge commits for faster analysis
- Uses efficient diff processing with early path filtering

Healthy ownership balance typically means:
- Clear primary maintainer (40-80% of commits)
- Shared knowledge with backup contributors
- Not too concentrated (avoiding bottlenecks)
- Not too diffuse (avoiding responsibility gaps)

Classifications:
- Healthy: Good ownership balance (40-80% primary ownership)
- Caution: Too concentrated (>80% single owner in teams >3)
- Warning: Diffuse ownership without clear primary maintainer
- Critical: Extremely diffuse ownership in large contributor base

"With collective ownership, anyone can change any part of the code at any time." — Martin Fowler`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		pathFilters, source := getConfigPaths(cmd, "ownership-clarity.paths")
		lastArg := getConfigLast(cmd, "ownership-clarity.last")
		limit, _ := cmd.Flags().GetInt("limit")

		// Print configuration scope
		printCommandScope(cmd, "ownership-clarity", lastArg, pathFilters, source)

		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Could not open repository: %v", err)
		}

		stats, err := analyzeOwnershipClarity(repo, pathFilters, lastArg)
		if err != nil {
			log.Fatalf("Error analyzing ownership clarity: %v", err)
		}

		printOwnershipClarityStats(stats, pathFilters, limit)
	},
}

func init() {
	ownershipClarityCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	ownershipClarityCmd.Flags().String("last", "", "Limit analysis to recent timeframe (e.g., '30d', '6m', '1y'). Defaults to '1y' for performance.")
	ownershipClarityCmd.Flags().Int("limit", 10, "Number of files to show in detailed analysis")
	rootCmd.AddCommand(ownershipClarityCmd)
}
