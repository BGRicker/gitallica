# Research Methodology

Gitallica implements 14 research-backed metrics based on authoritative sources from software engineering research, industry studies, and established best practices.

## Research Foundation

### Primary Sources

- **DORA (DevOps Research and Assessment)**: State of DevOps research and Accelerate book
- **Microsoft Research (MSR)**: Large-scale code analysis studies
- **Clean Code Principles**: Robert C. Martin's guidelines
- **Extreme Programming**: Kent Beck's methodologies
- **Refactoring**: Martin Fowler's principles
- **CodeScene**: Adam Tornhill's code analysis research

### Research Categories

1. **Code Evolution Metrics** (3 metrics)
2. **Team Performance Metrics** (3 metrics)  
3. **Quality Metrics** (3 metrics)
4. **Delivery Performance Metrics** (3 metrics)
5. **Architecture Metrics** (2 metrics)

**Note**: Review Bottlenecks (#13) requires GitHub API integration and is planned for future implementation.

## Detailed Methodology

### 1. Code Evolution Metrics

#### Additions vs. Deletions Ratio (Churn)
**Research Basis**: Microsoft Research, Edsger Dijkstra
**Formula**: `Churn = (Additions + Deletions) ÷ Total LOC`
**Threshold**: Keep churn under ~20% of codebase size
**Rationale**: 
> "If we wish to count lines of code, we should not regard them as 'lines produced' but as 'lines spent.'" — Edsger Dijkstra

High churn indicates instability and potential architectural issues.

#### Code Survival Rate
**Research Basis**: MSR, CodeScene research
**Methodology**: Track lines from addition to deletion/modification
**Threshold**: Investigate areas where <50% of lines survive 6-12 months
**Rationale**: 
> "A large-scale study of 3.3 billion code-element lifetimes found a median lifespan of ~2.4 years."

Short lifespans indicate unstable design or wasted effort.

#### High-Churn Files & Directories
**Research Basis**: Martin Fowler's Refactoring principles
**Methodology**: Track file modification frequency over time
**Threshold**: File churn >20% in a timeframe flags instability
**Rationale**:
> "Refactoring changes the program in small steps. If you make a mistake, it is easy to find the bug." — Martin Fowler

High-churn files deserve architectural attention.

### 2. Team Performance Metrics

#### Bus Factor
**Research Basis**: Martin Fowler's collective ownership principles
**Methodology**: Analyze commit authorship per directory/file
**Threshold**: Target bus factor of ~25-50% of team (e.g., 4-5 in a 10-person team)
**Rationale**:
> "With collective ownership, anyone can change any part of the code at any time." — Martin Fowler

Balances collective ownership with clear stewardship.

#### Ownership Clarity
**Research Basis**: Industry research on ownership patterns
**Methodology**: Analyze contributor distribution across files
**Threshold**: Flag files with no contributor ≥40-50% of commits (when >10 contributors exist)
**Rationale**: Diffuse ownership correlates with inconsistent quality; concentrated ownership risks bottlenecks.

#### Onboarding Footprint
**Research Basis**: Robert C. Martin's Clean Code principles
**Methodology**: Track new contributor file touch patterns
**Threshold**: New contributors touching >10-20 files in first 5 commits may signal steep onboarding
**Rationale**:
> "Developers spend much more time reading code than writing it, so making it easy to read makes it easier to write." — Robert C. Martin

Scoped, small first tasks help new developers succeed.

### 3. Quality Metrics

#### Code vs. Test Ratio
**Research Basis**: Robert C. Martin's Clean Code
**Methodology**: Analyze test file vs. source file ratios
**Threshold**: Test-to-code ratio ≈1:1, up to 2:1. Anything below 1:1 suggests lagging tests
**Rationale**:
> "Test code is just as important as production code." — Robert C. Martin

Healthy ratio shows rehearsal before performance.

#### High-Risk Commits
**Research Basis**: Kent Beck, Martin Fowler
**Methodology**: Analyze commit size (lines changed, files touched)
**Threshold**: >400 lines or >10-12 files per commit is high risk
**Rationale**:
> "All changes are small. There are only longer and shorter feedback cycles." — Kent Beck
> "Refactoring changes the program in small steps." — Martin Fowler

Large commits reduce review effectiveness and rollback safety.

#### Commit Size Analysis
**Research Basis**: Industry research on review effectiveness
**Methodology**: Categorize commits by size and complexity
**Thresholds**:
- Low: ≤100 lines, ≤5 files
- Medium: 100-400 lines, 5-10 files
- High: 400-800 lines, 10-15 files
- Critical: >800 lines, >15 files
**Rationale**: Reviews are most effective under 400 lines.

### 4. Delivery Performance Metrics

#### Change Lead Time
**Research Basis**: DORA State of DevOps, Accelerate research
**Methodology**: Measure time from commit to deployment
**Thresholds**:
- Elite: <1 day
- High: 1 day-1 week
- Medium: 1 week-1 month
- Low: >1 month
**Rationale**: These are the DORA benchmarks for lead time. Lead time reflects delivery health.

#### Commit Cadence Trends
**Research Basis**: Kent Beck's Extreme Programming principles
**Methodology**: Analyze commit frequency patterns over time
**Threshold**: Track trends, not absolutes. Spikes or dips may reveal crunch, burnout, or stagnation
**Rationale**:
> "Overtime is a symptom of a serious problem… you can't work a second week of overtime." — Kent Beck

Sustainable pace is better than bursts of frenzy.

#### Long-Lived Branches
**Research Basis**: DORA research, trunk-based development
**Methodology**: Analyze branch lifecycle and merge patterns
**Threshold**: Branches older than a few days are risky. Trunk-based development recommends daily merges
**Rationale**: Accelerate/DORA research shows elite teams merge frequently and keep branches short-lived.

### 5. Architecture Metrics

#### New Component Creation Rate
**Research Basis**: Industry benchmarks for component growth
**Methodology**: Track creation of new components and modules
**Threshold**: Context-aware analysis based on project type
**Rationale**: Rapid component creation may indicate architectural instability or growth.

#### Directory Entropy
**Research Basis**: Edsger Dijkstra's simplicity principles
**Methodology**: Measure directory organization and modularity
**Threshold**: Context-aware analysis with different rules for root vs. subdirectories
**Rationale**:
> "Simplicity is prerequisite for reliability." — Edsger Dijkstra

High entropy signals weak modularity and eroded boundaries.

#### Dead Zones
**Research Basis**: Robert C. Martin's Clean Code
**Methodology**: Identify files untouched for extended periods
**Threshold**: Files untouched for ≥12 months that remain active
**Rationale**:
> "The only way to go fast is to keep your code as clean as possible at all times." — Robert C. Martin

Untouched code becomes a liability; better to refactor, revive, or delete.

## Threshold Methodology

### Performance Level Classification

#### DORA Performance Levels
Based on Accelerate research and State of DevOps studies:

- **Elite Performance**: 70%+ commits achieve elite thresholds
- **High Performance**: 60%+ commits achieve elite or high thresholds
- **Medium Performance**: 50%+ commits achieve acceptable thresholds
- **Low Performance**: <50% commits achieve acceptable thresholds

#### Risk Classification
Based on industry research and best practices:

- **Low Risk**: Within recommended thresholds
- **Medium Risk**: Approaching threshold limits
- **High Risk**: Exceeding thresholds, requiring attention
- **Critical Risk**: Significantly exceeding thresholds, immediate action needed

### Statistical Analysis

#### Percentile Analysis
- **50th Percentile (Median)**: Middle value, less affected by outliers
- **95th Percentile**: Captures extreme cases and outliers
- **Distribution Analysis**: Breakdown across performance categories

#### Trend Analysis
- **Linear Regression**: Identifies trends over time
- **Spike/Dip Detection**: Identifies unusual patterns
- **Sustainability Assessment**: Evaluates long-term viability

## Validation and Calibration

### Industry Benchmarks
Thresholds are calibrated against:
- DORA State of DevOps research
- Microsoft Research studies
- Industry best practices
- Open source project analysis

### Context Awareness
Gitallica automatically detects project characteristics:
- Programming language and ecosystem
- Project size and complexity
- Team size and structure
- Development methodology

### Configurable Thresholds
All thresholds can be customized via:
- Configuration files
- Command-line flags
- Environment variables

## Research Citations

### Primary Sources
1. Forsgren, N., Humble, J., & Kim, G. (2018). *Accelerate: The Science of Lean Software and DevOps*. IT Revolution Press.
2. Martin, R. C. (2008). *Clean Code: A Handbook of Agile Software Craftsmanship*. Prentice Hall.
3. Fowler, M. (2018). *Refactoring: Improving the Design of Existing Code*. Addison-Wesley.
4. Beck, K. (2004). *Extreme Programming Explained: Embrace Change*. Addison-Wesley.
5. Dijkstra, E. W. (1972). "The Humble Programmer." *Communications of the ACM*.

### Research Studies
1. Microsoft Research. "A Large-Scale Study of Code Survival." (2019)
2. CodeScene. "Code Survival Analysis." Adam Tornhill (2020)
3. DORA. "State of DevOps Report." (2019-2023)
4. GitHub. "The State of the Octoverse." (2023)

### Industry Standards
1. DORA Metrics: Lead Time, Deployment Frequency, Change Failure Rate, Time to Recovery
2. Clean Code Principles: Test ratios, commit size, code organization
3. Extreme Programming: Sustainable pace, collective ownership
4. Trunk-Based Development: Branch lifecycle, merge frequency

## Methodology Validation

### Peer Review
- Research methodology reviewed by software engineering experts
- Thresholds validated against industry benchmarks
- Implementation verified through extensive testing

### Continuous Improvement
- Regular updates based on new research
- Community feedback integration
- Performance monitoring and optimization

### Open Source Validation
- Methodology open for community review
- Research citations provided for all thresholds
- Transparent implementation of all algorithms
