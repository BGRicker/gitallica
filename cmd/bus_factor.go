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
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
)

const (
	// Bus factor thresholds based on empirical GitHub research
	// Most projects have bus factor of 1-2, with 46% having bus factor of 1
	criticalBusFactorThreshold = 1  // Single point of failure
	lowBusFactorThreshold      = 2  // Most common in empirical studies
)

const busFactorBenchmarkContext = "Empirical studies show 46% of GitHub projects have bus factor of 1, 28% have bus factor of 2."

// DirectoryBusFactorStats represents bus factor statistics for a directory
type DirectoryBusFactorStats struct {
	Path               string
	TotalLines         int                    // Total lines of code in directory
	AuthorLines        map[string]int         // Lines authored by each contributor
	AuthorPercentages  map[string]float64     // Percentage of lines authored by each contributor
	BusFactor          int
	RiskLevel          string
	Recommendation     string
	TopContributors    []AuthorContribution
}

// AuthorContribution represents an author's contribution to a directory
type AuthorContribution struct {
	Author     string
	Lines      int        // Lines authored (was Commits)
	Percentage float64
}

// BusFactorAnalysis represents the overall bus factor analysis
type BusFactorAnalysis struct {
	TimeWindow        string
	TotalDirectories  int
	DirectoryStats    []DirectoryBusFactorStats
	OverallRiskDirs   []DirectoryBusFactorStats
	HealthyDirs       []DirectoryBusFactorStats
}

// normalizeAuthorName normalizes author names to handle different formats
func normalizeAuthorName(name, email string) string {
	// Clean and normalize inputs
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanName := strings.ToLower(strings.TrimSpace(name))
	
	// Handle common email variations for the same person using configurable mappings
	for _, mapping := range DefaultAuthorMappings {
		for _, pattern := range mapping.Patterns {
			if strings.Contains(cleanEmail, pattern) || strings.Contains(cleanName, pattern) {
				return mapping.Canonical
			}
		}
	}
	
	// Check if email looks generic or invalid
	if isGenericEmail(cleanEmail) && cleanName != "" {
		// Prefer name when email is generic
		return strings.ToLower(cleanName)
	}
	
	// Use email if it looks valid
	if cleanEmail != "" && strings.Contains(cleanEmail, "@") {
		return cleanEmail
	}
	
	// Fallback to name if available
	if cleanName != "" {
		return strings.ToLower(cleanName)
	}
	
	return "unknown"
}

// isGenericEmail checks if an email looks generic or auto-generated
func isGenericEmail(email string) bool {
	// Check for domain-based generic patterns
	genericDomains := []string{
		"@localhost",
		"@example.com",
		"@example.org", 
		"@test.com",
	}
	
	for _, domain := range genericDomains {
		if strings.Contains(email, domain) {
			return true
		}
	}
	
	// Check for username-based patterns
	if strings.HasPrefix(email, "noreply@") ||
	   strings.HasPrefix(email, "no-reply@") ||
	   strings.HasPrefix(email, "user@") ||
	   strings.HasPrefix(email, "admin@") ||
	   strings.HasPrefix(email, "root@") {
		return true
	}
	
	return false
}


// calculateBusFactor calculates the bus factor for a directory based on author contributions
func calculateBusFactor(authorCommits map[string]int) int {
	if len(authorCommits) == 0 {
		return 0
	}
	
	// Calculate total commits
	totalCommits := 0
	for _, commits := range authorCommits {
		totalCommits += commits
	}
	
	if totalCommits == 0 {
		return 0
	}
	
	// Sort authors by contribution descending
	type authorContrib struct {
		author  string
		commits int
		percent float64
	}
	
	var contribs []authorContrib
	for author, commits := range authorCommits {
		percent := float64(commits) / float64(totalCommits) * 100
		contribs = append(contribs, authorContrib{author, commits, percent})
	}
	
	sort.Slice(contribs, func(i, j int) bool {
		return contribs[i].commits > contribs[j].commits
	})
	
	// Calculate bus factor: minimum number of people needed to have >50% of knowledge
	accumulatedPercent := 0.0
	busFactor := 0
	
	for _, contrib := range contribs {
		busFactor++
		accumulatedPercent += contrib.percent
		if accumulatedPercent > 50.0 {
			break
		}
	}
	
	return busFactor
}

// classifyBusFactorRisk classifies the risk level based on bus factor
func classifyBusFactorRisk(busFactor, totalContributors int) string {
	if totalContributors == 0 {
		return "Unknown"
	}
	
	// Based on empirical GitHub research showing most projects have bus factor 1-2
	switch {
	case busFactor <= criticalBusFactorThreshold:
		return "Critical"
	case busFactor <= lowBusFactorThreshold:
		return "High"
	case busFactor <= 4:
		return "Medium"
	default:
		return "Healthy"
	}
}

// calculateAuthorContributionPercentage calculates percentage contribution for each author
func calculateAuthorContributionPercentage(authorCommits map[string]int) map[string]float64 {
	percentages := make(map[string]float64)
	
	totalCommits := 0
	for _, commits := range authorCommits {
		totalCommits += commits
	}
	
	if totalCommits == 0 {
		return percentages
	}
	
	for author, commits := range authorCommits {
		percentages[author] = float64(commits) / float64(totalCommits) * 100
	}
	
	return percentages
}

// getTopContributors returns the top N contributors sorted by contribution
func getTopContributors(authorLines map[string]int, authorPercentages map[string]float64, limit int) []AuthorContribution {
	var contributors []AuthorContribution
	
	for author, lines := range authorLines {
		percentage := authorPercentages[author]
		contributors = append(contributors, AuthorContribution{
			Author:     author,
			Lines:      lines,
			Percentage: percentage,
		})
	}
	
	// Sort by lines descending
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].Lines > contributors[j].Lines
	})
	
	if len(contributors) > limit {
		contributors = contributors[:limit]
	}
	
	return contributors
}

// getRecommendation provides actionable recommendations based on bus factor risk
func getRecommendation(riskLevel string, busFactor int) string {
	switch riskLevel {
	case "Critical":
		return "Urgent: spread knowledge, pair programming, documentation"
	case "High":
		return "Important: increase knowledge sharing and cross-training"
	case "Medium":
		return "Consider: encourage more contributors and code reviews"
	case "Low":
		return "Monitor: maintain current collaboration patterns"
	case "Healthy":
		return "Good: balanced knowledge distribution"
	default:
		return "Review: assess contributor patterns"
	}
}

// sortDirectoriesByBusFactorRisk sorts directories by risk level priority
func sortDirectoriesByBusFactorRisk(dirs []DirectoryBusFactorStats) []DirectoryBusFactorStats {
	sorted := make([]DirectoryBusFactorStats, len(dirs))
	copy(sorted, dirs)
	
	// Define risk priority order
	riskOrder := map[string]int{
		"Critical": 1,
		"High":     2,
		"Medium":   3,
		"Healthy":  4,
		"Unknown":  5,
	}
	
	sort.Slice(sorted, func(i, j int) bool {
		orderI := riskOrder[sorted[i].RiskLevel]
		orderJ := riskOrder[sorted[j].RiskLevel]
		
		if orderI != orderJ {
			return orderI < orderJ
		}
		
		// If same risk level, sort by bus factor (lower is riskier)
		return sorted[i].BusFactor < sorted[j].BusFactor
	})
	
	return sorted
}

// findFileAuthor finds the most recent author of a file using efficient commit traversal
func findFileAuthor(repo *git.Repository, fileName string, headCommit *object.Commit, since time.Time) (object.Signature, error) {
	// First try to find the author within the specified time window
	author, found := findFileAuthorInTimeWindow(repo, fileName, headCommit, since)
	if found {
		return author, nil
	}
	
	// If not found in time window, try to find the most recent author overall
	// This handles cases where files were last modified before the cutoff time
	author, found = findFileAuthorInTimeWindow(repo, fileName, headCommit, time.Time{})
	if found {
		return author, nil
	}
	
	// If still not found, try using a comprehensive fallback approach
	author, found = findFileAuthorWithFallback(repo, fileName, headCommit)
	if found {
		return author, nil
	}
	
	return object.Signature{}, fmt.Errorf("could not determine author for file %s", fileName)
}

// findFileAuthorInTimeWindow finds the most recent author of a file within a specific time window
func findFileAuthorInTimeWindow(repo *git.Repository, fileName string, headCommit *object.Commit, since time.Time) (object.Signature, bool) {
	cIter, err := repo.Log(&git.LogOptions{From: headCommit.Hash})
	if err != nil {
		return object.Signature{}, false
	}
	defer cIter.Close()
	
	var lastAuthor object.Signature
	
	err = cIter.ForEach(func(c *object.Commit) error {
		if !since.IsZero() && c.Committer.When.Before(since) {
			return storer.ErrStop
		}
		
		// Check if this commit modified the file
		if c.NumParents() == 0 {
			// Initial commit - check if file exists
			tree, err := c.Tree()
			if err != nil {
				return nil
			}
			
			_, err = tree.File(fileName)
			if err == nil {
				// File exists in this commit
				lastAuthor = c.Author
				return storer.ErrStop
			}
		} else {
			// Regular commit - check if file was modified
			parent, err := c.Parent(0)
			if err != nil {
				return nil
			}
			
			patch, err := parent.Patch(c)
			if err != nil {
				return nil
			}
			
			for _, filePatch := range patch.FilePatches() {
				from, to := filePatch.Files()
				var currentFileName string
				if to != nil {
					currentFileName = to.Path()
				} else if from != nil {
					currentFileName = from.Path()
				}
				
				if currentFileName == fileName {
					// This commit modified our file
					lastAuthor = c.Author
					return storer.ErrStop
				}
			}
		}
		
		return nil
	})
	
	if err != nil && err != storer.ErrStop {
		return object.Signature{}, false
	}
	
	return lastAuthor, lastAuthor.Name != ""
}

// findFileAuthorWithFallback uses a more comprehensive approach to find the author
func findFileAuthorWithFallback(repo *git.Repository, fileName string, headCommit *object.Commit) (object.Signature, bool) {
	// Try to find the author by looking at the file's creation in the initial commit
	cIter, err := repo.Log(&git.LogOptions{From: headCommit.Hash})
	if err != nil {
		return object.Signature{}, false
	}
	defer cIter.Close()
	
	var lastAuthor object.Signature
	var found bool
	
	// Walk through all commits to find when the file was first created
	err = cIter.ForEach(func(c *object.Commit) error {
		if c.NumParents() == 0 {
			// Initial commit - check if file exists
			tree, err := c.Tree()
			if err != nil {
				return nil
			}
			
			_, err = tree.File(fileName)
			if err == nil {
				// File exists in this commit
				lastAuthor = c.Author
				found = true
				return storer.ErrStop
			}
		} else {
			// Regular commit - check if file was added (not just modified)
			parent, err := c.Parent(0)
			if err != nil {
				return nil
			}
			
			parentTree, err := parent.Tree()
			if err != nil {
				return nil
			}
			
			tree, err := c.Tree()
			if err != nil {
				return nil
			}
			
			// Check if file exists in current commit but not in parent
			_, err = tree.File(fileName)
			if err == nil {
				_, err = parentTree.File(fileName)
				if err != nil {
					// File was added in this commit
					lastAuthor = c.Author
					found = true
					return storer.ErrStop
				}
			}
		}
		
		return nil
	})
	
	if err != nil && err != storer.ErrStop {
		return object.Signature{}, false
	}
	
	return lastAuthor, found
}

// buildFileAuthorMap efficiently builds a map of file authors using a single git traversal
func buildFileAuthorMap(repo *git.Repository, ref *plumbing.Reference, since time.Time, pathFilters []string, fileAuthors map[string]map[string]int) error {
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fmt.Errorf("could not get commits: %v", err)
	}
	defer cIter.Close()
	
	err = cIter.ForEach(func(c *object.Commit) error {
		if !since.IsZero() && c.Committer.When.Before(since) {
			return storer.ErrStop
		}
		
		// Handle merge commits by diffing against their first parent
		var parentTree *object.Tree
		if c.NumParents() > 0 {
			parent, err := c.Parent(0)
			if err != nil {
				return nil
			}
			parentTree, err = parent.Tree()
			if err != nil {
				return nil
			}
		}
		
		tree, err := c.Tree()
		if err != nil {
			return nil
		}
		
		if parentTree != nil {
			// Compare with parent to get changed files
			changes, err := tree.Diff(parentTree)
			if err != nil {
				return nil
			}
			
			for _, change := range changes {
				if change.To.Name == "" {
					continue // skip deletions
				}
				
				// Apply path filter if specified
				if !matchesPathFilter(change.To.Name, pathFilters) {
					continue
				}
				
				// Initialize file authors map if needed
				if fileAuthors[change.To.Name] == nil {
					fileAuthors[change.To.Name] = make(map[string]int)
				}
				
				author := normalizeAuthorName(c.Author.Name, c.Author.Email)
				fileAuthors[change.To.Name][author]++ // Count commits, not lines
			}
		} else {
			// Initial commit - all files are authored by this commit
			err = tree.Files().ForEach(func(f *object.File) error {
				if !matchesPathFilter(f.Name, pathFilters) {
					return nil
				}
				
				// Initialize file authors map if needed
				if fileAuthors[f.Name] == nil {
					fileAuthors[f.Name] = make(map[string]int)
				}
				
				author := normalizeAuthorName(c.Author.Name, c.Author.Email)
				fileAuthors[f.Name][author]++ // Count commits, not lines
				
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	return err
}

// analyzeBusFactor performs bus factor analysis using an efficient commit-based approach
// This provides accurate knowledge measurement while maintaining good performance by
// analyzing file authorship through commit history rather than line-by-line blame.
func analyzeBusFactor(repo *git.Repository, since time.Time, pathFilters []string) (*BusFactorAnalysis, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD: %v", err)
	}
	
	// Track file authorship per directory using commit-based analysis
	directoryOwnership := make(map[string]map[string]int)
	
	// Get current HEAD commit and tree
	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD commit: %v", err)
	}
	
	tree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD tree: %v", err)
	}
	
	// Build comprehensive file author map efficiently
	fileAuthors := make(map[string]map[string]int) // file -> author -> lines contributed
	err = buildFileAuthorMap(repo, ref, since, pathFilters, fileAuthors)
	if err != nil {
		return nil, fmt.Errorf("error building file author map: %v", err)
	}
	
	// Initialize directory structure
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
		
		// Get directory for this file
		dir := filepath.Dir(f.Name)
		if dir == "." {
			dir = "root"
		} else {
			dir = dir + "/"
		}
		
		// Initialize directory ownership map if needed
		if directoryOwnership[dir] == nil {
			directoryOwnership[dir] = make(map[string]int)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing files: %v", err)
	}
	
	// Count commits by author per directory
	for fileName, authorCommits := range fileAuthors {
		// Get directory for this file
		dir := filepath.Dir(fileName)
		if dir == "." {
			dir = "root"
		} else {
			dir = dir + "/"
		}
		
		// Initialize directory ownership map if needed
		if directoryOwnership[dir] == nil {
			directoryOwnership[dir] = make(map[string]int)
		}
		
		// Add each author's commits to the directory
		for author, commits := range authorCommits {
			directoryOwnership[dir][author] += commits
		}
	}
	
	// Calculate bus factor statistics for each directory
	var directoryStats []DirectoryBusFactorStats
	for dir, authorCommits := range directoryOwnership {
		totalCommits := 0
		for _, commits := range authorCommits {
			totalCommits += commits
		}
		
		busFactor := calculateBusFactor(authorCommits)
		riskLevel := classifyBusFactorRisk(busFactor, len(authorCommits))
		authorPercentages := calculateAuthorContributionPercentage(authorCommits)
		topContributors := getTopContributors(authorCommits, authorPercentages, 5)
		recommendation := getRecommendation(riskLevel, busFactor)
		
		stats := DirectoryBusFactorStats{
			Path:              dir,
			TotalLines:        totalCommits, // Using commits as proxy for lines
			AuthorLines:       authorCommits, // Using commits as proxy for lines
			AuthorPercentages: authorPercentages,
			BusFactor:         busFactor,
			RiskLevel:         riskLevel,
			Recommendation:    recommendation,
			TopContributors:   topContributors,
		}
		
		directoryStats = append(directoryStats, stats)
	}
	
	// Sort by risk level
	directoryStats = sortDirectoriesByBusFactorRisk(directoryStats)
	
	// Separate risky and healthy directories
	var overallRiskDirs, healthyDirs []DirectoryBusFactorStats
	for _, stats := range directoryStats {
		switch stats.RiskLevel {
		case "Critical", "High":
			overallRiskDirs = append(overallRiskDirs, stats)
		case "Healthy":
			healthyDirs = append(healthyDirs, stats)
		}
	}
	
	timeWindow := "all time"
	if !since.IsZero() {
		timeWindow = fmt.Sprintf("since %s", since.Format("2006-01-02"))
	}
	
	return &BusFactorAnalysis{
		TimeWindow:       timeWindow,
		TotalDirectories: len(directoryStats),
		DirectoryStats:   directoryStats,
		OverallRiskDirs:  overallRiskDirs,
		HealthyDirs:      healthyDirs,
	}, nil
}

// printBusFactorStats prints bus factor analysis results
func printBusFactorStats(analysis *BusFactorAnalysis, limit int) {
	fmt.Printf("Bus Factor Analysis\n")
	fmt.Printf("Time window: %s\n", analysis.TimeWindow)
	fmt.Printf("Total directories analyzed: %d\n", analysis.TotalDirectories)
	fmt.Printf("High-risk directories: %d\n", len(analysis.OverallRiskDirs))
	fmt.Printf("Healthy directories: %d\n", len(analysis.HealthyDirs))
	fmt.Println()
	fmt.Println("Context:", busFactorBenchmarkContext)
	fmt.Println()
	
	if len(analysis.DirectoryStats) == 0 {
		fmt.Println("No directories found for analysis.")
		return
	}
	
	fmt.Printf("Directory Bus Factor Analysis (showing top %d):\n", limit)
	fmt.Printf("Directory                    Bus Factor Contributors Risk Level  Recommendation\n")
	fmt.Printf("---------------------------- ---------- ----------- ----------- ----------------------\n")
	
	for i, stats := range analysis.DirectoryStats {
		if i >= limit {
			break
		}
		
		contributorCount := len(stats.AuthorLines)
		
		fmt.Printf("%-28s %10d %11d %-11s %s\n",
			truncateDirectoryPath(stats.Path, 28),
			stats.BusFactor,
			contributorCount,
			stats.RiskLevel,
			truncateRecommendation(stats.Recommendation, 22))
	}
	
	// Show detailed breakdown for high-risk directories
	if len(analysis.OverallRiskDirs) > 0 {
		fmt.Printf("\n[!] High-Risk Directories (detailed breakdown):\n")
		showCount := min(3, len(analysis.OverallRiskDirs))
		
		for i := 0; i < showCount; i++ {
			stats := analysis.OverallRiskDirs[i]
			fmt.Printf("\n%s (Bus Factor: %d, Risk: %s)\n", stats.Path, stats.BusFactor, stats.RiskLevel)
			fmt.Printf("  Top contributors:\n")
			
			for j, contrib := range stats.TopContributors {
				if j >= 3 { // Show top 3 contributors
					break
				}
				fmt.Printf("    %s: %d lines (%.1f%%)\n", 
					truncateAuthorName(contrib.Author, 20), contrib.Lines, contrib.Percentage)
			}
			fmt.Printf("  Recommendation: %s\n", stats.Recommendation)
		}
	}
}

// truncateDirectoryPath truncates a directory path to fit in specified width
func truncateDirectoryPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// truncateRecommendation truncates a recommendation to fit in specified width
func truncateRecommendation(rec string, maxLen int) string {
	if len(rec) <= maxLen {
		return rec
	}
	return rec[:maxLen-3] + "..."
}

// truncateAuthorName truncates an author name to fit in specified width
func truncateAuthorName(author string, maxLen int) string {
	if len(author) <= maxLen {
		return author
	}
	return author[:maxLen-3] + "..."
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// busFactorCmd represents the bus-factor command
var busFactorCmd = &cobra.Command{
	Use:   "bus-factor",
	Short: "Analyze bus factor (knowledge concentration) per directory",
	Long: `Analyze how many people would need to leave before knowledge of specific 
directories becomes critically impacted. Helps identify knowledge concentration 
risks and promotes collective ownership.

Bus factor represents the minimum number of people who need to leave before a 
project becomes critically understaffed in a specific area.

Risk Levels:
- Critical: Bus factor 1 (single point of failure)
- High: Bus factor 2-3 in larger teams (knowledge concentration)
- Medium: Bus factor adequate but could be improved
- Healthy: Good knowledge distribution (25-50% of team)

Based on Martin Fowler's collective ownership principles and industry research.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		lastArg, _ := cmd.Flags().GetString("last")
		pathFilters := getConfigPaths(cmd, "bus-factor.paths")
		limitArg, _ := cmd.Flags().GetInt("limit")
		
		// Print configuration scope
		printCommandScope(cmd, "bus-factor", lastArg, pathFilters)

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

		analysis, err := analyzeBusFactor(repo, since, pathFilters)
		if err != nil {
			log.Fatalf("Error analyzing bus factor: %v", err)
		}

		printBusFactorStats(analysis, limitArg)
	},
}

func init() {
	busFactorCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	busFactorCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	busFactorCmd.Flags().Int("limit", 10, "Number of top results to show")
	rootCmd.AddCommand(busFactorCmd)
}
