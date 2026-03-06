package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetState clears all package-level variables and re-runs environment
// detection. Tests call this after setting env vars via t.Setenv so that
// the package state reflects the current environment.
func resetState() {
	detectFromEnv()
}

func TestIsAgentMode(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		setFlag  *bool // nil = don't call SetFlag
		wantMode bool
	}{
		{
			name:     "no env vars set",
			wantMode: false,
		},
		{
			name:     "CLAUDE_CODE=1",
			envVars:  map[string]string{"CLAUDE_CODE": "1"},
			wantMode: true,
		},
		{
			name:     "CURSOR_AGENT=true",
			envVars:  map[string]string{"CURSOR_AGENT": "true"},
			wantMode: true,
		},
		{
			name:     "GITHUB_COPILOT=yes",
			envVars:  map[string]string{"GITHUB_COPILOT": "yes"},
			wantMode: true,
		},
		{
			name:     "AMAZON_Q=1",
			envVars:  map[string]string{"AMAZON_Q": "1"},
			wantMode: true,
		},
		{
			name:     "GRAFANACTL_AGENT_MODE=1",
			envVars:  map[string]string{"GRAFANACTL_AGENT_MODE": "1"},
			wantMode: true,
		},
		{
			name: "GRAFANACTL_AGENT_MODE=0 overrides CLAUDE_CODE=1",
			envVars: map[string]string{
				"GRAFANACTL_AGENT_MODE": "0",
				"CLAUDE_CODE":          "1",
			},
			wantMode: false,
		},
		{
			name: "GRAFANACTL_AGENT_MODE=false overrides CURSOR_AGENT=1",
			envVars: map[string]string{
				"GRAFANACTL_AGENT_MODE": "false",
				"CURSOR_AGENT":         "1",
			},
			wantMode: false,
		},
		{
			name:     "GRAFANACTL_AGENT_MODE=no disables agent mode",
			envVars:  map[string]string{"GRAFANACTL_AGENT_MODE": "no"},
			wantMode: false,
		},
		{
			name:     "SetFlag(true) with no env vars enables agent mode",
			setFlag:  boolPtr(true),
			wantMode: true,
		},
		{
			name:     "SetFlag(false) does not disable when env detected",
			envVars:  map[string]string{"CLAUDE_CODE": "1"},
			setFlag:  boolPtr(false),
			wantMode: true,
		},
		{
			name:     "case insensitive truthy: CLAUDE_CODE=TRUE",
			envVars:  map[string]string{"CLAUDE_CODE": "TRUE"},
			wantMode: true,
		},
		{
			name:     "case insensitive truthy: CLAUDE_CODE=Yes",
			envVars:  map[string]string{"CLAUDE_CODE": "Yes"},
			wantMode: true,
		},
		{
			name:     "case insensitive falsy: GRAFANACTL_AGENT_MODE=FALSE",
			envVars:  map[string]string{"GRAFANACTL_AGENT_MODE": "FALSE"},
			wantMode: false,
		},
		{
			name:     "case insensitive falsy: GRAFANACTL_AGENT_MODE=No",
			envVars:  map[string]string{"GRAFANACTL_AGENT_MODE": "No"},
			wantMode: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Setenv automatically restores env after the subtest.
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			resetState()

			if tc.setFlag != nil {
				SetFlag(*tc.setFlag)
			}

			assert.Equal(t, tc.wantMode, IsAgentMode())
		})
	}
}

func TestDetectedFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		setFlag  *bool
		wantEnv  bool
	}{
		{
			name:    "returns true when env var set",
			envVars: map[string]string{"CLAUDE_CODE": "1"},
			wantEnv: true,
		},
		{
			name:    "returns false when only SetFlag used",
			setFlag: boolPtr(true),
			wantEnv: false,
		},
		{
			name:    "returns false when no detection at all",
			wantEnv: false,
		},
		{
			name:    "returns false when GRAFANACTL_AGENT_MODE=0",
			envVars: map[string]string{"GRAFANACTL_AGENT_MODE": "0"},
			wantEnv: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			resetState()

			if tc.setFlag != nil {
				SetFlag(*tc.setFlag)
			}

			assert.Equal(t, tc.wantEnv, DetectedFromEnv())
		})
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1", true},
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"0", false},
		{"false", false},
		{"no", false},
		{"", false},
		{"random", false},
		{"enabled", false},
	}

	for _, tc := range tests {
		t.Run("isTruthy("+tc.input+")", func(t *testing.T) {
			assert.Equal(t, tc.want, isTruthy(tc.input))
		})
	}
}

func TestIsFalsy(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"0", true},
		{"false", true},
		{"False", true},
		{"FALSE", true},
		{"no", true},
		{"No", true},
		{"NO", true},
		{"1", false},
		{"true", false},
		{"yes", false},
		{"", false},
		{"random", false},
		{"disabled", false},
	}

	for _, tc := range tests {
		t.Run("isFalsy("+tc.input+")", func(t *testing.T) {
			assert.Equal(t, tc.want, isFalsy(tc.input))
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
