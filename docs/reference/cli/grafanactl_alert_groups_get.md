## grafanactl alert groups get

Get a single alert rule group.

### Synopsis

Get detailed information about a specific alert rule group by its name, including all rules in the group.

```
grafanactl alert groups get NAME [flags]
```

### Examples

```
	# Get alert rule group details
	grafanactl alert groups get "High Priority Alerts"

	# Output as JSON
	grafanactl alert groups get "CPU Alerts" -o json

	# Output as YAML
	grafanactl alert groups get "Memory Alerts" -o yaml
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

* [grafanactl alert groups](grafanactl_alert_groups.md)	 - Manage alert rule groups

