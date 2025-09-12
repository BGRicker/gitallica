# Changelog

All notable changes to Gitallica will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-01-10

### Added
- **Multiple Path Support**: All commands now support multiple `--path` flags for analyzing multiple directories/files simultaneously
  - Example: `gitallica churn --path src/ --path lib/ --path app/`
  - Works across all 14 metrics for flexible analysis scope
- **Configuration File System**: Complete YAML configuration support with `.gitallica.yaml`
  - Per-command path configuration (e.g., `churn.paths`, `bus-factor.paths`)
  - Global defaults for `last` and `limit` settings
  - Configuration priority: CLI flags > config file > defaults
  - Example configuration file provided (`.gitallica.yaml.example`)
- **Enhanced CLI Experience**: Improved command-line interface with better path handling
  - Cross-platform path normalization for Windows/macOS/Linux compatibility
  - Backward compatibility maintained for single-path usage

### Changed
- **Path Filtering**: Enhanced from single-path to multi-path support across all commands
- **Configuration Management**: Centralized configuration system using Viper
- **Documentation**: Updated README with comprehensive configuration examples and setup instructions

### Technical Implementation
- **Path Processing**: New `matchesPathFilter()` and `matchesSinglePathFilter()` utilities
- **Configuration Integration**: `getConfigPaths()` helper for unified config/CLI handling
- **Cross-Platform Support**: Proper path handling for different operating systems
- **Backward Compatibility**: All existing single-path usage continues to work unchanged

## [Unreleased]

## [1.0.0] - 2025-01-XX

### Added
- **Change Lead Time Analysis**: DORA-compliant lead time measurement with real Git calculations
  - Elite/High/Medium/Low classification based on Accelerate research
  - Merge and tag-based calculation methods
  - Statistical analysis with median, 95th percentile, and distribution metrics
  - Performance-based recommendations and benchmarking

- **Long-Lived Branches Analysis**: Branch lifecycle analysis and trunk-based development compliance
  - Divergence-based age calculation from merge base
  - Git ancestry traversal for proper merge detection
  - Unique commit counting for feature branches
  - Risk classification and compliance assessment

- **Commit Cadence Trends**: Commit frequency analysis and sustainability assessment
  - Continuous time series with zero-fill for missing periods
  - Median-based baseline for spike/dip detection
  - ISO week calculation for consistent timezone-independent grouping
  - Trend analysis and sustainability recommendations

- **High-Risk Commits Analysis**: Large commit identification and risk assessment
  - Lines changed and files touched analysis
  - Risk classification based on Kent Beck and Martin Fowler principles
  - Performance optimization with single-patch computation
  - Detailed commit attribution and recommendations

- **Onboarding Footprint Analysis**: New contributor behavior analysis
  - Single-pass Git log iteration for memory efficiency
  - File touch pattern analysis for new developers
  - Onboarding complexity assessment
  - Recommendations for improving new developer experience

- **Ownership Clarity Analysis**: Code ownership pattern analysis
  - File-level ownership concentration measurement
  - Contributor distribution analysis
  - Risk assessment for diffuse vs. concentrated ownership
  - Recommendations for balanced ownership patterns

- **Code vs. Test Ratio Analysis**: Test coverage measurement
  - Test file vs. source file ratio analysis
  - Robert C. Martin Clean Code principles
  - Classification levels and recommendations
  - Path-specific analysis capabilities

- **Bus Factor Analysis**: Knowledge concentration measurement per directory
  - Efficient file-level authorship analysis
  - Commit-based traversal for accurate metrics
  - Risk assessment for knowledge concentration
  - Recommendations for collective ownership

- **Dead Zones Analysis**: Untouched code identification
  - Files untouched for extended periods detection
  - Age analysis and risk assessment
  - Recommendations for code maintenance
  - Path filtering and time window analysis

- **Directory Entropy Analysis**: Directory structure complexity measurement
  - Context-aware analysis for different project types
  - Modularity and organization assessment
  - Risk classification based on entropy levels
  - Recommendations for structural improvements

- **Component Creation Analysis**: New component creation rate tracking
  - Component and module creation pattern analysis
  - Growth rate assessment
  - Risk indicators for architectural stability
  - Context-aware thresholds

- **High-Churn Files Analysis**: File modification frequency analysis
  - High-churn file and directory identification
  - Martin Fowler Refactoring principles
  - Risk assessment for architectural attention
  - Recommendations for refactoring priorities

- **Code Survival Analysis**: Code lifespan measurement
  - Line-level survival tracking from addition to deletion
  - MSR and CodeScene research methodology
  - Survival rate analysis and recommendations
  - Path-specific and time-window analysis

- **Churn Analysis**: Additions vs. deletions ratio measurement
  - Code volatility and stability pattern analysis
  - Microsoft Research methodology
  - Edsger Dijkstra principles
  - Recommendations for code stability

### Technical Implementation
- **Real Git Analysis**: All metrics use authentic Git repository analysis with go-git
- **No External Dependencies**: Pure Go implementation using only go-git library
- **Comprehensive Test Coverage**: Test-driven development with >90% test coverage
- **Performance Optimization**: Efficient Git operations and memory management
- **Error Handling**: Robust error handling with sentinel errors and graceful degradation
- **CLI Interface**: Cobra-based command-line interface with consistent patterns
- **Configuration Support**: YAML configuration files and environment variables
- **Cross-Platform**: Works on macOS, Linux, and Windows

### Research Foundation
- **DORA Metrics**: State of DevOps research and Accelerate book implementation
- **Clean Code Principles**: Robert C. Martin's guidelines and best practices
- **Microsoft Research**: Code survival and churn analysis methodology
- **Industry Benchmarks**: Bus factor, ownership patterns, and quality metrics
- **Academic Citations**: Proper research citations and threshold justification

### Documentation
- **User Guide**: Comprehensive usage guide with examples and best practices
- **Command Reference**: Complete command documentation with flags and options
- **Research Methodology**: Detailed research basis and threshold methodology
- **Contributing Guidelines**: Development setup and contribution process
- **Installation Instructions**: Ready for Homebrew integration

## [0.1.0] - 2025-01-XX

### Added
- Initial project structure and basic CLI framework
- Core Git analysis capabilities
- Basic churn and survival analysis
- Project documentation foundation

---

## Version History

- **v1.1.0**: Multiple path support and configuration file system
- **v1.0.0**: Complete implementation of 14/15 research-backed metrics
- **v0.1.0**: Initial project setup and basic functionality

## Future Roadmap

### Planned Features
- **Review Bottlenecks Analysis**: PR analysis (requires GitHub API integration) - Metric #13
- **Advanced Visualization**: Chart and graph generation
- **CI/CD Integration**: Automated analysis and reporting
- **Team Dashboards**: Web-based team performance dashboards
- **Custom Metrics**: User-defined metric creation
- **Export Capabilities**: JSON, CSV, and other format exports

### Homebrew Integration
- Formula creation and submission
- Automated releases and versioning
- Binary distribution and installation
- Community adoption and feedback

---

*For detailed information about each release, see the [GitHub Releases](https://github.com/bgricker/gitallica/releases) page.*
