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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

// isEmptyLine checks if a string contains only whitespace.
func isEmptyLine(s string) bool {
	return strings.TrimSpace(s) == ""
}

// countLines counts the number of lines in a string, handling edge cases.
func countLines(content string) int {
	lines := strings.Count(content, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lines++
	}
	return lines
}

// previewContent truncates a string to a maximum length for debug logging.
func previewContent(s string) string {
	const contentPreviewLength = 40
	const previewSuffix = "..."
	if len(s) > contentPreviewLength {
		return s[:contentPreviewLength] + previewSuffix
	}
	return s
}

// applyMergeCommitAdjustment applies ceiling division to line counts for merge commits.
// This function implements a heuristic to estimate the actual impact of a merge commit
// by dividing accumulated line changes by the number of parents. The ceiling division
// ensures we don't truncate small values to zero, which would be misleading.
//
// Rationale for this approach:
// 1. Merge commits often contain changes that appear in multiple parent diffs
// 2. Simple summation would overcount the actual impact of the merge
// 3. Division by parent count provides a reasonable estimate of the merge's true impact
// 4. Ceiling division prevents underestimation for small changes
//
// This is a pragmatic heuristic rather than a perfect solution, as true merge analysis
// would require complex conflict resolution analysis that's beyond the scope of this tool.
func applyMergeCommitAdjustment(additions, deletions, parentCount int) (int, int) {
	if parentCount <= 1 {
		return additions, deletions
	}
	
	// Apply ceiling division to both additions and deletions
	adjustedAdditions := (additions + parentCount - 1) / parentCount
	adjustedDeletions := (deletions + parentCount - 1) / parentCount
	
	return adjustedAdditions, adjustedDeletions
}

// matchesPathFilter checks if a file path matches any of the given filters using proper path handling
func matchesPathFilter(filePath string, pathFilters []string) bool {
	if len(pathFilters) == 0 {
		return true
	}
	
	// Normalize paths for cross-platform compatibility
	// Convert backslashes to forward slashes first for Windows compatibility
	cleanFilePath := strings.ReplaceAll(filePath, "\\", "/")
	cleanFilePath = filepath.ToSlash(filepath.Clean(cleanFilePath))
	
	for _, pathFilter := range pathFilters {
		if pathFilter == "" {
			continue
		}
		
		cleanPathFilter := strings.ReplaceAll(pathFilter, "\\", "/")
		cleanPathFilter = filepath.ToSlash(filepath.Clean(cleanPathFilter))
		
		// Exact match
		if cleanFilePath == cleanPathFilter {
			return true
		}
		
		// Check if file is under the specified directory (with proper directory boundary)
		if strings.HasPrefix(cleanFilePath, cleanPathFilter+"/") {
			return true
		}
	}
	
	return false
}

// matchesSinglePathFilter provides backward compatibility for single path filtering
func matchesSinglePathFilter(filePath, pathFilter string) bool {
	if pathFilter == "" {
		return true
	}
	return matchesPathFilter(filePath, []string{pathFilter})
}

// commitAffectsPath checks if a commit affects any of the specified path filters
func commitAffectsPath(commit *object.Commit, pathFilters []string) (bool, error) {
	if len(pathFilters) == 0 {
		return true, nil
	}

	// For initial commits, check if any files match the path filters
	if commit.NumParents() == 0 {
		tree, err := commit.Tree()
		if err != nil {
			return false, err
		}
		
		found := false
		err = tree.Files().ForEach(func(f *object.File) error {
			if matchesPathFilter(f.Name, pathFilters) {
				found = true
				return storer.ErrStop // Stop iteration
			}
			return nil
		})
		return found, err
	}

	// For regular commits, check the diff against parent
	parent, err := commit.Parent(0)
	if err != nil {
		return false, err
	}

	patch, err := parent.Patch(commit)
	if err != nil {
		return false, err
	}

	for _, stat := range patch.Stats() {
		if matchesPathFilter(stat.Name, pathFilters) {
			return true, nil
		}
	}

	return false, nil
}

// mergeViperConfig merges configuration from source viper into target viper
func mergeViperConfig(source, target *viper.Viper) {
	if source == nil || target == nil {
		return
	}
	
	for _, key := range source.AllKeys() {
		target.Set(key, source.Get(key))
	}
}

// getConfigPaths returns configured paths from CLI flags or config file, with CLI taking precedence
func getConfigPaths(cmd *cobra.Command, configKey string) ([]string, string) {
	// First try to get from command line flags (highest priority)
	pathArgs, _ := cmd.Flags().GetStringSlice("path")
	if len(pathArgs) > 0 {
		return pathArgs, "(from CLI)"
	}
	
	// Fall back to config file
	if viper.IsSet(configKey) {
		if paths := viper.GetStringSlice(configKey); len(paths) > 0 {
			return paths, "(from config)"
		}
	}
	
	return []string{}, ""
}

// titleCase converts a string to title case (first letter of each word capitalized)
func titleCase(s string) string {
	if s == "" {
		return s
	}
	
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// expandTimeWindow converts abbreviated time windows to readable format
func expandTimeWindow(lastArg string) string {
	if lastArg == "" {
		return "all time"
	}
	
	// Handle common abbreviations
	switch lastArg {
	case "1d":
		return "last 1 day"
	case "2d":
		return "last 2 days"
	case "3d":
		return "last 3 days"
	case "7d":
		return "last 7 days"
	case "14d":
		return "last 14 days"
	case "30d":
		return "last 30 days"
	case "1m":
		return "last 1 month"
	case "2m":
		return "last 2 months"
	case "3m":
		return "last 3 months"
	case "6m":
		return "last 6 months"
	case "1y":
		return "last 1 year"
	case "2y":
		return "last 2 years"
	case "3y":
		return "last 3 years"
	default:
		// For any other format, just prepend "last " to make it readable
		return "last " + lastArg
	}
}

// printCommandScope prints the configuration scope for a command
func printCommandScope(cmd *cobra.Command, commandName string, lastArg string, pathFilters []string, source string) {
	// Print config file information if available
	if viper.ConfigFileUsed() != "" {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
	
	// Print command scope header
	fmt.Fprintf(os.Stderr, "=== %s Analysis Scope ===\n", titleCase(commandName))
	
	// Print time window with expanded format
	fmt.Fprintf(os.Stderr, "Time window: %s\n", expandTimeWindow(lastArg))
	
	// Print path filters with source
	if len(pathFilters) > 0 {
		fmt.Fprintf(os.Stderr, "Path filter: %s %s\n", strings.Join(pathFilters, ", "), source)
	} else {
		fmt.Fprintf(os.Stderr, "Path filter: all files\n")
	}
	
	fmt.Fprintf(os.Stderr, "\n")
}

