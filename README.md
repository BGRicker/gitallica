# 🎸 gitallica  
*Shred your git history. Rock your repo insights.*  

---

## ⚡ Ride the Lightning by performing temporal diff analysis of distributed version control logs.  

---

## � Usage

Here are some examples of how to use the **gitallica** CLI to analyze churn:

```bash
gitallica churn
```
Shows churn across the entire history.

```bash
gitallica churn --last 6m
```
Shows churn for the last 6 months.

```bash
gitallica churn --path cmd/
```
Shows churn scoped to the `cmd/` directory.

The `--last` argument accepts a value in the format `#{number}{unit}`, where unit can be:
- `d` for days
- `m` for months
- `y` for years

For example, `--last 30d` means the last 30 days, `--last 1y` means the last year.

You can combine multiple flags, like:
```bash
gitallica churn --last 1y --path internal/
```
to show churn in the `internal/` directory over the last year.

---

Here are some examples of how to use the **gitallica** CLI to analyze code survival:

```bash
gitallica survival
```
Shows code survival rate across the entire history.

```bash
gitallica survival --last 3m
```
Shows code survival rate for the last 3 months.

```bash
gitallica survival --path src/
```
Shows code survival rate scoped to the `src/` directory.

```bash
gitallica survival --last 6m --path lib/ --debug
```
Shows code survival rate for the last 6 months in the `lib/` directory with debug output enabled.

**Available flags:**

- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--debug` : Enable debug output for more detailed logging during analysis.

---

Here are some examples of how to use the **gitallica** CLI to analyze high-churn files and directories:

```bash
gitallica churn-files
```
Shows files and directories with high churn rates.

```bash
gitallica churn-files --last 3m --limit 5
```
Shows top 5 high-churn files/directories in the last 3 months.

```bash
gitallica churn-files --path src/ --directories
```
Shows high-churn analysis for the `src/` directory, including directory-level statistics.

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--limit` : Number of top results to show (default 10).
- `--directories` : Also show directory-level churn statistics.

---

Here are some examples of how to use the **gitallica** CLI to analyze commit sizes:

```bash
gitallica commit-size
```
Shows commits by risk level based on size and file count.

```bash
gitallica commit-size --min-risk High --limit 5
```
Shows top 5 high-risk commits.

```bash
gitallica commit-size --last 30d --min-risk Medium
```
Shows medium and high-risk commits from the last 30 days.

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--min-risk` : Filter by minimum risk level (Low, Medium, High, Critical).
- `--limit` : Number of top results to show (default 10).

---

Here are some examples of how to use the **gitallica** CLI to analyze new component creation:

```bash
gitallica component-creation
```
Analyzes the rate of new component creation across different frameworks.

```bash
gitallica component-creation --framework javascript --last 30d
```
Shows component creation rate for JavaScript/TypeScript components in the last 30 days.

```bash
gitallica component-creation --framework ruby --limit 5
```
Shows top 5 Ruby component types by creation count.

**Supported frameworks:**
- `javascript` : JavaScript/TypeScript classes and React components
- `ruby` : Rails models, controllers, and service objects
- `python` : Python classes and modules
- `go` : Go structs and interfaces
- `java` : Java classes and interfaces
- `csharp` : C# classes and interfaces

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--framework` : Filter by specific framework (javascript, ruby, python, go, java, csharp).
- `--limit` : Number of top results to show (default 10).

---

Here are some examples of how to use the **gitallica** CLI to analyze directory entropy:

```bash
gitallica directory-entropy
```
Analyzes entropy across repository directories to identify areas with weak modularity. Automatically detects project type (Go, Node.js, Python, Ruby, etc.) for context-aware analysis.

```bash
gitallica directory-entropy --limit 5
```
Shows top 5 high and low entropy directories.

```bash
gitallica directory-entropy --last 30d
```
Shows directory entropy analysis for the last 30 days.

**Context-Aware Analysis:**
- **Project Type Detection**: Automatically identifies Go CLI, Node.js, Python, Ruby/Rails, and generic projects
- **Root Directory Rules**: Different entropy thresholds for root vs. subdirectories
- **Framework-Specific Patterns**: Understands expected file types for each project structure

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--limit` : Number of top results to show (default 10).

---

Here are some examples of how to use the **gitallica** CLI to analyze dead zones:

```bash
gitallica dead-zones
```
Identifies files untouched for ≥12 months that may represent technical debt.

```bash
gitallica dead-zones --limit 5
```
Shows top 5 oldest dead zone files.

```bash
gitallica dead-zones --last 6m --path src/
```
Shows dead zones in the `src/` directory, looking at the last 6 months of activity.

**Note**: Dead-zone age is based on time since last modification.
The `--last` window controls the activity period scanned to determine staleness,
which may surface files untouched for >6 months (e.g., 12+ months).

**Risk Classification:**
- **Low Risk**: 12-17 months untouched (consider reviewing)
- **Medium Risk**: 18-23 months untouched (needs attention)  
- **High Risk**: 24-35 months untouched (refactor or remove)
- **Critical**: 36+ months untouched (urgent: refactor or delete)

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--limit` : Number of top results to show (default 10).

---

Here are some examples of how to use the **gitallica** CLI to analyze bus factor:

```bash
gitallica bus-factor
```
Analyzes bus factor (knowledge concentration) across all repository directories using Git blame data for accurate line-level authorship analysis.

```bash
gitallica bus-factor --limit 5
```
Shows top 5 directories with highest bus factor risk.

```bash
gitallica bus-factor --last 6m --path src/
```
Shows bus factor analysis for the `src/` directory over the last 6 months.

**Intelligent Analysis:**
- **File-Level Ownership**: Uses efficient commit-based analysis to determine file authorship
- **Performance Optimized**: Fast analysis suitable for large repositories (O(n×c) complexity)
- **Accurate Knowledge Measurement**: Avoids misleading metrics from drive-by commits while maintaining speed
- **Research-Backed Thresholds**: Based on Martin Fowler's collective ownership principles

**Risk Classification:**
- **Critical**: Bus factor 1 (single point of failure)
- **High**: Bus factor 2-3 in larger teams (knowledge concentration)
- **Medium**: Bus factor adequate but could be improved
- **Healthy**: Good knowledge distribution (25-50% of team)

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--limit` : Number of top results to show (default 10).

---

Here are some examples of how to use the **gitallica** CLI to analyze test-to-code ratio:

```bash
gitallica test-ratio
```
Analyzes the ratio of test code to source code across the entire repository.

```bash
gitallica test-ratio --path src/
```
Analyzes test-to-code ratio for files in the `src/` directory only.

**Research-Based Analysis:**
- **Clean Code Principles**: Based on Robert C. Martin's Clean Code guidelines for test coverage
- **Comprehensive Detection**: Identifies test files across multiple languages and frameworks
- **Actionable Insights**: Provides specific recommendations and line count targets
- **Cross-Language Support**: Supports Go, JavaScript/TypeScript, Python, Ruby, Java, C#, and more

**Ratio Classifications:**
- **Excellent**: 1:1 to 2:1 ratio (ideal coverage with reasonable overhead)
- **Healthy**: 1:1 ratio (balanced test and source code)
- **Caution**: 0.75:1 to 1:1 ratio (adequate but could be improved)
- **Warning**: 0.5:1 to 0.75:1 ratio (insufficient test coverage)
- **Critical**: <0.5:1 ratio or no tests (urgent attention needed)

**Available flags:**
- `--path` : Limit analysis to a specific directory or path within the repository.

---

Here are some examples of how to use the **gitallica** CLI to analyze ownership clarity:

```bash
gitallica ownership-clarity
```
Analyzes ownership clarity across all repository files to identify unclear ownership patterns.

```bash
gitallica ownership-clarity --limit 5
```
Shows top 5 files with ownership clarity issues.

```bash
gitallica ownership-clarity --last 6m --path src/
```
Analyzes ownership clarity for the `src/` directory over the last 6 months.

**Intelligent Analysis:**
- **Balanced Ownership Detection**: Identifies both overly concentrated and diffuse ownership patterns
- **Team Size Awareness**: Applies different thresholds for small vs. large contributor teams
- **Research-Based Classification**: Based on Martin Fowler's collective ownership principles
- **Actionable Recommendations**: Provides specific guidance for improving ownership balance

**Ownership Classifications:**
- **Healthy**: Good ownership balance (40-80% primary ownership)
- **Caution**: Too concentrated (>80% single owner in teams >3)
- **Warning**: Diffuse ownership without clear primary maintainer
- **Critical**: Extremely diffuse ownership in large contributor base

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--limit` : Number of files to show in detailed analysis (default 10).

---

Here are some examples of how to use the **gitallica** CLI to analyze onboarding footprint:

```bash
gitallica onboarding-footprint
```
Analyzes what files new contributors touch first to understand onboarding complexity.

```bash
gitallica onboarding-footprint --limit 5 --commit-limit 3
```
Shows top 5 recent contributors analyzing their first 3 commits.

```bash
gitallica onboarding-footprint --last 6m --path src/
```
Analyzes onboarding patterns for the `src/` directory over the last 6 months.

**Research-Based Analysis:**
- **Clean Code Principles**: Based on Robert C. Martin's principles about code readability and developer experience
- **New Contributor Tracking**: Identifies first-time contributors and analyzes their early commits
- **Onboarding Complexity Assessment**: Measures files touched in first commits to gauge entry complexity
- **Entry Point Identification**: Discovers common files that new developers modify first

**Complexity Classifications:**
- **Simple**: 1-5 files (excellent focused onboarding)
- **Moderate**: 6-10 files (reasonable complexity)
- **Complex**: 11-20 files (consider simplifying)
- **Overwhelming**: >20 files (urgent improvement needed)

**Insights Provided:**
- **Average files touched** by new contributors in early commits
- **Distribution of onboarding complexity** across contributors
- **Most common entry point files** that new developers touch
- **Time-based analysis** of onboarding patterns
- **Actionable recommendations** for improving developer onboarding

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--limit` : Number of contributors to show in detailed analysis (default 10).
- `--commit-limit` : Number of initial commits to analyze per contributor (default 5).

---

Here are some examples of how to use the **gitallica** CLI to analyze high-risk commits:

```bash
gitallica high-risk-commits
```
Analyzes all commits to identify those with high risk due to size or complexity.

```bash
gitallica high-risk-commits --limit 5
```
Shows top 5 risky commits with detailed analysis.

```bash
gitallica high-risk-commits --last 30d --path src/
```
Analyzes high-risk commits in the `src/` directory over the last 30 days.

**Research-Based Analysis:**
- **Line Change Detection**: Identifies commits with >400 lines changed (high risk) or >800 lines (critical risk)
- **File Complexity Tracking**: Flags commits touching >12 files (high risk) or >20 files (critical risk)  
- **Review Impact Assessment**: Based on Kent Beck and Martin Fowler's principles of small, incremental changes
- **Rollback Risk Evaluation**: Large commits reduce review effectiveness and increase rollback complexity

**Risk Classifications:**
- **Low**: Small, focused commits (≤400 lines, ≤12 files)
- **Moderate**: Sizeable commits that warrant attention
- **High**: Large commits requiring careful review (>400 lines OR ≥12 files)
- **Critical**: Monster commits with extremely high risk (≥800 lines OR ≥20 files)

**Insights Provided:**
- **Risk distribution** across all analyzed commits
- **Average commit size** in lines and files changed
- **Largest commit identification** with detailed metrics
- **Risky commit details** with authors, dates, and change scope
- **Actionable recommendations** for improving commit hygiene

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--limit` : Number of risky commits to show in detailed output (default 10).

---

Here are some examples of how to use the **gitallica** CLI to analyze commit cadence trends:

```bash
gitallica commit-cadence
```
Analyzes commit frequency patterns over time using weekly grouping to identify trends and sustainability.

```bash
gitallica commit-cadence --period day --last 30d
```
Shows daily commit patterns for the last 30 days to identify recent pace changes.

```bash
gitallica commit-cadence --period month --path src/
```
Analyzes monthly commit trends for the `src/` directory to understand long-term patterns.

**Research-Based Analysis:**
- **Sustainable Pace Principles**: Based on Kent Beck's Extreme Programming principles about sustainable development
- **Trend Detection**: Uses linear regression to identify increasing, decreasing, or stable commit patterns
- **Spike/Dip Identification**: Detects periods of unusually high or low activity that may indicate crunch or stagnation
- **Team Health Assessment**: Evaluates overall sustainability based on pace volatility and trends

**Sustainability Classifications:**
- **Healthy**: Sustainable pace with minimal volatility (5-25 commits/period, stable trends)
- **Caution**: Some concerning patterns detected (decreasing trends, occasional dips)
- **Warning**: Multiple spikes indicating potential crunch periods
- **Critical**: High volatility with frequent spikes and dips indicating unsustainable practices

**Insights Provided:**
- **Overall trend direction** with statistical strength measurement
- **Commit spikes** that may indicate crunch periods or deadline pressure
- **Commit dips** that could signal blockers, burnout, or reduced team capacity
- **Sustainability assessment** based on pace and volatility patterns
- **Recent period analysis** showing latest development patterns
- **Actionable recommendations** for improving team pace and sustainability

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Scope the analysis to a specific directory or path within the repository.
- `--period` : Time period for grouping commits (day, week, month) - default is "week".

---

Here are some examples of how to use the **gitallica** CLI to analyze long-lived branches:

```bash
gitallica long-lived-branches
```
Analyzes all branches to identify trunk-based development compliance and long-lived branch risks.

```bash
gitallica long-lived-branches --last 30d --limit 5
```
Shows the 5 most risky branches from the last 30 days of development.

```bash
gitallica long-lived-branches --path src/ --show-merged
```
Analyzes branches affecting the `src/` directory, including recently merged branches.

**Research-Based Analysis:**
- **DORA Principles**: Based on Accelerate research showing elite teams merge frequently and keep branches short-lived
- **Trunk-Based Development**: Emphasizes small, frequent integrations to reduce deployment risk
- **Integration Risk Assessment**: Long-lived branches increase merge conflicts and delivery delays
- **Team Velocity Impact**: Branch age correlation with deployment frequency and team performance

**Risk Classifications:**
- **Healthy**: Branches ≤2 days old (aligned with elite team practices)
- **Warning**: Branches 3-7 days old (monitor for integration opportunities)
- **Risky**: Branches 8-21 days old (significant integration risk)
- **Critical**: Branches >21 days old (high risk of conflicts and delivery delays)

**Trunk-Based Compliance Levels:**
- **Excellent**: 80%+ branches are healthy (≤2 days)
- **Good**: 60-80% branches are healthy
- **Moderate**: 40-60% branches are healthy
- **Poor**: 20-40% branches are healthy
- **Critical**: <20% branches are healthy (high integration risk)

**Insights Provided:**
- **Branch age distribution** across different risk categories
- **Trunk-based development compliance** assessment with specific percentage
- **Oldest branch identification** with author and last commit details
- **Risky branch details** showing branches requiring immediate attention
- **Team integration health** based on branch lifecycle patterns
- **Actionable recommendations** for improving development practices

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Limit analysis to branches affecting a specific directory or path within the repository.
- `--limit` : Number of risky branches to show in detailed output (default 10).
- `--show-merged` : Include recently merged branches in analysis.

---

Here are some examples of how to use the **gitallica** CLI to analyze change lead time:

```bash
gitallica change-lead-time
```
Analyzes change lead time using DORA benchmarks to measure delivery performance.

```bash
gitallica change-lead-time --last 90d --limit 10
```
Shows lead time analysis for the last 90 days with top 10 slowest/fastest commits.

```bash
gitallica change-lead-time --path src/ --method tag
```
Analyzes lead time for the `src/` directory using release tags as deployment markers.

**Research-Based Analysis:**
- **DORA Metrics Foundation**: Based on State of DevOps research and Accelerate book findings
- **Delivery Performance Indicator**: Measures organizational capability and flow efficiency
- **Business Correlation**: Strong predictor of software delivery performance and business outcomes
- **Elite Team Benchmarking**: Enables comparison against industry-leading practices

**DORA Classification Levels:**
- **Elite**: <1 day lead time (industry-leading performance)
- **High**: 1 day to 1 week (strong performance)
- **Medium**: 1 week to 1 month (moderate performance requiring optimization)
- **Low**: >1 month (significant improvement needed)

**Lead Time Calculation Methods:**
- **Merge Method**: Time from commit creation to appearance in main branch
- **Tag Method**: Time from commit creation to inclusion in release tag
- **Configurable Approach**: Adaptable to different deployment and release strategies

**Performance Level Assessment:**
- **Elite Performance**: 70%+ commits achieve elite lead time (<1 day)
- **High Performance**: 60%+ commits achieve elite or high lead time
- **Medium Performance**: 50%+ commits achieve elite, high, or medium lead time
- **Low Performance**: <50% commits achieve acceptable lead time

**Insights Provided:**
- **DORA performance distribution** across all classification levels
- **Statistical analysis** including average, median, and 95th percentile lead times
- **Fastest and slowest commits** identification with detailed attribution
- **Performance benchmarking** against DORA research standards
- **Delivery bottleneck identification** through outlier analysis
- **Actionable recommendations** based on current performance level

**Available flags:**
- `--last` : Specify the time window to analyze, in the format `#{number}{unit}` (e.g., `30d`, `6m`, `1y`).
- `--path` : Limit analysis to commits affecting a specific directory or path within the repository.
- `--limit` : Number of slowest/fastest commits to show in detailed output (default 5).
- `--method` : Lead time calculation method - 'merge' (commit to main) or 'tag' (commit to release tag).

---

## Guiding Metrics & Research-Based Benchmarks  

Here are the **15 greatest hits**—each paired with rationale and a relevant quote from respected authors.  

---

### 1. **Additions vs. Deletions Ratio**  
*Is your repo getting leaner, or just bloated?*  

**Threshold:** Keep churn (added + deleted lines) under ~20% of codebase size.  

**Why:**  
> “If we wish to count lines of code, we should not regard them as ‘lines produced’ but as ‘lines spent.’” — *Edsger W. Dijkstra*  

Ken Thompson added: “One of my most productive days was throwing away 1,000 lines of code.”  

Churn is natural, but if more than ~20% of the codebase is rewritten in a short period, it often signals instability.  

#### How We Calculate Churn  

In gitallica, churn is calculated as the ratio of the sum of added and deleted lines relative to the total lines of code (LOC) in the codebase, expressed by the formula:  

`Churn = (Additions + Deletions) ÷ Total LOC`  

This approach aligns with definitions and methodologies used by Microsoft Research, Adam Tornhill's CodeScene, and other industry benchmarks. Measuring churn relative to the total codebase size rather than just raw additions or deletions provides a normalized view of code volatility. Established engineering research supports thresholds around 15–20% churn as indicators of potential instability or areas needing attention.  

---

### 2. **Code Survival Rate**  
*Do your changes stand the test of time?*  

**Threshold:** Investigate areas where <50% of lines survive 6–12 months.  

**Why:**  
A large-scale study of 3.3 billion code-element lifetimes found a median lifespan of ~2.4 years. Modules with unusually short lifespans often indicate unstable design or wasted effort.  

---

### 3. **High-Churn Files & Directories**  
*Which parts of your repo just won’t stay quiet?*  

**Threshold:** File churn >20% in a timeframe flags instability.  

**Why:**  
> “Refactoring changes the program in small steps. If you make a mistake, it is easy to find the bug.” — *Martin Fowler, Refactoring*  

Files with high churn are refactored (or hacked) repeatedly; they deserve architectural attention.  

---

### 4. **New Component Creation Rate**  
*Are you growing at the right pace, or sprawling out of control?*  

**Threshold:** Monitor for sudden spikes or persistent growth beyond baseline.  

**Why:**  
> “The design should have the fewest possible classes and methods.” — *Kent Beck’s rules of simple design*  

A spike in new models or services can indicate architectural sprawl.  

---

### 5. **Directory Entropy**  
*When clean albums turn into messy mixtapes.*  

**Threshold:** Context-aware analysis with different rules for root vs. subdirectories.  

**Why:**  
> "Simplicity is prerequisite for reliability." — *Edsger W. Dijkstra*  

High entropy signals weak modularity and eroded boundaries. **gitallica** automatically detects project type (Go, Node.js, Python, Ruby, etc.) and applies appropriate entropy thresholds based on expected directory structures and file type patterns.  

---

### 6. **Dead Zones**  
*Code that time forgot.*  

**Threshold:** Files untouched for ≥12 months that remain active.  

**Why:**  
> “The only way to go fast is to keep your code as clean as possible at all times.” — *Robert C. Martin, Clean Code*  

Untouched code becomes a liability; better to refactor, revive, or delete.  

---

### 7. **Bus Factor (per directory)**  
*Who's the last person standing if someone leaves?*

**Threshold:** Target bus factor of ~25–50% of team (e.g., 4–5 in a 10-person team).  

**Why:**  
Collective ownership is healthier than strong ownership:  
> "With collective ownership, anyone can change any part of the code at any time." — *Martin Fowler*  

But diffuse ownership without clear stewardship risks accountability gaps.

**Intelligent Analysis:** Uses efficient file-level authorship analysis with commit-based traversal, providing accurate knowledge concentration metrics while maintaining excellent performance for large repositories.

---

### 8. **Ownership Clarity**  
*Does anyone really own this code?*  

**Threshold:** Flag files with no contributor ≥40–50% of commits (when >10 contributors exist).  

**Why:** Diffuse ownership correlates with inconsistent quality; concentrated ownership risks bottlenecks. Balance matters.  

---

### 9. **Onboarding Footprint**  
*What new devs touch first.*  

**Threshold:** New contributors touching >10–20 files in first 5 commits may signal steep onboarding.  

**Why:**  
> “Developers spend much more time reading code than writing it, so making it easy to read makes it easier to write.” — *Robert C. Martin, Clean Code*  

Scoped, small first tasks help new developers succeed.  

---

### 10. **Code vs. Test Ratio**  
*Are you rehearsing as much as you’re performing?*  

**Threshold:** Test-to-code ratio ≈1:1, up to 2:1. Anything below 1:1 suggests lagging tests.  

**Why:**  
> “Test code is just as important as production code.” — *Robert C. Martin, Clean Code*  

A healthy ratio shows you’re rehearsing before the big show.  

---

### 11. **High-Risk Commits**  
*The monster commits that touch everything.*  

**Threshold:** >400 lines or >10–12 files per commit is high risk.  

**Why:**  
> “All changes are small. There are only longer and shorter feedback cycles.” — *Kent Beck*  
> “Refactoring changes the program in small steps.” — *Martin Fowler*  

Large commits reduce review effectiveness and rollback safety.  

---

### 12. **Commit Cadence Trends**  
*Is your repo speeding up or slowing down?*  

**Threshold:** Track trends, not absolutes. Spikes or dips may reveal crunch, burnout, or stagnation.  

**Why:**  
> “Overtime is a symptom of a serious problem… you can’t work a second week of overtime.” — *Kent Beck, Extreme Programming Explained*  

A sustainable pace is better than bursts of frenzy.  

---

### 13. **Review Bottlenecks**  
*Where pull requests go to die.*  

**Threshold:** PRs taking >3–5 days should be investigated.  

**Why:**  
Studies show reviews are most effective under 400 lines and under an hour. Longer reviews lose effectiveness.  

---

### 14. **Long-Lived Branches**  
*That one track nobody finishes.*  

**Threshold:** Branches older than a few days are risky. Trunk-based development recommends daily merges.  

**Why:**  
Accelerate/DORA research shows elite teams merge frequently and keep branches short-lived.  

---

### 15. **Change Lead Time**  
*How long does it take to ship?*  

**Threshold:**  
- Elite: <1 day  
- High: 1 day–1 week  
- Medium: 1 week–1 month  
- Low: >1 month  

**Why:**  
These are the DORA benchmarks for lead time. Lead time reflects delivery health.  

---

## How These Are Used in Development  

- **Scaling by team size:** Metrics like bus factor scale with org size.  
- **Configurable thresholds:** Override defaults via config or CLI flags.  
- **Tracing thresholds:** Each metric cites rationale from authoritative sources.  
- **Performance-conscious design:** Heavy metrics run lazily or on demand.  
- **Context-aware analysis:** Automatically detects project type and applies appropriate thresholds for accurate insights.  

---

## Why Use gitallica?  

Other tools give you vanity metrics like stars and forks. **gitallica** digs deeper—into how your codebase evolves, where risks lie, and how your team really works.  

**Intelligent Analysis:** Context-aware metrics that understand your project type (Go, Node.js, Python, Ruby, etc.) and apply appropriate thresholds—no more false positives from standard project structures.

- **Managers** → Spot architectural rot and churn early.  
- **Tech leads** → See review friction and risk hotspots.  
- **Developers** → Understand fragile areas and onboarding hurdles.  

⚡ *Don't just play your repo. Rock it—consciously.*  
