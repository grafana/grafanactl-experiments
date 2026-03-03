package config

import (
	"os"
	"testing"
)

func TestLoadCLIOptions_AutoApproveTrue(t *testing.T) {
	// Set environment variable
	os.Setenv("GRAFANACTL_AUTO_APPROVE", "1")
	defer os.Unsetenv("GRAFANACTL_AUTO_APPROVE")

	opts, err := LoadCLIOptions()
	if err != nil {
		t.Fatalf("LoadCLIOptions() error = %v, want nil", err)
	}

	if !opts.AutoApprove {
		t.Errorf("AutoApprove = %v, want true", opts.AutoApprove)
	}
}

func TestLoadCLIOptions_AutoApproveTrueString(t *testing.T) {
	// Set environment variable
	os.Setenv("GRAFANACTL_AUTO_APPROVE", "true")
	defer os.Unsetenv("GRAFANACTL_AUTO_APPROVE")

	opts, err := LoadCLIOptions()
	if err != nil {
		t.Fatalf("LoadCLIOptions() error = %v, want nil", err)
	}

	if !opts.AutoApprove {
		t.Errorf("AutoApprove = %v, want true", opts.AutoApprove)
	}
}

func TestLoadCLIOptions_AutoApproveFalse(t *testing.T) {
	// Set environment variable
	os.Setenv("GRAFANACTL_AUTO_APPROVE", "0")
	defer os.Unsetenv("GRAFANACTL_AUTO_APPROVE")

	opts, err := LoadCLIOptions()
	if err != nil {
		t.Fatalf("LoadCLIOptions() error = %v, want nil", err)
	}

	if opts.AutoApprove {
		t.Errorf("AutoApprove = %v, want false", opts.AutoApprove)
	}
}

func TestLoadCLIOptions_AutoApproveEmpty(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("GRAFANACTL_AUTO_APPROVE")

	opts, err := LoadCLIOptions()
	if err != nil {
		t.Fatalf("LoadCLIOptions() error = %v, want nil", err)
	}

	if opts.AutoApprove {
		t.Errorf("AutoApprove = %v, want false (default)", opts.AutoApprove)
	}
}