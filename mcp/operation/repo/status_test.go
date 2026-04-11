package repo

import (
	"errors"
	"testing"

	gitea_errors "gitea.com/gitea/gitea-mcp/pkg/errors"
	gitea_sdk "code.gitea.io/sdk/gitea"
)

func TestParseStatusState(t *testing.T) {
	tests := []struct {
		name        string
		state       string
		wantState   gitea_sdk.StatusState
		wantErr     bool
		errContains string
	}{
		{
			name:      "pending",
			state:     "pending",
			wantState: gitea_sdk.StatusPending,
			wantErr:   false,
		},
		{
			name:      "success",
			state:     "success",
			wantState: gitea_sdk.StatusSuccess,
			wantErr:   false,
		},
		{
			name:      "failure",
			state:     "failure",
			wantState: gitea_sdk.StatusFailure,
			wantErr:   false,
		},
		{
			name:      "error",
			state:     "error",
			wantState: gitea_sdk.StatusError,
			wantErr:   false,
		},
		{
			name:        "invalid state",
			state:       "invalid",
			wantErr:     true,
			errContains: "invalid state",
		},
		{
			name:        "empty state",
			state:       "",
			wantErr:     true,
			errContains: "invalid state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStatusState(tt.state)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseStatusState() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !errors.Is(err, errors.New(tt.errContains)) {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("parseStatusState() error = %v, should contain %v", err.Error(), tt.errContains)
					}
				}
				return
			}
			if err != nil {
				t.Errorf("parseStatusState() unexpected error = %v", err)
				return
			}
			if got != tt.wantState {
				t.Errorf("parseStatusState() = %v, want %v", got, tt.wantState)
			}
		})
	}
}

func TestSlimStatus(t *testing.T) {
	tests := []struct {
		name string
		status *gitea_sdk.Status
		want map[string]any
	}{
		{
			name:   "nil status",
			status: nil,
			want:   nil,
		},
		{
			name: "full status",
			status: &gitea_sdk.Status{
				ID:          123,
				State:       gitea_sdk.StatusSuccess,
				TargetURL:   "https://review.lumbridgecorp.com/project/commit/248ade7",
				Context:     "ci/metal",
				Description: "Build succeeded on Metal",
				CreatedAt:   "2024-01-15T10:30:00Z",
			},
			want: map[string]any{
				"id":          int64(123),
				"state":       gitea_sdk.StatusSuccess,
				"target_url":  "https://review.lumbridgecorp.com/project/commit/248ade7",
				"context":     "ci/metal",
				"description": "Build succeeded on Metal",
				"created_at":  "2024-01-15T10:30:00Z",
			},
		},
		{
			name: "pending status",
			status: &gitea_sdk.Status{
				ID:        456,
				State:     gitea_sdk.StatusPending,
				Context:   "ci/cloud-1",
				CreatedAt: "2024-01-15T10:31:00Z",
			},
			want: map[string]any{
				"id":         int64(456),
				"state":      gitea_sdk.StatusPending,
				"target_url": "",
				"context":    "ci/cloud-1",
				"created_at": "2024-01-15T10:31:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slimStatus(tt.status)
			if tt.want == nil {
				if got != nil {
					t.Errorf("slimStatus() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("slimStatus() = nil, want %v", tt.want)
				return
			}
			for key, wantVal := range tt.want {
				gotVal, ok := got[key]
				if !ok {
					t.Errorf("slimStatus() missing key %s", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("slimStatus()[%s] = %v, want %v", key, gotVal, wantVal)
				}
			}
		})
	}
}

func TestErrorTranslation_CreateCommitStatus(t *testing.T) {
	tests := []struct {
		name           string
		errMsg         string
		expectedOp     string
		expectedCtxKey string
	}{
		{
			name:           "404 not found",
			errMsg:         "CreateStatus: 404 Not Found",
			expectedOp:     "CreateCommitStatus",
			expectedCtxKey: "sha",
		},
		{
			name:           "401 unauthorized",
			errMsg:         "CreateStatus: 401 Unauthorized",
			expectedOp:     "CreateCommitStatus",
			expectedCtxKey: "context",
		},
		{
			name:           "403 forbidden",
			errMsg:         "CreateStatus: 403 Forbidden",
			expectedOp:     "CreateCommitStatus",
			expectedCtxKey: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, map[string]string{
				"operation": "CreateCommitStatus",
				"owner":     "karti-ai",
				"repo":      "gitea-mcp",
				"sha":       "248ade7a9c...",
				"state":     "success",
				"context":   "ci/metal",
			})

			var enhanced *gitea_errors.EnhancedError
			if !errors.As(translated, &enhanced) {
				t.Fatal("expected translated error to be EnhancedError")
			}

			if enhanced.Operation != tt.expectedOp {
				t.Errorf("expected operation %s, got %s", tt.expectedOp, enhanced.Operation)
			}

			if enhanced.Context[tt.expectedCtxKey] == "" {
				t.Errorf("expected context key %s to be set", tt.expectedCtxKey)
			}

			t.Logf("Translated error: %s", enhanced.Error())
		})
	}
}

func TestParseStatusState_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		state       string
		wantState   gitea_sdk.StatusState
		wantErr     bool
		errContains string
	}{
		{
			name:        "mixed case pending",
			state:       "Pending",
			wantErr:     true,
			errContains: "invalid state",
		},
		{
			name:        "mixed case success",
			state:       "Success",
			wantErr:     true,
			errContains: "invalid state",
		},
		{
			name:        "whitespace pending",
			state:       " pending",
			wantErr:     true,
			errContains: "invalid state",
		},
		{
			name:        "whitespace success",
			state:       "success ",
			wantErr:     true,
			errContains: "invalid state",
		},
		{
			name:        "long invalid string",
			state:       "this_is_not_a_valid_state",
			wantErr:     true,
			errContains: "invalid state",
		},
		{
			name:        "numeric string",
			state:       "123",
			wantErr:     true,
			errContains: "invalid state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStatusState(tt.state)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseStatusState() error = nil, wantErr %v", tt.wantErr)
					return
				}
				return
			}
			if err != nil {
				t.Errorf("parseStatusState() unexpected error = %v", err)
				return
			}
			if got != tt.wantState {
				t.Errorf("parseStatusState() = %v, want %v", got, tt.wantState)
			}
		})
	}
}

func TestSlimStatus_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		status *gitea_sdk.Status
		want   map[string]any
	}{
		{
			name:   "nil status",
			status: nil,
			want:   nil,
		},
		{
			name: "status with zero ID",
			status: &gitea_sdk.Status{
				ID:        0,
				State:     gitea_sdk.StatusPending,
				TargetURL: "",
				Context:   "",
				CreatedAt: "2024-01-15T10:30:00Z",
			},
			want: map[string]any{
				"id":         int64(0),
				"state":      gitea_sdk.StatusPending,
				"target_url": "",
				"context":    "",
				"created_at": "2024-01-15T10:30:00Z",
			},
		},
		{
			name: "status with long URL",
			status: &gitea_sdk.Status{
				ID:          789,
				State:       gitea_sdk.StatusSuccess,
				TargetURL:   "https://very-long-review-environment-url.example.com/path/to/project/commit/248ade7a9c.../build/12345/logs?filter=all#section-2",
				Context:     "continuous-integration/jenkins/build-and-test-all-platforms",
				Description: "Build succeeded on all platforms including Windows, macOS, and Linux with full test suite",
				CreatedAt:   "2024-01-15T10:30:00Z",
			},
			want: map[string]any{
				"id":         int64(789),
				"state":      gitea_sdk.StatusSuccess,
				"target_url": "https://very-long-review-environment-url.example.com/path/to/project/commit/248ade7a9c.../build/12345/logs?filter=all#section-2",
				"context":    "continuous-integration/jenkins/build-and-test-all-platforms",
				"created_at": "2024-01-15T10:30:00Z",
			},
		},
		{
			name: "failure status",
			status: &gitea_sdk.Status{
				ID:          101,
				State:       gitea_sdk.StatusFailure,
				TargetURL:   "https://ci.example.com/build/101",
				Context:     "ci/build",
				Description: "Build failed",
				CreatedAt:   "2024-01-15T11:00:00Z",
			},
			want: map[string]any{
				"id":         int64(101),
				"state":      gitea_sdk.StatusFailure,
				"target_url": "https://ci.example.com/build/101",
				"context":    "ci/build",
				"created_at": "2024-01-15T11:00:00Z",
			},
		},
		{
			name: "error status",
			status: &gitea_sdk.Status{
				ID:          102,
				State:       gitea_sdk.StatusError,
				TargetURL:   "",
				Context:     "ci/error",
				Description: "Error occurred",
				CreatedAt:   "2024-01-15T11:01:00Z",
			},
			want: map[string]any{
				"id":         int64(102),
				"state":      gitea_sdk.StatusError,
				"target_url": "",
				"context":    "ci/error",
				"created_at": "2024-01-15T11:01:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slimStatus(tt.status)
			if tt.want == nil {
				if got != nil {
					t.Errorf("slimStatus() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("slimStatus() = nil, want %v", tt.want)
				return
			}
			for key, wantVal := range tt.want {
				gotVal, ok := got[key]
				if !ok {
					t.Errorf("slimStatus() missing key %s", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("slimStatus()[%s] = %v, want %v", key, gotVal, wantVal)
				}
			}
		})
	}
}

func TestErrorTranslation_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		ctx        map[string]string
		checkField string
		wantValue  string
	}{
		{
			name:       "timeout error",
			errMsg:     "request timeout",
			ctx:        map[string]string{"operation": "CreateCommitStatus"},
			checkField: "operation",
			wantValue:  "CreateCommitStatus",
		},
		{
			name:       "network error",
			errMsg:     "connection refused",
			ctx:        map[string]string{"operation": "CreateCommitStatus"},
			checkField: "operation",
			wantValue:  "CreateCommitStatus",
		},
		{
			name:       "500 server error",
			errMsg:     "500 Internal Server Error",
			ctx:        map[string]string{"operation": "CreateCommitStatus"},
			checkField: "operation",
			wantValue:  "CreateCommitStatus",
		},
		{
			name:       "rate limit error",
			errMsg:     "429 Too Many Requests",
			ctx:        map[string]string{"operation": "CreateCommitStatus"},
			checkField: "operation",
			wantValue:  "CreateCommitStatus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, tt.ctx)

			var enhanced *gitea_errors.EnhancedError
			if !errors.As(translated, &enhanced) {
				t.Fatal("expected translated error to be EnhancedError")
			}

			if enhanced.Context[tt.checkField] != tt.wantValue {
				t.Errorf("expected context[%s] = %s, got %s", tt.checkField, tt.wantValue, enhanced.Context[tt.checkField])
			}
		})
	}
}

func TestSlimStatusFieldTypes(t *testing.T) {
	status := &gitea_sdk.Status{
		ID:          int64(999),
		State:       gitea_sdk.StatusSuccess,
		TargetURL:   "https://example.com",
		Context:     "test",
		Description: "desc",
		CreatedAt:   "2024-01-15T10:30:00Z",
	}

	slimmed := slimStatus(status)

	if slimmed == nil {
		t.Fatal("slimStatus returned nil")
	}

	if id, ok := slimmed["id"].(int64); !ok {
		t.Errorf("id should be int64, got %T", slimmed["id"])
	} else if id != 999 {
		t.Errorf("id = %d, want 999", id)
	}

	if state, ok := slimmed["state"].(gitea_sdk.StatusState); !ok {
		t.Errorf("state should be StatusState, got %T", slimmed["state"])
	} else if state != gitea_sdk.StatusSuccess {
		t.Errorf("state = %v, want %v", state, gitea_sdk.StatusSuccess)
	}

	for _, key := range []string{"target_url", "context", "description", "created_at"} {
		if val, ok := slimmed[key].(string); !ok && slimmed[key] != nil {
			t.Errorf("%s should be string, got %T", key, slimmed[key])
		} else if !ok {
			t.Errorf("%s should not be nil", key)
		}
	}
}
