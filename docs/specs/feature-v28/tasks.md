---
type: feature-tasks
title: "Stage 1: MVP plugin"
status: draft
spec: docs/specs/feature-v28/spec.md
plan: docs/specs/feature-v28/plan.md
created: 2026-03-06
---

# Implementation Tasks

## Dependency Graph

```
Wave 1: Scaffold
  T1: plugin scaffold + plugin.json
        |
        +---------------------------+------+
        |                    |      |      |
Wave 2+3: (all parallel from T1)    |      |
  T2: setup-grafanactl       T3:    T4:    T5: grafana-debugger stub
  (skill from scratch)       conf.md explore (can run in parallel w/ T2-T4)
        |                    |      |      |
        +---------------------------+------+
        |
Wave 4: Quality Gates (depends on T2, T3, T4, T5)
  T6: plugin-dev validation + bug verification
```

---

## Wave 1: Plugin Scaffold

### T1: Create plugin directory structure and manifest
**Priority**: P0
**Effort**: Small
**Depends on**: none
**Type**: chore
**FRs**: FR-001, FR-002, FR-003, FR-027

Create the plugin directory tree and write `plugin.json`. Invoke `plugin-dev:plugin-structure` before writing the manifest to ensure it follows current Claude Code plugin conventions.

**Procedure:**

1. Invoke `plugin-dev:plugin-structure` to get current manifest schema guidance.
2. Create all directories:
   - `claude-plugin/.claude-plugin/`
   - `claude-plugin/agents/`
   - `claude-plugin/skills/setup-grafanactl/references/`
   - `claude-plugin/skills/explore-datasources/references/`
3. Write `claude-plugin/.claude-plugin/plugin.json` with required fields.
4. Verify JSON is valid and contains no `mcpServers`, `commands`, or `hooks` keys.

**Deliverables:**
- `claude-plugin/.claude-plugin/plugin.json`
- Directory tree: `claude-plugin/agents/`, `claude-plugin/skills/setup-grafanactl/references/`, `claude-plugin/skills/explore-datasources/references/`

**Acceptance criteria:**
- GIVEN the repository root WHEN I inspect `claude-plugin/.claude-plugin/plugin.json` THEN it is valid JSON with `name` = "grafanactl", `version` = "0.1.0", `description` present, `keywords` containing "grafana", "observability", "prometheus", "loki"
- GIVEN the plugin.json file WHEN I search for `mcpServers`, `commands`, or `hooks` keys THEN zero matches are found
- GIVEN the directory tree WHEN I list all paths THEN `claude-plugin/agents/`, `claude-plugin/skills/setup-grafanactl/references/`, `claude-plugin/skills/explore-datasources/references/` all exist

---

## Wave 2: Content (parallel tasks)

### T2: Write setup-grafanactl skill from scratch
**Priority**: P0
**Effort**: Medium
**Depends on**: T1
**Type**: task
**FRs**: FR-004, FR-005, FR-007, FR-008, FR-009, FR-010, FR-011

Write `claude-plugin/skills/setup-grafanactl/SKILL.md` entirely from scratch using `agent-docs/config-system.md` as the authoritative source for all config paths. Invoke `plugin-dev:skill-development` for structural guidance before writing.

**Procedure:**

1. Read `agent-docs/config-system.md` (already read; key struct: `contexts.<name>.grafana.{server,token,user,password,org-id,stack-id}`).
2. Invoke `plugin-dev:skill-development` for SKILL.md structural guidance.
3. Write SKILL.md with YAML frontmatter (`name`, `description`).
4. Document three configuration paths:
   - Path A (Grafana Cloud): `grafana.token`, auto-discovered stack-id
   - Path B (On-premise): `grafana.user`/`grafana.password`, `grafana.org-id`
   - Path C (Environment variables): `GRAFANA_SERVER`, `GRAFANA_TOKEN`, `GRAFANA_ORG_ID`/`GRAFANA_STACK_ID`
5. Include default datasource configuration (`default-prometheus-datasource`, `default-loki-datasource`).
6. Include troubleshooting section: config check failures, 401/403, connection refused/timeout, namespace resolution.
7. Verify zero instances of Bug 1-4 patterns in the output file.

**Deliverables:**
- `claude-plugin/skills/setup-grafanactl/SKILL.md`

**Acceptance criteria:**
- GIVEN `claude-plugin/skills/setup-grafanactl/SKILL.md` WHEN I read the frontmatter THEN `name` and `description` fields are present; description mentions setup, configuration, authentication, connection, and first-time use
- GIVEN the skill content WHEN I search for `auth.type`, `auth.token`, `auth.username`, `contexts.<name>.namespace` as config set targets THEN zero matches are found
- GIVEN the skill content WHEN I search for `grafanactl graph` as standalone command or pipe target THEN zero matches are found
- GIVEN the skill content WHEN I search for `--all-versions` THEN zero matches are found
- GIVEN the skill content WHEN I read it THEN all three config paths (Cloud, on-prem, env vars) are documented
- GIVEN the skill content WHEN I read it THEN default datasource configuration and troubleshooting sections are present

---

### T3: Write configuration.md reference from scratch
**Priority**: P0
**Effort**: Medium
**Depends on**: T1
**Type**: task
**FRs**: FR-004, FR-012, FR-013

Write `claude-plugin/skills/setup-grafanactl/references/configuration.md` from scratch using `agent-docs/config-system.md` as the sole authoritative source. This file MUST NOT be derived from `.claude/skills/grafanactl/references/configuration.md` (which contains Bug 1 throughout).

**Procedure:**

1. Read `agent-docs/config-system.md` data model section.
2. Extract all `config set` paths from the struct hierarchy:
   - `contexts.<name>.grafana.server`
   - `contexts.<name>.grafana.token`
   - `contexts.<name>.grafana.user`
   - `contexts.<name>.grafana.password`
   - `contexts.<name>.grafana.org-id`
   - `contexts.<name>.grafana.stack-id`
   - `contexts.<name>.grafana.tls.insecure-skip-verify`
   - `contexts.<name>.grafana.tls.ca-data`
   - `contexts.<name>.grafana.tls.cert-data`
   - `contexts.<name>.grafana.tls.key-data`
   - `contexts.<name>.default-prometheus-datasource`
   - `contexts.<name>.default-loki-datasource`
3. Document environment variable names and precedence (env vars override current context only).
4. Document config file location priority (5 levels).
5. Document namespace resolution logic (stack-id auto-discovery, org-id fallback).
6. Document multi-context management (use-context, list-contexts, --context flag).
7. Cross-reference every path against `agent-docs/config-system.md` struct hierarchy to verify correctness.

**Deliverables:**
- `claude-plugin/skills/setup-grafanactl/references/configuration.md`

**Acceptance criteria:**
- GIVEN `claude-plugin/skills/setup-grafanactl/references/configuration.md` WHEN I compare every `config set` path to the data model in `agent-docs/config-system.md` THEN every path matches the actual struct hierarchy
- GIVEN the file WHEN I search for `auth.type`, `auth.token`, `auth.username`, `namespace` as a config set target THEN zero matches are found
- GIVEN the file WHEN I read it THEN it documents: config set paths for all fields, environment variable names and precedence, config file location, namespace resolution logic, and multi-context management
- GIVEN the file WHEN I compare it to `.claude/skills/grafanactl/references/configuration.md` THEN the content is structurally different (not a patch of the old file)

---

### T4: Adapt explore-datasources skill and copy references
**Priority**: P0
**Effort**: Medium
**Depends on**: T1
**Type**: task
**FRs**: FR-005, FR-006, FR-007, FR-014, FR-015, FR-016, FR-017, FR-018

Adapt `skills/explore-datasources/SKILL.md` from the existing `.claude/skills/discover-datasources/SKILL.md`. Copy and fix the two reference files. This task owns all Bug 2 and Bug 3 fixes in copied content.

**Procedure:**

1. Read `.claude/skills/discover-datasources/SKILL.md` (already read; confirmed mostly clean of Bug 1-4 except the SKILL.md itself has no bugs).
2. Adapt SKILL.md:
   - Update YAML frontmatter: set `name` to "explore-datasources", write keyword-rich `description` mentioning datasource discovery, metrics, labels, log streams, datasource UIDs.
   - Preserve all four steps, all four examples, troubleshooting entries, Advanced Usage section, and Output Formats section.
   - Add cross-reference to `setup-grafanactl` skill (e.g., "If grafanactl is not configured, see the setup-grafanactl skill first").
   - Verify no `grafanactl graph` pipe references exist (none found in source).
   - Verify no `--all-versions` references exist (none found in source).
3. Copy `references/discovery-patterns.md` with targeted fixes:
   - **Bug 2 fix**: Replace "Visualizing Query Results" section -- remove `| grafanactl graph` pipe examples, replace with `-o graph` codec examples.
   - **Bug 3 fix**: In "Saving Datasource UIDs" section, fix `jq -r '.[] | select(.type=="prometheus") | .uid'` to `jq -r '.datasources[] | select(.type=="prometheus") | .uid'` (and same for loki).
   - Verify no `--all-versions` references.
4. Copy `references/logql-syntax.md` as-is (verified: zero Bug 1-4 content).
5. Run final grep across all three output files for Bug 1-4 patterns.

**Deliverables:**
- `claude-plugin/skills/explore-datasources/SKILL.md`
- `claude-plugin/skills/explore-datasources/references/discovery-patterns.md`
- `claude-plugin/skills/explore-datasources/references/logql-syntax.md`

**Acceptance criteria:**
- GIVEN `claude-plugin/skills/explore-datasources/SKILL.md` WHEN I compare it to `.claude/skills/discover-datasources/SKILL.md` THEN all four steps, all four examples, troubleshooting entries, and Advanced Usage section are present
- GIVEN the SKILL.md frontmatter WHEN I read it THEN `name` and `description` are present; description mentions datasource discovery, metrics, labels, log streams, datasource UIDs
- GIVEN the SKILL.md content WHEN I search for `setup-grafanactl` THEN at least one cross-reference is found
- GIVEN `claude-plugin/skills/explore-datasources/references/discovery-patterns.md` WHEN I search for `| grafanactl graph` THEN zero matches are found
- GIVEN `claude-plugin/skills/explore-datasources/references/discovery-patterns.md` WHEN I search for `jq -r '.\[\]` (bare array access on datasource list) THEN zero matches are found; all datasource list jq examples use `.datasources[]`
- GIVEN `claude-plugin/skills/explore-datasources/references/logql-syntax.md` WHEN I diff against the source THEN files are identical (no modifications needed)
- GIVEN all three output files WHEN I search for `--all-versions` THEN zero matches are found

---

## Wave 3: Agent Stub

### T5: Write grafana-debugger agent stub
**Priority**: P1
**Effort**: Small
**Depends on**: T1
**Type**: task
**FRs**: FR-019, FR-020, FR-021, FR-022, FR-029

Write `agents/grafana-debugger.md` as a system prompt stub. Invoke `plugin-dev:agent-development` for structural guidance on frontmatter fields and description quality.

**Procedure:**

1. Invoke `plugin-dev:agent-development` for agent file guidance.
2. Write `agents/grafana-debugger.md` with:
   - YAML frontmatter: `name` ("grafana-debugger"), `description` (mentioning diagnose, errors, latency, service degradation, Grafana observability data).
   - System prompt body establishing diagnostic approach: discover datasources, confirm scraping, query error rates, correlate logs, summarize findings.
   - Note: always use `-o json` for machine-parseable output.
   - Note: always use datasource UIDs, not names.
   - Note: if grafanactl not configured, guide user through setup first.
3. Keep the stub generic -- do not reference specific skill names to avoid coupling to Stage 2 content.

**Deliverables:**
- `claude-plugin/agents/grafana-debugger.md`

**Acceptance criteria:**
- GIVEN `claude-plugin/agents/grafana-debugger.md` WHEN I read the frontmatter THEN `name` is "grafana-debugger" and `description` mentions diagnosing application issues, errors, latency, service degradation, Grafana observability
- GIVEN the agent body WHEN I read it THEN the diagnostic approach is described (discover datasources, confirm scraping, query error rates, correlate logs, summarize findings)
- GIVEN the agent body WHEN I search for `-o json` THEN at least one mention is found
- GIVEN the agent body WHEN I search for "UID" THEN at least one mention is found about using datasource UIDs not names
- GIVEN the agent body WHEN I search for "setup" or "configured" THEN at least one mention is found about guiding users through setup if grafanactl is not configured

---

## Wave 4: Quality Gates

### T6: Validate plugin with plugin-dev meta-skills and verify all bugs fixed
**Priority**: P0
**Effort**: Medium
**Depends on**: T2, T3, T4, T5
**Type**: chore
**FRs**: FR-004, FR-005, FR-006, FR-007, FR-023, FR-024, FR-025, FR-026, FR-028, FR-030

Run all quality gate validations: plugin-dev:skill-reviewer on each skill, plugin-dev:plugin-validator on the full plugin, bug pattern verification via grep, and plugin loading test.

**Procedure:**

1. **Bug verification (all 4 bugs across all plugin files):**
   - Grep all files under `skills/`, `agents/`, `.claude-plugin/` for `auth.type`, `auth.token`, `auth.username`, `contexts.*.namespace` as config targets -- expect zero matches (Bug 1).
   - Grep for `| grafanactl graph` and `grafanactl graph` as standalone -- expect zero matches (Bug 2).
   - Grep for `'.\[0\].uid'` and `'.\[\]` on datasource list output -- expect zero matches (Bug 3).
   - Grep for `--all-versions` -- expect zero matches (Bug 4).

2. **Skill reviews:**
   - Invoke `plugin-dev:skill-reviewer` on `skills/setup-grafanactl/SKILL.md`. Address any major issues found.
   - Invoke `plugin-dev:skill-reviewer` on `skills/explore-datasources/SKILL.md`. Address any major issues found.

3. **Plugin validation:**
   - Invoke `plugin-dev:plugin-validator` on the complete plugin. Address any critical errors.

4. **Plugin loading test:**
   - Run `claude --plugin-dir ./claude-plugin` and verify it loads without errors.
   - Verify both skills appear in the skill list.
   - Test auto-triggering: "help me set up grafanactl" should trigger setup-grafanactl.
   - Test auto-triggering: "what datasources does my Grafana have" should trigger explore-datasources.

5. **Fix any issues** found during validation. Re-run validation after fixes until all gates pass.

**Deliverables:**
- All plugin files passing validation (no new files; this task modifies existing files from T1-T5 if issues are found)
- Verification log confirming all 4 bug patterns have zero matches

**Acceptance criteria:**
- GIVEN all plugin content files WHEN I grep for Bug 1-4 patterns THEN zero matches across all four bug categories
- GIVEN `skills/setup-grafanactl/SKILL.md` WHEN reviewed by `plugin-dev:skill-reviewer` THEN no major issues reported
- GIVEN `skills/explore-datasources/SKILL.md` WHEN reviewed by `plugin-dev:skill-reviewer` THEN no major issues reported
- GIVEN the complete plugin WHEN validated by `plugin-dev:plugin-validator` THEN passes without critical errors
- GIVEN the plugin at repo root WHEN loaded with `claude --plugin-dir ./claude-plugin` THEN plugin loads without errors and both skills appear in the skill list
- GIVEN the plugin is loaded WHEN user says "help me set up grafanactl" THEN setup-grafanactl skill triggers
- GIVEN the plugin is loaded WHEN user says "what datasources does my Grafana have" THEN explore-datasources skill triggers
