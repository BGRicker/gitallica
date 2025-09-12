/*
Copyright © 2025 Ben Ricker <ben@jumboturbo.com>

*/
package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	diff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/spf13/cobra"
)

// lineKeySeparator is used to join file paths and (a hash of) line content to create unique keys.
// The null byte (\x00) is chosen as a separator because it is unlikely to appear in file paths,
// minimizing the risk of accidental key collisions.
//
// We purposefully use a cryptographic hash of the line content in keys (see makeKey) rather than the
// raw line text. This avoids any edge cases where the separator could be present in content and
// drastically lowers the chance of collisions. If a collision ever did occur (i.e., two different
// (file path, line content) pairs produced the same key), survival tracking could count distinct
// lines as the same, leading to false positives/negatives in the statistics. Using SHA-256 makes
// such collisions *extremely* unlikely in practice.
const lineKeySeparator = "\x00"



// researchNote contains a short reference used in CLI output about healthy survival rates.
const researchNote = "Large-scale study of 3.3 billion code-line lifetimes shows median lifespan of ~2.4 years (Spinellis et al.)."

func makeKey(filename, line string) string {
	sum := sha256.Sum256([]byte(line))
	return filename + lineKeySeparator + hex.EncodeToString(sum[:])
}


func printSurvivalStats(totalAdded, survived int, percent float64) {
	fmt.Printf("Survival rate:\n")
	fmt.Printf("  Lines added:    %d\n", totalAdded)
	fmt.Printf("  Still present:  %d\n", survived)
	fmt.Printf("  Survival rate:  %.2f%%\n", percent)
	fmt.Println(researchNote)
	fmt.Println()
}

// survivalCmd represents the survival command
var survivalCmd = &cobra.Command{
	Use:   "survival",
	Short: "Analyze code survival rate",
	Long: `Check how many lines survive over time compared to how many were added. 
Helps spot unstable areas where code gets rewritten too frequently.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		lastArg, _ := cmd.Flags().GetString("last")
		pathFilters := getConfigPaths(cmd, "survival.paths")
		debugArg, _ := cmd.Flags().GetBool("debug")
		
		// Print configuration scope
		printCommandScope(cmd, "survival", lastArg, pathFilters)

		// Parse --last argument
		var cutoff time.Time
		var err error
		if lastArg != "" {
			cutoff, err = parseDurationArg(lastArg)
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

		// Map to track added lines: key = file + hash(line content), value = occurrence count
		added := make(map[string]int)

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
			if debugArg {
				log.Printf("[survival] Commit %s at %v, parents: %d", commit.Hash.String(), commitTime, commit.NumParents())
			}
			if !cutoff.IsZero() && commitTime.Before(cutoff) {
				if debugArg {
					log.Printf("[survival] Skipping commit %s: before cutoff", commit.Hash.String())
				}
				continue
			}
			if commit.NumParents() > 1 {
				if debugArg {
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
				if debugArg {
					log.Printf("[survival] Generating patch for commit %s vs parent %s", commit.Hash.String(), parent.Hash.String())
				}
				patch, err = parent.Patch(commit)
				if err != nil {
					continue
				}
			} else {
				// Initial commit, diff with empty tree
				if debugArg {
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
				if debugArg {
					chunks := fileStat.Chunks()
					log.Printf("[survival] Entering file patch for %s with %d chunks", filename, len(chunks))
					for i, chunk := range chunks {
						contentPreview := previewContent(chunk.Content())
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
				if !matchesPathFilter(filename, pathFilters) {
					continue
				}
				for _, chunk := range fileStat.Chunks() {
					if chunk.Type() == diff.Add {
						if debugArg {
							log.Printf("[survival] Addition chunk in file %s", filename)
						}
						lines := strings.Split(chunk.Content(), "\n")
						for _, l := range lines {
							if isEmptyLine(l) {
								continue
							}
							key := makeKey(filename, l)
							added[key]++
							if debugArg {
								log.Printf("[survival] Added line: %q", strings.TrimSpace(l))
							}
						}
					}
				}
			}
		}

		// Sum counts so duplicates are accounted for accurately
		totalAdded := 0
		for _, c := range added {
			totalAdded += c
		}
		if debugArg {
			log.Printf("[survival] Total added lines tracked (counted): %d", totalAdded)
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
			if !matchesPathFilter(f.Name, pathFilters) {
				return nil
			}
			// Only text files
			content, err := f.Contents()
			if err != nil {
				return nil
			}
			lines := strings.Split(content, "\n")
			for _, l := range lines {
				if isEmptyLine(l) {
					continue
				}
				key := makeKey(f.Name, l)
				if count, ok := added[key]; ok {
					survived++
					if count == 1 {
						delete(added, key)
					} else {
						added[key] = count - 1
					}
				}
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Failed to walk HEAD tree: %v", err)
		}

		// Sanity check: the number of surviving lines must never exceed the number of lines added.
		if survived > totalAdded {
			log.Fatalf("[survival][error] Surviving lines (%d) exceed lines added (%d). This indicates a counting bug or data inconsistency.", survived, totalAdded)
		}
		percent := float64(survived) / float64(totalAdded) * 100
		// A negative percentage would indicate a serious bug (should be mathematically impossible here).
		if percent < 0 {
			log.Fatalf("[survival][error] Negative survival rate (%.2f%%). This indicates a counting bug or data inconsistency.", percent)
		}

		if debugArg {
			log.Printf("[survival] Introduced: %d, Surviving: %d, Rate: %.2f%%",
				totalAdded, survived, percent)
		}

		printSurvivalStats(totalAdded, survived, percent)
	},
}

func init() {
	survivalCmd.Flags().String("last", "", "Time window to consider (e.g. 7d, 2m, 1y)")
	survivalCmd.Flags().StringSlice("path", []string{}, "Limit analysis to specific paths (can be specified multiple times)")
	survivalCmd.Flags().Bool("debug", false, "Enable debug logging for survival analysis")
	rootCmd.AddCommand(survivalCmd)
}