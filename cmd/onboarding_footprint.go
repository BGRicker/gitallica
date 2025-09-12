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
	// Research-backed thresholds based on Clean Code principles for onboarding complexity
	
	// onboardingSimpleThreshold: 1-5 files in first commits indicates focused onboarding
	onboardingSimpleThreshold = 5
	
	// onboardingModerateThreshold: 6-10 files indicates reasonable complexity  
	onboardingModerateThreshold = 10
	
	// onboardingComplexThreshold: 11-20 files may signal steep onboarding curve
	onboardingComplexThreshold = 20
	
	// onboardingDefaultCommitLimit: analyze first 5 commits for onboarding patterns
	onboardingDefaultCommitLimit = 5
)

const onboardingBenchmarkContext = "Scoped, small first tasks help new developers succeed (Clean Code - Robert C. Martin)."

// OnboardingFootprintStats represents the onboarding footprint analysis
type OnboardingFootprintStats struct {
	TotalContributors      int
	AnalyzedContributors   int
	AverageFilesTouched    float64
	SimpleOnboarding       int
	ModerateOnboarding     int
	ComplexOnboarding      int
	OverwhelmingOnboarding int
	Contributors           []NewContributor
	CommonFiles           []FilePopularity
	TimeWindow            string
}

// NewContributor represents a new contributor's onboarding pattern
type NewContributor struct {
	Email              string
	FirstCommitTime    time.Time
	FilesTouched       int
	CommitsAnalyzed    int
	Status             string
	Recommendation     string
	FilesModified      []string
}

// FilePopularity represents how often a file is touched by new contributors
type FilePopularity struct {
	FilePath    string
	TouchCount  int
	Percentage  float64
}

// classifyOnboardingComplexity classifies onboarding complexity based on files touched
func classifyOnboardingComplexity(filesCount int) (string, string) {
	switch {
	case filesCount <= onboardingSimpleThreshold:
		return "Simple", "Excellent focused onboarding"
	case filesCount <= onboardingModerateThreshold:
		return "Moderate", "Reasonable onboarding complexity"
	case filesCount <= onboardingComplexThreshold:
		return "Complex", "Consider simplifying initial tasks"
	default:
		return "Overwhelming", "Urgent: simplify onboarding process"
	}
}

// analyzeOnboardingFootprint analyzes onboarding patterns in the repository
func analyzeOnboardingFootprint(repo *git.Repository, pathFilters []string, lastArg string, commitLimit int) (*OnboardingFootprintStats, error) {
	var since *time.Time
	timeWindow := "all time"
	
	if lastArg != "" {
		sinceTime, err := parseDurationArg(lastArg)
		if err != nil {
			return nil, fmt.Errorf("invalid time window: %v", err)
		}
		since = &sinceTime
		timeWindow = fmt.Sprintf("since %s", since.Format("2006-01-02"))
	}
	
	// Single-pass analysis: find first commits AND gather commit data efficiently
	// This prevents memory issues from loading full history twice
	commitIter, err := repo.Log(&git.LogOptions{
		// No Since filter - we need full history to find true first commits
	})
	if err != nil {
		return nil, fmt.Errorf("could not get commit log: %v", err)
	}
	defer commitIter.Close()
	
	// Track data during single pass
	authorTrueFirstCommit := make(map[string]time.Time)
	allCommitData := make(map[string][]*CommitInfo) // Store all commits by author
	
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Skip commits without author information
		if commit.Author.Email == "" {
			return nil
		}
		
		author := commit.Author.Email
		commitTime := commit.Author.When
		
		// Skip merge commits for cleaner analysis
		if commit.NumParents() > 1 {
			return nil
		}
		
		// Track the earliest commit time for each author across ALL history
		if firstTime, exists := authorTrueFirstCommit[author]; !exists || commitTime.Before(firstTime) {
			authorTrueFirstCommit[author] = commitTime
		}
		
		// Get files changed in this commit
		var filesChanged []string
		
		if commit.NumParents() == 0 {
			// Initial commit - treat as adding all files
			tree, err := commit.Tree()
			if err != nil {
				return err
			}
			
			err = tree.Files().ForEach(func(file *object.File) error {
				if matchesPathFilter(file.Name, pathFilters) {
					filesChanged = append(filesChanged, file.Name)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			// Regular commit - analyze diff with parent
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
			
			changes, err := parentTree.Diff(currentTree)
			if err != nil {
				return err
			}
			
			for _, change := range changes {
				var filePath string
				if change.To.Name != "" {
					filePath = change.To.Name
				} else if change.From.Name != "" {
					filePath = change.From.Name
				}
				
				if filePath != "" && matchesPathFilter(filePath, pathFilters) {
					filesChanged = append(filesChanged, filePath)
				}
			}
		}
		
		// Store commit info for all authors during single pass
		allCommitData[author] = append(allCommitData[author], &CommitInfo{
			Hash:    commit.Hash.String(),
			Time:    commitTime,
			Files:   filesChanged,
			Message: commit.Message,
		})
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing commits: %v", err)
	}
	
	// After single pass: filter to only "new" contributors whose first commit falls within time window
	var newContributors []string
	if since != nil {
		for author, firstCommit := range authorTrueFirstCommit {
			if firstCommit.After(*since) || firstCommit.Equal(*since) {
				newContributors = append(newContributors, author)
			}
		}
	} else {
		// No time window specified - all contributors are considered "new"
		for author := range authorTrueFirstCommit {
			newContributors = append(newContributors, author)
		}
	}
	
	// Filter commit data to only new contributors for analysis
	contributorCommits := make(map[string][]*CommitInfo)
	for _, author := range newContributors {
		if commits, exists := allCommitData[author]; exists {
			contributorCommits[author] = commits
		}
	}
	
	// Sort commits by time for each new contributor
	for author := range contributorCommits {
		sort.Slice(contributorCommits[author], func(i, j int) bool {
			return contributorCommits[author][i].Time.Before(contributorCommits[author][j].Time)
		})
	}
	
	// Analyze onboarding patterns
	var contributors []NewContributor
	filePopularity := make(map[string]int)
	totalFilesTouched := 0
	
	for author, commits := range contributorCommits {
		if len(commits) == 0 {
			continue
		}
		
		// Analyze first N commits for onboarding pattern
		filesTouched := make(map[string]bool)
		commitsAnalyzed := 0
		
		for i, commit := range commits {
			if i >= commitLimit {
				break
			}
			commitsAnalyzed++
			
			for _, file := range commit.Files {
				filesTouched[file] = true
				filePopularity[file]++
			}
		}
		
		filesCount := len(filesTouched)
		totalFilesTouched += filesCount
		status, recommendation := classifyOnboardingComplexity(filesCount)
		
		// Convert map to slice for storage
		var filesModified []string
		for file := range filesTouched {
			filesModified = append(filesModified, file)
		}
		sort.Strings(filesModified)
		
		contributors = append(contributors, NewContributor{
			Email:              author,
			FirstCommitTime:    authorTrueFirstCommit[author],
			FilesTouched:       filesCount,
			CommitsAnalyzed:    commitsAnalyzed,
			Status:             status,
			Recommendation:     recommendation,
			FilesModified:      filesModified,
		})
	}
	
	// Sort contributors by first commit time (newest first)
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].FirstCommitTime.After(contributors[j].FirstCommitTime)
	})
	
	// Calculate statistics
	var averageFilesTouched float64
	if len(contributors) > 0 {
		averageFilesTouched = float64(totalFilesTouched) / float64(len(contributors))
	}
	
	simpleCount := 0
	moderateCount := 0
	complexCount := 0
	overwhelmingCount := 0
	
	for _, contributor := range contributors {
		switch contributor.Status {
		case "Simple":
			simpleCount++
		case "Moderate":
			moderateCount++
		case "Complex":
			complexCount++
		case "Overwhelming":
			overwhelmingCount++
		}
	}
	
	// Calculate file popularity - use only contributors who actually touched files in scope
	var commonFiles []FilePopularity
	contributorsWithFiles := 0
	for _, contributor := range contributors {
		if contributor.FilesTouched > 0 {
			contributorsWithFiles++
		}
	}
	
	for file, count := range filePopularity {
		var percentage float64
		if contributorsWithFiles > 0 {
			percentage = float64(count) / float64(contributorsWithFiles) * 100
		}
		commonFiles = append(commonFiles, FilePopularity{
			FilePath:   file,
			TouchCount: count,
			Percentage: percentage,
		})
	}
	
	// Sort by popularity
	sort.Slice(commonFiles, func(i, j int) bool {
		return commonFiles[i].TouchCount > commonFiles[j].TouchCount
	})
	
	return &OnboardingFootprintStats{
		TotalContributors:      len(contributorCommits), // All contributors who made commits
		AnalyzedContributors:   len(contributors),       // Contributors with sufficient data for analysis
		AverageFilesTouched:    averageFilesTouched,
		SimpleOnboarding:       simpleCount,
		ModerateOnboarding:     moderateCount,
		ComplexOnboarding:      complexCount,
		OverwhelmingOnboarding: overwhelmingCount,
		Contributors:           contributors,
		CommonFiles:           commonFiles,
		TimeWindow:            timeWindow,
	}, nil
}

// CommitInfo represents commit information for analysis
type CommitInfo struct {
	Hash    string
	Time    time.Time
	Files   []string
	Message string
	Author  string // Optional: author email for analysis
}

// printOnboardingFootprintStats prints the onboarding footprint analysis results
func printOnboardingFootprintStats(stats *OnboardingFootprintStats, pathFilters []string, limit int, commitLimit int) {
	fmt.Printf("Onboarding Footprint Analysis\n")
	if len(pathFilters) > 0 {
		fmt.Printf("Path filters: %s\n", strings.Join(pathFilters, ", "))
	}
	fmt.Printf("Time window: %s\n", stats.TimeWindow)
	fmt.Printf("Contributors found: %d\n", stats.TotalContributors)
	fmt.Printf("Contributors analyzed: %d\n", stats.AnalyzedContributors)
	fmt.Printf("Average files touched in first %d commits: %.1f\n", commitLimit, stats.AverageFilesTouched)
	fmt.Println()
	
	// Summary by complexity
	fmt.Printf("Onboarding Complexity Distribution:\n")
	if stats.AnalyzedContributors == 0 {
		fmt.Printf("  Simple: %d contributors (0.0%%)\n", stats.SimpleOnboarding)
		fmt.Printf("  Moderate: %d contributors (0.0%%)\n", stats.ModerateOnboarding)
		fmt.Printf("  Complex: %d contributors (0.0%%)\n", stats.ComplexOnboarding)
		fmt.Printf("  Overwhelming: %d contributors (0.0%%)\n", stats.OverwhelmingOnboarding)
	} else {
		fmt.Printf("  Simple: %d contributors (%.1f%%)\n", stats.SimpleOnboarding,
			float64(stats.SimpleOnboarding)/float64(stats.AnalyzedContributors)*100)
		fmt.Printf("  Moderate: %d contributors (%.1f%%)\n", stats.ModerateOnboarding,
			float64(stats.ModerateOnboarding)/float64(stats.AnalyzedContributors)*100)
		fmt.Printf("  Complex: %d contributors (%.1f%%)\n", stats.ComplexOnboarding,
			float64(stats.ComplexOnboarding)/float64(stats.AnalyzedContributors)*100)
		fmt.Printf("  Overwhelming: %d contributors (%.1f%%)\n", stats.OverwhelmingOnboarding,
			float64(stats.OverwhelmingOnboarding)/float64(stats.AnalyzedContributors)*100)
	}
	fmt.Println()
	
	fmt.Println("Context:", onboardingBenchmarkContext)
	fmt.Println()
	
	// Show detailed contributor analysis
	displayCount := limit
	if displayCount > len(stats.Contributors) {
		displayCount = len(stats.Contributors)
	}
	
	if displayCount > 0 {
		fmt.Printf("Recent Contributors (showing %d):\n", displayCount)
		for i := 0; i < displayCount; i++ {
			contributor := stats.Contributors[i]
			fmt.Printf("\n%d. %s — %s\n", i+1, contributor.Email, contributor.Status)
			fmt.Printf("   First commit: %s\n", contributor.FirstCommitTime.Format("2006-01-02"))
			fmt.Printf("   Files touched: %d (in first %d commits)\n", 
				contributor.FilesTouched, contributor.CommitsAnalyzed)
			fmt.Printf("   Recommendation: %s\n", contributor.Recommendation)
			
			if len(contributor.FilesModified) > 0 {
				fmt.Printf("   Files: ")
				for j, file := range contributor.FilesModified {
					if j >= 5 { // Show only first 5 files
						fmt.Printf("... (%d more)", len(contributor.FilesModified)-j)
						break
					}
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", file)
				}
				fmt.Println()
			}
		}
	}
	
	// Show common entry point files
	if len(stats.CommonFiles) > 0 {
		fmt.Printf("\nMost Common Entry Point Files:\n")
		commonFilesLimit := 10
		if commonFilesLimit > len(stats.CommonFiles) {
			commonFilesLimit = len(stats.CommonFiles)
		}
		
		for i := 0; i < commonFilesLimit; i++ {
			file := stats.CommonFiles[i]
			fmt.Printf("  %d. %s (%d contributors, %.1f%%)\n", 
				i+1, file.FilePath, file.TouchCount, file.Percentage)
		}
	}
	
	// Provide actionable insights
	fmt.Printf("\nRecommendations:\n")
	
	recommendations := []struct {
		count   int
		message string
	}{
		{stats.OverwhelmingOnboarding, "contributors had overwhelming onboarding - urgent process improvement needed"},
		{stats.ComplexOnboarding, "contributors had complex onboarding - consider simplifying initial tasks"},
		{stats.ModerateOnboarding, "contributors had moderate onboarding complexity"},
		{stats.SimpleOnboarding, "contributors had simple, focused onboarding - excellent!"},
	}
	
	for _, rec := range recommendations {
		if rec.count > 0 {
			fmt.Printf("  • %d %s\n", rec.count, rec.message)
		}
	}
	
	if stats.AverageFilesTouched > float64(onboardingComplexThreshold) {
		fmt.Printf("  • Average files touched (%.1f) exceeds recommended threshold (%d)\n", 
			stats.AverageFilesTouched, onboardingComplexThreshold)
		fmt.Printf("  • Consider creating simpler, more focused first issues for new contributors\n")
	} else if stats.AverageFilesTouched <= float64(onboardingSimpleThreshold) {
		fmt.Printf("  • Excellent onboarding complexity - new contributors have focused entry points\n")
	}
}

// onboardingFootprintCmd represents the onboarding-footprint command
var onboardingFootprintCmd = &cobra.Command{
	Use:   "onboarding-footprint",
	Short: "Analyze what new contributors touch first",
	Long: `Analyze onboarding footprint to understand what files new contributors 
touch in their first commits and identify onboarding complexity patterns.

Helps evaluate whether the codebase provides good entry points for new 
developers or if the onboarding process is too complex.

Research-backed analysis:
- Analyzes first 5 commits by default (configurable with --commit-limit)
- Tracks files touched by new contributors in their early commits
- Identifies common entry point files and onboarding patterns

Complexity Classifications:
- Simple: 1-5 files (excellent focused onboarding)
- Moderate: 6-10 files (reasonable complexity)
- Complex: 11-20 files (consider simplifying)
- Overwhelming: >20 files (urgent improvement needed)

"Developers spend much more time reading code than writing it, so making it 
easy to read makes it easier to write." — Robert C. Martin, Clean Code`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		pathFilters := getConfigPaths(cmd, "onboarding-footprint.paths")
		lastArg, _ := cmd.Flags().GetString("last")
		limit, _ := cmd.Flags().GetInt("limit")
		commitLimit, _ := cmd.Flags().GetInt("commit-limit")
		
		// Print configuration scope
		printCommandScope(cmd, "onboarding-footprint", lastArg, pathFilters)

		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Could not open repository: %v", err)
		}

		stats, err := analyzeOnboardingFootprint(repo, pathFilters, lastArg, commitLimit)
		if err != nil {
			log.Fatalf("Error analyzing onboarding footprint: %v", err)
		}

		printOnboardingFootprintStats(stats, pathFilters, limit, commitLimit)
	},
}

func init() {
	onboardingFootprintCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	onboardingFootprintCmd.Flags().String("last", "", "Limit analysis to recent timeframe (e.g., '30d', '6m', '1y')")
	onboardingFootprintCmd.Flags().Int("limit", 10, "Number of contributors to show in detailed analysis")
	onboardingFootprintCmd.Flags().Int("commit-limit", onboardingDefaultCommitLimit, "Number of initial commits to analyze per contributor")
	rootCmd.AddCommand(onboardingFootprintCmd)
}
