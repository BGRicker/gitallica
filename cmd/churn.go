package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
)

const churnBenchmarkContext = "Healthy codebases typically maintain churn below ~15% (KPI Depot; Opsera benchmarks)."

// parseDurationArg parses a string like "7d", "2m", "1y" and returns a cutoff time.Time from now.
func parseDurationArg(arg string) (time.Time, error) {
	if len(arg) < 2 {
		return time.Time{}, fmt.Errorf("invalid duration argument: %s", arg)
	}
	unit := arg[len(arg)-1]
	numStr := arg[:len(arg)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid number in duration: %s", arg)
	}
	now := time.Now()
	switch unit {
	case 'd':
		return now.AddDate(0, 0, -num), nil
	case 'm':
		return now.AddDate(0, -num, 0), nil
	case 'y':
		return now.AddDate(-num, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid unit in duration: %c", unit)
	}
}

func processCommitDiffs(c *object.Commit, pathArg string) (int, int) {
	var additions, deletions int
	if c.NumParents() == 0 {
		return 0, 0
	}
	parents := c.Parents()
	for {
		parent, err := parents.Next()
		if err != nil {
			break
		}
		patch, err := parent.Patch(c)
		if err != nil {
			continue
		}
		for _, stat := range patch.Stats() {
			if pathArg != "" && !strings.HasPrefix(stat.Name, pathArg) {
				continue
			}
			additions += stat.Addition
			deletions += stat.Deletion
		}
	}
	return additions, deletions
}

func countLines(content string) int {
	lines := strings.Count(content, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lines++
	}
	return lines
}

// churnCmd represents the churn command
var churnCmd = &cobra.Command{
	Use:   "churn",
	Short: "Show additions vs deletions ratio in the repo",
	Long: `Analyze git history to show how much code was added vs deleted. 
This helps you understand whether your repo is growing sustainably 
or accumulating complexity.`,
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Could not open repository: %v", err)
		}

		ref, err := repo.Head()
		if err != nil {
			log.Fatalf("Could not get HEAD: %v", err)
		}

		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			log.Fatalf("Could not get commits: %v", err)
		}

		lastArg, _ := cmd.Flags().GetString("last")
		pathArg, _ := cmd.Flags().GetString("path")
		since := time.Time{}
		if lastArg != "" {
			cutoff, err := parseDurationArg(lastArg)
			if err != nil {
				log.Fatalf("Could not parse --last argument: %v", err)
			}
			since = cutoff
		}

		var additions, deletions int
		err = cIter.ForEach(func(c *object.Commit) error {
			if !since.IsZero() && c.Committer.When.Before(since) {
				return storer.ErrStop
			}
			a, d := processCommitDiffs(c, pathArg)
			additions += a
			deletions += d
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking commits: %v", err)
		}

		// Calculate total LOC from latest commit's tree
		headCommit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			log.Fatalf("Could not get HEAD commit: %v", err)
		}
		tree, err := headCommit.Tree()
		if err != nil {
			log.Fatalf("Could not get HEAD tree: %v", err)
		}
		var totalLOC int
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
			totalLOC += countLines(content)
			return nil
		})
		if err != nil {
			log.Fatalf("Could not count LOC: %v", err)
		}

		var churnPercent float64
		if totalLOC > 0 {
			churnPercent = float64(additions+deletions) / float64(totalLOC) * 100
		} else {
			churnPercent = 0
		}
		status := ""
		if churnPercent <= 5 {
			status = "Healthy (≤5%)"
		} else if churnPercent <= 15 {
			status = "Caution (5–15%)"
		} else {
			status = "Warning (>15%)"
		}

		fmt.Printf("Additions vs Deletions:\n")
		fmt.Printf("- Additions: %d lines\n", additions)
		fmt.Printf("- Deletions: %d lines\n", deletions)
		fmt.Printf("- Total LOC: %d lines\n", totalLOC)
		fmt.Printf("Churn = (Additions + Deletions) / Total LOC\n")
		fmt.Printf("Churn: %.2f%% — %s\n", churnPercent, status)
		fmt.Println("Context:", churnBenchmarkContext)
	},
}

func init() {
	rootCmd.AddCommand(churnCmd)
	churnCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	churnCmd.Flags().String("path", "", "Limit analysis to a specific path")
}