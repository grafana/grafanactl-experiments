---
name: add-provider
description: Add a new Grafana product provider to grafanactl. Guides through API discovery, design decisions, implementation, and verification. Use when adding support for a new Grafana Cloud product (SLO, OnCall, Synthetic Monitoring, k6, ML, etc.) or when the user says "add provider", "new provider", or "integrate [product]".
---

# Add Provider

Add a new Grafana product provider to grafanactl — from API discovery through
verified implementation.

## When to Use

- User wants to add CLI support for a Grafana Cloud product
- User says "add provider", "new provider", "integrate [product]"
- A bead task references provider implementation (e.g., Wave 2/3 providers)

**First**: Check the decision tree in `references/decision-tree.md` to confirm
a provider is the right approach (vs extending the existing resources command).

## Workflow

The workflow has four phases. Each phase has a gate — don't proceed until the
gate condition is met.

```
Phase 1: Discover    → Design doc produced
Phase 2: Design      → Decision framework answered
Phase 3: Implement   → Code compiles, tests pass
Phase 4: Verify      → All checklists green
```

### Prerequisites

Before starting, confirm with the user:
- **Product name** — which Grafana product to integrate
- **Access** — do they have a running Grafana instance with the product enabled?
- **Scope** — full provider or single resource type first?

---

## Phase 1: Discover

> **Guide**: `agent-docs/provider-discovery-guide.md` — follow Sections 1.1–1.6

Research the product's API surface. Work through these steps in order:

### Step 1: Map the API Surface (Section 1.1)

Search for the product's OpenAPI spec:
```bash
# Check for OpenAPI spec repos
gh search repos "org:grafana {product}-openapi" --json name,url
# Check plugin API paths
curl -s -H "Authorization: Bearer $TOKEN" "$GRAFANA_URL/api/plugins" | jq '.[] | select(.id | contains("{product}"))'
```

Capture: base path, auth scheme, endpoints, response wrappers, pagination.

### Step 2: Check Existing Tooling (Section 1.2)

```bash
# Check Terraform provider for resource schemas
gh api repos/grafana/terraform-provider-grafana/contents/internal/resources --jq '.[].name'
# Check for Go SDK
gh search repos "org:grafana {product}-go-client OR {product}-api-go" --json name,url
```

Extract: schema fields, types, validation rules, CRUD patterns.

### Step 3: Inspect Source Code (Section 1.3)

```bash
# Find the product's API handlers
gh search code "org:grafana repo:grafana/{product} path:pkg/api" --json path,repository
```

Look for: undocumented endpoints, enum values, validation rules, RBAC requirements.

### Step 4: Identify Auth Model (Section 1.4)

Determine if the product reuses `grafana.token` or needs separate credentials.
This drives `ConfigKeys()` in Phase 3.

### Step 5: Map Resource Relationships (Section 1.5)

Document how resources reference each other and existing grafanactl resources.

### Step 6: Test API Behavior (Section 1.6)

Make real API calls to validate assumptions:
```bash
# List resources
curl -H "Authorization: Bearer $TOKEN" \
  "$GRAFANA_URL/api/plugins/{plugin-id}/resources/v1/{resource}"
```

Verify: response shape, duplicate handling, ID generation, error format.

### Gate: Discovery Complete

Present findings to user. Confirm:
- [ ] API endpoints documented
- [ ] Auth model identified
- [ ] Resource relationships mapped
- [ ] At least one successful API call made

---

## Phase 2: Design

> **Guide**: `agent-docs/provider-discovery-guide.md` Section 2 (Decision Framework)

Answer each decision question. Use the tables in the guide for reference.

### Decision 1: Auth Strategy (Section 2.1)

| Scenario | ConfigKeys | Token Source |
|----------|-----------|--------------|
| Same Grafana SA token | `[]` (empty) | `curCtx.Grafana.Token` |
| Separate product token | `[{Name: "token", Secret: true}]` | Provider config |
| Separate URL + token | `[{Name: "url"}, {Name: "token", Secret: true}]` | Provider config |

### Decision 2: API Client Type (Section 2.2)

| API Type | Client Approach |
|----------|----------------|
| Plugin API (`/api/plugins/...`) | Custom `http.Client` with Bearer token |
| K8s-compatible API (`/apis/...`) | grafanactl's existing dynamic client |
| External service API | Custom `http.Client` with product-specific auth |

**Warning**: Always verify K8s APIs are externally accessible before choosing that path.

### Decision 3: Envelope Mapping (Section 2.3)

Map API objects to grafanactl's K8s envelope:
```yaml
apiVersion: {product}.ext.grafana.app/v1alpha1
kind: {ResourceKind}    # PascalCase singular
metadata:
  name: {unique-id}     # UUID or slug from API
  namespace: default
spec:
  {fields}              # User-editable fields only
```

### Decision 4: Command Surface (Section 2.4)

Standard set (always include):
```
grafanactl {provider}
├── {resource-group}
│   ├── list
│   ├── get <id>
│   ├── push [path...]
│   ├── pull
│   └── delete <id...>
```

Consider adding: `status` (if operational health data exists).

### Decision 5: Package Layout (Section 2.5)

Actual convention (from SLO reference implementation):

```
internal/{provider}/            ← top-level package, NOT internal/providers/{provider}/
├── provider.go                 # Provider interface impl + configLoader
├── provider_test.go            # Contract tests
├── {resource}/                 # One subpackage per resource type
│   ├── types.go
│   ├── client.go
│   ├── adapter.go
│   ├── commands.go
│   └── *_test.go
```

Single resource type → flat package. Multiple → subpackage per type.

**Note**: Provider implementations live at `internal/{name}/`, NOT
`internal/providers/{name}/`. The `internal/providers/` package contains only
the interface definition and registry — keeping provider implementations
separate avoids import cycles.

### Decision 6: Implementation Staging (Section 2.6)

Break into independently shippable stages with a design doc for each:

```
docs/designs/{provider}/
├── {date}-{provider}-plan.md           # Top-level plan (all stages)
├── 1-{resource}-crud/
│   └── {date}-{resource}-crud.md       # Stage 1 design
├── 2-{secondary}-crud/
│   └── {date}-{secondary}-crud.md      # Stage 2 design
└── 3-status/
    └── {date}-status.md                # Stage 3 design
```

Common stage sequence:
1. Core CRUD for primary resource (~1,300 LOC for SLO)
2. Secondary resource types, if any (~500 LOC)
3. Status/monitoring (~350 LOC)
4. Advanced features (graph, timeline, etc.)

### Gate: Design Complete

Write a top-level plan doc in `docs/designs/{provider}/` capturing all
decisions, file tree, and stage breakdown. Create per-stage docs for each
stage. **Get user approval before implementing.** The SLO plan is the template:
`docs/designs/slo-provider/2026-03-04-slo-provider-plan.md`.

---

## Phase 3: Implement

> **Guide**: `agent-docs/provider-guide.md` — follow Steps 1–7
> **UX Guide**: `agent-docs/design-guide.md` — comply with all [CURRENT] and [ADOPT] items

Implement one stage at a time. For each stage:

### Step 1: Provider Interface (`provider-guide.md` Step 1)

Create `internal/{name}/provider.go`. Include the full `configLoader` —
providers cannot import `cmd/grafanactl/config` (import cycle):

```go
type {Name}Provider struct{}
var _ providers.Provider = &{Name}Provider{}

func (p *{Name}Provider) Name() string      { return "{name}" }
func (p *{Name}Provider) ShortDesc() string { return "Manage Grafana {Product} resources." }

// configLoader avoids importing cmd/grafanactl/config (import cycle).
// Copy from internal/slo/provider.go and update as needed.
type configLoader struct {
    configFile string
    ctxName    string
}

func (l *configLoader) bindFlags(flags *pflag.FlagSet) { ... }
func (l *configLoader) LoadRESTConfig(ctx context.Context) (config.NamespacedRESTConfig, error) { ... }
```

**Important**: Copy the full `configLoader` from `internal/slo/provider.go` —
it handles env vars (`GRAFANA_TOKEN`, `GRAFANA_PROVIDER_*`), context switching,
and validation. Don't simplify it; the full implementation is required.

### Step 2: Config Keys (`provider-guide.md` Step 2)

Declare all keys the provider reads. Secret keys get `Secret: true`.
SLO uses empty `[]` because it reuses `grafana.token` — most plugin API
providers can do the same.

### Step 3: Validate (`provider-guide.md` Step 3)

Return actionable errors pointing to `grafanactl config set ...`.

### Step 4: Commands (`provider-guide.md` Step 4)

```go
func (p *{Name}Provider) Commands() []*cobra.Command {
    loader := &configLoader{}
    cmd := &cobra.Command{Use: "{name}", Short: p.ShortDesc()}
    loader.bindFlags(cmd.PersistentFlags())
    cmd.AddCommand({resource}.Commands(loader))
    return []*cobra.Command{cmd}
}
```

**UX requirements** (from `design-guide.md`):
- Register `text` table codec as default for list/get commands
- Use `cmdio.Success/Warning/Error/Info` for status messages
- Include `-o json/yaml` support via `io.Options`
- Include help text with examples (3-5 per command)
- Push is idempotent (create-or-update)
- Data fetching is format-agnostic (Pattern 13)
- PromQL via `promql-builder`, not string formatting (Pattern 14)

**Client decision**: Hand-roll the HTTP client (don't import generated OpenAPI
clients). grafanactl's pattern is direct HTTP calls with `Authorization: Bearer`
headers. Generated clients use awkward types that break the adapter's
`encoding/json` round-trip. ~200 LOC for a typical CRUD client.

### Step 5: Types + Client + Adapter

For each resource type, create:
- `types.go` — Go structs matching API schema. Use camelCase field names
  matching the API (ensures lossless pull → edit → push round-trips)
- `client.go` — HTTP client (List, Get, Create, Update, Delete) with `httptest`
  unit tests
- `adapter.go` — Translate between API objects and K8s `Unstructured`. Test
  with round-trip property tests

Use `internal/slo/definitions/` as the reference for all three files.

### Step 6: Register

**Registration is in `cmd/grafanactl/root/command.go`**, NOT
`internal/providers/registry.go`. This avoids import cycles between
`internal/providers` and provider implementations (which import `internal/config`):

```go
// cmd/grafanactl/root/command.go

import {name}provider "github.com/grafana/grafanactl/internal/{name}"

func allProviders() []providers.Provider {
    return append(
        providers.All(),
        &sloprovider.SLOProvider{},
        &{name}provider.{Name}Provider{},  // add here
    )
}
```

### Step 7: Tests (`provider-guide.md` Step 7)

Write contract tests for the provider interface + unit tests for each component:
- Provider interface compliance
- Adapter round-trip (API → K8s → API preserves data)
- Client HTTP behavior (use httptest)
- Command integration (flag parsing, output format)

### Gate: Stage Complete

```bash
make build    # Binary compiles
make tests    # All tests pass with race detection
make lint     # No lint errors
grafanactl providers   # New provider listed
grafanactl config view # Secrets redacted correctly
```

---

## Phase 4: Verify

> **Checklist**: `agent-docs/design-guide.md` Section 7 + `agent-docs/provider-guide.md` Checklist

Run through both checklists:

### Interface Compliance
- [ ] All five `Provider` methods implemented
- [ ] `Name()` lowercase, unique, stable
- [ ] All config keys declared in `ConfigKeys()`
- [ ] Secret keys have `Secret: true`
- [ ] `Validate()` returns actionable error with `config set` command
- [ ] Provider in `allProviders()` in `cmd/grafanactl/root/command.go`

### UX Compliance
- [ ] All data-display commands support `-o json/yaml`
- [ ] List/get register `text` table codec as default
- [ ] Error messages include actionable suggestions
- [ ] No `os.Exit()` in command code
- [ ] Status messages use `cmdio` functions
- [ ] `--config` and `--context` inherited via persistent flags
- [ ] Help text: Short (verb, period-terminated), Long, Examples
- [ ] Push is idempotent
- [ ] Data fetching is format-agnostic
- [ ] PromQL uses `promql-builder` (if applicable)

### Build Verification
- [ ] `make build` succeeds
- [ ] `make tests` passes
- [ ] `make lint` passes
- [ ] `grafanactl providers` lists the new provider
- [ ] `grafanactl config view` redacts secrets

---

## Reference Implementation

The SLO provider was built as the Wave 1 reference implementation (PR #13).
Key files:

| Component | Path |
|-----------|------|
| Provider struct + configLoader | `internal/slo/provider.go` |
| Definitions commands | `internal/slo/definitions/commands.go` |
| API client | `internal/slo/definitions/client.go` |
| K8s adapter | `internal/slo/definitions/adapter.go` |
| Status (Prometheus hybrid) | `internal/slo/definitions/status.go` |
| Timeline (range query + graph) | `internal/slo/definitions/timeline.go` |
| Registration | `cmd/grafanactl/root/command.go` (`allProviders()`) |
| Top-level plan | `docs/designs/slo-provider/2026-03-04-slo-provider-plan.md` |
| Stage 1 design | `docs/designs/slo-provider/1-slo-definitions-crud/` |
| Stage 2 design | `docs/designs/slo-provider/2-reports-crud/` |

**What the SLO commit taught us** (from the actual implementation experience):
- Design docs were produced for each stage before coding began — the plan
  drove implementation, not the other way around
- The `configLoader` needed full env var resolution (`GRAFANA_TOKEN`,
  `GRAFANA_PROVIDER_SLO_*`) — not just flag binding
- Hand-rolling the HTTP client (~200 LOC) was cleaner than importing the
  OpenAPI-generated client (awkward types, poor adapter fit)
- Status commands require a hybrid pattern: REST API for resource data +
  Prometheus instant queries for live metrics
- K8s CRDs for SLO exist internally but are NOT accessible externally —
  verified by real API call, drove the plugin API choice

## Common Pitfalls

| Pitfall | Mitigation |
|---------|------------|
| Incomplete OpenAPI specs | Cross-reference with source code route handlers |
| K8s CRDs not externally accessible | Always verify with real API call before choosing K8s client path |
| readOnly fields in POST/PUT | Adapter must strip server-generated fields on Create/Update |
| Different list response envelopes | Define response types per product (no universal wrapper) |
| configLoader is non-trivial | Copy full implementation from `internal/slo/provider.go`, don't simplify |
| Registration in wrong place | Use `cmd/grafanactl/root/command.go`, not `internal/providers/registry.go` |
| Package path confusion | Provider code lives in `internal/{name}/`, not `internal/providers/{name}/` |
