---
type: feature-plan
title: "Stage 1: MVP plugin"
status: draft
spec: docs/specs/feature-v28/spec.md
created: 2026-03-06
---

# Architecture and Design Decisions

## Plugin Layout Architecture

All plugin content is nested inside a `claude-plugin/` subdirectory at the repository root. This isolates plugin files from the rest of the repo, groups them for coherent `git` change tracking, and enables loading with `claude --plugin-dir ./claude-plugin`.

```
grafanactl (repo root)
|
+-- .claude/                         <-- Existing: contributor-facing skills (unchanged)
|   +-- skills/
|       +-- grafanactl/              <-- Contains known bugs; NOT a source for plugin
|       +-- discover-datasources/    <-- Source for explore-datasources adaptation
|       +-- add-provider/            <-- Contributor-facing; excluded from plugin
|
+-- claude-plugin/                   <-- NEW: all plugin content in one directory
|   +-- .claude-plugin/
|   |   +-- plugin.json              <-- Plugin manifest
|   |
|   +-- agents/
|   |   +-- grafana-debugger.md      <-- Stub (full workflow in Stage 2)
|   |
|   +-- skills/
|       +-- setup-grafanactl/
|       |   +-- SKILL.md             <-- Written from scratch
|       |   +-- references/
|       |       +-- configuration.md <-- Written from scratch (source: agent-docs/config-system.md)
|       |
|       +-- explore-datasources/
|           +-- SKILL.md             <-- Adapted from .claude/skills/discover-datasources/
|           +-- references/
|               +-- discovery-patterns.md  <-- Copied + verified (bug-free)
|               +-- logql-syntax.md        <-- Copied + verified (bug-free)
|
+-- agent-docs/                      <-- Existing: authoritative source of truth
    +-- config-system.md             <-- Primary source for configuration.md rewrite
```

### Content Flow

```
                         AUTHORITATIVE SOURCES
                         =====================

  agent-docs/config-system.md --------> claude-plugin/skills/setup-grafanactl/SKILL.md
         |                                      |
         +------------------------------------> claude-plugin/skills/setup-grafanactl/references/configuration.md
                                                (written from scratch)

  .claude/skills/discover-datasources/
         |
         +-- SKILL.md ----(adapt)-----------> claude-plugin/skills/explore-datasources/SKILL.md
         |                                    (add cross-ref to setup-grafanactl,
         |                                     remove any graph pipe refs)
         |
         +-- references/discovery-patterns.md -(copy + fix)-> ...references/discovery-patterns.md
         |                                                     (fix Bug 2: graph pipe, Bug 3: jq path)
         |
         +-- references/logql-syntax.md -------(copy)-------> ...references/logql-syntax.md
                                                               (no bugs found; verbatim copy)

  Research report Section 2 ----------> claude-plugin/agents/grafana-debugger.md
                                        (stub from research agent template)

  Research report Section 2 ----------> claude-plugin/.claude-plugin/plugin.json
                                        (manifest from research template)
```

### Bug Impact Map

Each bug has a specific blast radius that determines which files need attention:

```
Bug 1 (config paths):    .claude/skills/grafanactl/ (throughout)
                         .claude/skills/grafanactl/references/configuration.md (throughout)
                         --> Mitigation: rewrite from scratch, never copy from these files

Bug 2 (graph pipe):      .claude/skills/grafanactl/SKILL.md (lines 72, 117, 152)
                         .claude/skills/discover-datasources/references/discovery-patterns.md (lines 207-214)
                         --> Mitigation: fix in discovery-patterns.md before copying

Bug 3 (jq envelope):    .claude/skills/discover-datasources/references/discovery-patterns.md (lines 84, 87)
                         --> Mitigation: fix .datasources[] wrapper in discovery-patterns.md

Bug 4 (--all-versions): .claude/skills/grafanactl/references/selectors.md
                         --> Mitigation: not copying selectors.md; verify no leakage into other files
```

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| Plugin nested in `claude-plugin/` subdirectory | All plugin files live under `claude-plugin/` at the repo root. `claude --plugin-dir ./claude-plugin` loads the plugin. This isolates plugin content from the rest of the repo, avoids `agents/` and `skills/` directories at the repo root, and enables coherent `git` tracking (all plugin changes appear under `claude-plugin/`). See reference: https://github.com/steveyegge/beads/tree/main/claude-plugin |
| Write configuration.md from config-system.md struct hierarchy | The existing `configuration.md` has Bug 1 embedded in every example (fictional `auth.type`/`auth.token`/`namespace` schema). The correct struct path is `contexts.<name>.grafana.{server,token,user,password,org-id,stack-id}`. Patching would risk missing instances; a clean rewrite from the authoritative source is safer. |
| Copy discovery-patterns.md with targeted fixes rather than rewrite | The file is 90% correct. Only the "Visualizing Query Results" section (Bug 2: `grafanactl graph` pipe) and "Saving Datasource UIDs" section (Bug 3: `.[]` vs `.datasources[]`) need fixes. A targeted edit preserves tested content. |
| Copy logql-syntax.md as-is | Verified: contains zero instances of Bug 1-4 patterns. No config paths, no graph pipe, no jq envelope paths, no --all-versions. Safe to copy verbatim. |
| SKILL.md `description` field drives auto-triggering | Claude Code matches user intent against skill descriptions. The descriptions must be keyword-rich but non-overlapping. `setup-grafanactl` owns "setup/config/auth/connection"; `explore-datasources` owns "datasource/metrics/labels/log streams/UIDs". |
| grafana-debugger as minimal stub | The agent references a diagnostic workflow that depends on `debug-with-grafana` skill (Stage 2). Shipping a full agent workflow without its supporting skill would produce hallucinated tool invocations. The stub establishes the agent's identity and approach without committing to specific skill references. |
| Omit `allowed-tools` from SKILL.md frontmatter | The existing discover-datasources skill uses `allowed-tools: grafanactl`, but grafanactl is invoked via the Bash tool, not as a named tool. This field's semantics for CLI-wrapping plugins need verification. Omitting avoids breakage; can be added post-verification. |
| Use plugin-dev meta-skills as quality gates, not generators | The plugin-dev skills provide review checklists and validation rubrics. The content itself is domain-specific (grafanactl workflows) and must be authored from domain knowledge. Meta-skills catch structural issues (missing frontmatter, description quality, manifest schema). |

## Compatibility

### What Continues Working
- `.claude/skills/` contributor-facing skills remain untouched. They continue functioning for contributors working on grafanactl itself.
- `agent-docs/` reference documentation is read-only; no modifications needed.
- All existing CLI behavior, tests, and builds are unaffected (this is a content-only change).

### What Is New
- `.claude-plugin/plugin.json` -- Plugin manifest enabling `claude --plugin-dir` loading.
- `skills/setup-grafanactl/` -- New skill teaching agents to configure grafanactl correctly.
- `skills/explore-datasources/` -- Adapted skill for datasource discovery (corrected version of the contributor-facing skill).
- `agents/grafana-debugger.md` -- Stub agent for future diagnostic workflows.
- `skills/setup-grafanactl/references/configuration.md` -- Authoritative config reference for agents.

### What Is Explicitly Excluded
- No MCP server, hooks, or slash commands (by design; see spec negative constraints).
- No modifications to `.claude/skills/` (bug fixes go into plugin copies only).
- No changes to Go source code, Makefile, or CI/CD.
