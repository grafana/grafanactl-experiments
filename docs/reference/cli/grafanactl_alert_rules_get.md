## grafanactl alert rules get

Get a single alert rule.

### Synopsis

Get detailed information about a specific alert rule by its UID.

```
grafanactl alert rules get UID [flags]
```

### Examples

```
	# Get alert rule details
	grafanactl alert rules get abc123xyz

	# Output as JSON
	grafanactl alert rules get abc123xyz -o json

	# Output as YAML
	grafanactl alert rules get abc123xyz -o yaml
```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of: json, table, yaml (default "table")
```

### Options inherited from parent commands

```
      --config string    Path to the configuration file to use
      --context string   Name of the context to use
      --no-color         Disable color output
  -v, --verbose count    Verbose mode. Multiple -v options increase the verbosity (maximum: 3).
```

### SEE ALSO

* [grafanactl alert rules](grafanactl_alert_rules.md)	 - Manage alert rules

