# Command Reference

Complete reference for all Gitallica commands, flags, and options.

## Global Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--config` | Config file path | `--config ~/.gitallica.yaml` |
| `--help` | Show help for command | `gitallica churn --help` |

## Commands Overview

### Code Evolution Commands

#### `churn`
Analyzes additions vs. deletions ratio to measure code volatility.

**Flags:**
- `--last string`: Time window (e.g., `30d`, `6m`, `1y`)
- `--path string`: Limit to specific directory or file (can be specified multiple times)

**Examples:**
```bash
gitallica churn
gitallica churn --last 30d
gitallica churn --path src/
gitallica churn --last 6m --path lib/
```

**Output:**
- Total additions and deletions
- Churn percentage
- Research-based thresholds
- Recommendations

#### `survival`
Measures code survival rate - how long code survives before changes.

**Flags:**
- `--last string`: Time window for analysis
- `--path string`: Limit to specific path
- `--debug`: Enable debug output

**Examples:**
```bash
gitallica survival
gitallica survival --last 6m
gitallica survival --path src/ --debug
```

**Output:**
- Lines added vs. still present
- Survival rate percentage
- Research context
- Recommendations

#### `churn-files`
Identifies high-churn files and directories.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results to show (default 10)

**Examples:**
```bash
gitallica churn-files
gitallica churn-files --last 3m --limit 20
gitallica churn-files --path src/
```

**Output:**
- Files with highest churn rates
- Churn percentages
- Risk classifications
- Recommendations

### Team Performance Commands

#### `bus-factor`
Analyzes knowledge concentration per directory.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit to specific directory
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica bus-factor
gitallica bus-factor --path src/
gitallica bus-factor --last 1y --limit 5
```

**Output:**
- Bus factor per directory
- Knowledge concentration analysis
- Risk assessment
- Recommendations

#### `ownership-clarity`
Analyzes code ownership patterns across files.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica ownership-clarity
gitallica ownership-clarity --last 6m
gitallica ownership-clarity --path src/
```

**Output:**
- Ownership distribution
- Clarity classifications
- Risk indicators
- Recommendations

#### `onboarding-footprint`
Analyzes what new contributors touch first.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica onboarding-footprint
gitallica onboarding-footprint --last 1y
gitallica onboarding-footprint --path src/
```

**Output:**
- New contributor patterns
- File touch analysis
- Onboarding complexity
- Recommendations

### Quality Commands

#### `test-ratio`
Analyzes test-to-code ratio.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica test-ratio
gitallica test-ratio --path src/
gitallica test-ratio --last 6m
```

**Output:**
- Test vs. code ratios
- Classification levels
- Research context
- Recommendations

#### `high-risk-commits`
Identifies commits that pose high risk due to size or complexity.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica high-risk-commits
gitallica high-risk-commits --last 30d
gitallica high-risk-commits --path src/ --limit 20
```

**Output:**
- High-risk commits
- Risk classifications
- Research thresholds
- Recommendations

#### `commit-size`
Analyzes commit sizes and identifies risky commits.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)
- `--min-risk string`: Minimum risk level (Low, Medium, High, Critical)
- `--summary`: Show risk distribution summary

**Examples:**
```bash
gitallica commit-size
gitallica commit-size --last 30d --summary
gitallica commit-size --min-risk High --limit 5
```

**Output:**
- Commit size analysis
- Risk level distribution
- Detailed commit information
- Recommendations

### Delivery Performance Commands

#### `change-lead-time`
Analyzes change lead time using DORA metrics.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 5)
- `--method string`: Calculation method (`merge` or `tag`)

**Examples:**
```bash
gitallica change-lead-time
gitallica change-lead-time --method merge --limit 10
gitallica change-lead-time --last 90d --path src/
```

**Output:**
- Lead time statistics
- DORA performance levels
- Fastest/slowest commits
- Recommendations

#### `commit-cadence`
Analyzes commit frequency trends and sustainability.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 5)

**Examples:**
```bash
gitallica commit-cadence
gitallica commit-cadence --last 6m
gitallica commit-cadence --path src/ --limit 10
```

**Output:**
- Commit frequency trends
- Sustainability analysis
- Spike/dip detection
- Recommendations

#### `long-lived-branches`
Analyzes branch lifecycles and trunk-based development compliance.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)
- `--show-merged`: Include recently merged branches

**Examples:**
```bash
gitallica long-lived-branches
gitallica long-lived-branches --last 3m
gitallica long-lived-branches --show-merged --limit 20
```

**Output:**
- Branch age analysis
- Risk classifications
- Compliance levels
- Recommendations

### Architecture Commands

#### `component-creation`
Analyzes new component creation rate.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica component-creation
gitallica component-creation --last 1y
gitallica component-creation --path src/
```

**Output:**
- Component creation patterns
- Growth rate analysis
- Risk indicators
- Recommendations

#### `directory-entropy`
Analyzes directory structure entropy and modularity.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica directory-entropy
gitallica directory-entropy --last 30d
gitallica directory-entropy --path src/ --limit 5
```

**Output:**
- Entropy measurements
- Modularity analysis
- Risk classifications
- Recommendations

#### `dead-zones`
Identifies files untouched for extended periods.

**Flags:**
- `--last string`: Time window
- `--path string`: Limit analysis scope
- `--limit int`: Number of results (default 10)

**Examples:**
```bash
gitallica dead-zones
gitallica dead-zones --last 1y
gitallica dead-zones --path src/
```

**Output:**
- Untouched files
- Age analysis
- Risk assessments
- Recommendations

## Time Window Format

All commands support the `--last` flag with the following format:

### Syntax
`--last #{number}{unit}`

### Units
- `d` - Days
- `m` - Months
- `y` - Years

### Examples
```bash
--last 7d     # Last week
--last 30d    # Last month
--last 3m     # Last quarter
--last 6m     # Last half year
--last 1y     # Last year
--last 2y     # Last two years
```

## Path Filtering

All commands support the `--path` flag for filtering analysis scope. **Multiple paths are supported** for analyzing multiple directories or files simultaneously.

### Single Path Filtering
```bash
--path src/           # Analyze src/ directory
--path lib/           # Analyze lib/ directory
--path cmd/           # Analyze cmd/ directory
--path src/main.go    # Analyze specific file
```

### Multiple Path Filtering (NEW!)
```bash
# Multiple directories
gitallica churn --path src/ --path lib/ --path app/

# Mixed directories and files
gitallica bus-factor --path cmd/ --path main.go --path README.md

# Multiple specific files
gitallica test-ratio --path src/main.go --path tests/main_test.go
```

### Combined Filtering
```bash
gitallica churn --last 30d --path src/ --path lib/
gitallica survival --last 6m --path lib/ --path app/ --limit 5
```

### Configuration File Integration
```bash
# Set up default paths in ~/.gitallica.yaml
churn:
  paths:
    - "src/"
    - "lib/"
    - "app/"

# Then run without specifying paths
gitallica churn  # Uses config file paths

# Or override with CLI flags
gitallica churn --path README.md  # Overrides config, analyzes only README.md
```

## Output Control

### Limiting Results
```bash
--limit 5     # Show top 5 results
--limit 10    # Show top 10 results
--limit 20    # Show top 20 results
```

### Summary Mode
```bash
--summary     # Show summary statistics only
```

### Debug Mode
```bash
--debug       # Enable debug output
```

## Configuration

### Config File
Gitallica supports configuration via YAML file with **per-command path settings** and global defaults:

**Default location**: `~/.gitallica.yaml`

**Example configuration**:
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
  last: "30d"
  limit: 10
```

**Configuration Priority:**
1. Command-line flags (highest priority)
2. Project-specific `.gitallica.yaml` or `.gitallica.yml` in current directory
3. Home directory `~/.gitallica.yaml` (lowest priority)
4. Default values (fallback)

**Setup:**
```bash
# Global configuration (applies to all projects)
cp .gitallica.yaml.example ~/.gitallica.yaml

# Project-specific configuration (overrides global settings)
cp .gitallica.yaml.example .gitallica.yaml

# Customize your paths and settings
# Then run commands without specifying paths
gitallica churn  # Uses project config, falls back to global config
gitallica churn --path README.md  # Overrides all config files
```

**Configuration Hierarchy:**
- **Project-specific**: `.gitallica.yaml` or `.gitallica.yml` in your project root
- **Global**: `~/.gitallica.yaml` in your home directory  
- **Explicit**: `--config /path/to/config.yaml` flag
- **CLI flags**: Always override configuration files

## Error Handling

### Common Errors

#### Repository Not Found
```
Error: could not open repository: repository not found
```
**Solution**: Ensure you're running gitallica from within a Git repository.

#### Invalid Time Window
```
Error: invalid time window 'xyz': invalid format
```
**Solution**: Use correct format: `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).

#### Path Not Found
```
Error: path 'nonexistent/' not found
```
**Solution**: Verify the path exists in your repository.

### Debugging

Enable debug output for troubleshooting:
```bash
gitallica churn --debug
gitallica survival --last 30d --debug
```

## Performance Tips

### Large Repositories
- Use smaller time windows (`--last 30d` instead of `--last 1y`)
- Limit output with `--limit` flag
- Use path filtering to focus analysis

### Frequent Analysis
- Cache results for repeated queries
- Use configuration files for common settings
- Consider automated analysis scripts

### CI/CD Integration
- Use `--summary` flag for automated checks
- Set appropriate `--limit` values
- Combine with `--path` filtering for focused analysis
