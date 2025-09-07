/*
Copyright © 2025 Ben Ricker <ben@jumboturbo.com>

*/
package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	diff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/spf13/cobra"
)

var survivalLast string
var survivalPath string
var survivalDebug bool

func printSurvivalStats(totalAdded, survived int, percent float64) {
	fmt.Printf("Survival rate:\n")
	fmt.Printf("  Lines added:    %d\n", totalAdded)
	fmt.Printf("  Still present:  %d\n", survived)
	fmt.Printf("  Survival rate:  %.2f%%\n", percent)
	fmt.Println("Healthy survival rates are typically above ~50% after 12 months (MSR, CodeScene research).")
	fmt.Println()
}

// survivalCmd represents the survival command
var survivalCmd = &cobra.Command{
	Use:   "survival",
	Short: "Analyze code survival rate",
	Long: `Check how many lines survive over time compared to how many were added. 
Helps spot unstable areas where code gets rewritten too frequently.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse --last argument
		var cutoff time.Time
		var err error
		if survivalLast != "" {
			cutoff, err = parseDurationArg(survivalLast)
			if err != nil {
				log.Fatalf("Invalid --last value: %v", err)
			}
		}

		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Failed to open git repo: %v", err)
		}
		ref, err := repo.Head()
		if err != nil {
			log.Fatalf("Failed to get HEAD: %v", err)
		}
		headCommit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			log.Fatalf("Failed to get HEAD commit: %v", err)
		}

		// Map to track added lines: key = file+line content, value = struct{when, file}
		type addedLine struct {
			File string
			Line string
			Time time.Time
		}
		added := make(map[string]addedLine)

		// Iterate commits, collect all added lines after cutoff
		commitsIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			log.Fatalf("Failed to iterate commits: %v", err)
		}
		defer commitsIter.Close()

		for {
			commit, err := commitsIter.Next()
			if err != nil {
				break
			}
			commitTime := commit.Committer.When
			if survivalDebug {
				log.Printf("[survival] Commit %s at %v, parents: %d", commit.Hash.String(), commitTime, commit.NumParents())
			}
			if !cutoff.IsZero() && commitTime.Before(cutoff) {
				if survivalDebug {
					log.Printf("[survival] Skipping commit %s: before cutoff", commit.Hash.String())
				}
				continue
			}
			if commit.NumParents() > 1 {
				if survivalDebug {
					log.Printf("[survival] Skipping commit %s: merge commit", commit.Hash.String())
				}
				continue
			}
			var patch *object.Patch
			if commit.NumParents() == 1 {
				parent, err := commit.Parent(0)
				if err != nil {
					continue
				}
				if survivalDebug {
					log.Printf("[survival] Generating patch for commit %s vs parent %s", commit.Hash.String(), parent.Hash.String())
				}
				patch, err = parent.Patch(commit)
				if err != nil {
					continue
				}
			} else {
				// Initial commit, diff with empty tree
				if survivalDebug {
					log.Printf("[survival] Generating patch for initial commit %s", commit.Hash.String())
				}
				emptyTree := &object.Tree{}
				t, err := commit.Tree()
				if err != nil {
					continue
				}
				patch, err = emptyTree.Patch(t)
				if err != nil {
					continue
				}
			}
			for _, fileStat := range patch.FilePatches() {
				from, to := fileStat.Files()
				var filename string
				if to != nil {
					filename = to.Path()
				} else if from != nil {
					filename = from.Path()
				}
				if survivalDebug {
					chunks := fileStat.Chunks()
					log.Printf("[survival] Entering file patch for %s with %d chunks", filename, len(chunks))
					for i, chunk := range chunks {
						contentPreview := chunk.Content()
						if len(contentPreview) > 40 {
							contentPreview = contentPreview[:40] + "..."
						}
						var chunkType string
						switch chunk.Type() {
						case diff.Add:
							chunkType = "Add"
						case diff.Delete:
							chunkType = "Delete"
						case diff.Equal:
							chunkType = "Equal"
						default:
							chunkType = fmt.Sprintf("Unknown(%v)", chunk.Type())
						}
						log.Printf("[survival] Chunk %d: type %s, content preview: %q", i, chunkType, contentPreview)
					}
				}
				if survivalPath != "" && !strings.HasPrefix(filename, survivalPath) {
					continue
				}
				for _, chunk := range fileStat.Chunks() {
					if chunk.Type() == diff.Add {
						if survivalDebug {
							log.Printf("[survival] Addition chunk in file %s", filename)
						}
						lines := strings.Split(chunk.Content(), "\n")
						for _, l := range lines {
							if strings.TrimSpace(l) == "" {
								continue
							}
							key := filename + "\x00" + l
							added[key] = addedLine{
								File: filename,
								Line: l,
								Time: commitTime,
							}
							if survivalDebug {
								log.Printf("[survival] Added line: %q", strings.TrimSpace(l))
							}
						}
					}
				}
			}
		}

		totalAdded := len(added)
		if survivalDebug {
			log.Printf("[survival] Total added lines tracked: %d", totalAdded)
		}
		if totalAdded == 0 {
			fmt.Println("No lines added in the specified window.")
			return
		}

		// Now, walk HEAD tree, check which lines survived
		survived := 0
		headTree, err := headCommit.Tree()
		if err != nil {
			log.Fatalf("Failed to get HEAD tree: %v", err)
		}
		err = headTree.Files().ForEach(func(f *object.File) error {
			if survivalPath != "" && !strings.HasPrefix(f.Name, survivalPath) {
				return nil
			}
			// Only text files
			r, err := f.Reader()
			if err != nil {
				return nil
			}
			defer r.Close()
			content, err := f.Contents()
			if err != nil {
				return nil
			}
			lines := strings.Split(content, "\n")
			for _, l := range lines {
				if strings.TrimSpace(l) == "" {
					continue
				}
				key := f.Name + "\x00" + l
				if _, ok := added[key]; ok {
					survived++
				}
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Failed to walk HEAD tree: %v", err)
		}

		percent := float64(survived) / float64(totalAdded) * 100
		if percent > 100 {
			percent = 100
		}
		if percent < 0 {
			percent = 0
		}

		if survivalDebug {
			log.Printf("[survival] Introduced: %d, Deleted: %d, Surviving: %d, Rate: %.2f%%",
				totalAdded, 0, survived, percent)
		}

		printSurvivalStats(totalAdded, survived, percent)
	},
}

func init() {
	survivalCmd.Flags().StringVar(&survivalLast, "last", "", "Time window to consider (e.g. 7d, 2m, 1y)")
	survivalCmd.Flags().StringVar(&survivalPath, "path", "", "Restrict to file path prefix")
	survivalCmd.Flags().BoolVar(&survivalDebug, "debug", false, "Enable debug logging for survival analysis")
	rootCmd.AddCommand(survivalCmd)
}