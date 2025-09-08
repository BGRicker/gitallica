package cmd

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

// Thresholds for sustainable development pace analysis
const (
	// Trend detection thresholds
	stableTrendThreshold = 0.1 // Slope within ±0.1 is considered stable
	
	// Spike/dip detection multipliers
	spikeThresholdMultiplier = 2.0 // 2x average is a spike
	dipThresholdMultiplier   = 0.3 // <30% of average is a dip
	
	// Sustainability thresholds
	healthyAvgCommitsLow    = 5.0  // Minimum for healthy pace
	healthyAvgCommitsHigh   = 25.0 // Maximum for healthy pace
	warningSpikesThreshold  = 2    // 2+ spikes indicates warning
	criticalSpikesThreshold = 3    // 3+ spikes indicates critical
)

// TimePeriod represents a time window with commit count
type TimePeriod struct {
	Start       time.Time
	End         time.Time
	CommitCount int
	Severity    string // For spikes/dips: "Low", "Medium", "High"
}

// CommitCadenceStats contains analysis results
type CommitCadenceStats struct {
	TotalCommits            int
	TotalPeriods            int
	AverageCommitsPerPeriod float64
	TrendDirection          string // "Increasing", "Decreasing", "Stable", "Unknown"
	TrendStrength           float64
	SustainabilityLevel     string // "Healthy", "Caution", "Warning", "Critical"
	Spikes                  []TimePeriod
	Dips                    []TimePeriod
	TimePeriods             []TimePeriod
}


var commitCadenceCmd = &cobra.Command{
	Use:   "commit-cadence",
	Short: "Analyze commit cadence trends to identify pace and sustainability patterns",
	Long: `Analyzes commit frequency patterns over time to identify trends, spikes, and 
sustainable development pace indicators.

Track trends, not absolutes - spikes or dips may reveal crunch, burnout, or stagnation.
This analysis helps teams maintain sustainable development practices and identify
potential process issues before they become critical.

Research basis:
- "Overtime is a symptom of a serious problem... you can't work a second week of overtime." — Kent Beck
- DORA metrics emphasize sustainable deployment frequency and team health
- Agile principles promote sustainable development pace

The analysis groups commits by time periods and identifies:
- Overall trend direction (increasing/decreasing/stable)
- Commit spikes that may indicate crunch periods
- Commit dips that may indicate stagnation or burnout
- Sustainability assessment based on pace and volatility`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.PlainOpen(".")
		if err != nil {
			return fmt.Errorf("could not open repository: %v", err)
		}

		pathArg, _ := cmd.Flags().GetString("path")
		lastArg, _ := cmd.Flags().GetString("last")
		periodArg, _ := cmd.Flags().GetString("period")

		stats, err := analyzeCommitCadence(repo, pathArg, lastArg, periodArg)
		if err != nil {
			return err
		}

		printCommitCadenceStats(stats, periodArg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCadenceCmd)
	commitCadenceCmd.Flags().String("last", "", "Specify the time window to analyze (e.g., 30d, 6m, 1y)")
	commitCadenceCmd.Flags().String("path", "", "Limit analysis to a specific directory or path")
	commitCadenceCmd.Flags().String("period", "week", "Time period for grouping (day, week, month)")
}

// analyzeCommitCadence performs the main cadence analysis
func analyzeCommitCadence(repo *git.Repository, pathArg string, lastArg string, periodArg string) (*CommitCadenceStats, error) {
	var since *time.Time
	if lastArg != "" {
		sinceTime, err := parseDurationArg(lastArg)
		if err != nil {
			return nil, fmt.Errorf("invalid time window: %v", err)
		}
		since = &sinceTime
	}
	
	// Get commits within time window
	commitIter, err := repo.Log(&git.LogOptions{
		Since: since,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get commit log: %v", err)
	}
	defer commitIter.Close()
	
	var commits []CommitInfo
	
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Skip commits without author information
		if commit.Author.Email == "" {
			return nil
		}
		
		// Skip merge commits for cleaner analysis
		if commit.NumParents() > 1 {
			return nil
		}
		
		// If path filtering is specified, check if commit affects the path
		if pathArg != "" {
			affectsPath, err := commitAffectsPath(commit, pathArg)
			if err != nil {
				return err
			}
			if !affectsPath {
				return nil
			}
		}
		
		commits = append(commits, CommitInfo{
			Hash:    commit.Hash.String()[:8],
			Time:    commit.Author.When,
			Author:  commit.Author.Email,
			Message: commit.Message,
			Files:   []string{}, // Not needed for cadence analysis
		})
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error analyzing commits: %v", err)
	}
	
	// Group commits by time periods
	timePeriods := groupCommitsByTimePeriod(commits, periodArg)
	
	// Calculate comprehensive statistics
	stats := calculateCommitCadenceStats(timePeriods)
	
	return stats, nil
}

// calculateISOWeekPeriod calculates the start, end, and key for an ISO week period
func calculateISOWeekPeriod(t time.Time) (start, end time.Time, key string) {
	utcTime := t.UTC()
	year, week := utcTime.ISOWeek()
	
	// Calculate the Monday of this ISO week
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	jan1Weekday := jan1.Weekday()
	if jan1Weekday == time.Sunday {
		jan1Weekday = 7 // Adjust Sunday to be day 7 instead of 0
	}
	daysToFirstMonday := 1 - int(jan1Weekday)
	if daysToFirstMonday > 0 {
		daysToFirstMonday -= 7
	}
	
	start = jan1.AddDate(0, 0, daysToFirstMonday+(week-1)*7)
	end = start.Add(7*24*time.Hour - time.Nanosecond)
	key = start.Format("2006-W02")
	
	return start, end, key
}

// groupCommitsByTimePeriod groups commits into time-based buckets with zero-fill for missing periods
func groupCommitsByTimePeriod(commits []CommitInfo, period string) []TimePeriod {
	if len(commits) == 0 {
		return []TimePeriod{}
	}
	
	// Sort commits by time
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.Before(commits[j].Time)
	})
	
	// First pass: collect commit counts by period
	periodMap := make(map[string]int)
	periodStarts := make(map[string]time.Time)
	periodEnds := make(map[string]time.Time)
	
	for _, commit := range commits {
		var periodKey string
		var periodStart, periodEnd time.Time
		
		switch period {
		case "day":
			year, month, day := commit.Time.Date()
			periodStart = time.Date(year, month, day, 0, 0, 0, 0, commit.Time.Location())
			periodEnd = periodStart.Add(24*time.Hour - time.Nanosecond)
			periodKey = periodStart.Format("2006-01-02")
		case "week":
			// Use helper function for ISO week calculation
			periodStart, periodEnd, periodKey = calculateISOWeekPeriod(commit.Time)
		case "month":
			year, month, _ := commit.Time.Date()
			periodStart = time.Date(year, month, 1, 0, 0, 0, 0, commit.Time.Location())
			periodEnd = periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
			periodKey = periodStart.Format("2006-01")
		default:
			// Default to week - use helper function for ISO week calculation  
			periodStart, periodEnd, periodKey = calculateISOWeekPeriod(commit.Time)
		}
		
		periodMap[periodKey]++
		if _, exists := periodStarts[periodKey]; !exists {
			periodStarts[periodKey] = periodStart
			periodEnds[periodKey] = periodEnd
		}
	}
	
	// Second pass: create continuous time series with zero-fill for missing periods
	if len(periodStarts) == 0 {
		return []TimePeriod{}
	}
	
	// Find the time range from earliest to latest commit
	earliestTime := commits[0].Time
	latestTime := commits[len(commits)-1].Time
	
	// Generate complete time series from earliest to latest period
	var periods []TimePeriod
	current := earliestTime
	
	for current.Before(latestTime) || current.Equal(latestTime) {
		var periodKey string
		var periodStart, periodEnd time.Time
		
		// Use same period calculation logic as in first pass
		switch period {
		case "day":
			year, month, day := current.Date()
			periodStart = time.Date(year, month, day, 0, 0, 0, 0, current.Location())
			periodEnd = periodStart.Add(24*time.Hour - time.Nanosecond)
			periodKey = periodStart.Format("2006-01-02")
		case "week":
			periodStart, periodEnd, periodKey = calculateISOWeekPeriod(current)
		case "month":
			year, month, _ := current.Date()
			periodStart = time.Date(year, month, 1, 0, 0, 0, 0, current.Location())
			periodEnd = periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
			periodKey = periodStart.Format("2006-01")
		default:
			// Default to week
			periodStart, periodEnd, periodKey = calculateISOWeekPeriod(current)
		}
		
		// Get commit count for this period (zero if no commits)
		commitCount := periodMap[periodKey]
		
		periods = append(periods, TimePeriod{
			Start:       periodStart,
			End:         periodEnd,
			CommitCount: commitCount,
		})
		
		// Move to next period
		switch period {
		case "day":
			current = periodStart.Add(24 * time.Hour)
		case "week":
			current = periodStart.Add(7 * 24 * time.Hour)
		case "month":
			current = periodStart.AddDate(0, 1, 0)
		default:
			current = periodStart.Add(7 * 24 * time.Hour)
		}
		
		// Safety check to prevent infinite loops
		if len(periods) > 1000 {
			break
		}
	}
	
	return periods
}

// calculateCommitCadenceStats computes comprehensive cadence statistics
func calculateCommitCadenceStats(periods []TimePeriod) *CommitCadenceStats {
	stats := &CommitCadenceStats{
		TotalPeriods: len(periods),
		TimePeriods:  periods,
	}
	
	if len(periods) == 0 {
		stats.TrendDirection = "Unknown"
		stats.SustainabilityLevel = "Unknown"
		return stats
	}
	
	// Calculate basic statistics
	totalCommits := 0
	for _, period := range periods {
		totalCommits += period.CommitCount
	}
	stats.TotalCommits = totalCommits
	stats.AverageCommitsPerPeriod = float64(totalCommits) / float64(len(periods))
	
	// Calculate trend direction and strength
	if len(periods) >= 2 {
		slope := calculateTrendSlope(periods)
		stats.TrendDirection, stats.TrendStrength = classifyTrendDirection(slope)
	} else {
		stats.TrendDirection = "Unknown"
		stats.TrendStrength = 0.0
	}
	
	// Detect spikes and dips
	stats.Spikes = detectCommitSpikes(periods)
	stats.Dips = detectCommitDips(periods)
	
	// Assess sustainability level
	stats.SustainabilityLevel = classifySustainabilityLevel(
		stats.AverageCommitsPerPeriod,
		len(stats.Spikes),
		len(stats.Dips),
		stats.TrendDirection,
	)
	
	return stats
}

// calculateTrendSlope calculates the linear regression slope for trend analysis
func calculateTrendSlope(periods []TimePeriod) float64 {
	n := float64(len(periods))
	if n < 2 {
		return 0
	}
	
	var sumX, sumY, sumXY, sumX2 float64
	
	for i, period := range periods {
		x := float64(i)
		y := float64(period.CommitCount)
		
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Linear regression: slope = (n*ΣXY - ΣX*ΣY) / (n*ΣX² - (ΣX)²)
	numerator := n*sumXY - sumX*sumY
	denominator := n*sumX2 - sumX*sumX
	
	if denominator == 0 {
		return 0
	}
	
	return numerator / denominator
}

// classifyTrendDirection determines trend direction and strength from slope
func classifyTrendDirection(slope float64) (string, float64) {
	absSlope := math.Abs(slope)
	
	if absSlope <= stableTrendThreshold {
		return "Stable", absSlope
	} else if slope > 0 {
		return "Increasing", absSlope
	} else {
		return "Decreasing", absSlope
	}
}

// detectCommitSpikes identifies periods with unusually high commit activity
func detectCommitSpikes(periods []TimePeriod) []TimePeriod {
	// Require minimum periods for robust detection
	if len(periods) < 5 {
		return []TimePeriod{}
	}
	
	// Use median for more robust baseline (less affected by extremes)
	counts := make([]int, len(periods))
	for i, period := range periods {
		counts[i] = period.CommitCount
	}
	sort.Ints(counts)
	
	var baseline float64
	mid := len(counts) / 2
	if len(counts)%2 == 0 {
		baseline = float64(counts[mid-1]+counts[mid]) / 2
	} else {
		baseline = float64(counts[mid])
	}
	
	// Guard against zero baseline
	if baseline == 0 {
		return []TimePeriod{}
	}
	
	var spikes []TimePeriod
	spikeThreshold := baseline * spikeThresholdMultiplier
	
	for _, period := range periods {
		if float64(period.CommitCount) > spikeThreshold {
			spike := period
			
			// Classify spike severity
			ratio := float64(period.CommitCount) / baseline
			if ratio >= 2.5 {
				spike.Severity = "High"
			} else if ratio >= 2.0 {
				spike.Severity = "Medium"
			} else {
				spike.Severity = "Low"
			}
			
			spikes = append(spikes, spike)
		}
	}
	
	return spikes
}

// detectCommitDips identifies periods with unusually low commit activity
func detectCommitDips(periods []TimePeriod) []TimePeriod {
	// Require minimum periods for robust detection
	if len(periods) < 5 {
		return []TimePeriod{}
	}
	
	// Use median for more robust baseline (less affected by extremes)
	counts := make([]int, len(periods))
	for i, period := range periods {
		counts[i] = period.CommitCount
	}
	sort.Ints(counts)
	
	var baseline float64
	mid := len(counts) / 2
	if len(counts)%2 == 0 {
		baseline = float64(counts[mid-1]+counts[mid]) / 2
	} else {
		baseline = float64(counts[mid])
	}
	
	// Guard against zero baseline
	if baseline == 0 {
		return []TimePeriod{}
	}
	
	var dips []TimePeriod
	dipThreshold := baseline * dipThresholdMultiplier
	
	for _, period := range periods {
		if float64(period.CommitCount) < dipThreshold {
			dip := period
			
			// Classify dip severity
			ratio := float64(period.CommitCount) / baseline
			if ratio <= 0.15 {
				dip.Severity = "High"
			} else if ratio <= 0.25 {
				dip.Severity = "Medium"
			} else {
				dip.Severity = "Low"
			}
			
			dips = append(dips, dip)
		}
	}
	
	return dips
}

// classifySustainabilityLevel assesses overall development pace sustainability
func classifySustainabilityLevel(avgCommits float64, spikeCount, dipCount int, trendDirection string) string {
	// Critical conditions
	if spikeCount >= criticalSpikesThreshold && dipCount >= 2 {
		return "Critical"
	}
	if avgCommits > healthyAvgCommitsHigh && spikeCount >= warningSpikesThreshold {
		return "Critical"
	}
	
	// Warning conditions
	if spikeCount >= warningSpikesThreshold {
		return "Warning"
	}
	if avgCommits > healthyAvgCommitsHigh && trendDirection == "Increasing" {
		return "Warning"
	}
	
	// Caution conditions
	if dipCount >= 2 {
		return "Caution"
	}
	if avgCommits < healthyAvgCommitsLow && trendDirection == "Decreasing" {
		return "Caution"
	}
	if trendDirection == "Decreasing" && spikeCount > 0 {
		return "Caution"
	}
	if trendDirection == "Decreasing" && dipCount > 0 {
		return "Caution"
	}
	
	// Healthy conditions
	if avgCommits >= healthyAvgCommitsLow && avgCommits <= healthyAvgCommitsHigh {
		if spikeCount <= 1 && dipCount <= 1 && trendDirection != "Decreasing" {
			return "Healthy"
		}
	}
	
	// Default to caution if unclear
	return "Caution"
}

// commitAffectsPath checks if a commit affects the specified path
// Optimized for performance with early returns and minimal diff computation
func commitAffectsPath(commit *object.Commit, pathArg string) (bool, error) {
	if commit.NumParents() == 0 {
		// Initial commit - check if it has files in the path
		tree, err := commit.Tree()
		if err != nil {
			return false, err
		}
		
		found := false
		err = tree.Files().ForEach(func(file *object.File) error {
			if matchesPathFilter(file.Name, pathArg) {
				found = true
				// Note: could use early exit here, but for simplicity we let the loop complete
			}
			return nil
		})
		return found, err
	}
	
	// Regular commit - optimized diff check
	parent, err := commit.Parent(0)
	if err != nil {
		return false, err
	}
	
	parentTree, err := parent.Tree()
	if err != nil {
		return false, err
	}
	
	currentTree, err := commit.Tree()
	if err != nil {
		return false, err
	}
	
	// Use efficient patch computation (single patch, not per-file)
	patch, err := parentTree.Patch(currentTree)
	if err != nil {
		return false, err
	}
	
	// Early return on first match for better performance
	stats := patch.Stats()
	for _, fileStat := range stats {
		if matchesPathFilter(fileStat.Name, pathArg) {
			return true, nil
		}
	}
	
	return false, nil
}

// printCommitCadenceStats displays the analysis results
func printCommitCadenceStats(stats *CommitCadenceStats, period string) {
	fmt.Printf("Commit Cadence Trends Analysis\n")
	fmt.Printf("Time period grouping: %s\n", period)
	fmt.Printf("Total commits analyzed: %d\n", stats.TotalCommits)
	fmt.Printf("Total periods: %d\n", stats.TotalPeriods)
	
	if stats.TotalPeriods == 0 {
		fmt.Printf("No commits found in the specified criteria.\n")
		return
	}
	
	fmt.Printf("Average commits per %s: %.1f\n", period, stats.AverageCommitsPerPeriod)
	fmt.Printf("\n")
	
	// Trend analysis
	fmt.Printf("Trend Analysis:\n")
	fmt.Printf("  Direction: %s", stats.TrendDirection)
	if stats.TrendDirection != "Unknown" {
		fmt.Printf(" (strength: %.2f)", stats.TrendStrength)
	}
	fmt.Printf("\n")
	
	// Sustainability assessment
	fmt.Printf("  Sustainability: %s\n", stats.SustainabilityLevel)
	fmt.Printf("\n")
	
	// Context and research
	fmt.Printf("Context: Track trends, not absolutes - spikes/dips may reveal crunch, burnout, or stagnation (Kent Beck).\n\n")
	
	// Spikes analysis
	if len(stats.Spikes) > 0 {
		fmt.Printf("Commit Spikes Detected (%d):\n", len(stats.Spikes))
		for i, spike := range stats.Spikes {
			fmt.Printf("  %d. %s - %s (%d commits, %s severity)\n", 
				i+1, 
				spike.Start.Format("2006-01-02"), 
				spike.End.Format("2006-01-02"),
				spike.CommitCount,
				spike.Severity)
		}
		fmt.Printf("\n")
	}
	
	// Dips analysis
	if len(stats.Dips) > 0 {
		fmt.Printf("Commit Dips Detected (%d):\n", len(stats.Dips))
		for i, dip := range stats.Dips {
			fmt.Printf("  %d. %s - %s (%d commits, %s severity)\n", 
				i+1, 
				dip.Start.Format("2006-01-02"), 
				dip.End.Format("2006-01-02"),
				dip.CommitCount,
				dip.Severity)
		}
		fmt.Printf("\n")
	}
	
	// Recent periods (last 5)
	if len(stats.TimePeriods) > 0 {
		recentCount := 5
		if len(stats.TimePeriods) < 5 {
			recentCount = len(stats.TimePeriods)
		}
		fmt.Printf("Recent Periods (last %d):\n", recentCount)
		
		for i := len(stats.TimePeriods) - recentCount; i < len(stats.TimePeriods); i++ {
			period := stats.TimePeriods[i]
			fmt.Printf("  %s - %s: %d commits\n", 
				period.Start.Format("2006-01-02"), 
				period.End.Format("2006-01-02"),
				period.CommitCount)
		}
		fmt.Printf("\n")
	}
	
	// Recommendations
	fmt.Printf("Recommendations:\n")
	switch stats.SustainabilityLevel {
	case "Critical":
		fmt.Printf("  • URGENT: High spike/dip volatility suggests unsustainable pace\n")
		fmt.Printf("  • Consider workload balancing and process improvements\n")
		fmt.Printf("  • Review team capacity and project planning\n")
	case "Warning":
		fmt.Printf("  • Multiple spikes detected - monitor for crunch periods\n")
		fmt.Printf("  • Consider more consistent development pace\n")
		fmt.Printf("  • Review sprint planning and task estimation\n")
	case "Caution":
		fmt.Printf("  • Some concerning patterns detected\n")
		if stats.TrendDirection == "Decreasing" {
			fmt.Printf("  • Decreasing trend may indicate reduced team velocity\n")
		}
		if len(stats.Dips) > 0 {
			fmt.Printf("  • Low activity periods may indicate blockers or burnout\n")
		}
	case "Healthy":
		fmt.Printf("  • Good sustainable development pace!\n")
		fmt.Printf("  • Continue current practices and monitor trends\n")
	case "Unknown":
		fmt.Printf("  • Insufficient data for sustainability assessment\n")
		fmt.Printf("  • Consider analyzing a longer time period\n")
	}
	
	// Overall pace assessment
	if stats.AverageCommitsPerPeriod > 0 {
		if stats.AverageCommitsPerPeriod < healthyAvgCommitsLow {
			fmt.Printf("  • Low average pace (%.1f commits/%s) - consider if adequate\n", 
				stats.AverageCommitsPerPeriod, period)
		} else if stats.AverageCommitsPerPeriod > healthyAvgCommitsHigh {
			fmt.Printf("  • High average pace (%.1f commits/%s) - ensure sustainability\n", 
				stats.AverageCommitsPerPeriod, period)
		}
	}
}
