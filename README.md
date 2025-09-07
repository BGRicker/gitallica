# 🎸 gitallica  
*Shred your git history. Rock your repo insights.*  

---

## ⚡ Ride the Lightning by performing temporal diff analysis of distributed version control logs.  

---

Most teams think they know their codebase. But under the surface, there are subtle patterns of churn, ownership concentration, and architectural sprawl that are easy to miss. **gitallica** helps you crank the amp to 11 and hear the *real story* of your repo—backed by research-informed thresholds and best practices.

By mining your git history, gitallica transforms raw commits into insights that help teams stay tight, efficient, and in rhythm. This README includes our guiding thresholds and their research citations.

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

**Threshold:** Compare entropy across directories; flag outliers relative to team norms.  

**Why:**  
> “Simplicity is prerequisite for reliability.” — *Edsger W. Dijkstra*  

High entropy signals weak modularity and eroded boundaries.  

---

### 6. **Dead Zones**  
*Code that time forgot.*  

**Threshold:** Files untouched for ≥12 months that remain active.  

**Why:**  
> “The only way to go fast is to keep your code as clean as possible at all times.” — *Robert C. Martin, Clean Code*  

Untouched code becomes a liability; better to refactor, revive, or delete.  

---

### 7. **Bus Factor (per directory)**  
*Who’s the last person standing if someone leaves?*  

**Threshold:** Target bus factor of ~25–50% of team (e.g., 4–5 in a 10-person team).  

**Why:**  
Collective ownership is healthier than strong ownership:  
> “With collective ownership, anyone can change any part of the code at any time.” — *Martin Fowler*  

But diffuse ownership without clear stewardship risks accountability gaps.  

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

---

## Why Use gitallica?  

Other tools give you vanity metrics like stars and forks. **gitallica** digs deeper—into how your codebase evolves, where risks lie, and how your team really works.  

- **Managers** → Spot architectural rot and churn early.  
- **Tech leads** → See review friction and risk hotspots.  
- **Developers** → Understand fragile areas and onboarding hurdles.  

⚡ *Don’t just play your repo. Rock it—consciously.*  
