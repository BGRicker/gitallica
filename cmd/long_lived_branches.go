package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

// Thresholds for long-lived branch analysis based on trunk-based development guidance
const (
	// Branch age thresholds (in days) - based on trunk-based development principles
	healthyBranchMaxAge  = 1.0  // Daily merges recommended
	warningBranchMaxAge  = 3.0  // Warning after a few days
	criticalBranchMaxAge = 7.0  // Critical after a week
	
	// Trunk-based compliance thresholds (percentage of healthy branches)
	excellentComplianceThreshold = 0.8 // 80%+ healthy branches
	goodComplianceThreshold      = 0.6 // 60%+ healthy branches  
	moderateComplianceThreshold  = 0.4 // 40%+ healthy branches
	poorComplianceThreshold      = 0.2 // 20%+ healthy branches
	// Below 20% is Critical
	
	// Performance limits
	maxCommitCountLimit = 1000 // Limit to avoid excessive counting for very large branches
)

// Sentinel errors for control flow clarity
var (
	ErrIterationComplete = errors.New("iteration complete")
	ErrReachedMergeBase  = errors.New("reached merge base")
	ErrBranchTooLarge    = errors.New("branch too large")
)

// BranchInfo represents information about a Git branch
type BranchInfo struct {
	Name               string
	AgeInDays          float64
	Status             string // "active", "merged", "stale"
	Risk               string // "Healthy", "Warning", "Risky", "Critical"
	LastCommitAuthor   string
	LastCommitTime     time.Time
	CommitCount        int
	DivergencePoint    string // Hash of the divergence commit
}

// LongLivedBranchesStats contains analysis results for branch lifespans
type LongLivedBranchesStats struct {
	TotalBranches         int
	AverageBranchAge      float64
	HealthyBranches       int
	WarningBranches       int
	RiskyBranches         int
	CriticalBranches      int
	TrunkBasedCompliance  string // "Excellent", "Good", "Moderate", "Poor", "Critical", "Unknown"
	OldestBranch          *BranchInfo
	RiskyBranchDetails    []BranchInfo
	Branches              []BranchInfo
}

// longLivedBranchesCmd represents the long-lived-branches command
var longLivedBranchesCmd = &cobra.Command{
	Use:   "long-lived-branches",
	Short: "Analyze long-lived branches to identify trunk-based development compliance",
	Long: `Analyzes branch lifespans to identify long-lived branches that may indicate
departures from trunk-based development practices.

Long-lived branches increase integration risk, reduce deployment frequency, and
can lead to merge conflicts and delivery delays.

Research basis:
- "Teams should merge into trunk at least once or several times per day" — DORA State of DevOps
- Accelerate research shows elite teams merge frequently and keep branches short-lived
- Trunk-based development principles emphasize small, frequent integrations

The analysis identifies:
- Branch age distribution and risk classification
- Compliance with trunk-based development practices  
- Specific long-lived branches requiring attention
- Recommendations for improving integration practices`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.PlainOpen(".")
		if err != nil {
			return fmt.Errorf("could not open repository: %v", err)
		}

		lastArg, _ := cmd.Flags().GetString("last")
		pathFilters := getConfigPaths(cmd, "long-lived-branches.paths")
		limitArg, _ := cmd.Flags().GetInt("limit")
		showMergedArg, _ := cmd.Flags().GetBool("show-merged")
		
		// Print configuration scope
		printCommandScope(cmd, "long-lived-branches", lastArg, pathFilters)

		stats, err := analyzeLongLivedBranches(repo, pathFilters, lastArg, limitArg, showMergedArg)
		if err != nil {
			return err
		}

		printLongLivedBranchesStats(stats, limitArg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(longLivedBranchesCmd)
	longLivedBranchesCmd.Flags().String("last", "", "Specify the time window to analyze (e.g., 30d, 6m, 1y)")
	longLivedBranchesCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	longLivedBranchesCmd.Flags().Int("limit", 10, "Number of risky branches to show in detailed output")
	longLivedBranchesCmd.Flags().Bool("show-merged", false, "Include recently merged branches in analysis")
}

// analyzeLongLivedBranches performs the main branch analysis
func analyzeLongLivedBranches(repo *git.Repository, pathFilters []string, lastArg string, limitArg int, showMerged bool) (*LongLivedBranchesStats, error) {
	// Parse time window if provided
	var cutoffTime time.Time
	var err error
	if lastArg != "" {
		cutoffTime, err = parseDurationArg(lastArg)
		if err != nil {
			return nil, fmt.Errorf("invalid time window '%s': %v", lastArg, err)
		}
	}

	// Get all branches (local and remote)
	branches, err := getAllBranches(repo, cutoffTime, pathFilters, showMerged)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze branches: %v", err)
	}

	// Calculate comprehensive statistics
	stats := calculateLongLivedBranchesStats(branches)
	
	return stats, nil
}

// getAllBranches retrieves and analyzes all branches in the repository
func getAllBranches(repo *git.Repository, cutoffTime time.Time, pathFilters []string, showMerged bool) ([]BranchInfo, error) {
	var branches []BranchInfo
	now := time.Now()

	// Get HEAD commit for divergence analysis
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %v", err)
	}

	// Get all references (branches)
	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %v", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Skip non-branch references
		if !ref.Name().IsBranch() && !ref.Name().IsRemote() {
			return nil
		}

		// Skip HEAD and current branch reference
		branchName := getBranchDisplayName(ref.Name().String())
		if branchName == "HEAD" || ref.Name() == head.Name() {
			return nil
		}

		// Get the commit this branch points to
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return nil // Skip problematic branches
		}

		// Calculate branch age based on divergence point  
		branchAge, err := calculateBranchAgeFromDivergence(repo, ref.Hash(), head.Hash(), now)
		if err != nil {
			// Fallback to commit timestamp if divergence calculation fails
			branchAge = calculateBranchAge(commit.Author.When, now)
		}

		// Skip if outside time window
		if !cutoffTime.IsZero() && commit.Author.When.Before(cutoffTime) {
			return nil
		}

		// Skip merged branches unless requested
		if !showMerged {
			isMerged, err := isBranchMerged(repo, ref.Hash(), head.Hash())
			if err == nil && isMerged {
				return nil
			}
		}

		// Skip if path filter doesn't match
		if len(pathFilters) > 0 {
			affects, err := branchAffectsPath(repo, ref.Hash(), head.Hash(), pathFilters)
			if err != nil || !affects {
				return nil
			}
		}

		// Get commit count for this branch
		commitCount, err := getBranchCommitCount(repo, ref.Hash(), head.Hash())
		if err != nil {
			commitCount = 0 // Default if we can't calculate
		}

		// Create branch info
		branchInfo := BranchInfo{
			Name:             branchName,
			AgeInDays:        branchAge,
			Status:           "active",
			Risk:             classifyBranchRisk(branchAge),
			LastCommitAuthor: commit.Author.Name,
			LastCommitTime:   commit.Author.When,
			CommitCount:      commitCount,
			DivergencePoint:  ref.Hash().String()[:8],
		}

		branches = append(branches, branchInfo)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort branches by age (oldest first)
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].AgeInDays > branches[j].AgeInDays
	})

	return branches, nil
}

// calculateBranchAge calculates the age of a branch in days
func calculateBranchAge(createdTime, currentTime time.Time) float64 {
	duration := currentTime.Sub(createdTime)
	return duration.Hours() / 24.0
}

// calculateBranchAgeFromDivergence calculates branch age based on divergence point
func calculateBranchAgeFromDivergence(repo *git.Repository, branchHash, mainHash plumbing.Hash, currentTime time.Time) (float64, error) {
	branchCommit, err := repo.CommitObject(branchHash)
	if err != nil {
		return 0, err
	}

	mainCommit, err := repo.CommitObject(mainHash)
	if err != nil {
		return 0, err
	}

	// Find merge base (divergence point)
	mergeBase, err := branchCommit.MergeBase(mainCommit)
	if err != nil || len(mergeBase) == 0 {
		// If no merge base found, use branch commit time
		return calculateBranchAge(branchCommit.Author.When, currentTime), nil
	}

	// Calculate age from divergence point
	divergenceTime := mergeBase[0].Author.When
	return calculateBranchAge(divergenceTime, currentTime), nil
}

// classifyBranchRisk classifies a branch based on its age
func classifyBranchRisk(ageInDays float64) string {
	if ageInDays <= healthyBranchMaxAge {
		return "Healthy"
	} else if ageInDays <= warningBranchMaxAge {
		return "Warning"
	} else if ageInDays <= criticalBranchMaxAge {
		return "Risky"
	}
	return "Critical"
}

// classifyTrunkBasedCompliance assesses adherence to trunk-based development
func classifyTrunkBasedCompliance(healthyBranches, totalBranches int) string {
	if totalBranches == 0 {
		return "Unknown"
	}

	healthyRatio := float64(healthyBranches) / float64(totalBranches)

	if healthyRatio >= excellentComplianceThreshold {
		return "Excellent"
	} else if healthyRatio >= goodComplianceThreshold {
		return "Good"
	} else if healthyRatio >= moderateComplianceThreshold {
		return "Moderate"
	} else if healthyRatio >= poorComplianceThreshold {
		return "Poor"
	}
	return "Critical"
}

// calculateLongLivedBranchesStats computes comprehensive branch statistics
func calculateLongLivedBranchesStats(branches []BranchInfo) *LongLivedBranchesStats {
	stats := &LongLivedBranchesStats{
		TotalBranches: len(branches),
		Branches:      branches,
	}

	if len(branches) == 0 {
		stats.TrunkBasedCompliance = "Unknown"
		return stats
	}

	// Calculate basic statistics
	totalAge := 0.0
	var riskyBranches []BranchInfo
	var oldestBranch *BranchInfo

	for i, branch := range branches {
		totalAge += branch.AgeInDays

		// Ensure risk is classified (in case it wasn't set in input data)
		if branch.Risk == "" {
			branch.Risk = classifyBranchRisk(branch.AgeInDays)
			stats.Branches[i].Risk = branch.Risk // Update the original slice
		}

		// Track oldest branch
		if oldestBranch == nil || branch.AgeInDays > oldestBranch.AgeInDays {
			branchCopy := branch // Create a copy to avoid pointer issues
			oldestBranch = &branchCopy
		}

		// Count by risk classification
		switch branch.Risk {
		case "Healthy":
			stats.HealthyBranches++
		case "Warning":
			stats.WarningBranches++
		case "Risky":
			stats.RiskyBranches++
			riskyBranches = append(riskyBranches, branch)
		case "Critical":
			stats.CriticalBranches++
			riskyBranches = append(riskyBranches, branch)
		}
	}

	stats.AverageBranchAge = totalAge / float64(len(branches))
	stats.OldestBranch = oldestBranch
	stats.RiskyBranchDetails = riskyBranches
	stats.TrunkBasedCompliance = classifyTrunkBasedCompliance(stats.HealthyBranches, stats.TotalBranches)

	return stats
}

// Helper functions for Git operations

// getBranchDisplayName extracts a clean branch name from a reference
func getBranchDisplayName(refName string) string {
	// Remove refs/heads/ or refs/remotes/origin/ prefixes
	name := strings.TrimPrefix(refName, "refs/heads/")
	name = strings.TrimPrefix(name, "refs/remotes/origin/")
	return name
}

// isBranchMerged checks if a branch has been merged into the main branch
func isBranchMerged(repo *git.Repository, branchHash, mainHash plumbing.Hash) (bool, error) {
	// Check if branchHash is reachable from mainHash using commit ancestry
	mainCommit, err := repo.CommitObject(mainHash)
	if err != nil {
		return false, err
	}
	
	// Walk through main branch history to find if branch commit exists
	commitIter := object.NewCommitPreorderIter(mainCommit, nil, nil)
	defer commitIter.Close()
	
	return commitIter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == branchHash {
			return ErrIterationComplete // Use sentinel error to break iteration
		}
		return nil
	}) != nil, nil
}

// branchAffectsPath checks if a branch has changes affecting the specified path
func branchAffectsPath(repo *git.Repository, branchHash, mainHash plumbing.Hash, pathFilters []string) (bool, error) {
	// Get commits unique to this branch
	branchCommit, err := repo.CommitObject(branchHash)
	if err != nil {
		return false, err
	}

	// Simple path check - in practice, you'd want to diff against the divergence point
	tree, err := branchCommit.Tree()
	if err != nil {
		return false, err
	}

	found := false
	err = tree.Files().ForEach(func(file *object.File) error {
		if matchesPathFilter(file.Name, pathFilters) {
			found = true
		}
		return nil
	})

	return found, err
}

// getBranchCommitCount counts commits unique to this branch
func getBranchCommitCount(repo *git.Repository, branchHash, mainHash plumbing.Hash) (int, error) {
	// Get merge base to find divergence point
	branchCommit, err := repo.CommitObject(branchHash)
	if err != nil {
		return 0, err
	}

	mainCommit, err := repo.CommitObject(mainHash)
	if err != nil {
		return 0, err
	}

	// Find common ancestor (simplified approach)
	mergeBase, err := branchCommit.MergeBase(mainCommit)
	if err != nil || len(mergeBase) == 0 {
		return 0, err
	}

	// Count commits from branch tip to merge base
	count := 0
	commitIter := object.NewCommitPreorderIter(branchCommit, nil, nil)
	defer commitIter.Close()

	err = commitIter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == mergeBase[0].Hash {
			return ErrReachedMergeBase // Stop counting
		}
		count++
		// Limit to avoid excessive counting
		if count > maxCommitCountLimit {
			return ErrBranchTooLarge
		}
		return nil
	})

	return count, nil
}

// printLongLivedBranchesStats displays the analysis results
func printLongLivedBranchesStats(stats *LongLivedBranchesStats, limit int) {
	fmt.Println("Long-Lived Branches Analysis")
	fmt.Printf("Total branches analyzed: %d\n", stats.TotalBranches)

	if stats.TotalBranches == 0 {
		fmt.Println("No branches found for analysis.")
		return
	}

	fmt.Printf("Average branch age: %.1f days\n", stats.AverageBranchAge)
	fmt.Println()

	// Risk distribution
	fmt.Println("Branch Risk Distribution:")
	fmt.Printf("  Healthy (≤%.0f days): %d branches\n", healthyBranchMaxAge, stats.HealthyBranches)
	fmt.Printf("  Warning (%.0f-%.0f days): %d branches\n", healthyBranchMaxAge+1, warningBranchMaxAge, stats.WarningBranches)
	fmt.Printf("  Risky (%.0f-%.0f days): %d branches\n", warningBranchMaxAge+1, criticalBranchMaxAge, stats.RiskyBranches)
	fmt.Printf("  Critical (>%.0f days): %d branches\n", criticalBranchMaxAge, stats.CriticalBranches)
	fmt.Println()

	// Trunk-based compliance
	fmt.Printf("Trunk-Based Development Compliance: %s\n", stats.TrunkBasedCompliance)
	healthyPercentage := float64(stats.HealthyBranches) / float64(stats.TotalBranches) * 100
	fmt.Printf("  %.1f%% of branches are healthy (≤%.0f days)\n", healthyPercentage, healthyBranchMaxAge)
	fmt.Println()

	// Context and research
	fmt.Println("Context: DORA research shows elite teams merge frequently and keep branches short-lived.")
	fmt.Println()

	// Oldest branch
	if stats.OldestBranch != nil {
		fmt.Printf("Oldest Branch: %s (%.1f days old)\n", stats.OldestBranch.Name, stats.OldestBranch.AgeInDays)
		fmt.Printf("  Last commit by: %s\n", stats.OldestBranch.LastCommitAuthor)
		fmt.Printf("  Last commit: %s\n", stats.OldestBranch.LastCommitTime.Format("2006-01-02 15:04"))
		fmt.Println()
	}

	// Risky branches details
	if len(stats.RiskyBranchDetails) > 0 {
		fmt.Printf("Risky Branches (showing up to %d):\n", limit)
		displayCount := len(stats.RiskyBranchDetails)
		if displayCount > limit {
			displayCount = limit
		}

		for i := 0; i < displayCount; i++ {
			branch := stats.RiskyBranchDetails[i]
			fmt.Printf("  %s: %.1f days (%s risk)\n", branch.Name, branch.AgeInDays, branch.Risk)
			fmt.Printf("    Last commit by: %s (%s)\n", branch.LastCommitAuthor, branch.LastCommitTime.Format("2006-01-02"))
		}
		fmt.Println()
	}

	// Recommendations
	fmt.Println("Recommendations:")
	if stats.TrunkBasedCompliance == "Critical" || stats.TrunkBasedCompliance == "Poor" {
		fmt.Println("  • Consider adopting trunk-based development practices")
		fmt.Println("  • Merge feature branches daily or every few days")
		fmt.Println("  • Break large features into smaller, incremental changes")
	}
	if stats.CriticalBranches > 0 {
		fmt.Println("  • Review and merge critical long-lived branches immediately")
		fmt.Println("  • Consider breaking large changes into smaller pull requests")
	}
	if stats.RiskyBranches > 0 {
		fmt.Println("  • Schedule merging of risky branches to reduce integration risk")
	}
	if stats.TrunkBasedCompliance == "Excellent" {
		fmt.Println("  • Excellent trunk-based development practices! Keep it up.")
	}
	if len(stats.RiskyBranchDetails) > limit {
		fmt.Printf("  • %d additional risky branches not shown - use --limit to see more\n", len(stats.RiskyBranchDetails)-limit)
	}
}
