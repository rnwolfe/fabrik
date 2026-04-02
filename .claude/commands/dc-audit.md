---
description: >
  Evaluate fabrik's UX through the lens of an expert datacenter network architect persona.
  Validates Critical User Journeys (CUJs), surfaces friction points, and produces a
  prioritized findings report. Use for audit cadences, before major UI changes, or
  when designing new features — invoke as /dc-audit [focus-area]. focus-area is
  optional and defaults to auditing all areas when omitted (valid values: topology | catalog | metrics | racks).
---

You are performing a UX audit of **fabrik** — a datacenter network topology design tool.
Your evaluation framework is grounded in a realistic expert user persona. By default,
work through all phases in order and do not skip any. If a specific [focus-area]
argument is provided, run only the CUJs and phases that are relevant to that focus area
and skip non-relevant phases.
---

## The Persona: Morgan, Staff Network Architect

Morgan is the persona against whom you evaluate every UX decision in fabrik.

**Background**
- 12 years designing and operating large-scale datacenter networks at cloud and
  co-location providers
- Deep expertise in Clos/fat-tree topologies: understands oversubscription ratios,
  ECMP, port budgets, and fabric stage math intuitively
- Currently responsible for capacity planning, hardware RFQs, and network design reviews
  for a mid-size cloud provider (tens of thousands of server ports across multiple sites)
- Has used Visio, Lucidchart, NetBox, and a half-dozen internal bespoke tools —
  none of which are good enough

**Goals**
1. Quickly model "what does this fabric look like at N pods?" before writing a design doc
2. Validate that a design is physically realizable (port counts, RU, power fit the hardware)
3. Explore hardware options and their cost/density/power trade-offs
4. Produce artifacts (diagrams, metrics summaries) to present to stakeholders
5. Check a proposed design against known-good Clos math without doing it by hand

**Mental model**
- Thinks in tiers: leaf → spine → super-spine. The number of stages matters.
- A "block" is a failure domain, a "pod" is a unit of scale-out
- Port utilization and oversubscription ratio are the two numbers that tell you if a design is sane
- Rack space and power are physical constraints that kill otherwise-good designs
- Terminology matters: uses "leaf/spine", "uplink/downlink", "radix", "ECMP", "bisection bandwidth"

**Pain points with current tooling**
- Generic diagramming tools don't understand Clos math — you have to compute everything manually
- NetBox is inventory, not design — you can't model a future state
- Spreadsheets work but are fragile and hard to share
- Tools that don't know real hardware force you to abstract away the details that matter

**What Morgan does NOT care about**
- Aesthetic customization (colors, themes) — function over form
- Step-by-step wizards for things she already knows how to do
- Tooltips explaining what a "spine switch" is

---

## Phase 1 — Orient: Read the current app state

Read the following to understand what is actually in the codebase today.
Do NOT rely on memory — read the actual files.

1. Route definitions to know what screens exist:
   - `frontend/src/App.tsx` or `frontend/src/main.tsx` (find the router)
   - Look for route files in `frontend/src/`

2. Key feature components — skim (don't read fully) to understand what UI exists:
   - `frontend/src/` — list all top-level directories and key files

3. The core data models:
   - `frontend/src/` — find TypeScript model/type files

4. Any existing UX copy: page titles, button labels, empty states, error messages —
   sample these from 3-4 component files

After reading, write a brief (bullet list) inventory of:
- Screens that exist
- Key actions available per screen
- Data models central to the design flow

---

## Phase 2 — CUJ Walkthrough

Walk through each Critical User Journey below as Morgan would experience it.
For each CUJ, read the relevant component(s) and evaluate the UX against the
criteria listed. Note findings as you go.

### CUJ 1: Start a new fabric design

**Morgan's goal**: Quickly spin up a new design and get to the point where she can
see a Clos topology taking shape.

**Steps to evaluate**:
1. Can Morgan create a new design from the home/dashboard screen?
2. Is the entry point obvious without reading instructions?
3. How many clicks/steps before she sees something topology-shaped?
4. What defaults are chosen, and are they sensible for an expert?

**Criteria**:
- [ ] New design creation is reachable in ≤2 clicks from home
- [ ] No mandatory wizard or multi-step form before the design canvas appears
- [ ] Default values (if any) reflect real-world Clos starting points, not arbitrary numbers
- [ ] The design is named automatically or with a sensible placeholder (not "Untitled")

### CUJ 2: Define fabric hierarchy (blocks, pods, tiers)

**Morgan's goal**: Model a specific fabric structure — e.g., 4 blocks of 48 servers
each, 2-stage Clos.

**Steps to evaluate**:
1. How does Morgan add blocks/pods/tiers?
2. Is the hierarchy (block → superblock → site) surfaced clearly?
3. Can she see the computed fabric math (switch counts, port utilization) as she builds?
4. Does the UI use the right vocabulary (leaf/spine, not "node/group")?

**Criteria**:
- [ ] Hierarchy is editable in the topology view, not a separate settings panel
- [ ] Fabric math (leaf count, spine count, oversubscription) updates live as she edits
- [ ] Labels use network-architecture vocabulary: leaf, spine, uplink, downlink, radix
- [ ] The relationship between hierarchy level and fabric tier is visually apparent

### CUJ 3: Select and assign hardware

**Morgan's goal**: Pick a real switch model for the leaf and spine roles, verify the
port counts work.

**Steps to evaluate**:
1. How does Morgan browse and select device models?
2. Are port groups (speed + count) clearly displayed?
3. Can she see how a device's port budget affects the fabric math?
4. Is it obvious which devices are suitable for which tier?

**Criteria**:
- [ ] Device selection is accessible from within the topology flow (not a separate app section only)
- [ ] Port group details (e.g., "48×25GbE + 6×100GbE") are displayed, not hidden behind clicks
- [ ] Filtering by vendor, type, or port speed is available
- [ ] The catalog-to-topology connection is clear: assigning a device model updates fabric math

### CUJ 4: Validate design constraints

**Morgan's goal**: Confirm the design is physically realizable — ports fit, RU fits,
power is within budget.

**Steps to evaluate**:
1. Where does Morgan see oversubscription ratios?
2. Are power and RU constraints surfaced proactively (before they're exceeded)?
3. Are constraint violations clearly flagged, and do they link back to the source?
4. Is the metrics view integrated into the design flow or siloed?

**Criteria**:
- [ ] Oversubscription ratio is visible without navigating away from the topology
- [ ] Constraint violations (port exhaustion, power overrun, RU overflow) show inline warnings
- [ ] Metrics are tied to the active design — no ambiguity about which design is being evaluated
- [ ] The numbers shown match standard Clos math (verify with a known topology)

### CUJ 5: Understand what the tool is telling her (empty states, errors, guidance)

**Morgan's goal**: When something is wrong or missing, understand what to do next —
without reading documentation.

**Steps to evaluate**:
1. What do empty states look like (no designs, no devices in catalog, no metrics data)?
2. Are error messages specific and actionable for an expert?
3. Is guidance appropriately calibrated — not over-explaining basics, not under-explaining edge cases?

**Criteria**:
- [ ] Empty states explain why there's nothing to show and what action to take
- [ ] Error messages use domain vocabulary and specify the failing constraint
- [ ] The app does not explain what a "spine switch" is, but does explain *why* port counts don't add up
- [ ] Help/documentation links go to relevant articles, not the top of the knowledge base

---

## Phase 3 — Vocabulary & Terminology Scan

Read the UI copy across key components (button labels, page titles, section headings,
placeholder text, error messages, empty states). Flag:

| Finding | Severity |
|---------|----------|
| Wrong term for a well-defined concept (e.g. "node" instead of "switch", "group" instead of "pod") | HIGH |
| Ambiguous term that could mean multiple things to a network architect | MEDIUM |
| Missing label where one is needed (unlabeled inputs, orphan numbers) | MEDIUM |
| Overly generic or consumer-app language ("Get started!", "No items yet") | LOW |

---

## Phase 4 — Information Architecture

Evaluate the navigation and screen structure against Morgan's mental model:

1. **Does the nav reflect the workflow?** A DC designer goes: design → hardware → validate.
   Does the app's nav reinforce this or fight it?

2. **Is the active design context always clear?** Morgan may have multiple designs.
   Which one is she editing right now? Is it obvious?

3. **Are the right things co-located?** Topology and metrics should be closely related.
   Catalog and topology assignment should be frictionless. Are they?

4. **Depth vs. breadth trade-offs**: Is information appropriately nested (expert detail
   one click away) rather than either hidden or dumped on the surface?

---

## Phase 5 — Format the Report

Output a markdown report with this exact structure:

```markdown
# DC Designer UX Audit
*Persona: Morgan, Staff Network Architect*
*Date: <date>*
*Scope: <all | specific area>*

## Executive Summary
<2–4 sentences: the overall UX health from Morgan's perspective, top theme, recommended focus>

## CUJ Scorecard
| CUJ | Status | Friction Score (1–5) | Key Issue |
|-----|--------|----------------------|-----------|
| 1. Start a new design | ✅ Pass / ⚠️ Partial / ❌ Fail | N | ... |
| 2. Define fabric hierarchy | ... | N | ... |
| 3. Select hardware | ... | N | ... |
| 4. Validate constraints | ... | N | ... |
| 5. Guidance & empty states | ... | N | ... |

*Friction score: 1 = frictionless, 5 = Morgan would abandon the task*

---

## Findings

### HIGH Priority

#### [H1] <Short title>
**Where**: `ComponentName` / screen name
**What Morgan experiences**: <1–2 sentences describing the friction from her perspective>
**Why it matters**: <impact on her goal or mental model>
**Suggested fix**: <specific, concrete recommendation>

[repeat for each HIGH finding]

### MEDIUM Priority
[same structure]

### LOW Priority
[same structure]

---

## Vocabulary Findings
| Current Term | Correct Term | Location | Severity |
|-------------|-------------|----------|----------|
| ...          | ...          | ...       | HIGH/MED/LOW |

---

## Information Architecture Notes
<Bulleted observations about nav, context, co-location, depth/breadth>

---

## What's Working Well
<3–5 specific things that Morgan would find genuinely good — be honest, not diplomatic>

---

## Recommended Next Actions (Top 3)
1. **[Priority]** `File:component` — <specific change> — *Why: <Morgan's perspective>*
2. ...
3. ...
```

---

## Phase 6 — Calibration notes

Apply these when making judgment calls:

- **Morgan is an expert.** Do not credit the app for explaining basic concepts.
  Credit it for exposing expert detail that other tools hide.
- **Friction score honestly.** A score of 1 should be rare — it means zero resistance.
  Most working flows are 2–3. Score 4–5 only when Morgan would actually give up or
  reach for a spreadsheet instead.
- **Vocabulary is a proxy for domain understanding.** If the app uses generic terms,
  it signals the tool doesn't understand the domain. This erodes expert trust quickly.
- **Co-location matters more than completeness.** A metric that exists but is 3 screens
  away from where the decision is made is effectively absent for Morgan's workflow.
- **Empty states are the first impression for new designs.** They're often the worst
  part of tool UX and deserve scrutiny.
- **Do not invent findings.** If a CUJ actually works well, say so. The value of this
  audit is honest signal, not a list of problems to seem thorough.
- **If invoked with a specific focus area** (e.g., `/dc-audit topology`), still run all
  phases in order, but limit CUJs and examples to that area; always include the vocabulary
  scan for that focus area.
- **If invoked to inform a specific decision** (user provides context before running),
  anchor findings to that decision — lead the report with how findings affect that choice.
