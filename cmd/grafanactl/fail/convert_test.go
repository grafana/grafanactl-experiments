package fail

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sapi "k8s.io/apimachinery/pkg/api/errors"
)

func TestConvertContextCanceled(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantMatch    bool
		wantExitCode int
	}{
		{
			name:         "bare context.Canceled returns ExitCancelled",
			err:          context.Canceled,
			wantMatch:    true,
			wantExitCode: ExitCancelled,
		},
		{
			name:         "wrapped context.Canceled returns ExitCancelled",
			err:          fmt.Errorf("operation failed: %w", context.Canceled),
			wantMatch:    true,
			wantExitCode: ExitCancelled,
		},
		{
			name:      "non-canceled error returns nil",
			err:       fmt.Errorf("some other error"),
			wantMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, matched := convertContextCanceled(tc.err)

			assert.Equal(t, tc.wantMatch, matched)

			if tc.wantMatch {
				require.NotNil(t, got)
				require.NotNil(t, got.ExitCode)
				assert.Equal(t, tc.wantExitCode, *got.ExitCode)
				assert.Equal(t, "Operation cancelled", got.Summary)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestConvertAPIErrors_AuthExitCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantExitCode int
	}{
		{
			name: "401 Unauthorized returns ExitAuthFailure",
			err: &k8sapi.StatusError{
				ErrStatus: metav1.Status{
					Status:  metav1.StatusFailure,
					Code:    401,
					Reason:  metav1.StatusReasonUnauthorized,
					Message: "Unauthorized",
				},
			},
			wantExitCode: ExitAuthFailure,
		},
		{
			name: "403 Forbidden returns ExitAuthFailure",
			err: &k8sapi.StatusError{
				ErrStatus: metav1.Status{
					Status:  metav1.StatusFailure,
					Code:    403,
					Reason:  metav1.StatusReasonForbidden,
					Message: "Forbidden",
				},
			},
			wantExitCode: ExitAuthFailure,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, matched := convertAPIErrors(tc.err)

			require.True(t, matched)
			require.NotNil(t, got)
			require.NotNil(t, got.ExitCode, "ExitCode should be set for auth errors")
			assert.Equal(t, tc.wantExitCode, *got.ExitCode)
		})
	}
}

func TestErrorToDetailedError_ConverterOrdering(t *testing.T) {
	// A context.Canceled wrapping a 401 error should be classified as
	// cancelled (exit 5), not as auth failure (exit 3), because the
	// cancellation converter runs first in the chain.
	unauthorizedErr := &k8sapi.StatusError{
		ErrStatus: metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    401,
			Reason:  metav1.StatusReasonUnauthorized,
			Message: "Unauthorized",
		},
	}
	wrappedErr := fmt.Errorf("request failed: %w: %w", context.Canceled, unauthorizedErr)

	got := ErrorToDetailedError(wrappedErr)

	require.NotNil(t, got)
	require.NotNil(t, got.ExitCode, "ExitCode should be set")
	assert.Equal(t, ExitCancelled, *got.ExitCode, "context.Canceled should take precedence over auth errors")
}

func TestErrorToDetailedError_UnrecognizedError(t *testing.T) {
	// An error that matches no specific converter should return a
	// DetailedError with nil ExitCode (caller defaults to 1).
	err := fmt.Errorf("something completely unexpected")

	got := ErrorToDetailedError(err)

	require.NotNil(t, got)
	assert.Nil(t, got.ExitCode, "unrecognized errors should have nil ExitCode")
	assert.Equal(t, "Unexpected error", got.Summary)
}
