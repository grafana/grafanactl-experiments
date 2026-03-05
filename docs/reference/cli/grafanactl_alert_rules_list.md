## grafanactl alert rules list

List alert rules.

### Synopsis

List all alert rules configured in Grafana. Results can be filtered by group, folder, or state.

```
grafanactl alert rules list [flags]
```

### Examples

```
	# List all alert rules
	grafanactl alert rules list

	# List rules in a specific group
	grafanactl alert rules list --group "High Priority"

	# List rules in a specific folder
	grafanactl alert rules list --folder abc123

	# List only firing rules
	grafanactl alert rules list --status firing

	# List pending rules in a specific group
	grafanactl alert rules list --group "CPU Alerts" --status pending

	# Output as JSON
	grafanactl alert rules list -o json
```

### Options

```
      --folder string   Filter by folder UID
      --group string    Filter by group name
  -h, --help            help for list
  -o, --output string   Output format. One of: json, table, yaml (default "table")
      --status string   Filter by rule state (firing, pending, inactive)
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

