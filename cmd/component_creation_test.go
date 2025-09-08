package cmd

import (
	"sort"
	"testing"
)

func TestDetectComponentsInFile(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		content        string
		expectedTypes  map[string]int
	}{
		{
			name:     "JavaScript class",
			filePath: "test.js",
			content:  "class MyClass {\n  constructor() {}\n}",
			expectedTypes: map[string]int{
				"javascript-class": 1,
			},
		},
		{
			name:     "React component",
			filePath: "Component.jsx",
			content:  "function MyComponent() {\n  return <div>Hello</div>;\n}",
			expectedTypes: map[string]int{
				"react-component": 1,
			},
		},
		{
			name:     "Ruby model",
			filePath: "user.rb",
			content:  "class User < ApplicationRecord\n  validates :email, presence: true\nend",
			expectedTypes: map[string]int{
				"ruby-model": 1,
			},
		},
		{
			name:     "Ruby controller",
			filePath: "users_controller.rb",
			content:  "class UsersController < ApplicationController\n  def index\n  end\nend",
			expectedTypes: map[string]int{
				"ruby-controller": 1,
			},
		},
		{
			name:     "Python class",
			filePath: "models.py",
			content:  "class User:\n    def __init__(self):\n        pass",
			expectedTypes: map[string]int{
				"python-class": 1,
			},
		},
		{
			name:     "Go struct",
			filePath: "user.go",
			content:  "type User struct {\n    Name string\n    Email string\n}",
			expectedTypes: map[string]int{
				"go-struct": 1,
			},
		},
		{
			name:     "Go interface",
			filePath: "interface.go",
			content:  "type Reader interface {\n    Read([]byte) (int, error)\n}",
			expectedTypes: map[string]int{
				"go-interface": 1,
			},
		},
		{
			name:     "Java class",
			filePath: "User.java",
			content:  "public class User {\n    private String name;\n}",
			expectedTypes: map[string]int{
				"java-class": 1,
			},
		},
		{
			name:     "C# class",
			filePath: "User.cs",
			content:  "public class User {\n    public string Name { get; set; }\n}",
			expectedTypes: map[string]int{
				"csharp-class": 1,
			},
		},
		{
			name:     "Multiple components",
			filePath: "mixed.js",
			content:  "class MyClass {}\nfunction MyComponent() {}\nconst AnotherComponent = () => {}",
			expectedTypes: map[string]int{
				"javascript-class": 1,
				"react-component":  2,
			},
		},
		{
			name:          "No components",
			filePath:      "plain.txt",
			content:       "This is just plain text",
			expectedTypes: map[string]int{},
		},
		{
			name:          "Wrong extension",
			filePath:      "test.txt",
			content:       "class MyClass {}",
			expectedTypes: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectComponentsInFile(tt.filePath, tt.content)
			
			if len(result) != len(tt.expectedTypes) {
				t.Errorf("Expected %d component types, got %d", len(tt.expectedTypes), len(result))
			}
			
			for expectedType, expectedCount := range tt.expectedTypes {
				if actualCount, exists := result[expectedType]; !exists {
					t.Errorf("Expected component type %s not found", expectedType)
				} else if actualCount != expectedCount {
					t.Errorf("Expected %d components of type %s, got %d", expectedCount, expectedType, actualCount)
				}
			}
		})
	}
}

func TestCalculateCreationRate(t *testing.T) {
	tests := []struct {
		name           string
		stats          []ComponentCreationStats
		timeWindow     string
		expectedTotal  int
		expectedSpike  bool
		expectedReason string
	}{
		{
			name: "Low creation rate",
			stats: []ComponentCreationStats{
				{ComponentType: "javascript-class", Count: 3},
				{ComponentType: "react-component", Count: 2},
			},
			timeWindow:     "last 30d",
			expectedTotal:  5,
			expectedSpike:  false,
			expectedReason: "",
		},
		{
			name: "High creation rate (spike detected)",
			stats: []ComponentCreationStats{
				{ComponentType: "javascript-class", Count: 8},
				{ComponentType: "react-component", Count: 5},
			},
			timeWindow:     "last 7d",
			expectedTotal:  13,
			expectedSpike:  true,
			expectedReason: "High component creation rate: 13 components",
		},
		{
			name:           "No components",
			stats:          []ComponentCreationStats{},
			timeWindow:     "last 30d",
			expectedTotal:  0,
			expectedSpike:  false,
			expectedReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := calculateCreationRate(tt.stats, tt.timeWindow)
			
			if rate.TotalCreated != tt.expectedTotal {
				t.Errorf("Expected total %d, got %d", tt.expectedTotal, rate.TotalCreated)
			}
			
			if rate.SpikeDetected != tt.expectedSpike {
				t.Errorf("Expected spike %v, got %v", tt.expectedSpike, rate.SpikeDetected)
			}
			
			if rate.SpikeReason != tt.expectedReason {
				t.Errorf("Expected reason %q, got %q", tt.expectedReason, rate.SpikeReason)
			}
			
			if rate.TimeWindow != tt.timeWindow {
				t.Errorf("Expected time window %q, got %q", tt.timeWindow, rate.TimeWindow)
			}
		})
	}
}

func TestComponentTypes(t *testing.T) {
	// Test that all component types have valid patterns
	for typeKey, componentType := range componentTypes {
		t.Run(typeKey, func(t *testing.T) {
			if componentType.Name == "" {
				t.Errorf("Component type %s has empty name", typeKey)
			}
			
			if len(componentType.Patterns) == 0 {
				t.Errorf("Component type %s has no patterns", typeKey)
			}
			
			if len(componentType.Extensions) == 0 {
				t.Errorf("Component type %s has no extensions", typeKey)
			}
			
			if componentType.Description == "" {
				t.Errorf("Component type %s has no description", typeKey)
			}
			
			// Test that patterns compile
			for i, pattern := range componentType.Patterns {
				if pattern == nil {
					t.Errorf("Component type %s has nil pattern at index %d", typeKey, i)
				}
			}
		})
	}
}

func TestComponentCreationStatsSorting(t *testing.T) {
	stats := []ComponentCreationStats{
		{ComponentType: "low-count", Count: 1},
		{ComponentType: "high-count", Count: 10},
		{ComponentType: "medium-count", Count: 5},
	}
	
	// Sort by count (descending)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})
	
	expectedOrder := []string{"high-count", "medium-count", "low-count"}
	for i, expected := range expectedOrder {
		if stats[i].ComponentType != expected {
			t.Errorf("Expected %s at position %d, got %s", expected, i, stats[i].ComponentType)
		}
	}
}
