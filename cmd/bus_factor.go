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
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
)

const (
	healthyBusFactorThreshold  = 4
	mediumBusFactorThreshold   = 3
	criticalBusFactorThreshold = 1
)

const busFactorBenchmarkContext = "Target bus factor of 25-50% of team size ensures collective ownership without accountability gaps (Martin Fowler)."

// DirectoryBusFactorStats represents bus factor statistics for a directory
type DirectoryBusFactorStats struct {
	Path               string
	TotalCommits       int
	AuthorCommits      map[string]int
	AuthorPercentages  map[string]float64
	BusFactor          int
	RiskLevel          string
	Recommendation     string
	TopContributors    []AuthorContribution
}

// AuthorContribution represents an author's contribution to a directory
type AuthorContribution struct {
	Author     string
	Commits    int
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
func normalizeAuthorName(author string) string {
	if author == "" {
		return "unknown"
	}
	
	// Extract email from "Name <email>" format
	emailRegex := regexp.MustCompile(`<([^>]+)>`)
	if matches := emailRegex.FindStringSubmatch(author); len(matches) > 1 {
		return strings.ToLower(strings.TrimSpace(matches[1]))
	}
	
	// If it looks like an email, normalize it
	if strings.Contains(author, "@") {
		return strings.ToLower(strings.TrimSpace(author))
	}
	
	// Otherwise, normalize as name
	return strings.ToLower(strings.TrimSpace(author))
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

// classifyBusFactorRisk classifies the risk level based on bus factor and team size
func classifyBusFactorRisk(busFactor, totalContributors int) string {
	if totalContributors == 0 {
		return "Unknown"
	}
	
	// Research-backed thresholds from Martin Fowler's collective ownership principles
	// Target: 25-50% of team size for healthy ownership
	healthyMinimum := max(healthyBusFactorThreshold, totalContributors/4) // At least 25% of team
	
	switch {
	case busFactor <= criticalBusFactorThreshold:
		return "Critical"
	case busFactor <= mediumBusFactorThreshold && totalContributors > 6:
		return "High"
	case busFactor < healthyMinimum:
		if totalContributors <= 6 {
			return "Medium"
		}
		return "High"
	case busFactor >= healthyMinimum:
		return "Healthy"
	default:
		return "Unknown"
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
func getTopContributors(authorCommits map[string]int, authorPercentages map[string]float64, limit int) []AuthorContribution {
	var contributors []AuthorContribution
	
	for author, commits := range authorCommits {
		percentage := authorPercentages[author]
		contributors = append(contributors, AuthorContribution{
			Author:     author,
			Commits:    commits,
			Percentage: percentage,
		})
	}
	
	// Sort by commits descending
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].Commits > contributors[j].Commits
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

// analyzeBusFactor performs bus factor analysis on the repository
func analyzeBusFactor(repo *git.Repository, since time.Time, pathArg string) (*BusFactorAnalysis, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD: %v", err)
	}
	
	// Track author commits per directory
	directoryAuthors := make(map[string]map[string]int)
	
	// Iterate through commits to collect author data
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not get commits: %v", err)
	}
	defer cIter.Close()
	
	err = cIter.ForEach(func(c *object.Commit) error {
		if !since.IsZero() && c.Committer.When.Before(since) {
			return storer.ErrStop
		}
		
		// Normalize author name
		author := normalizeAuthorName(c.Author.Email)
		if author == "" {
			author = normalizeAuthorName(c.Author.Name)
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
			
			// Track directories of changed files
			affectedDirs := make(map[string]bool)
			for _, change := range changes {
				if change.To.Name == "" {
					continue // skip deletions
				}
				
				// Apply path filter if specified
				if pathArg != "" && !strings.HasPrefix(change.To.Name, pathArg) {
					continue
				}
				
				dir := filepath.Dir(change.To.Name)
				if dir == "." {
					dir = "root"
				} else {
					dir = dir + "/"
				}
				
				affectedDirs[dir] = true
			}
			
			// Increment author commits for each affected directory
			for dir := range affectedDirs {
				if directoryAuthors[dir] == nil {
					directoryAuthors[dir] = make(map[string]int)
				}
				directoryAuthors[dir][author]++
			}
		} else {
			// Initial commit - count all files
			err = tree.Files().ForEach(func(f *object.File) error {
				if pathArg != "" && !strings.HasPrefix(f.Name, pathArg) {
					return nil
				}
				
				dir := filepath.Dir(f.Name)
				if dir == "." {
					dir = "root"
				} else {
					dir = dir + "/"
				}
				
				if directoryAuthors[dir] == nil {
					directoryAuthors[dir] = make(map[string]int)
				}
				directoryAuthors[dir][author]++
				
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing commits: %v", err)
	}
	
	// Calculate bus factor statistics for each directory
	var directoryStats []DirectoryBusFactorStats
	for dir, authorCommits := range directoryAuthors {
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
			TotalCommits:      totalCommits,
			AuthorCommits:     authorCommits,
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
		
		contributorCount := len(stats.AuthorCommits)
		
		fmt.Printf("%-28s %10d %11d %-11s %s\n",
			truncateDirectoryPath(stats.Path, 28),
			stats.BusFactor,
			contributorCount,
			stats.RiskLevel,
			truncateRecommendation(stats.Recommendation, 22))
	}
	
	// Show detailed breakdown for high-risk directories
	if len(analysis.OverallRiskDirs) > 0 {
		fmt.Printf("\n⚠️  High-Risk Directories (detailed breakdown):\n")
		showCount := min(3, len(analysis.OverallRiskDirs))
		
		for i := 0; i < showCount; i++ {
			stats := analysis.OverallRiskDirs[i]
			fmt.Printf("\n%s (Bus Factor: %d, Risk: %s)\n", stats.Path, stats.BusFactor, stats.RiskLevel)
			fmt.Printf("  Top contributors:\n")
			
			for j, contrib := range stats.TopContributors {
				if j >= 3 { // Show top 3 contributors
					break
				}
				fmt.Printf("    %s: %d commits (%.1f%%)\n", 
					truncateAuthorName(contrib.Author, 20), contrib.Commits, contrib.Percentage)
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
		pathArg, _ := cmd.Flags().GetString("path")
		limitArg, _ := cmd.Flags().GetInt("limit")

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

		analysis, err := analyzeBusFactor(repo, since, pathArg)
		if err != nil {
			log.Fatalf("Error analyzing bus factor: %v", err)
		}

		printBusFactorStats(analysis, limitArg)
	},
}

func init() {
	busFactorCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	busFactorCmd.Flags().String("path", "", "Limit analysis to a specific path")
	busFactorCmd.Flags().Int("limit", 10, "Number of top results to show")
	rootCmd.AddCommand(busFactorCmd)
}
