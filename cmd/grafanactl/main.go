package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/grafana/grafanactl/cmd/grafanactl/fail"
	"github.com/grafana/grafanactl/cmd/grafanactl/root"
)

// Version variables which are set at build time.
var (
	version string
	//nolint:gochecknoglobals
	commit string
	//nolint:gochecknoglobals
	date string
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	handleError(root.Command(formatVersion()).ExecuteContext(ctx))
}

func handleError(err error) {
	if err == nil {
		return
	}

	// Fast-path: context cancellation (e.g., SIGINT).
	// Skip detailed error formatting — exit cleanly and quickly.
	if errors.Is(err, context.Canceled) {
		os.Exit(fail.ExitCancelled)
	}

	exitCode := 1
	detailedErr := fail.ErrorToDetailedError(err)

	if detailedErr != nil {
		fmt.Fprintln(os.Stderr, detailedErr.Error())

		if detailedErr.ExitCode != nil {
			exitCode = *detailedErr.ExitCode
		}
	}

	os.Exit(exitCode)
}

func formatVersion() string {
	if version == "" {
		version = "SNAPSHOT"
	}

	return fmt.Sprintf("%s built from %s on %s", version, commit, date)
}
