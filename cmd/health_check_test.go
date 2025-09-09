package cmd

import (
	"testing"
)

func TestGetSeverityScore(t *testing.T) {
	tests := []struct {
		severity string
		expected int
	}{
		{"Critical", 100},
		{"High", 75},
		{"Medium", 50},
		{"Low", 25},
		{"Warning", 60},
		{"Caution", 40},
		{"Unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := getSeverityScore(tt.severity)
			if result != tt.expected {
				t.Errorf("Expected severity score %d for %s, got %d", tt.expected, tt.severity, result)
			}
		})
	}
}

func TestCategorizeIssue(t *testing.T) {
	tests := []struct {
		metric   string
		expected string
	}{
		{"churn", "Code Stability"},
		{"bus-factor", "Knowledge Management"},
		{"test-ratio", "Code Quality"},
		{"dead-zones", "Technical Debt"},
		{"commit-cadence", "Development Practices"},
		{"change-lead-time", "DORA Performance"},
		{"unknown-metric", "General"},
	}

	for _, tt := range tests {
		t.Run(tt.metric, func(t *testing.T) {
			result := categorizeIssue(tt.metric)
			if result != tt.expected {
				t.Errorf("Expected category %s for metric %s, got %s", tt.expected, tt.metric, result)
			}
		})
	}
}

func TestGenerateHealthSummary(t *testing.T) {
	tests := []struct {
		name     string
		report   *HealthReport
		expected string
	}{
		{
			name: "no issues",
			report: &HealthReport{
				TotalIssues: 0,
			},
			expected: "✅ Excellent! No significant issues detected. Your codebase appears healthy.",
		},
		{
			name: "critical issues only",
			report: &HealthReport{
				TotalIssues:    2,
				CriticalIssues: 2,
			},
			expected: "🚨 2 critical issues require immediate attention. ",
		},
		{
			name: "mixed issues",
			report: &HealthReport{
				TotalIssues:    5,
				CriticalIssues: 1,
				HighIssues:     2,
				MediumIssues:   1,
				LowIssues:      1,
			},
			expected: "🚨 1 critical issue requires immediate attention. ⚠️ 2 high-priority issues should be addressed soon. 📋 1 medium-priority issue needs attention. 💡 1 low-priority issue can be addressed when convenient.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateHealthSummary(tt.report)
			if result != tt.expected {
				t.Errorf("Expected summary %q, got %q", tt.expected, result)
			}
		})
	}
}
