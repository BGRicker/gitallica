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
	diff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/spf13/cobra"
)

const componentCreationContext = "Sudden spikes in component creation often indicate architectural sprawl or lack of design discipline."

// componentCreationSpikeThreshold defines the threshold for detecting spikes in component creation.
// Based on industry research, creating more than 10 components in a short period often indicates
// architectural sprawl or lack of design discipline (Kent Beck's simple design principle).
const componentCreationSpikeThreshold = 10

// ComponentType represents different types of components across frameworks
type ComponentType struct {
	Name        string
	Patterns    []*regexp.Regexp
	Extensions  []string
	Description string
}

// ComponentCreationStats tracks component creation statistics
type ComponentCreationStats struct {
	ComponentType string
	Count         int
	Files         []string
	FirstSeen     time.Time
	LastSeen      time.Time
}

// ComponentCreationRate tracks creation rate over time
type ComponentCreationRate struct {
	TimeWindow    string
	TotalCreated  int
	ByType        map[string]int
	SpikeDetected bool
	SpikeReason   string
}

// Define component patterns for different frameworks
var componentTypes = map[string]ComponentType{
	"javascript-class": {
		Name:        "JavaScript Class",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`class\s+\w+`)},
		Extensions:  []string{".js", ".jsx", ".ts", ".tsx"},
		Description: "ES6+ classes and TypeScript classes",
	},
	"react-component": {
		Name:        "React Component",
		Patterns:    []*regexp.Regexp{
			// Match class components: class Foo extends React.Component
			regexp.MustCompile(`\bclass\s+\w+\s+extends\s+(React\.)?Component\b`),
			// Match functions that return JSX: return <Something ... or
			regexp.MustCompile(`return\s+\(<\w+`),
			// Match functions that return JSX directly: return <Something
			regexp.MustCompile(`return\s+<\w+`),
		},
		Extensions:  []string{".js", ".jsx", ".ts", ".tsx"},
		Description: "React functional and class components",
	},
	"ruby-model": {
		Name:        "Ruby Model",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`\bclass\s+\w+\s*<\s*ApplicationRecord\b`)},
		Extensions:  []string{".rb"},
		Description: "Rails ActiveRecord models",
	},
	"ruby-controller": {
		Name:        "Ruby Controller",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`\bclass\s+\w+Controller\s*<\s*ApplicationController\b`)},
		Extensions:  []string{".rb"},
		Description: "Rails controllers",
	},
	"ruby-service": {
		Name:        "Ruby Service",
		Patterns:    []*regexp.Regexp{
			regexp.MustCompile(`class\s+\w+Service`),
			regexp.MustCompile(`class\s+\w+\s*<\s*Service`),
		},
		Extensions:  []string{".rb"},
		Description: "Ruby service objects",
	},
	"python-class": {
		Name:        "Python Class",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`class\s+\w+.*:`)},
		Extensions:  []string{".py"},
		Description: "Python classes",
	},
	"go-struct": {
		Name:        "Go Struct",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`type\s+\w+\s+struct`)},
		Extensions:  []string{".go"},
		Description: "Go structs",
	},
	"go-interface": {
		Name:        "Go Interface",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`type\s+\w+\s+interface`)},
		Extensions:  []string{".go"},
		Description: "Go interfaces",
	},
	"java-class": {
		Name:        "Java Class",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`public\s+class\s+\w+`)},
		Extensions:  []string{".java"},
		Description: "Java classes",
	},
	"csharp-class": {
		Name:        "C# Class",
		Patterns:    []*regexp.Regexp{regexp.MustCompile(`public\s+class\s+\w+`)},
		Extensions:  []string{".cs"},
		Description: "C# classes",
	},
}

// detectComponentsInFile analyzes a file to detect component definitions
func detectComponentsInFile(filePath string, content string) map[string]int {
	detected := make(map[string]int)
	ext := strings.ToLower(filepath.Ext(filePath))
	
	for typeKey, componentType := range componentTypes {
		// Check if file extension matches
		extensionMatch := false
		for _, allowedExt := range componentType.Extensions {
			if ext == allowedExt {
				extensionMatch = true
				break
			}
		}
		
		if !extensionMatch {
			continue
		}
		
		// Check patterns
		for _, pattern := range componentType.Patterns {
			matches := pattern.FindAllString(content, -1)
			if len(matches) > 0 {
				detected[typeKey] += len(matches)
			}
		}
	}
	
	return detected
}

// analyzeComponentCreation analyzes component creation patterns in the repository
func analyzeComponentCreation(repo *git.Repository, since time.Time, framework string) ([]ComponentCreationStats, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD: %v", err)
	}
	
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("could not get commits: %v", err)
	}
	defer cIter.Close()
	
	componentStats := make(map[string]*ComponentCreationStats)
	
	err = cIter.ForEach(func(c *object.Commit) error {
		if !since.IsZero() && c.Committer.When.Before(since) {
			return storer.ErrStop
		}
		
		// Get parent commit to compare changes
		var parentTree *object.Tree
		if len(c.ParentHashes) > 0 {
			parent, err := repo.CommitObject(c.ParentHashes[0])
			if err == nil {
				parentTree, _ = parent.Tree()
			}
		}
		
		// Get current tree
		tree, err := c.Tree()
		if err != nil {
			return nil
		}
		
		// Skip initial commit (no parent) to avoid false positives in component detection
		if parentTree == nil {
			return nil
		}
		
		// Analyze only added/modified files using diff
		changes, err := tree.Diff(parentTree)
		if err != nil {
			return nil
		}
		
		for _, change := range changes {
			if change.To.Name == "" {
				continue // skip deletions
			}
			
			file, err := tree.File(change.To.Name)
			if err != nil {
				continue
			}
			
			// Skip binary files
			isBinary, err := file.IsBinary()
			if err != nil || isBinary {
				continue
			}
			
			// Filter by framework if specified - check file extension
			if framework != "" {
				ext := strings.ToLower(filepath.Ext(file.Name))
				frameworkMatch := false
				
				// Check if file extension matches framework
				switch framework {
				case "javascript":
					frameworkMatch = ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx"
				case "ruby":
					frameworkMatch = ext == ".rb"
				case "python":
					frameworkMatch = ext == ".py"
				case "go":
					frameworkMatch = ext == ".go"
				case "java":
					frameworkMatch = ext == ".java"
				case "csharp":
					frameworkMatch = ext == ".cs"
				}
				
				if !frameworkMatch {
					continue
				}
			}
			
			// Get the diff patch to analyze only added lines
			patch, err := change.Patch()
			if err != nil {
				continue
			}
			
			// Extract only added lines from the patch
			var addedContent strings.Builder
			for _, filePatch := range patch.FilePatches() {
				for _, chunk := range filePatch.Chunks() {
					// Check chunk type before processing content for efficiency
					if chunk.Type() == diff.Add {
						addedContent.WriteString(chunk.Content())
					}
				}
			}
			
			// Only analyze if there are added lines
			if addedContent.Len() == 0 {
				continue
			}
			
			// Detect components in added content only
			detected := detectComponentsInFile(file.Name, addedContent.String())
			for componentType, count := range detected {
				if stats, exists := componentStats[componentType]; exists {
					stats.Count += count
					// Avoid duplicate file entries
					alreadyPresent := false
					for _, fname := range stats.Files {
						if fname == file.Name {
							alreadyPresent = true
							break
						}
					}
					if !alreadyPresent {
						stats.Files = append(stats.Files, file.Name)
					}
					if c.Committer.When.Before(stats.FirstSeen) {
						stats.FirstSeen = c.Committer.When
					}
					if c.Committer.When.After(stats.LastSeen) {
						stats.LastSeen = c.Committer.When
					}
				} else {
					componentStats[componentType] = &ComponentCreationStats{
						ComponentType: componentType,
						Count:         count,
						Files:         []string{file.Name},
						FirstSeen:     c.Committer.When,
						LastSeen:      c.Committer.When,
					}
				}
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing commits: %v", err)
	}
	
	// Convert to slice and sort by count
	var result []ComponentCreationStats
	for _, stats := range componentStats {
		result = append(result, *stats)
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	
	return result, nil
}

// calculateCreationRate calculates the component creation rate over time
func calculateCreationRate(stats []ComponentCreationStats, timeWindow string) ComponentCreationRate {
	totalCreated := 0
	byType := make(map[string]int)
	
	for _, stat := range stats {
		totalCreated += stat.Count
		byType[stat.ComponentType] = stat.Count
	}
	
	rate := ComponentCreationRate{
		TimeWindow:   timeWindow,
		TotalCreated: totalCreated,
		ByType:       byType,
	}
	
	// Simple spike detection: if more than threshold components created in recent period
	if totalCreated > componentCreationSpikeThreshold {
		rate.SpikeDetected = true
		rate.SpikeReason = fmt.Sprintf("High component creation rate: %d components (threshold: %d)", totalCreated, componentCreationSpikeThreshold)
	}
	
	return rate
}

// printComponentCreationStats prints component creation statistics
func printComponentCreationStats(stats []ComponentCreationStats, rate ComponentCreationRate, framework string) {
	fmt.Printf("New Component Creation Rate Analysis\n")
	fmt.Printf("Time window: %s\n", rate.TimeWindow)
	if framework != "" {
		fmt.Printf("Framework filter: %s\n", framework)
	}
	fmt.Printf("Total components created: %d\n", rate.TotalCreated)
	
	if rate.SpikeDetected {
		fmt.Printf("⚠️  %s\n", rate.SpikeReason)
	} else {
		fmt.Printf("✅ Healthy component creation rate\n")
	}
	
	fmt.Printf("\nContext: %s\n", componentCreationContext)
	fmt.Printf("\nTop components by creation count:\n")
	fmt.Printf("Component Type                    Count Files\n")
	fmt.Printf("---------------------------------- ----- -----\n")
	
	for _, stat := range stats {
		if len(stat.Files) > 0 {
			componentName := componentTypes[stat.ComponentType].Name
			fmt.Printf("%-32s %5d %5d\n", componentName, stat.Count, len(stat.Files))
		}
	}
}

// componentCreationCmd represents the component-creation command
var componentCreationCmd = &cobra.Command{
	Use:   "component-creation",
	Short: "Analyze new component creation rate",
	Long: `Track the rate of new component creation across different frameworks.
Helps identify architectural sprawl and design discipline issues.

Supports multiple frameworks:
- JavaScript/TypeScript: Classes, React components
- Ruby/Rails: Models, controllers, services
- Python: Classes and modules
- Go: Structs and interfaces
- Java: Classes and interfaces
- C#: Classes and interfaces`,
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Could not open repository: %v", err)
		}
		
		lastArg, _ := cmd.Flags().GetString("last")
		frameworkArg, _ := cmd.Flags().GetString("framework")
		limitArg, _ := cmd.Flags().GetInt("limit")
		
		since := time.Time{}
		if lastArg != "" {
			cutoff, err := parseDurationArg(lastArg)
			if err != nil {
				log.Fatalf("Could not parse --last argument: %v", err)
			}
			since = cutoff
		}
		
		stats, err := analyzeComponentCreation(repo, since, frameworkArg)
		if err != nil {
			log.Fatalf("Error analyzing component creation: %v", err)
		}
		
		// Limit results if specified
		if limitArg > 0 && len(stats) > limitArg {
			stats = stats[:limitArg]
		}
		
		timeWindow := "all time"
		if !since.IsZero() {
			timeWindow = fmt.Sprintf("last %s", lastArg)
		}
		
		rate := calculateCreationRate(stats, timeWindow)
		printComponentCreationStats(stats, rate, frameworkArg)
	},
}

func init() {
	rootCmd.AddCommand(componentCreationCmd)
	componentCreationCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	componentCreationCmd.Flags().String("framework", "", "Filter by framework (javascript, ruby, python, go, java, csharp)")
	componentCreationCmd.Flags().Int("limit", 10, "Number of top results to show")
}
