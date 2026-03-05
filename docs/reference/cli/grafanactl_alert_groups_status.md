## grafanactl alert groups status

Show alert rule group status.

### Synopsis

Show the current status of alert rule groups including rule counts by state (firing, pending, inactive) and last evaluation time. If a name is provided, shows status for that specific group; otherwise shows status for all groups.

```
grafanactl alert groups status [NAME] [flags]
```

### Examples

```
	# Show status for all alert rule groups
	grafanactl alert groups status

	# Show status for a specific group
	grafanactl alert groups status "High Priority Alerts"

	# Output as JSON
	grafanactl alert groups status -o json
```

### Options

```
  -h, --help            help for status
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

