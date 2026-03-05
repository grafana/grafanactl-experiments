---
name: grafanactl
description: Manage Grafana resources (dashboards, folders, alerts) from the command line using kubectl-style patterns. Supports GitOps workflows, multi-environment promotion, and dashboards-as-code. Use when the user wants to pull/push/manage Grafana resources, set up contexts, or work with dashboards in version control. For datasource discovery and querying, use the discover-datasources skill instead.
---

# grafanactl

Manage Grafana resources from the command line. This skill focuses on configuration, resource management (dashboards/folders), and GitOps workflows.

# Instructions

## Scope

This skill covers:
- Configuring Grafana contexts (connection settings, auth)
- Listing available resource types
- Pulling resources to local files for version control
- Pushing resources back to Grafana
- GitOps workflows and multi-environment promotion
- Dashboard development with live reload

**Not covered here**: Datasource discovery, querying metrics/logs, exploring telemetry → use the **discover-datasources** skill for that.

## Approach

Explain the workflow and commands clearly. If you can't execute commands directly, provide clear instructions and examples so the user can run them. Focus on helping the user understand the grafanactl patterns and accomplish their goal.

**Command examples:**
- Assume `grafanactl` is on the user's PATH (use `grafanactl`, not `./bin/grafanactl`)
- If working in the grafanactl repo itself, note that they may need to use `./bin/grafanactl` or install it first

**When querying metrics/logs:**
- If the user's query doesn't include labels, suggest useful labels they could add (only suggest labels that actually exist in the datasource)
- Help them understand how to filter and aggregate their data effectively

### Step 1: Configure Context

Before interacting with Grafana, set up a context with your instance credentials:

```bash
# For Grafana Cloud (auto-discovers stack ID)
grafanactl config set contexts.mystack.server https://mystack.grafana.net
grafanactl config set contexts.mystack.auth.type token
grafanactl config set contexts.mystack.auth.token <api-token>
grafanactl config use-context mystack

# For on-premise Grafana
grafanactl config set contexts.local.server http://localhost:3000
grafanactl config set contexts.local.auth.type basic
grafanactl config set contexts.local.auth.username admin
grafanactl config set contexts.local.auth.password admin
grafanactl config set contexts.local.namespace 1  # org-id
grafanactl config use-context local
```

**Expected output:** "Switched to context 'mystack'"

Verify your configuration:
```bash
grafanactl config check
```

### Step 2: Discover Available Resources

List what resource types are available in your Grafana instance:

```bash
grafanactl resources list
```

**Expected output:** Table of available resources (dashboards, folders, datasources, alerts, etc.)

### Step 3: Pull Resources

Fetch resources from Grafana to local files:

```bash
# Pull all dashboards and folders
grafanactl resources pull dashboards folders -p ./grafana-resources

# Pull specific resources by UID
grafanactl resources pull dashboards/my-dashboard-uid
```

**Expected output:** Resources saved to `./grafana-resources/` directory organized by kind

### Step 4: Push or Modify Resources

Make changes and push back to Grafana:

```bash
# Preview changes first
grafanactl resources push --dry-run

# Push to Grafana
grafanactl resources push

# Edit resource interactively
grafanactl resources edit dashboards/my-dashboard-uid
```

**Expected output:** Summary of created/updated resources

# Examples

## Example 1: GitOps Workflow

**User says:** "Pull production dashboards, commit to git, and deploy to staging"

**Actions:**
1. Switch to production context: `grafanactl config use-context production`
2. Pull dashboards: `grafanactl resources pull dashboards -p ./prod-dashboards`
3. Commit to git: `git add prod-dashboards && git commit -m "Snapshot prod dashboards"`
4. Switch to staging: `grafanactl config use-context staging`
5. Deploy: `grafanactl resources push -p ./prod-dashboards --dry-run` then `grafanactl resources push -p ./prod-dashboards`

**Result:** Dashboards from production are now in staging, with full git history

## Example 2: Dashboard Development with Live Reload

**User says:** "I want to develop dashboards locally with live preview"

**Actions:**
1. Create dashboard generator script (Go, Python, etc.)
2. Start serve command: `grafanactl resources serve --script 'go run dashboard-gen/*.go' --watch ./dashboard-gen --script-format yaml`
3. Open http://localhost:8080 in browser
4. Edit generator code - changes appear immediately

**Result:** Local Grafana instance with live reload, no need to push to remote instance

## Example 3: Multi-Environment Promotion

**User says:** "Promote dashboards from dev to staging to production"

**Actions:**
1. Configure contexts for all environments (dev, staging, prod)
2. Pull from dev: `grafanactl config use-context dev && grafanactl resources pull -p ./dashboards`
3. Push to staging: `grafanactl config use-context staging && grafanactl resources push -p ./dashboards --dry-run && grafanactl resources push -p ./dashboards`
4. After validation, push to prod: `grafanactl config use-context prod && grafanactl resources push -p ./dashboards --dry-run && grafanactl resources push -p ./dashboards`

**Result:** Controlled promotion through environments with dry-run validation

# Troubleshooting

## Error: "resource not managed by grafanactl"

**Cause:** Trying to modify a resource created by Grafana UI or another tool

**Solution:**
- To view only: No action needed
- To modify: Add `--include-managed` flag (use with caution - may conflict with other managers)
```bash
grafanactl resources push --include-managed
```

## Error: "Connection refused" or "Connection timeout"

**Cause:** Cannot reach Grafana server

**Solution:**
1. Verify server URL: `grafanactl config view`
2. Check network connectivity: `curl -I <grafana-server-url>`
3. Verify authentication: `grafanactl config check`
4. For Grafana Cloud, ensure stack ID is correct (auto-discovered from URL)

## Error: "401 Unauthorized" or "403 Forbidden"

**Cause:** Invalid credentials or insufficient permissions

**Solution:**
1. Verify token is valid: Check in Grafana UI under Service Accounts
2. For basic auth, verify username/password
3. Ensure token has required permissions (Editor or Admin for push operations)
4. Re-configure auth: `grafanactl config set contexts.mycontext.auth.token <new-token>`

## Error: "Folder not found" when pushing dashboards

**Cause:** Dashboard references a folder that doesn't exist yet

**Solution:** grafanactl automatically handles folder dependencies by pushing folders first. If this fails:
1. Manually create folder: Push folder resource first
2. Check folder UID in dashboard spec matches existing folder
3. Use `--dry-run` to preview what will be created

## Performance: Slow pulls with many resources

**Cause:** Pulling thousands of resources can be slow

**Solution:**
1. Filter to specific resource types: `grafanactl resources pull dashboards` (not all resources)
2. Use selectors to pull specific resources: `grafanactl resources pull dashboards/uid1,uid2`
3. Adjust concurrency if needed (default is 10 concurrent operations)

## Serve command: "Feature toggle kubernetesDashboards not enabled"

**Cause:** Grafana instance doesn't have required feature flag enabled

**Solution:**
Enable feature flag in Grafana configuration:
```ini
[feature_toggles]
enable = kubernetesDashboards
```
Or use environment variable: `GF_FEATURE_TOGGLES_ENABLE=kubernetesDashboards`

---

**For detailed reference documentation, see:**
- `references/resource-model.md` - Grafana resource relationships and dependencies
- `references/query-patterns.md` - Advanced query patterns and examples
- `references/configuration.md` - Complete configuration options
- `references/selectors.md` - Resource selector syntax guide
