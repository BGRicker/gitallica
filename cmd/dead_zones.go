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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
)

const (
	// Dead zone threshold - teams should set context-specific thresholds
	deadZoneThresholdMonths = 12  // Default threshold, teams should customize
	deadZoneLowRiskThresholdMonths = 24  // Low risk threshold
	deadZoneHighRiskThresholdMonths = 36 // High risk threshold
)

const deadZonesBenchmarkContext = "Code age should guide architectural decisions; teams set context-specific thresholds (CodeScene)."

// DeadZoneFileStats represents statistics for a potentially stale file
type DeadZoneFileStats struct {
	Path         string
	LastModified time.Time
	AgeInMonths  int
	Size         int64
	RiskLevel    string
	Recommendation string
}

// DeadZoneAnalysis represents the overall dead zone analysis
type DeadZoneAnalysis struct {
	TimeWindow     string
	TotalFiles     int
	DeadZoneFiles  []DeadZoneFileStats
	ActiveFiles    int
	DeadZoneCount  int
	DeadZonePercent float64
}

// calculateFileAge calculates the age of a file in months
func calculateFileAge(lastModified, referenceTime time.Time) int {
	if lastModified.IsZero() || lastModified.After(referenceTime) {
		return 0
	}
	
	years := referenceTime.Year() - lastModified.Year()
	months := int(referenceTime.Month()) - int(lastModified.Month())
	
	totalMonths := years*12 + months
	
	// Adjust if we haven't reached the day of the month yet
	if referenceTime.Day() < lastModified.Day() {
		totalMonths--
	}
	
	if totalMonths < 0 {
		return 0
	}
	
	return totalMonths
}

// isDeadZone determines if a file qualifies as a dead zone
func isDeadZone(lastModified, referenceTime time.Time) bool {
	ageInMonths := calculateFileAge(lastModified, referenceTime)
	return ageInMonths >= deadZoneThresholdMonths
}

// classifyDeadZoneRisk classifies the risk level of a dead zone file
func classifyDeadZoneRisk(ageInMonths int) (string, string) {
	if ageInMonths < deadZoneThresholdMonths {
		return "Active", "regularly maintained"
	} else if ageInMonths < deadZoneLowRiskThresholdMonths {
		return "Low Risk", "Consider reviewing"
	} else if ageInMonths < deadZoneHighRiskThresholdMonths {
		return "Medium Risk", "Needs attention"
	} else {
		return "High Risk", "Refactor or remove"
	}
}

// fileChangeHandler defines a callback for when a file change is found
type fileChangeHandler func(c *object.Commit) error

// findFileInCommitHistory searches git history for a specific file and calls handler when found
func findFileInCommitHistory(repo *git.Repository, fileName string, handler fileChangeHandler) error {
	ref, err := repo.Head()
	if err != nil {
		return err
	}
	
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}
	defer cIter.Close()
	
	return cIter.ForEach(func(c *object.Commit) error {
		// Handle merge commits by diffing against their first parent.
		// This is a common strategy to identify changes introduced by a merge.
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
				if change.To.Name == fileName {
					return handler(c)
				}
			}
		} else {
			// Initial commit - check if file exists
			_, err = tree.File(fileName)
			if err == nil {
				return handler(c)
			}
		}
		
		return nil
	})
}

// getLastModifiedFromGitHistory gets the last modification time for a file from full git history
func getLastModifiedFromGitHistory(repo *git.Repository, fileName string) (time.Time, bool, error) {
	var lastCommitTime time.Time
	var found bool
	
	err := findFileInCommitHistory(repo, fileName, func(c *object.Commit) error {
		lastCommitTime = c.Committer.When
		found = true
		return storer.ErrStop
	})
	
	if err != nil && err != storer.ErrStop {
		return time.Time{}, false, err
	}
	
	return lastCommitTime, found, nil
}

// sortDeadZonesByAge sorts dead zone files by age (oldest first)
func sortDeadZonesByAge(files []DeadZoneFileStats) []DeadZoneFileStats {
	sorted := make([]DeadZoneFileStats, len(files))
	copy(sorted, files)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].AgeInMonths > sorted[j].AgeInMonths
	})
	
	return sorted
}

// analyzeDeadZones performs dead zone analysis on the repository
func analyzeDeadZones(repo *git.Repository, since time.Time, pathFilters []string) (*DeadZoneAnalysis, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD: %v", err)
	}
	
	// Track the last modification time for each file
	fileLastModified := make(map[string]time.Time)
	
	// Iterate through all commits to find last modification times
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not get commits: %v", err)
	}
	defer cIter.Close()
	
	err = cIter.ForEach(func(c *object.Commit) error {
		if !since.IsZero() && c.Committer.When.Before(since) {
			return storer.ErrStop
		}
		
		// Handle merge commits by diffing against their first parent.
		// This is a common strategy to identify changes introduced by a merge.
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
				
				// Update last modified time for this file (only if not already set, since we iterate newest to oldest)
				if _, exists := fileLastModified[change.To.Name]; !exists {
					fileLastModified[change.To.Name] = c.Committer.When
				}
			}
		} else {
			// Initial commit - all files are "modified" at this time
			err = tree.Files().ForEach(func(f *object.File) error {
				if !matchesPathFilter(f.Name, pathFilters) {
					return nil
				}
				
				if _, exists := fileLastModified[f.Name]; !exists {
					fileLastModified[f.Name] = c.Committer.When
				}
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
	
	// Get current file tree to check which files still exist
	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD commit: %v", err)
	}
	
	tree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD tree: %v", err)
	}
	
	// Build analysis for current files
	var deadZoneFiles []DeadZoneFileStats
	var activeFiles int
	var totalFiles int
	now := time.Now()
	
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
		
		totalFiles++
		
		// Get last modified time for this file
		lastModified, exists := fileLastModified[f.Name]
		if !exists {
			// File exists but wasn't modified in the analysis window - get last modification from full Git history
			var found bool
			lastModified, found, err = getLastModifiedFromGitHistory(repo, f.Name)
			if err != nil {
				log.Printf("Warning: failed to get git history for %s: %v", f.Name, err)
				// Use a reasonable fallback - 3 years ago to indicate old but not infinitely old
				lastModified = time.Now().AddDate(-3, 0, 0)
			} else if !found {
				log.Printf("Warning: no git history found for %s, treating as old file", f.Name)
				// Use a reasonable fallback - 3 years ago to indicate old but not infinitely old
				lastModified = time.Now().AddDate(-3, 0, 0)
			}
		}
		
		ageInMonths := calculateFileAge(lastModified, now)
		isDead := isDeadZone(lastModified, now)
		
		if isDead {
			// Get file size from blob metadata (best effort - not critical for dead zone analysis)
			var size int64
			if f.Blob.Hash.IsZero() {
				size = 0 // Empty file or no content
			} else {
				// Only attempt blob lookup for dead zone files to minimize performance impact
				if blob, err := repo.BlobObject(f.Blob.Hash); err == nil && blob != nil {
					size = blob.Size
				} else {
					// Size lookup failed - not critical, use 0 and continue
					size = 0
				}
			}
			
			riskLevel, recommendation := classifyDeadZoneRisk(ageInMonths)
			
			deadZoneFiles = append(deadZoneFiles, DeadZoneFileStats{
				Path:           f.Name,
				LastModified:   lastModified,
				AgeInMonths:    ageInMonths,
				Size:           size,
				RiskLevel:      riskLevel,
				Recommendation: recommendation,
			})
		} else {
			activeFiles++
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing current files: %v", err)
	}
	
	// Sort dead zones by age (oldest first)
	deadZoneFiles = sortDeadZonesByAge(deadZoneFiles)
	
	deadZoneCount := len(deadZoneFiles)
	var deadZonePercent float64
	if totalFiles > 0 {
		deadZonePercent = float64(deadZoneCount) / float64(totalFiles) * 100
	}
	
	timeWindow := "all time"
	if !since.IsZero() {
		timeWindow = fmt.Sprintf("since %s", since.Format("2006-01-02"))
	}
	
	return &DeadZoneAnalysis{
		TimeWindow:      timeWindow,
		TotalFiles:      totalFiles,
		DeadZoneFiles:   deadZoneFiles,
		ActiveFiles:     activeFiles,
		DeadZoneCount:   deadZoneCount,
		DeadZonePercent: deadZonePercent,
	}, nil
}

// formatFileSize formats file size in human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// printDeadZoneStats prints dead zone analysis results
func printDeadZoneStats(analysis *DeadZoneAnalysis, limit int) {
	fmt.Printf("Dead Zones Analysis\n")
	fmt.Printf("Time window: %s\n", analysis.TimeWindow)
	fmt.Printf("Total files analyzed: %d\n", analysis.TotalFiles)
	fmt.Printf("Active files: %d\n", analysis.ActiveFiles)
	fmt.Printf("Dead zone files: %d (%.1f%%)\n", analysis.DeadZoneCount, analysis.DeadZonePercent)
	fmt.Printf("Threshold: Files untouched for ≥%d months\n", deadZoneThresholdMonths)
	fmt.Println()
	fmt.Println("Context:", deadZonesBenchmarkContext)
	fmt.Println()
	
	if len(analysis.DeadZoneFiles) == 0 {
		fmt.Println("✅ No dead zones found! All files are actively maintained.")
		return
	}
	
	fmt.Printf("⚠️  Dead Zone Files (showing top %d):\n", limit)
	fmt.Printf("File                                  Age     Size      Risk Level    Recommendation\n")
	fmt.Printf("------------------------------------- ------- --------- ------------- -----------------------\n")
	
	for i, file := range analysis.DeadZoneFiles {
		if i >= limit {
			break
		}
		
		ageStr := fmt.Sprintf("%d months", file.AgeInMonths)
		sizeStr := formatFileSize(file.Size)
		
		fmt.Printf("%-37s %7s %9s %-13s %s\n",
			truncateFilePath(file.Path, 37), ageStr, sizeStr, file.RiskLevel, file.Recommendation)
	}
	
	if len(analysis.DeadZoneFiles) > limit {
		fmt.Printf("\n... and %d more dead zone files\n", len(analysis.DeadZoneFiles)-limit)
	}
}

// truncateFilePath truncates a file path to fit in a specified width
func truncateFilePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// deadZonesCmd represents the dead-zones command
var deadZonesCmd = &cobra.Command{
	Use:   "dead-zones",
	Short: "Identify files untouched for ≥12 months",
	Long: `Analyze git history to identify files that haven't been modified for 12+ months.
These "dead zones" often indicate technical debt, abandoned features, or code that
should be refactored or removed.

Risk Levels:
- Low Risk: 12-17 months (consider reviewing)
- Medium Risk: 18-23 months (needs attention)  
- High Risk: 24-35 months (refactor or remove)
- Critical: 36+ months (urgent: refactor or delete)

Based on Clean Code principles - untouched code becomes a liability over time.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		lastArg, _ := cmd.Flags().GetString("last")
		pathFilters := getConfigPaths(cmd, "dead-zones.paths")
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

		analysis, err := analyzeDeadZones(repo, since, pathFilters)
		if err != nil {
			log.Fatalf("Error analyzing dead zones: %v", err)
		}

		printDeadZoneStats(analysis, limitArg)
	},
}

func init() {
	deadZonesCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	deadZonesCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	deadZonesCmd.Flags().Int("limit", 10, "Number of top results to show")
	rootCmd.AddCommand(deadZonesCmd)
}
