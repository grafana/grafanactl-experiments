## grafanactl alert groups list

List alert rule groups.

### Synopsis

List all alert rule groups configured in Grafana.

```
grafanactl alert groups list [flags]
```

### Examples

```
	# List all alert rule groups
	grafanactl alert groups list

	# Output as JSON
	grafanactl alert groups list -o json

	# Output as YAML
	grafanactl alert groups list -o yaml
```

### Options

```
  -h, --help            help for list
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

