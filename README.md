# 🎸 gitallica  
*Shred your git history. Rock your repo insights.*  

---

## ⚡ Ride the Lightning by performing temporal diff analysis of distributed version control logs.  

---

Most teams think they know their codebase. But under the surface, there’s chaos, churn, and a few solos that go on way too long. **gitallica** helps you crank the amp to 11 and hear the *real story* of your repo.  

By mining your git history, gitallica transforms raw commits into insights that help teams stay tight, efficient, and in rhythm.  

---

## 📦 Installation (coming soon)  
```bash
brew install gitallica
```

---

## 🔍 What gitallica Tracks  

Here are the **15 greatest hits** — the metrics every engineering team needs to keep their repo from turning into a gitwreck.  

---

### 1. **Additions vs. Deletions Ratio**  
*Is your repo getting leaner, or just bloated?*  

Tracks lines of code added vs. removed. A healthy ratio suggests ongoing refactoring and thoughtful evolution of the codebase. A persistently high additions-to-deletions ratio indicates unchecked growth, complexity creep, and potential maintainability issues.  

```bash
$ gitallica churn
Additions vs Deletions (last 90 days):
- Additions: 52,134 lines
- Deletions: 14,876 lines
- Ratio: 3.5 : 1  (bloated — time for a remix)
```

---

### 2. **Code Survival Rate**  
*Do your changes stand the test of time?*  

Calculates how much code written in a given period survives after 6–12 months. High survival suggests architectural stability and durable work. Low survival rates point to rework, shifting priorities, or fragile implementations.  

```bash
$ gitallica survival --window 12m
Code Survival (12 months):
- Lines added: 102,487
- Lines still present: 73,905 (72%)
- Churned away: 28% (unstable areas in /lib)
```

---

### 3. **High-Churn Files & Directories**  
*Which parts of your repo just won’t stay quiet?*  

Identifies the most frequently modified files and directories. High churn often signals unclear ownership, weak abstractions, or hotspots prone to bugs. These are candidates for refactoring, stronger patterns, or clearer ownership models.  

```bash
$ gitallica hotspots
Top Hotspots (last 6 months):
1. components/NavBar.tsx (87 commits)
2. pages/api/auth.ts (72 commits)
3. hooks/useSearch.ts (65 commits)
```

---

### 4. **New Component Creation Rate**  
*Are you growing at the right pace, or sprawling out of control?*  

Monitors the creation of new React components, hooks, API routes, and modules. A steady pace suggests healthy growth, while spikes can indicate rapid complexity expansion that may outpace architectural guardrails.  

```bash
$ gitallica components
New components (last 12 months):
- React Components: 33
- Hooks: 14
- API Routes: 9
Trend: +45% YoY (expanding fast)
```

---

### 5. **Directory Entropy**  
*When clean albums turn into messy mixtapes.*  

Measures how evenly changes are distributed within directories. High entropy means code boundaries are blurred and responsibilities unclear, often leading to brittle or tangled dependencies. Low entropy indicates stronger modularity and architectural discipline.  

```bash
$ gitallica entropy
High-Entropy Directories:
1. lib/utils/ (0.81)
2. components/shared/ (0.74)
3. scripts/ (0.68)
```

---

### 6. **Dead Zones**  
*Code that time forgot.*  

Highlights files untouched for long periods. Legacy code isn’t always harmful, but stale code often harbors technical debt, unpatched vulnerabilities, or business logic nobody fully understands anymore. These should be reviewed or retired.  

```bash
$ gitallica deadzones --age 2y
Untouched >2 years:
- lib/legacyAuth.ts
- scripts/dbMigration2019.sql
```

---

### 7. **Bus Factor (per directory)**  
*Who’s the last person standing if someone leaves?*  

Counts contributors per directory. A low bus factor signals risk — critical knowledge concentrated in one or two people. Increasing contributor spread in these areas reduces single points of failure and improves resilience.  

```bash
$ gitallica busfactor
Directory Ownership:
- components/ : 12 contributors (healthy band)
- lib/payments/ : 1 contributor (solo act — risky)
- hooks/ : 7 contributors
```

---

### 8. **Ownership Clarity**  
*Does anyone really own this code?*  

Examines commit distribution per file. Clear ownership (most changes by a few people) means accountability and consistency. Diffuse ownership (many contributors spread thin) often correlates with inconsistent quality and harder maintenance.  

```bash
$ gitallica ownership
Ownership:
- components/NavBar.tsx : 80% by 2 devs (clear)
- pages/api/* : 50+ contributors (chaos)
```

---

### 9. **Onboarding Footprint**  
*What do new devs touch first?*  

Analyzes which files new contributors modify in their first commits. A narrow footprint suggests a structured onboarding path. A broad footprint indicates a steep learning curve or poorly scoped starter tasks, both of which slow down ramp-up.  

```bash
$ gitallica onboarding --limit 5
First 5 commits by new devs:
- Alice: 2 files (focused start)
- Bob: 47 files (overexposed)
```

---

### 10. **Code vs. Test Ratio**  
*Are you rehearsing as much as you’re performing?*  

Compares code added to test code added. A healthy balance reflects consistent testing discipline. A high imbalance suggests risks in quality assurance and signals the need to reinforce testing practices.  

```bash
$ gitallica test-ratio
Last 90 days:
- Code: 12,843 lines
- Tests: 2,114 lines
- Ratio: 6.1 : 1  (tests lagging)
```

---

### 11. **High-Risk Commits**  
*The monster commits that touch everything.*  

Flags commits that change an unusually large number of files. These are difficult to review, more likely to introduce bugs, and typically reflect poor commit hygiene or weak branching strategies.  

```bash
$ gitallica risky
Commits touching >25 files:
- abc1234: "Refactor auth flow" (32 files)
- def5678: "Checkout + Payments merge" (41 files)
```

---

### 12. **Commit Cadence Trends**  
*Is your repo speeding up or slowing down?*  

Shows commit activity over time. Sustained dips may reflect reduced capacity, bottlenecks, or morale issues. Sudden spikes often correlate with crunch periods or looming deadlines. Both are valuable signals for engineering leadership.  

```bash
$ gitallica cadence
Commits per month (last 12 months):
- Peak: 458 (Apr 2024)
- Current: 312 (Jul 2024)
Trend: -31% (slowing tempo)
```

---

### 13. **Review Bottlenecks**  
*Where pull requests go to die.*  

Analyzes average merge times per file or directory. Files that consistently delay reviews are high-friction areas — often too complex, poorly documented, or organizationally sensitive. These need extra focus to keep velocity healthy.  

```bash
$ gitallica bottlenecks
Slowest-to-merge files:
- lib/invoice.ts : 4.2 days avg
- components/Checkout.tsx : 3.9 days avg
```

---

### 14. **Long-Lived Branches**  
*That one track nobody finishes.*  

Flags branches that linger without merging. The longer a branch lives, the greater the chance of merge conflicts, architectural drift, and wasted effort. Monitoring these helps enforce healthier integration practices.  

```bash
$ gitallica branches
Long-lived branches:
- feature/checkout-revamp : 94 days
- spike/graphql-test : 73 days
```

---

### 15. **Change Lead Time**  
*How long does it take to ship?*  

Measures the time from the first commit on a branch to its merge. Short lead times correlate with agility and strong DevOps practices. Long lead times highlight delivery bottlenecks or inefficient workflows.  

```bash
$ gitallica leadtime
Average Lead Time (6 months): 3.8 days
90th percentile: 12.4 days
```

---

## 🚀 Why Use gitallica?  
Other tools give you vanity metrics like stars and forks. **gitallica** digs deeper — into how your codebase evolves, where risks lie, and how your team really works.  

- **Managers** → Spot bottlenecks and risks early.  
- **Tech leads** → Track architecture drift in real time.  
- **Developers** → Know which files are fragile before touching them.  

⚡ *Don’t just play your repo. Rock it.*  
