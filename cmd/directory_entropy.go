package cmd

import (
	"fmt"
	"log"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

const directoryEntropyContext = "High entropy signals weak modularity and eroded boundaries. Clean directories have focused purpose."

// DirectoryEntropyStats represents entropy statistics for a directory
type DirectoryEntropyStats struct {
	Path           string
	FileCount      int
	FileTypes      map[string]int
	Entropy        float64
	EntropyLevel   string
	Recommendation string
}

// DirectoryEntropyAnalysis represents the overall analysis
type DirectoryEntropyAnalysis struct {
	TimeWindow     string
	ProjectType    ProjectType
	TotalDirs      int
	AvgEntropy     float64
	HighEntropyDirs []DirectoryEntropyStats
	LowEntropyDirs  []DirectoryEntropyStats
}

// calculateEntropy calculates Shannon entropy for file type distribution
func calculateEntropy(fileTypes map[string]int) float64 {
	totalFiles := 0
	for _, count := range fileTypes {
		totalFiles += count
	}
	
	if totalFiles == 0 {
		return 0.0
	}
	
	entropy := 0.0
	for _, count := range fileTypes {
		if count > 0 {
			probability := float64(count) / float64(totalFiles)
			entropy -= probability * math.Log2(probability)
		}
	}
	
	return entropy
}

// ProjectType represents different types of software projects
type ProjectType struct {
	Name        string
	RootPatterns []string
	ExpectedDirs map[string][]string
	Description string
}

// Detect project type based on file patterns and structure
func detectProjectType(tree *object.Tree) ProjectType {
	var fileExtensions []string
	var fileNames []string
	var dirNames []string
	
	tree.Files().ForEach(func(f *object.File) error {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext != "" {
			fileExtensions = append(fileExtensions, ext)
		}
		
		fileName := strings.ToLower(filepath.Base(f.Name))
		fileNames = append(fileNames, fileName)
		
		dir := filepath.Dir(f.Name)
		if dir != "." && dir != "" {
			dirNames = append(dirNames, dir)
		}
		return nil
	})
	
	// Detect Go project - check for go.mod file specifically
	if contains(fileExtensions, ".go") && contains(fileNames, "go.mod") {
		return ProjectType{
			Name: "Go CLI/Application",
			RootPatterns: []string{".go", ".mod", ".sum", ".md", ".txt", ".yml", ".yaml"},
			ExpectedDirs: map[string][]string{
				"cmd":     {".go"},
				"internal": {".go"},
				"pkg":     {".go"},
				"docs":    {".md", ".rst"},
				"scripts": {".sh", ".py", ".bat"},
				"configs": {".yml", ".yaml", ".json", ".toml"},
			},
			Description: "Go project with standard layout",
		}
	}
	
	// Detect Node.js project - check for package.json specifically
	if contains(fileExtensions, ".js") && contains(fileNames, "package.json") {
		return ProjectType{
			Name: "Node.js Application",
			RootPatterns: []string{".js", ".json", ".md", ".txt", ".yml", ".yaml"},
			ExpectedDirs: map[string][]string{
				"src":     {".js", ".ts", ".jsx", ".tsx"},
				"lib":     {".js", ".ts"},
				"test":    {".js", ".ts", ".spec.js", ".test.js"},
				"docs":    {".md", ".rst"},
				"scripts": {".js", ".sh"},
				"config":  {".js", ".json", ".yml"},
			},
			Description: "Node.js project with standard layout",
		}
	}
	
	// Detect Python project
	if contains(fileExtensions, ".py") {
		return ProjectType{
			Name: "Python Application",
			RootPatterns: []string{".py", ".txt", ".md", ".yml", ".yaml", ".cfg", ".ini"},
			ExpectedDirs: map[string][]string{
				"src":     {".py"},
				"tests":   {".py"},
				"docs":    {".md", ".rst"},
				"scripts": {".py", ".sh"},
				"config":  {".py", ".yml", ".yaml", ".cfg"},
			},
			Description: "Python project with standard layout",
		}
	}
	
	// Detect Ruby/Rails project
	if contains(fileExtensions, ".rb") {
		return ProjectType{
			Name: "Ruby/Rails Application",
			RootPatterns: []string{".rb", ".gemspec", ".md", ".txt", ".yml", ".yaml"},
			ExpectedDirs: map[string][]string{
				"app":     {".rb", ".erb", ".haml"},
				"lib":     {".rb"},
				"spec":    {".rb"},
				"test":    {".rb"},
				"config":  {".rb", ".yml", ".yaml"},
				"docs":    {".md", ".rst"},
			},
			Description: "Ruby/Rails project with standard layout",
		}
	}
	
	// Default generic project
	return ProjectType{
		Name: "Generic Project",
		RootPatterns: []string{".md", ".txt", ".yml", ".yaml", ".json"},
		ExpectedDirs: map[string][]string{
			"src":     {},
			"docs":    {".md", ".rst"},
			"scripts": {},
			"config":  {},
		},
		Description: "Generic project structure",
	}
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isExpectedFileType checks if a file type is expected in a directory for the project type
func isExpectedFileType(projectType ProjectType, dirPath string, fileExt string) bool {
	// Root directory has different rules
	if dirPath == "root" || dirPath == "." {
		return contains(projectType.RootPatterns, fileExt)
	}
	
	// Check if directory has expected patterns
	if expectedExts, exists := projectType.ExpectedDirs[dirPath]; exists {
		if len(expectedExts) == 0 {
			return true // Directory allows any file type
		}
		return contains(expectedExts, fileExt)
	}
	
	// Unknown directory - be permissive
	return true
}

// classifyEntropyLevelWithContext provides context-aware entropy classification
func classifyEntropyLevelWithContext(entropy float64, avgEntropy float64, dirPath string, projectType ProjectType) (string, string) {
	isRoot := dirPath == "root" || dirPath == "."
	
	if isRoot {
		// Root directory has different rules based on project type
		switch {
		case entropy >= 3.0:
			return "High", "Consider organizing: too many file types in root"
		case entropy >= 2.0:
			return "Medium", "Acceptable: root directory with mixed concerns"
		default:
			return "Low", "Good: well-organized root directory"
		}
	}
	
	// Subdirectories follow strict rules
	switch {
	case entropy >= 2.0:
		return "Critical", "Urgent refactoring needed: severe boundary violations"
	case entropy >= 1.5:
		return "High", "Consider refactoring: mixed concerns detected"
	case entropy >= 0.8:
		return "Medium", "Monitor: some boundary erosion"
	default:
		return "Low", "Good: clear modular boundaries"
	}
}

// analyzeDirectoryEntropy analyzes entropy across repository directories
func analyzeDirectoryEntropy(repo *git.Repository, since time.Time) (*DirectoryEntropyAnalysis, error) {
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
	
	// Detect project type for context-aware analysis
	projectType := detectProjectType(tree)
	
	// Collect directory statistics
	dirStats := make(map[string]*DirectoryEntropyStats)
	
	err = tree.Files().ForEach(func(f *object.File) error {
		// Skip binary files
		isBinary, err := f.IsBinary()
		if err != nil || isBinary {
			return nil
		}
		
		// Get directory path
		dir := filepath.Dir(f.Name)
		if dir == "." {
			dir = "root"
		}
		
		// Get file extension
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == "" {
			ext = "no-extension"
		}
		
		// Initialize directory stats if needed
		if dirStats[dir] == nil {
			dirStats[dir] = &DirectoryEntropyStats{
				Path:      dir,
				FileTypes: make(map[string]int),
			}
		}
		
		// Update statistics
		dirStats[dir].FileCount++
		dirStats[dir].FileTypes[ext]++
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing files: %v", err)
	}
	
	// Calculate entropy for each directory
	var allEntropies []float64
	for _, stats := range dirStats {
		stats.Entropy = calculateEntropy(stats.FileTypes)
		allEntropies = append(allEntropies, stats.Entropy)
	}
	
	// Calculate average entropy
	avgEntropy := 0.0
	if len(allEntropies) > 0 {
		sum := 0.0
		for _, entropy := range allEntropies {
			sum += entropy
		}
		avgEntropy = sum / float64(len(allEntropies))
	}
	
	// Classify entropy levels with context awareness
	for _, stats := range dirStats {
		level, recommendation := classifyEntropyLevelWithContext(stats.Entropy, avgEntropy, stats.Path, projectType)
		stats.EntropyLevel = level
		stats.Recommendation = recommendation
	}
	
	// Sort directories by entropy
	var dirs []DirectoryEntropyStats
	for _, stats := range dirStats {
		dirs = append(dirs, *stats)
	}
	
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Entropy > dirs[j].Entropy
	})
	
	// Separate high and low entropy directories
	var highEntropyDirs, lowEntropyDirs []DirectoryEntropyStats
	for _, dir := range dirs {
		if dir.EntropyLevel == "Critical" || dir.EntropyLevel == "High" {
			highEntropyDirs = append(highEntropyDirs, dir)
		} else if dir.EntropyLevel == "Low" {
			lowEntropyDirs = append(lowEntropyDirs, dir)
		}
		// Medium entropy directories are not shown in either category
	}
	
	timeWindow := "all time"
	if !since.IsZero() {
		timeWindow = fmt.Sprintf("since %s", since.Format("2006-01-02"))
	}
	
	return &DirectoryEntropyAnalysis{
		TimeWindow:     timeWindow,
		ProjectType:    projectType,
		TotalDirs:      len(dirStats),
		AvgEntropy:     avgEntropy,
		HighEntropyDirs: highEntropyDirs,
		LowEntropyDirs:  lowEntropyDirs,
	}, nil
}

// printDirectoryEntropyStats prints directory entropy analysis
func printDirectoryEntropyStats(analysis *DirectoryEntropyAnalysis) {
	fmt.Printf("Directory Entropy Analysis\n")
	fmt.Printf("Time window: %s\n", analysis.TimeWindow)
	fmt.Printf("Project type: %s (%s)\n", analysis.ProjectType.Name, analysis.ProjectType.Description)
	fmt.Printf("Total directories analyzed: %d\n", analysis.TotalDirs)
	fmt.Printf("Average entropy: %.3f\n", analysis.AvgEntropy)
	fmt.Println()
	fmt.Println("Context:", directoryEntropyContext)
	fmt.Println()
	
	if len(analysis.HighEntropyDirs) > 0 {
		fmt.Printf("⚠️  High Entropy Directories (Need Attention):\n")
		fmt.Printf("Directory                    Files Types Entropy Level Recommendation\n")
		fmt.Printf("---------------------------- ----- ----- ---------- ----------------\n")
		for _, dir := range analysis.HighEntropyDirs {
			fmt.Printf("%-28s %5d %5d %10.3f %s\n", 
				dir.Path, dir.FileCount, len(dir.FileTypes), dir.Entropy, dir.Recommendation)
		}
		fmt.Println()
	}
	
	if len(analysis.LowEntropyDirs) > 0 {
		fmt.Printf("✅ Low Entropy Directories (Well Organized):\n")
		fmt.Printf("Directory                    Files Types Entropy Level Recommendation\n")
		fmt.Printf("---------------------------- ----- ----- ---------- ----------------\n")
		for _, dir := range analysis.LowEntropyDirs {
			fmt.Printf("%-28s %5d %5d %10.3f %s\n", 
				dir.Path, dir.FileCount, len(dir.FileTypes), dir.Entropy, dir.Recommendation)
		}
		fmt.Println()
	}
}

// directoryEntropyCmd represents the directory-entropy command
var directoryEntropyCmd = &cobra.Command{
	Use:   "directory-entropy",
	Short: "Analyze directory structure entropy",
	Long: `Analyze entropy across repository directories to identify areas with 
weak modularity and eroded boundaries. High entropy signals mixed concerns 
and unclear architectural boundaries.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags
		lastArg, _ := cmd.Flags().GetString("last")
		limitArg, _ := cmd.Flags().GetInt("limit")
		
		// Parse --last argument
		var since time.Time
		var err error
		if lastArg != "" {
			since, err = parseDurationArg(lastArg)
			if err != nil {
				log.Fatalf("Invalid --last value: %v", err)
			}
		}
		
		repo, err := git.PlainOpen(".")
		if err != nil {
			log.Fatalf("Failed to open git repo: %v", err)
		}
		
		analysis, err := analyzeDirectoryEntropy(repo, since)
		if err != nil {
			log.Fatalf("Failed to analyze directory entropy: %v", err)
		}
		
		// Apply limit if specified
		if limitArg > 0 {
			if len(analysis.HighEntropyDirs) > limitArg {
				analysis.HighEntropyDirs = analysis.HighEntropyDirs[:limitArg]
			}
			if len(analysis.LowEntropyDirs) > limitArg {
				analysis.LowEntropyDirs = analysis.LowEntropyDirs[:limitArg]
			}
		}
		
		printDirectoryEntropyStats(analysis)
	},
}

func init() {
	directoryEntropyCmd.Flags().String("last", "", "Limit analysis to a timeframe (e.g. 7d, 2m, 1y)")
	directoryEntropyCmd.Flags().Int("limit", 10, "Number of top results to show (default 10)")
	rootCmd.AddCommand(directoryEntropyCmd)
}
