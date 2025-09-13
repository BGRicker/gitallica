package cmd

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

// DORA lead time thresholds in hours based on 2021 State of DevOps report
const (
	eliteLeadTimeThreshold  = 24.0  // <1 day (elite performers)
	highLeadTimeThreshold   = 168.0 // <1 week (high performers)
	mediumLeadTimeThreshold = 720.0 // <1 month (medium performers)
	// Above 1 month is Low performance
)

// Performance level thresholds (percentage of commits in each category)
const (
	elitePerformanceThreshold  = 0.7 // 70%+ elite commits for Elite performance
	highPerformanceThreshold   = 0.6 // 60%+ elite+high commits for High performance
	mediumPerformanceThreshold = 0.5 // 50%+ elite+high+medium commits for Medium performance
	// Below 50% elite+high+medium is Low performance
)

// Sentinel errors for iteration control
var (
	ErrFoundTarget = errors.New("target found")
	ErrFoundCommit = errors.New("commit found")
)

// CommitLeadTime represents a commit with its lead time measurement
type CommitLeadTime struct {
	Hash           string
	Author         string
	CommitTime     time.Time
	DeployTime     time.Time
	LeadTimeHours  float64
	Classification string // "Elite", "High", "Medium", "Low"
	Message        string
}

// ChangeLeadTimeStats contains analysis results for change lead time
type ChangeLeadTimeStats struct {
	TotalCommits         int
	AverageLeadTimeHours float64
	MedianLeadTimeHours  float64
	P95LeadTimeHours     float64
	EliteCommits         int
	HighCommits          int
	MediumCommits        int
	LowCommits           int
	DORAPerformanceLevel string // "Elite", "High", "Medium", "Low", "Unknown"
	SlowestCommits       []CommitLeadTime
	FastestCommits       []CommitLeadTime
	Commits              []CommitLeadTime
}

// changeLeadTimeCmd represents the change-lead-time command
var changeLeadTimeCmd = &cobra.Command{
	Use:   "change-lead-time",
	Short: "Analyze change lead time from commit to deployment using DORA metrics",
	Long: `Analyzes change lead time to measure delivery performance using DORA benchmarks.

Change lead time measures the time from when code is committed to when it's 
running in production. This is a key indicator of delivery performance and
organizational capability.

DORA Benchmarks (2021):
- Elite: <1 hour lead time
- High: 1 hour to 1 day 
- Medium: 1 day to 1 week
- Low: >1 week

Research basis:
- Based on DORA State of DevOps research and Accelerate book findings
- Measures organizational delivery performance and flow efficiency
- Correlates with overall software delivery performance and business outcomes

The analysis identifies:
- Lead time distribution across DORA performance levels
- Delivery flow bottlenecks and optimization opportunities
- Team performance benchmarking against industry standards
- Recommendations for improving delivery velocity`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.PlainOpen(".")
		if err != nil {
			return fmt.Errorf("could not open repository: %v", err)
		}

		pathFilters, source := getConfigPaths(cmd, "change-lead-time.paths")
		lastArg := getConfigLast(cmd, "change-lead-time.last")
		limitArg, _ := cmd.Flags().GetInt("limit")
		methodArg, _ := cmd.Flags().GetString("method")

		// Print configuration scope
		printCommandScope(cmd, "change-lead-time", lastArg, pathFilters, source)

		stats, err := analyzeChangeLeadTime(repo, pathFilters, lastArg, limitArg, methodArg)
		if err != nil {
			return err
		}

		printChangeLeadTimeStats(stats, limitArg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(changeLeadTimeCmd)
	changeLeadTimeCmd.Flags().String("last", "", "Specify the time window to analyze (e.g., 30d, 6m, 1y)")
	changeLeadTimeCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	changeLeadTimeCmd.Flags().Int("limit", 5, "Number of slowest/fastest commits to show in detailed output")
	changeLeadTimeCmd.Flags().String("method", "merge", "Lead time calculation method: 'merge' (commit to main) or 'tag' (commit to release tag)")
}

// analyzeChangeLeadTime performs the main lead time analysis
func analyzeChangeLeadTime(repo *git.Repository, pathFilters []string, lastArg string, limitArg int, method string) (*ChangeLeadTimeStats, error) {
	// Parse time window if provided
	var cutoffTime time.Time
	var err error
	if lastArg != "" {
		cutoffTime, err = parseDurationArg(lastArg)
		if err != nil {
			return nil, fmt.Errorf("invalid time window '%s': %v", lastArg, err)
		}
	}

	// Get commits with lead time measurements
	commits, err := getCommitsWithLeadTime(repo, cutoffTime, pathFilters, method)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze lead time: %v", err)
	}

	// Calculate comprehensive statistics
	stats := calculateChangeLeadTimeStats(commits)

	return stats, nil
}

// getCommitsWithLeadTime retrieves commits and calculates their lead times
func getCommitsWithLeadTime(repo *git.Repository, cutoffTime time.Time, pathFilters []string, method string) ([]CommitLeadTime, error) {
	var commits []CommitLeadTime

	// Get the default branch (usually main/master)
	defaultBranch, err := getDefaultBranch(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get default branch: %v", err)
	}

	// Get commit iterator from default branch
	commitIter, err := repo.Log(&git.LogOptions{From: defaultBranch.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %v", err)
	}
	defer commitIter.Close()

	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Skip if outside time window
		if !cutoffTime.IsZero() && commit.Author.When.Before(cutoffTime) {
			return nil
		}

		// Skip merge commits for cleaner analysis
		if len(commit.ParentHashes) > 1 {
			return nil
		}

		// Skip if path filter doesn't match
		if len(pathFilters) > 0 {
			affects, err := commitAffectsPath(commit, pathFilters)
			if err != nil || !affects {
				return nil
			}
		}

		// Calculate lead time based on method
		var deployTime time.Time
		var leadTime float64

		switch method {
		case "merge":
			// For merge method, find when this commit was merged to main branch
			mergeTime, err := findCommitMergeTime(repo, commit.Hash)
			if err != nil {
				// If merge time cannot be determined, use commit time as fallback
				deployTime = commit.Author.When
				leadTime = 0 // Essentially no lead time for direct commits
			} else {
				deployTime = mergeTime
				leadTime = calculateLeadTime(commit.Author.When, mergeTime)
			}
		case "tag":
			// For tag method, find when this commit was tagged
			tagTime, err := findCommitInTags(repo, commit.Hash)
			if err != nil {
				return nil // Skip commits not found in tags
			}
			deployTime = tagTime
			leadTime = calculateLeadTime(commit.Author.When, tagTime)
		default:
			// Default to merge method with proper calculation
			mergeTime, err := findCommitMergeTime(repo, commit.Hash)
			if err != nil {
				deployTime = commit.Author.When
				leadTime = 0
			} else {
				deployTime = mergeTime
				leadTime = calculateLeadTime(commit.Author.When, mergeTime)
			}
		}

		commitLeadTime := CommitLeadTime{
			Hash:           commit.Hash.String()[:8],
			Author:         commit.Author.Name,
			CommitTime:     commit.Author.When,
			DeployTime:     deployTime,
			LeadTimeHours:  leadTime,
			Classification: classifyDORALeadTime(leadTime),
			Message:        commit.Message,
		}

		commits = append(commits, commitLeadTime)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by lead time (fastest first)
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].LeadTimeHours < commits[j].LeadTimeHours
	})

	return commits, nil
}

// getDefaultBranch resolves the repository's default branch
func getDefaultBranch(repo *git.Repository) (*plumbing.Reference, error) {
	// Try to get the default branch from HEAD
	head, err := repo.Head()
	if err == nil {
		return head, nil
	}

	// Fallback: try common default branch names
	defaultNames := []string{"refs/heads/main", "refs/heads/master"}
	for _, name := range defaultNames {
		ref, err := repo.Reference(plumbing.ReferenceName(name), true)
		if err == nil {
			return ref, nil
		}
	}

	return nil, fmt.Errorf("could not determine default branch")
}

// findCommitMergeTime finds when a commit was merged to the main branch
func findCommitMergeTime(repo *git.Repository, commitHash plumbing.Hash) (time.Time, error) {
	// Get the default branch
	defaultBranch, err := getDefaultBranch(repo)
	if err != nil {
		return time.Time{}, err
	}

	// Get all commits on the default branch
	commitIter, err := repo.Log(&git.LogOptions{From: defaultBranch.Hash()})
	if err != nil {
		return time.Time{}, err
	}
	defer commitIter.Close()

	var foundMergeTime time.Time
	found := false

	// Look for merge commits that include our target commit
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// If this is our target commit directly on main, use its commit time
		if commit.Hash == commitHash {
			foundMergeTime = commit.Author.When
			found = true
			return ErrFoundTarget // Break iteration
		}

		// Check merge commits (have multiple parents)
		if len(commit.ParentHashes) > 1 {
			// Check if our target commit is in any of the parent branches
			for _, parentHash := range commit.ParentHashes {
				isReachable, err := isCommitReachableFrom(repo, commitHash, parentHash)
				if err == nil && isReachable {
					// Check if this parent is not the main branch (first parent)
					if parentHash != commit.ParentHashes[0] {
						foundMergeTime = commit.Author.When
						found = true
						return ErrFoundTarget // Break iteration
					}
				}
			}
		}

		return nil
	})

	// Check if we found the target (ignore sentinel errors used for flow control)
	if err != nil && err != ErrFoundTarget {
		return time.Time{}, err
	}

	if found {
		return foundMergeTime, nil
	}

	// If no merge found, this might be a direct commit or we can't determine merge time
	return time.Time{}, fmt.Errorf("commit merge time not determinable")
}

// isCommitReachableFrom checks if targetCommit is reachable from fromCommit
func isCommitReachableFrom(repo *git.Repository, targetCommit, fromCommit plumbing.Hash) (bool, error) {
	// Start from fromCommit and traverse backwards
	commitIter, err := repo.Log(&git.LogOptions{From: fromCommit})
	if err != nil {
		return false, err
	}
	defer commitIter.Close()

	found := false
	err = commitIter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == targetCommit {
			found = true
			return ErrFoundCommit // Break iteration
		}
		return nil
	})

	// Check for actual errors (ignore sentinel error used for flow control)
	if err != nil && err != ErrFoundCommit {
		return false, err
	}

	return found, nil
}

// findCommitInTags finds when a commit was first tagged (basic implementation)
func findCommitInTags(repo *git.Repository, commitHash plumbing.Hash) (time.Time, error) {
	tagRefs, err := repo.Tags()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get tags: %w", err)
	}
	defer tagRefs.Close()

	var earliestTagTime time.Time
	found := false

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		tagObj, err := repo.TagObject(tagRef.Hash())
		if err != nil {
			// Might be a lightweight tag, try commit object
			commit, err := repo.CommitObject(tagRef.Hash())
			if err != nil {
				return nil // Skip this tag
			}

			// Check if our commit is reachable from this tag
			isReachable, err := isCommitReachableFrom(repo, commitHash, commit.Hash)
			if err == nil && isReachable {
				if !found || commit.Author.When.Before(earliestTagTime) {
					earliestTagTime = commit.Author.When
					found = true
				}
			}
			return nil
		}

		// Annotated tag
		if tagObj.TargetType == plumbing.CommitObject && tagObj.Target == commitHash {
			if !found || tagObj.Tagger.When.Before(earliestTagTime) {
				earliestTagTime = tagObj.Tagger.When
				found = true
			}
		}
		return nil
	})

	if err != nil {
		return time.Time{}, err
	}

	if !found {
		return time.Time{}, fmt.Errorf("commit not found in tags")
	}

	return earliestTagTime, nil
}

// calculateLeadTime calculates lead time in hours between two timestamps
func calculateLeadTime(commitTime, deployTime time.Time) float64 {
	duration := deployTime.Sub(commitTime)
	return duration.Hours()
}

// classifyDORALeadTime classifies lead time according to DORA benchmarks
func classifyDORALeadTime(leadTimeHours float64) string {
	if leadTimeHours < eliteLeadTimeThreshold {
		return "Elite"
	} else if leadTimeHours < highLeadTimeThreshold {
		return "High"
	} else if leadTimeHours < mediumLeadTimeThreshold {
		return "Medium"
	}
	return "Low"
}

// classifyDORAPerformanceLevel assesses overall team performance level
func classifyDORAPerformanceLevel(eliteCommits, highCommits, mediumCommits, lowCommits, totalCommits int) string {
	if totalCommits == 0 {
		return "Unknown"
	}

	eliteRatio := float64(eliteCommits) / float64(totalCommits)
	eliteHighRatio := float64(eliteCommits+highCommits) / float64(totalCommits)
	eliteHighMediumRatio := float64(eliteCommits+highCommits+mediumCommits) / float64(totalCommits)

	if eliteRatio >= elitePerformanceThreshold {
		return "Elite"
	} else if eliteHighRatio >= highPerformanceThreshold {
		return "High"
	} else if eliteHighMediumRatio >= mediumPerformanceThreshold {
		return "Medium"
	}
	return "Low"
}

// calculateChangeLeadTimeStats computes comprehensive lead time statistics
func calculateChangeLeadTimeStats(commits []CommitLeadTime) *ChangeLeadTimeStats {
	stats := &ChangeLeadTimeStats{
		TotalCommits: len(commits),
		Commits:      commits,
	}

	if len(commits) == 0 {
		stats.DORAPerformanceLevel = "Unknown"
		return stats
	}

	// Calculate basic statistics
	totalHours := 0.0
	var leadTimes []float64

	for _, commit := range commits {
		totalHours += commit.LeadTimeHours
		leadTimes = append(leadTimes, commit.LeadTimeHours)

		// Ensure classification is set (in case it wasn't set in input data)
		classification := commit.Classification
		if classification == "" {
			classification = classifyDORALeadTime(commit.LeadTimeHours)
		}

		// Count by DORA classification
		switch classification {
		case "Elite":
			stats.EliteCommits++
		case "High":
			stats.HighCommits++
		case "Medium":
			stats.MediumCommits++
		case "Low":
			stats.LowCommits++
		}
	}

	stats.AverageLeadTimeHours = totalHours / float64(len(commits))
	stats.MedianLeadTimeHours = calculatePercentile(leadTimes, 50)
	stats.P95LeadTimeHours = calculatePercentile(leadTimes, 95)
	stats.DORAPerformanceLevel = classifyDORAPerformanceLevel(stats.EliteCommits, stats.HighCommits, stats.MediumCommits, stats.LowCommits, stats.TotalCommits)

	// Sort commits by lead time for robustness (don't assume external sorting)
	sortedCommits := make([]CommitLeadTime, len(commits))
	copy(sortedCommits, commits)
	sort.Slice(sortedCommits, func(i, j int) bool {
		return sortedCommits[i].LeadTimeHours < sortedCommits[j].LeadTimeHours
	})

	// Get fastest and slowest commits from sorted list
	limit := 5
	if len(sortedCommits) < limit {
		limit = len(sortedCommits)
	}
	stats.FastestCommits = sortedCommits[:limit]
	stats.SlowestCommits = sortedCommits[len(sortedCommits)-limit:]

	// Reverse slowest commits to show highest lead times first
	for i, j := 0, len(stats.SlowestCommits)-1; i < j; i, j = i+1, j-1 {
		stats.SlowestCommits[i], stats.SlowestCommits[j] = stats.SlowestCommits[j], stats.SlowestCommits[i]
	}

	return stats
}

// calculatePercentile calculates the nth percentile of a sorted slice
// commitAffectsPath is now defined in utils.go

func calculatePercentile(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Calculate percentile index
	index := float64(percentile) / 100.0 * float64(len(sorted)-1)

	// Handle edge cases
	if index <= 0 {
		return sorted[0]
	}
	if index >= float64(len(sorted)-1) {
		return sorted[len(sorted)-1]
	}

	// Interpolate between values
	lower := int(index)
	upper := lower + 1
	weight := index - float64(lower)

	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// printChangeLeadTimeStats displays the analysis results
func printChangeLeadTimeStats(stats *ChangeLeadTimeStats, limit int) {
	fmt.Println("Change Lead Time Analysis")
	fmt.Printf("Total commits analyzed: %d\n", stats.TotalCommits)

	if stats.TotalCommits == 0 {
		fmt.Println("No commits found for analysis.")
		return
	}

	fmt.Printf("Average lead time: %.1f hours (%.1f days)\n", stats.AverageLeadTimeHours, stats.AverageLeadTimeHours/24)
	fmt.Printf("Median lead time: %.1f hours (%.1f days)\n", stats.MedianLeadTimeHours, stats.MedianLeadTimeHours/24)
	fmt.Printf("95th percentile: %.1f hours (%.1f days)\n", stats.P95LeadTimeHours, stats.P95LeadTimeHours/24)
	fmt.Println()

	// DORA classification distribution
	fmt.Println("DORA Performance Distribution:")
	fmt.Printf("  Elite (<1 day): %d commits (%.1f%%)\n", stats.EliteCommits, float64(stats.EliteCommits)/float64(stats.TotalCommits)*100)
	fmt.Printf("  High (1-7 days): %d commits (%.1f%%)\n", stats.HighCommits, float64(stats.HighCommits)/float64(stats.TotalCommits)*100)
	fmt.Printf("  Medium (1-4 weeks): %d commits (%.1f%%)\n", stats.MediumCommits, float64(stats.MediumCommits)/float64(stats.TotalCommits)*100)
	fmt.Printf("  Low (>1 month): %d commits (%.1f%%)\n", stats.LowCommits, float64(stats.LowCommits)/float64(stats.TotalCommits)*100)
	fmt.Println()

	// Overall DORA performance level
	fmt.Printf("Overall DORA Performance Level: %s\n", stats.DORAPerformanceLevel)
	elitePercentage := float64(stats.EliteCommits) / float64(stats.TotalCommits) * 100
	fmt.Printf("  %.1f%% of commits achieve elite performance (<1 day)\n", elitePercentage)
	fmt.Println()

	// Context and research
	fmt.Println("Context: DORA research shows lead time is a key predictor of software delivery performance.")
	fmt.Println()

	// Fastest commits
	if len(stats.FastestCommits) > 0 {
		displayLimit := len(stats.FastestCommits)
		if displayLimit > limit {
			displayLimit = limit
		}

		fmt.Printf("Fastest %d Commits:\n", displayLimit)
		for i := 0; i < displayLimit; i++ {
			commit := stats.FastestCommits[i]
			fmt.Printf("  %s: %.1f hours (%s)\n", commit.Hash, commit.LeadTimeHours, commit.Classification)
			fmt.Printf("    Author: %s (%s)\n", commit.Author, commit.CommitTime.Format("2006-01-02 15:04"))
		}
		fmt.Println()
	}

	// Slowest commits
	if len(stats.SlowestCommits) > 0 {
		displayLimit := len(stats.SlowestCommits)
		if displayLimit > limit {
			displayLimit = limit
		}

		fmt.Printf("Slowest %d Commits:\n", displayLimit)
		for i := 0; i < displayLimit; i++ {
			commit := stats.SlowestCommits[i]
			fmt.Printf("  %s: %.1f hours (%.1f days) (%s)\n", commit.Hash, commit.LeadTimeHours, commit.LeadTimeHours/24, commit.Classification)
			fmt.Printf("    Author: %s (%s)\n", commit.Author, commit.CommitTime.Format("2006-01-02 15:04"))
		}
		fmt.Println()
	}

	// Recommendations
	fmt.Println("Recommendations:")
	switch stats.DORAPerformanceLevel {
	case "Elite":
		fmt.Println("  • Excellent delivery performance! Maintain current practices")
		fmt.Println("  • Share learnings with other teams to scale elite practices")
	case "High":
		fmt.Println("  • Good delivery performance with room for optimization")
		fmt.Printf("  • Focus on moving more commits to elite level (currently %.1f%%)\n", elitePercentage)
	case "Medium":
		fmt.Println("  • Moderate delivery performance - significant improvement opportunity")
		fmt.Println("  • Review deployment pipeline for bottlenecks and automation gaps")
		fmt.Println("  • Consider trunk-based development and feature flags")
	case "Low":
		fmt.Println("  • Low delivery performance - critical improvement needed")
		fmt.Println("  • Implement continuous integration and deployment practices")
		fmt.Println("  • Reduce batch sizes and increase deployment frequency")
		fmt.Println("  • Focus on automation and reducing manual processes")
	default:
		fmt.Println("  • Insufficient data for performance assessment")
	}

	if stats.P95LeadTimeHours > 720 { // >30 days
		fmt.Println("  • 95th percentile lead time is very high - investigate outliers")
	}
}
