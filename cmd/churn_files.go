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
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
)

const (
	churnFilesHealthyThreshold = 5
	churnFilesCautionThreshold  = 20
)

const churnFilesBenchmarkContext = "Files with >20% churn often indicate architectural instability or frequent refactoring needs."

// FileChurnStats represents churn statistics for a single file.
type FileChurnStats struct {
	Path         string
	Additions    int
	Deletions    int
	TotalLOC     int
	ChurnPercent float64
	Status       string
}

// DirectoryChurnStats represents aggregated churn statistics for a directory.
type DirectoryChurnStats struct {
	Path         string
	Additions    int
	Deletions    int
	TotalLOC     int
	ChurnPercent float64
	Status       string
	FileCount    int
}

// calculateFileChurn calculates churn percentage and determines status for a file.
func calculateFileChurn(additions, deletions, totalLOC int) (float64, string) {
	if totalLOC == 0 {
		return 0.0, "Healthy"
	}
	
	churnPercent := float64(additions+deletions) / float64(totalLOC) * 100
	
	var status string
	if churnPercent <= float64(churnFilesHealthyThreshold) {
		status = "Healthy"
	} else if churnPercent <= float64(churnFilesCautionThreshold) {
		status = "Caution"
	} else {
		status = "Warning"
	}
	
	return churnPercent, status
}

// sortFilesByChurn sorts files by churn percentage in descending order.
func sortFilesByChurn(files []FileChurnStats) []FileChurnStats {
	sorted := make([]FileChurnStats, len(files))
	copy(sorted, files)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ChurnPercent > sorted[j].ChurnPercent
	})
	
	return sorted
}

// aggregateDirectoryChurn aggregates file-level churn into directory-level statistics.
func aggregateDirectoryChurn(files []FileChurnStats) []DirectoryChurnStats {
	dirMap := make(map[string]*DirectoryChurnStats)
	
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if dir == "." {
			dir = "root/"
		} else {
			dir = dir + "/"
		}
		
		if dirMap[dir] == nil {
			dirMap[dir] = &DirectoryChurnStats{
				Path: dir,
			}
		}
		
		dirMap[dir].Additions += file.Additions
		dirMap[dir].Deletions += file.Deletions
		dirMap[dir].TotalLOC += file.TotalLOC
		dirMap[dir].FileCount++
	}
	
	// Convert map to slice and calculate percentages
	var dirs []DirectoryChurnStats
	for _, dir := range dirMap {
		churnPercent, status := calculateFileChurn(dir.Additions, dir.Deletions, dir.TotalLOC)
		dir.ChurnPercent = churnPercent
		dir.Status = status
		dirs = append(dirs, *dir)
	}
	
	// Sort by churn percentage descending
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].ChurnPercent > dirs[j].ChurnPercent
	})
	
	return dirs
}

// processCommitForFileChurn processes a single commit to extract file-level churn data.
func processCommitForFileChurn(c *object.Commit, pathArg string) (map[string]FileChurnStats, error) {
	fileStats := make(map[string]FileChurnStats)
	
	if c.NumParents() == 0 {
		return fileStats, nil
	}
	
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
		
		for _, stat := range patch.Stats() {
			if pathArg != "" && !strings.HasPrefix(stat.Name, pathArg) {
				continue
			}
			
			if existing, exists := fileStats[stat.Name]; exists {
				existing.Additions += stat.Addition
				existing.Deletions += stat.Deletion
				fileStats[stat.Name] = existing
			} else {
				fileStats[stat.Name] = FileChurnStats{
					Path:      stat.Name,
					Additions: stat.Addition,
					Deletions: stat.Deletion,
				}
			}
		}
		return nil
	})
	
	// For merge commits, we need to avoid double-counting line changes
	// that appear in multiple parent diffs. Since we're already tracking unique files,
	// we only need to adjust line counts to estimate actual changes in the merge.
	if parentCount > 1 {
		for path := range fileStats {
			stats := fileStats[path]
			stats.Additions = (stats.Additions + parentCount - 1) / parentCount
			stats.Deletions = (stats.Deletions + parentCount - 1) / parentCount
			fileStats[path] = stats
		}
	}
	
	return fileStats, err
}

// getCurrentFileSizes gets the current size (LOC) of all files in the repository.
func getCurrentFileSizes(repo *git.Repository, pathArg string) (map[string]int, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD: %v", err)
	}
	
	headCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD commit: %v", err)
	}
	
	tree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD tree: %v", err)
	}
	
	fileSizes := make(map[string]int)
	err = tree.Files().ForEach(func(f *object.File) error {
		// Skip binary files
		isBinary, err := f.IsBinary()
		if err != nil {
			return nil
		}
		if isBinary {
			return nil
		}
		
		// Optionally filter by path
		if pathArg != "" && !strings.HasPrefix(f.Name, pathArg) {
			return nil
		}
		
		content, err := f.Contents()
		if err != nil {
			return nil
		}
		
		fileSizes[f.Name] = countLines(content)
		return nil
	})
	
	return fileSizes, err
}

// printFileChurnStats prints file-level churn statistics.
func printFileChurnStats(files []FileChurnStats, limit int) {
	fmt.Printf("\nTop %d files by churn:\n", limit)
	fmt.Printf("%-50s %8s %8s %8s %8s %s\n", "File", "Added", "Deleted", "Total LOC", "Churn %", "Status")
	fmt.Printf("%-50s %8s %8s %8s %8s %s\n", strings.Repeat("-", 50), "------", "-------", "---------", "-------", "------")
	
	for i, file := range files {
		if i >= limit {
			break
		}
		fmt.Printf("%-50s %8d %8d %8d %7.1f%% %s\n", 
			file.Path, file.Additions, file.Deletions, file.TotalLOC, file.ChurnPercent, file.Status)
	}
}

// printDirectoryChurnStats prints directory-level churn statistics.
func printDirectoryChurnStats(dirs []DirectoryChurnStats, limit int) {
	fmt.Printf("\nTop %d directories by churn:\n", limit)
	fmt.Printf("%-30s %8s %8s %8s %8s %6s %s\n", "Directory", "Added", "Deleted", "Total LOC", "Churn %", "Files", "Status")
	fmt.Printf("%-30s %8s %8s %8s %8s %6s %s\n", strings.Repeat("-", 30), "------", "-------", "---------", "-------", "------", "------")
	
	for i, dir := range dirs {
		if i >= limit {
			break
		}
		fmt.Printf("%-30s %8d %8d %8d %7.1f%% %6d %s\n", 
			dir.Path, dir.Additions, dir.Deletions, dir.TotalLOC, dir.ChurnPercent, dir.FileCount, dir.Status)
	}
}

// churnFilesCmd represents the churn-files command
var churnFilesCmd = &cobra.Command{
	Use:   "churn-files",
	Short: "Identify high-churn files and directories",
	Long: `Analyze which files and directories have the highest churn rates.
Files with high churn are refactored (or hacked) repeatedly and deserve architectural attention.

Thresholds:
- Healthy: ≤5% churn
- Caution: 5-20% churn  
- Warning: >20% churn`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		lastArg, _ := cmd.Flags().GetString("last")
		pathArg, _ := cmd.Flags().GetString("path")
		limitArg, _ := cmd.Flags().GetInt("limit")
		showDirsArg, _ := cmd.Flags().GetBool("directories")

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

		// Get current file sizes
		fileSizes, err := getCurrentFileSizes(repo, pathArg)
		if err != nil {
			log.Fatalf("Could not get current file sizes: %v", err)
		}

		// Iterate through commits to collect churn data
		ref, err := repo.Head()
		if err != nil {
			log.Fatalf("Could not get HEAD: %v", err)
		}

		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			log.Fatalf("Could not get commits: %v", err)
		}

		allFileStats := make(map[string]FileChurnStats)
		err = cIter.ForEach(func(c *object.Commit) error {
			if !since.IsZero() && c.Committer.When.Before(since) {
				return storer.ErrStop
			}
			
			commitFileStats, err := processCommitForFileChurn(c, pathArg)
			if err != nil {
				log.Printf("Error processing commit %s: %v", c.Hash.String(), err)
				return nil
			}
			for path, stats := range commitFileStats {
				if existing, exists := allFileStats[path]; exists {
					existing.Additions += stats.Additions
					existing.Deletions += stats.Deletions
					allFileStats[path] = existing
				} else {
					allFileStats[path] = stats
				}
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking commits: %v", err)
		}

		// Calculate churn percentages and status for each file
		var files []FileChurnStats
		for path, stats := range allFileStats {
			totalLOC := fileSizes[path]
			churnPercent, status := calculateFileChurn(stats.Additions, stats.Deletions, totalLOC)
			
			stats.TotalLOC = totalLOC
			stats.ChurnPercent = churnPercent
			stats.Status = status
			files = append(files, stats)
		}

		// Sort files by churn percentage
		files = sortFilesByChurn(files)

		// Print results
		fmt.Printf("High-Churn Files & Directories Analysis\n")
		fmt.Printf("Time window: %s\n", func() string {
			if since.IsZero() {
				return "all time"
			}
			return fmt.Sprintf("since %s", since.Format("2006-01-02"))
		}())
		if pathArg != "" {
			fmt.Printf("Path filter: %s\n", pathArg)
		}
		fmt.Printf("Threshold: >%d%% churn flags instability\n", churnFilesCautionThreshold)
		fmt.Println("Context:", churnFilesBenchmarkContext)

		printFileChurnStats(files, limitArg)

		if showDirsArg {
			dirs := aggregateDirectoryChurn(files)
			printDirectoryChurnStats(dirs, limitArg)
		}
	},
}

func init() {
	churnFilesCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	churnFilesCmd.Flags().String("path", "", "Limit analysis to a specific path")
	churnFilesCmd.Flags().Int("limit", 10, "Number of top results to show")
	churnFilesCmd.Flags().Bool("directories", false, "Also show directory-level churn statistics")
	rootCmd.AddCommand(churnFilesCmd)
}
