package cmd

import (
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

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

		var additions, deletions int
		err = cIter.ForEach(func(c *object.Commit) error {
			if c.NumParents() == 0 {
				return nil
			}

			parent, err := c.Parents().Next()
			if err != nil {
				return nil
			}

			patch, err := parent.Patch(c)
			if err != nil {
				return nil
			}

			for _, stat := range patch.Stats() {
				additions += stat.Addition
				deletions += stat.Deletion
			}

			return nil
		})
		if err != nil {
			log.Fatalf("Error walking commits: %v", err)
		}

		ratio := float64(additions) / float64(deletions+1) // +1 to avoid div by zero
		fmt.Printf("Additions vs Deletions (all history):\n")
		fmt.Printf("- Additions: %d lines\n", additions)
		fmt.Printf("- Deletions: %d lines\n", deletions)
		fmt.Printf("- Ratio: %.2f : 1\n", ratio)
	},
}

func init() {
	rootCmd.AddCommand(churnCmd)
}