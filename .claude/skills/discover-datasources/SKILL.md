---
name: discover-datasources
description: Discover all the datasources in a Grafana instance, and what information is available in each one. Supports Prometheus (metrics, labels, metadata, targets) and Loki (labels, log streams/series). This is useful when the user needs to understand where to find the data they need. Use when the user asks "where is the data for this dashboard?" or "what datasources are available?" or "what metrics are available?" or "what log streams exist?" or "where can i find metrics/logs for this system?"
allowed-tools: grafanactl
---

# Datasource Discovery

## Instructions

### Step 1: List Available Datasources

Start by identifying all datasources in the Grafana instance.

```bash
# List all datasources
grafanactl datasources list

# Filter by type if you know what you need
grafanactl datasources list --type prometheus
grafanactl datasources list --type loki
```

**Expected output:** Table showing UID, NAME, TYPE, URL, and DEFAULT columns.

**Important:** Always use the UID (not the name) in subsequent commands.

### Step 2: Explore Datasource Contents

Choose the appropriate exploration path based on datasource type.

#### For Prometheus Datasources

```bash
# List all available labels
grafanactl datasources prometheus labels -d <datasource-uid>

# Get values for a specific label to understand what's being monitored
grafanactl datasources prometheus labels -d <datasource-uid> --label job

# List all available metrics with descriptions
grafanactl datasources prometheus metadata -d <datasource-uid>

# Check what systems are being scraped
grafanactl datasources prometheus targets -d <datasource-uid>
```

**Expected output:** Tables showing labels, metrics, or targets depending on command.

#### For Loki Datasources

```bash
# List all available labels
grafanactl datasources loki labels -d <datasource-uid>

# Get values for a specific label
grafanactl datasources loki labels -d <datasource-uid> --label job

# List log streams matching a selector (required)
grafanactl datasources loki series -d <datasource-uid> -M '{job="varlogs"}'
```

**Expected output:** Tables showing labels or log stream series.

**Note:** The `series` command requires at least one `-M` (match) selector using LogQL syntax.

### Step 3: Summarize Discovery and Guide Next Steps

After exploring datasources, provide a **high-level summary** (not exhaustive lists) and ask targeted questions to help the user find what they need.

**Keep summaries concise:**
- Datasource types and what they contain (metrics vs logs)
- Key label dimensions available (cluster, namespace, app, job, etc.) with 2-3 example values
- High-level categories (e.g., "HTTP metrics", "Kubernetes logs", "application metrics")

**Be interactive - ask targeted questions based on what exists:**
- "I see telemetry from clusters: [cluster-a, cluster-b, cluster-c]. Are you looking for a specific cluster?"
- "Available apps: [app-1, app-2, app-3]. Are you looking for one of these?"
- "I found HTTP metrics with labels: code, handler, instance. Would you like example queries for monitoring HTTP traffic?"

**For metrics**: Focus on metric names, types (counter/gauge/histogram), and available labels. Skip infrastructure details like API endpoints.

**For logs**: Focus on label combinations that identify log streams (job, namespace, cluster, app). Skip formatting details.

**Only suggest things that actually exist** in their datasources - don't make generic suggestions.

**Keep formatting simple** - avoid excessive markdown, headers, or visual formatting. Focus on clear, concise text.

The goal is helping users understand what telemetry exists and where to find data for specific systems/apps/contexts.

## Examples

### Example 1: Finding HTTP metrics

**User says:** "What HTTP metrics are available?"

**Actions:**
1. List Prometheus datasources: `grafanactl datasources list --type prometheus`
2. Get datasource UID from output
3. Search for HTTP metrics: `grafanactl datasources prometheus metadata -d <uid> -o json | jq '.data | to_entries[] | select(.key | contains("http"))'`
4. List labels for one key metric: `grafanactl datasources prometheus labels -d <uid>`

**Concise summary:**
"Found 3 HTTP metrics:
- http_requests_total (counter) - labels: code, handler, job, instance
- http_request_duration_seconds (histogram) - labels: handler, job
- http_response_size_bytes (histogram) - labels: handler, job

Available jobs: [grafana, prometheus, node-exporter]"

**Interactive follow-up:**
"Are you looking for HTTP metrics for a specific service (job)? Or would you like example queries for request rates, error rates, or latency?"

**Result:** User gets actionable summary and targeted next-step questions.

### Example 2: Discovering which services are logging to Loki

**User says:** "What applications are sending logs to Loki?"

**Actions:**
1. List Loki datasources: `grafanactl datasources list --type loki`
2. Get datasource UID from output
3. Get job values: `grafanactl datasources loki labels -d <uid> --label job`
4. Optionally check other useful labels: `grafanactl datasources loki labels -d <uid> --label namespace`

**Concise summary:**
"Found logs from 15 applications across 3 clusters:

Jobs: grafana, loki, mimir, prometheus, node-exporter, app-server, api-gateway, ...
Clusters: dev-us-east-0, dev-eu-west-2, prod-us-central
Namespaces: default, monitoring, applications"

**Interactive follow-up:**
"Are you looking for logs from a specific application? Or a specific cluster/namespace?"

**Result:** User gets overview of what's logging and can drill down based on their needs.

### Example 3: Troubleshooting missing dashboard data

**User says:** "My dashboard shows no data for service X"

**Actions:**
1. Check available jobs: `grafanactl datasources prometheus labels -d <uid> --label job`
2. Check scrape targets: `grafanactl datasources prometheus targets -d <uid>`

**Concise summary:**
"Service X is not in the job labels. Available services: [service-a, service-b, service-c]

Scrape targets show 12 active endpoints, none matching 'service-x'."

**Interactive follow-up:**
"Did you mean one of these services? Or is 'service-x' expected but not being scraped yet?"

**Result:** User quickly understands the service isn't monitored and sees what is available.

### Example 4: Finding logs for a specific namespace

**User says:** "Show me all log streams from the production namespace"

**Actions:**
1. Get Loki datasource UID: `grafanactl datasources list --type loki`
2. Check namespace values: `grafanactl datasources loki labels -d <uid> --label namespace`
3. Get apps in that namespace: `grafanactl datasources loki labels -d <uid> --label app` (to understand what's there)

**Concise summary:**
"Found 8 applications logging in 'production' namespace:
- api-gateway, auth-service, billing, user-service, notification-service, ...

Also available: cluster, pod, container labels for filtering"

**Interactive follow-up:**
"Would you like to see logs from a specific app in production? Or example queries for aggregating across all production services?"

**Result:** User understands what's logging in production and can narrow down to specific apps.

## Troubleshooting

### Error: "datasource UID is required"

**Cause:** The `-d` flag was omitted and no default datasource is configured.

**Solution:**
```bash
# Option 1: Pass the UID explicitly
grafanactl datasources prometheus labels -d <datasource-uid>

# Option 2: Set a default datasource
grafanactl config set contexts.<context-name>.default-prometheus-datasource <uid>
```

### Error: "at least one --match selector is required"

**Cause:** The `loki series` command was called without a `-M` flag.

**Solution:** Loki series requires at least one LogQL selector:
```bash
# Correct
grafanactl datasources loki series -d <uid> -M '{job="varlogs"}'

# Wrong - will fail
grafanactl datasources loki series -d <uid>
```

### Error: "parse error on line 1, column X: bare \" in non-quoted-field"

**Cause:** Shell is interpreting quotes in the LogQL selector incorrectly.

**Solution:** Use single quotes around the entire selector:
```bash
# Correct - single quotes outside
grafanactl datasources loki series -d <uid> -M '{name="value", cluster="prod"}'

# Wrong - shell interprets quotes incorrectly
grafanactl datasources loki series -d <uid> -M {name="value"}
```

### Error: "datasource.prometheus.datasource.grafana.app \"<uid>\" not found"

**Cause:** The datasource UID doesn't exist or you don't have access to it.

**Solution:**
1. List datasources to verify UID: `grafanactl datasources list`
2. Check you're using the correct context: `grafanactl config current-context`
3. Verify datasource exists: `grafanactl datasources get <uid>`

### No output from labels/series commands

**Cause:** Datasource has no data or hasn't scraped/ingested anything yet.

**Solution:**
1. For Prometheus: Check targets are active: `grafanactl datasources prometheus targets -d <uid>`
2. For Loki: Verify labels exist: `grafanactl datasources loki labels -d <uid>`
3. Check datasource URL is reachable: `grafanactl datasources get <uid>`

## Advanced Usage

For detailed patterns, LogQL syntax guide, and advanced discovery workflows, see:
- [`references/discovery-patterns.md`](references/discovery-patterns.md) - Common discovery patterns and workflows
- [`references/logql-syntax.md`](references/logql-syntax.md) - LogQL selector syntax guide

## Output Formats

All commands support `-o json` or `-o yaml` for programmatic use:

```bash
# Get JSON output for piping to jq
grafanactl datasources prometheus labels -d <uid> -o json

# Example: Count total metrics
grafanactl datasources prometheus metadata -d <uid> -o json | jq '.data | length'
```

Default output is `table` format for human readability.
