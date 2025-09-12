# Gitallica

![Gitallica](docs/lars.png)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Build Status](https://img.shields.io/badge/Build-Passing-green.svg)](https://github.com/bgricker/gitallica)

**Gitallica** performs temporal diff analysis of distributed version control logs to help you understand code evolution, identify risks, and optimize team workflows. Analyze churn patterns, code survival rates, and other engineering metrics to make data-driven decisions about your codebase health.

## Features

- **14 Research-Backed Metrics**: Based on industry studies from Microsoft Research, DORA, and Clean Code principles
- **DORA Compliance**: Elite/High/Medium/Low performance classification using Accelerate benchmarks
- **Real Git Analysis**: Authentic repository analysis using go-git (no external dependencies)
- **Comprehensive Coverage**: From code survival to change lead time, bus factor to commit cadence
- **CLI-First Design**: Simple commands with powerful filtering and time-window analysis

## Quick Start

```bash
# Run comprehensive health check (recommended first step)
gitallica health-check --last 30d

# Analyze code churn patterns
gitallica churn --last 30d

# Check code survival rates
gitallica survival --last 6m

# Measure DORA performance
gitallica change-lead-time --limit 10
```

## Installation

**Coming soon**: Homebrew installation will be available after project stabilization.

For now, build from source:
```bash
git clone https://github.com/bgricker/gitallica.git
cd gitallica
go build -o gitallica .
```


## Available Commands

| Command | Description | Research Basis |
|---------|-------------|----------------|
| `health-check` | Comprehensive health check and issue identification | All metrics combined |
| `churn` | Additions vs. deletions ratio | Microsoft Research |
| `survival` | Code survival rate analysis | Spinellis et al. |
| `churn-files` | High-churn files identification | Nagappan & Ball |
| `component-creation` | New component creation rate | Kent Beck |
| `directory-entropy` | Directory structure entropy | Edsger Dijkstra |
| `dead-zones` | Untouched code identification | CodeScene |
| `bus-factor` | Knowledge concentration analysis | GitHub empirical studies |
| `ownership-clarity` | Code ownership patterns | Microsoft Research |
| `onboarding-footprint` | New contributor analysis | Robert C. Martin |
| `test-ratio` | Test-to-code ratio | TSP study |
| `high-risk-commits` | Large commit identification | Nokia Bell Labs |
| `commit-cadence` | Commit frequency trends | Kent Beck |
| `long-lived-branches` | Branch lifecycle analysis | DORA research |
| `change-lead-time` | DORA lead time metrics | DORA State of DevOps |

**Note**: Review Bottlenecks (#13) requires GitHub API integration and is planned for future implementation.

## Usage Examples

### Basic Analysis
```bash
# Run comprehensive health check (recommended first step)
gitallica health-check --last 30d

# Analyze entire repository
gitallica churn

# Time-scoped analysis
gitallica survival --last 3m

# Path-specific analysis
gitallica bus-factor --path src/
```

### Advanced Filtering
```bash
# Combined filters
gitallica churn --last 90d --path lib/

# Multiple paths at once (NEW!)
gitallica churn-files --path src/ --path lib/ --path app/
gitallica bus-factor --path cmd/ --path main.go --path README.md

# Detailed output
gitallica change-lead-time --limit 20 --method tag
```

### Configuration File

Create a `.gitallica.yaml` file to avoid repeating common options:

```yaml
# Default paths to analyze for each command
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

# Global settings
defaults:
  last: "6m"  # Default time window
  limit: 20  # Default number of results to show
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

**Configuration Visibility:**
Every command now shows you exactly what configuration is being used:

```bash
❯ gitallica churn --last 7d --path README.md
Using config file: /Users/benricker/code/gitallica/.gitallica.yaml
=== Churn Analysis Scope ===
Time window: last 7 days
Path filter: README.md (from CLI)
```

This helps you:
- **Remember project settings** when working with multiple repositories
- **Spot unintended configurations** when CLI flags override YAML settings
- **Debug configuration issues** by seeing exactly what's being used

## Documentation

- **[User Guide](docs/USER_GUIDE.md)** - Comprehensive usage guide
- **[Command Reference](docs/COMMANDS.md)** - Detailed command documentation
- **[Research Methodology](docs/RESEARCH.md)** - Research basis and thresholds

## Research Foundation

Gitallica implements 14 metrics based on authoritative sources:

- **DORA Metrics**: Elite/High/Medium/Low performance classification
- **Clean Code Principles**: Robert C. Martin's guidelines
- **Microsoft Research**: Code survival and churn analysis
- **Accelerate Research**: Lead time and deployment frequency
- **Industry Benchmarks**: Bus factor, ownership patterns, and more

Each metric includes:
- Research-backed thresholds
- Industry-standard classifications
- Actionable recommendations
- Performance benchmarking

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **DORA Research**: DevOps Research and Assessment team
- **Microsoft Research**: Code survival and churn studies
- **Clean Code**: Robert C. Martin's principles
- **Accelerate**: Nicole Forsgren, Jez Humble, Gene Kim
- **go-git**: Git implementation for Go

---

*Gitallica - Shred your git history. Rock your repo insights.* 🎸
