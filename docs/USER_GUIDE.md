# Gitallica User Guide

This comprehensive guide covers all aspects of using Gitallica for code analysis and team workflow optimization. Gitallica implements 14 research-backed metrics for comprehensive code analysis.

## Table of Contents

- [Getting Started](#getting-started)
- [Command Overview](#command-overview)
- [Time Windows](#time-windows)
- [Path Filtering](#path-filtering)
- [Advanced Usage](#advanced-usage)
- [Understanding Results](#understanding-results)
- [Best Practices](#best-practices)

## Getting Started

### Basic Commands

```bash
# Analyze code churn across entire history
gitallica churn

# Check code survival rates
gitallica survival

# Measure team knowledge distribution
gitallica bus-factor
```

### Time-Scoped Analysis

```bash
# Last 30 days
gitallica churn --last 30d

# Last 6 months
gitallica survival --last 6m

# Last year
gitallica bus-factor --last 1y
```

### Path-Specific Analysis

```bash
# Analyze specific directory
gitallica churn --path src/

# Analyze specific file
gitallica survival --path src/main.go

# Multiple paths (NEW!)
gitallica churn --path src/ --path lib/ --path app/
gitallica bus-factor --path cmd/ --path main.go --path README.md

# Combined filters
gitallica bus-factor --last 90d --path lib/
```

## Command Overview

### Code Evolution Metrics

#### `churn` - Additions vs. Deletions Ratio
Analyzes code volatility and stability patterns.

```bash
gitallica churn --last 30d
```

**Research Basis**: Microsoft Research, Edsger Dijkstra
**Threshold**: Keep churn under ~20% of codebase size

#### `survival` - Code Survival Rate
Measures how long code survives before being changed or deleted.

```bash
gitallica survival --last 6m --path src/
```

**Research Basis**: MSR, CodeScene research
**Threshold**: Investigate areas where <50% of lines survive 6-12 months

#### `churn-files` - High-Churn Files
Identifies files that are frequently modified.

```bash
gitallica churn-files --last 3m --limit 10
```

**Research Basis**: Martin Fowler's Refactoring principles
**Threshold**: File churn >20% flags instability

### Team Performance Metrics

#### `bus-factor` - Knowledge Concentration
Measures how many team members understand each part of the codebase.

```bash
gitallica bus-factor --path src/
```

**Research Basis**: Martin Fowler's collective ownership principles
**Threshold**: Target 25-50% of team size (e.g., 4-5 in a 10-person team)

#### `ownership-clarity` - Code Ownership Patterns
Analyzes ownership concentration and distribution.

```bash
gitallica ownership-clarity --last 6m
```

**Research Basis**: Industry research on ownership patterns
**Threshold**: Flag files with no contributor ≥40-50% of commits

#### `onboarding-footprint` - New Contributor Analysis
Analyzes what new developers touch first.

```bash
gitallica onboarding-footprint --last 1y
```

**Research Basis**: Robert C. Martin's Clean Code principles
**Threshold**: New contributors touching >10-20 files in first 5 commits

### Quality Metrics

#### `test-ratio` - Test Coverage Analysis
Measures test-to-code ratio.

```bash
gitallica test-ratio --path src/
```

**Research Basis**: Robert C. Martin's Clean Code
**Threshold**: Test-to-code ratio ≈1:1, up to 2:1

#### `high-risk-commits` - Large Commit Analysis
Identifies commits that pose review and rollback risks.

```bash
gitallica high-risk-commits --last 30d --limit 20
```

**Research Basis**: Kent Beck, Martin Fowler
**Threshold**: >400 lines or >10-12 files per commit is high risk

### Delivery Performance Metrics

#### `change-lead-time` - DORA Lead Time
Measures time from commit to deployment using DORA benchmarks.

```bash
gitallica change-lead-time --method merge --limit 10
```

**Research Basis**: DORA State of DevOps, Accelerate research
**Thresholds**:
- Elite: <1 day
- High: 1 day-1 week
- Medium: 1 week-1 month
- Low: >1 month

#### `commit-cadence` - Commit Frequency Trends
Analyzes commit patterns and sustainability.

```bash
gitallica commit-cadence --last 6m
```

**Research Basis**: Kent Beck's Extreme Programming principles
**Threshold**: Track trends, spikes/dips reveal crunch or stagnation

#### `long-lived-branches` - Branch Lifecycle Analysis
Identifies branches that live too long.

```bash
gitallica long-lived-branches --last 3m
```

**Research Basis**: DORA research, trunk-based development
**Threshold**: Branches older than a few days are risky

### Architecture Metrics

#### `component-creation` - New Component Rate
Tracks creation of new components and modules.

```bash
gitallica component-creation --last 1y
```

**Research Basis**: Industry benchmarks for component growth
**Threshold**: Context-aware analysis based on project type

#### `directory-entropy` - Structure Complexity
Measures directory organization and modularity.

```bash
gitallica directory-entropy --limit 10
```

**Research Basis**: Edsger Dijkstra's simplicity principles
**Threshold**: Context-aware analysis with different rules for root vs. subdirectories

#### `dead-zones` - Untouched Code
Identifies code that hasn't been modified recently.

```bash
gitallica dead-zones --last 1y
```

**Research Basis**: Robert C. Martin's Clean Code
**Threshold**: Files untouched for ≥12 months that remain active

## Time Windows

Gitallica supports flexible time window analysis:

### Format
`--last #{number}{unit}`

### Units
- `d` - Days
- `m` - Months  
- `y` - Years

### Examples
```bash
gitallica churn --last 7d    # Last week
gitallica survival --last 3m # Last quarter
gitallica bus-factor --last 1y # Last year
```

## Path Filtering

Filter analysis to specific parts of your repository:

### Directory Analysis
```bash
gitallica churn --path src/
gitallica survival --path lib/
gitallica bus-factor --path cmd/
```

### File Analysis
```bash
gitallica churn --path src/main.go
gitallica survival --path package.json
```

### Combined Filters
```bash
gitallica churn --last 30d --path src/
gitallica survival --last 6m --path lib/ --limit 5
```

## Configuration

### Setup Configuration File

Create configuration files to avoid repeating common options:

```bash
# Global configuration (applies to all projects)
cp .gitallica.yaml.example ~/.gitallica.yaml

# Project-specific configuration (overrides global settings)
cp .gitallica.yaml.example .gitallica.yaml
```

**Configuration Hierarchy:**
- **Project-specific**: `.gitallica.yaml` or `.gitallica.yml` in your project root
- **Global**: `~/.gitallica.yaml` in your home directory  
- **Explicit**: `--config /path/to/config.yaml` flag
- **CLI flags**: Always override configuration files

### Configuration Example

```yaml
# Per-command path configuration
churn:
  paths:
    - "src/"
    - "lib/"
    - "app/"

bus-factor:
  paths:
    - "src/"
    - "lib/"

test-ratio:
  paths:
    - "src/"
    - "tests/"

# Global defaults
defaults:
  last: "6m"  # Default time window
  limit: 20  # Default number of results to show
```

### Using Configuration

```bash
# Uses your config file settings
gitallica churn

# Overrides config with CLI flags
gitallica churn --path README.md

# Multiple paths override config
gitallica churn --path src/ --path lib/
```

**Configuration Priority:**
1. Command-line flags (highest priority)
2. Project-specific `.gitallica.yaml` or `.gitallica.yml` in current directory
3. Home directory `~/.gitallica.yaml` (lowest priority)
4. Default values (fallback)

## Advanced Usage

### Output Control
```bash
# Limit detailed output
gitallica high-risk-commits --limit 5

# Show summary only
gitallica commit-size --summary
```

### Method Selection
```bash
# Different calculation methods
gitallica change-lead-time --method merge
gitallica change-lead-time --method tag
```

### Debugging
```bash
# Enable debug output
gitallica survival --debug
```

## Understanding Results

### Performance Classifications

#### DORA Performance Levels
- **Elite**: Industry-leading performance
- **High**: Strong performance with room for optimization
- **Medium**: Moderate performance requiring improvement
- **Low**: Significant improvement needed

#### Risk Classifications
- **Low Risk**: Safe, well-contained changes
- **Medium Risk**: Moderate complexity requiring attention
- **High Risk**: Complex changes needing extra review
- **Critical Risk**: Major changes requiring architectural consideration

### Statistical Measures
- **Average**: Mean value across all data points
- **Median**: Middle value (50th percentile)
- **95th Percentile**: Value below which 95% of data falls
- **Distribution**: Breakdown across performance categories

## Best Practices

### Regular Analysis
```bash
# Weekly team review
gitallica churn --last 7d
gitallica high-risk-commits --last 7d

# Monthly architecture review
gitallica bus-factor --last 1m
gitallica directory-entropy --last 1m

# Quarterly performance review
gitallica change-lead-time --last 3m
gitallica survival --last 3m
```

### CI/CD Integration
```bash
# Automated quality gates
gitallica test-ratio --path src/
gitallica high-risk-commits --last 1d --limit 0
```

### Team Onboarding
```bash
# New team member analysis
gitallica onboarding-footprint --last 1m
gitallica bus-factor --path src/
```

### Performance Monitoring
```bash
# DORA metrics tracking
gitallica change-lead-time --method merge
gitallica commit-cadence --last 1m
gitallica long-lived-branches --last 1m
```

## Troubleshooting

### Common Issues

#### No Data Returned
- Check if the time window contains commits
- Verify path exists in repository
- Ensure repository has sufficient history

#### Unexpected Results
- Review command flags and syntax
- Check if path filter is too restrictive
- Verify time window format

#### Performance Issues
- Use smaller time windows for large repositories
- Limit output with `--limit` flag
- Consider path filtering for focused analysis

### Getting Help

- Check command help: `gitallica <command> --help`
- Review this user guide
- Check [Command Reference](COMMANDS.md) for detailed options
- See [Research Methodology](RESEARCH.md) for threshold explanations
