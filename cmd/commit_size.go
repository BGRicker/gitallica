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
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
)

const (
	commitSizeLowThreshold      = 100
	commitSizeMediumThreshold   = 400
	commitSizeHighThreshold     = 800
	commitSizeFilesLowThreshold = 5
	commitSizeFilesHighThreshold = 15
)

const commitSizeBenchmarkContext = "Large commits reduce review effectiveness and rollback safety. Studies show reviews are most effective under 400 lines."

// CommitSizeStats represents size and risk statistics for a single commit.
type CommitSizeStats struct {
	Hash         string
	Message      string
	Author       string
	Date         time.Time
	Additions    int
	Deletions    int
	FilesChanged int
	RiskScore    int
	RiskLevel    string
}

// calculateCommitRisk determines the risk level and score for a commit based on its size.
// Uses a weighted scoring system that prioritizes both line changes and file count.
func calculateCommitRisk(additions, deletions, filesChanged int) (string, int) {
	totalChanges := additions + deletions
	
	// Calculate risk score: prioritize large changes and many files
	// Files get 10 points each to reflect increased complexity
	riskScore := totalChanges + (filesChanged * 10)
	
	var riskLevel string
	// Use hybrid logic: score-based with file count modifiers
	switch {
	case riskScore >= commitSizeHighThreshold:
		if filesChanged >= commitSizeFilesHighThreshold {
			riskLevel = "Critical"
		} else {
			riskLevel = "High"
		}
	case riskScore >= commitSizeMediumThreshold:
		if filesChanged >= commitSizeFilesLowThreshold {
			riskLevel = "High"
		} else {
			riskLevel = "Medium"
		}
	case riskScore >= commitSizeLowThreshold:
		if filesChanged < commitSizeFilesLowThreshold {
			riskLevel = "Low"
		} else {
			riskLevel = "Medium"
		}
	default:
		riskLevel = "Low"
	}
	
	return riskLevel, riskScore
}

// sortCommitsByRisk sorts commits by risk score in descending order.
func sortCommitsByRisk(commits []CommitSizeStats) []CommitSizeStats {
	sorted := make([]CommitSizeStats, len(commits))
	copy(sorted, commits)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RiskScore > sorted[j].RiskScore
	})
	
	return sorted
}

// filterCommitsByRisk filters commits by minimum risk level.
func filterCommitsByRisk(commits []CommitSizeStats, minRisk string) []CommitSizeStats {
	riskLevels := map[string]int{
		"Low":      1,
		"Medium":   2,
		"High":     3,
		"Critical": 4,
	}
	
	minLevel := riskLevels[minRisk]
	var filtered []CommitSizeStats
	
	for _, commit := range commits {
		if riskLevels[commit.RiskLevel] >= minLevel {
			filtered = append(filtered, commit)
		}
	}
	
	return filtered
}

// truncateMessage truncates a commit message to a maximum length, handling multi-line messages.
func truncateMessage(message string, maxLen int) string {
	// First, get the first line only (before any newlines)
	firstLine := strings.Split(message, "\n")[0]
	
	if len(firstLine) <= maxLen {
		return firstLine
	}
	return firstLine[:maxLen-3] + "..."
}

// processCommitForSize processes a single commit to extract size statistics.
func processCommitForSize(c *object.Commit, pathArg string) (int, int, int, error) {
	var additions, deletions, filesChanged int
	
	if c.NumParents() == 0 {
		// Initial commit - count all files as additions
		tree, err := c.Tree()
		if err != nil {
			return 0, 0, 0, err
		}
		
		err = tree.Files().ForEach(func(f *object.File) error {
			if pathArg != "" && !strings.HasPrefix(f.Name, pathArg) {
				return nil
			}
			content, err := f.Contents()
			if err != nil {
				return nil
			}
			additions += countLines(content)
			filesChanged++
			return nil
		})
		return additions, deletions, filesChanged, err
	}
	
	// Regular commit - calculate diff
	parents := c.Parents()
	defer parents.Close()
	
	parentCount := 0
	err := parents.ForEach(func(parent *object.Commit) error {
		parentCount++
		patch, err := parent.Patch(c)
		if err != nil {
			log.Printf("failed to generate patch between parent %s and commit %s: %v", parent.Hash.String(), c.Hash.String(), err)
			return nil
		}
		
		fileSet := make(map[string]bool)
		for _, stat := range patch.Stats() {
			if pathArg != "" && !strings.HasPrefix(stat.Name, pathArg) {
				continue
			}
			
			additions += stat.Addition
			deletions += stat.Deletion
			fileSet[stat.Name] = true
		}
		
		// Count unique files changed
		for range fileSet {
			filesChanged++
		}
		
		return nil
	})
	
	// For merge commits (commits with multiple parents), the diff is calculated
	// against each parent, which can result in overcounting the number of files changed. To estimate the
	// number of unique files changed in the merge commit, we divide the total by the number of
	// parents. We use ceiling division to ensure that the result is not truncated to zero, which could happen
	// if the total is less than the number of parents.
	if parentCount > 1 {
		filesChanged = (filesChanged + parentCount - 1) / parentCount
	}
	
	return additions, deletions, filesChanged, err
}

// printCommitSizeStats prints commit size statistics in a formatted table.
func printCommitSizeStats(commits []CommitSizeStats, limit int) {
	fmt.Printf("\nTop %d commits by risk:\n", limit)
	fmt.Printf("%-12s %-50s %-20s %8s %8s %6s %6s %s\n", 
		"Hash", "Message", "Author", "Added", "Deleted", "Files", "Score", "Risk")
	fmt.Printf("%-12s %-50s %-20s %8s %8s %6s %6s %s\n", 
		strings.Repeat("-", 12), strings.Repeat("-", 50), strings.Repeat("-", 20), 
		"------", "-------", "-----", "-----", "----")
	
	for i, commit := range commits {
		if i >= limit {
			break
		}
		
		truncatedMessage := truncateMessage(commit.Message, 50)
		truncatedAuthor := truncateMessage(commit.Author, 20)
		
		fmt.Printf("%-12s %-50s %-20s %8d %8d %6d %6d %s\n",
			commit.Hash[:12], truncatedMessage, truncatedAuthor,
			commit.Additions, commit.Deletions, commit.FilesChanged, commit.RiskScore, commit.RiskLevel)
	}
}

// printRiskSummary prints a summary of risk distribution.
func printRiskSummary(commits []CommitSizeStats) {
	riskCounts := make(map[string]int)
	for _, commit := range commits {
		riskCounts[commit.RiskLevel]++
	}
	
	fmt.Printf("\nRisk Distribution:\n")
	fmt.Printf("  Low:      %d commits\n", riskCounts["Low"])
	fmt.Printf("  Medium:   %d commits\n", riskCounts["Medium"])
	fmt.Printf("  High:     %d commits\n", riskCounts["High"])
	fmt.Printf("  Critical: %d commits\n", riskCounts["Critical"])
}

// commitSizeCmd represents the commit-size command
var commitSizeCmd = &cobra.Command{
	Use:   "commit-size",
	Short: "Analyze commit sizes and identify risky commits",
	Long: `Analyze commit sizes to identify potentially risky commits that are hard to review,
debug, or rollback. Large commits reduce review effectiveness and increase risk.

Risk Levels:
- Low: ≤100 lines, ≤5 files
- Medium: 100-400 lines, 5-10 files  
- High: 400-800 lines, 10-15 files
- Critical: >800 lines, >15 files

Thresholds are based on research showing reviews are most effective under 400 lines.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		lastArg, _ := cmd.Flags().GetString("last")
		pathArg, _ := cmd.Flags().GetString("path")
		limitArg, _ := cmd.Flags().GetInt("limit")
		minRiskArg, _ := cmd.Flags().GetString("min-risk")
		summaryArg, _ := cmd.Flags().GetBool("summary")

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

		// Iterate through commits to collect size data
		ref, err := repo.Head()
		if err != nil {
			log.Fatalf("Could not get HEAD: %v", err)
		}

		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			log.Fatalf("Could not get commits: %v", err)
		}

		var commits []CommitSizeStats
		err = cIter.ForEach(func(c *object.Commit) error {
			if !since.IsZero() && c.Committer.When.Before(since) {
				return storer.ErrStop
			}
			
			additions, deletions, filesChanged, err := processCommitForSize(c, pathArg)
			if err != nil {
				log.Printf("Error processing commit %s: %v", c.Hash.String(), err)
				return nil
			}
			
			riskLevel, riskScore := calculateCommitRisk(additions, deletions, filesChanged)
			
			commit := CommitSizeStats{
				Hash:         c.Hash.String(),
				Message:      strings.TrimSpace(c.Message),
				Author:       c.Author.Name,
				Date:         c.Committer.When,
				Additions:    additions,
				Deletions:    deletions,
				FilesChanged: filesChanged,
				RiskScore:    riskScore,
				RiskLevel:    riskLevel,
			}
			
			commits = append(commits, commit)
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking commits: %v", err)
		}

		// Filter by minimum risk level if specified
		if minRiskArg != "" {
			commits = filterCommitsByRisk(commits, minRiskArg)
		}

		// Sort by risk score
		commits = sortCommitsByRisk(commits)

		// Print results
		fmt.Printf("Commit Size Analysis\n")
		fmt.Printf("Time window: %s\n", func() string {
			if since.IsZero() {
				return "all time"
			}
			return fmt.Sprintf("since %s", since.Format("2006-01-02"))
		}())
		if pathArg != "" {
			fmt.Printf("Path filter: %s\n", pathArg)
		}
		if minRiskArg != "" {
			fmt.Printf("Min risk level: %s\n", minRiskArg)
		}
		fmt.Printf("Total commits analyzed: %d\n", len(commits))
		fmt.Println("Context:", commitSizeBenchmarkContext)

		if summaryArg {
			printRiskSummary(commits)
		}

		if len(commits) > 0 {
			printCommitSizeStats(commits, limitArg)
		} else {
			fmt.Println("\nNo commits found matching the criteria.")
		}
	},
}

func init() {
	commitSizeCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	commitSizeCmd.Flags().String("path", "", "Limit analysis to a specific path")
	commitSizeCmd.Flags().Int("limit", 10, "Number of top results to show")
	commitSizeCmd.Flags().String("min-risk", "", "Minimum risk level to show (Low, Medium, High, Critical)")
	commitSizeCmd.Flags().Bool("summary", false, "Show risk distribution summary")
	rootCmd.AddCommand(commitSizeCmd)
}
