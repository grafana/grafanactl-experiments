## grafanactl alert rules status

Show alert rule status.

### Synopsis

Show the current status of alert rules including state, health, last evaluation time, and whether they are paused. If a UID is provided, shows status for that specific rule; otherwise shows status for all rules.

```
grafanactl alert rules status [UID] [flags]
```

### Examples

```
	# Show status for all alert rules
	grafanactl alert rules status

	# Show status for a specific rule
	grafanactl alert rules status abc123xyz

	# Show detailed status with wide output
	grafanactl alert rules status -o wide

	# Output as JSON
	grafanactl alert rules status -o json
```

### Options

```
  -h, --help            help for status
  -o, --output string   Output format. One of: json, table, wide, yaml (default "table")
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

