package operation

import (
	"errors"
	"testing"

	gitea_errors "gitea.com/gitea/gitea-mcp/pkg/errors"
)

func TestErrorScenarios_NotFound(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		ctx        map[string]string
		isNotFound bool
	}{
		{
			name:       "404 file not found",
			errMsg:     "request failed with status 404: GetContents error",
			ctx:        map[string]string{"operation": "GetFile", "path": "README.md"},
			isNotFound: true,
		},
		{
			name:       "404 repo not found",
			errMsg:     "request failed with status 404: GetRepo error",
			ctx:        map[string]string{"operation": "GetRepo"},
			isNotFound: true,
		},
		{
			name:       "404 issue not found",
			errMsg:     "request failed with status 404: GetIssue error",
			ctx:        map[string]string{"operation": "GetIssue"},
			isNotFound: true,
		},
		{
			name:       "404 pull request not found",
			errMsg:     "request failed with status 404: GetPullRequest error",
			ctx:        map[string]string{"operation": "GetPullRequest"},
			isNotFound: true,
		},
		{
			name:       "401 auth error not found",
			errMsg:     "request failed with status 401: unauthorized",
			ctx:        map[string]string{"operation": "GetFile"},
			isNotFound: false,
		},
		{
			name:       "500 server error not found",
			errMsg:     "request failed with status 500: internal server error",
			ctx:        map[string]string{"operation": "GetFile"},
			isNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, tt.ctx)

			if gitea_errors.IsNotFound(translated) != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", gitea_errors.IsNotFound(translated), tt.isNotFound)
			}
		})
	}
}

func TestErrorScenarios_AuthErrors(t *testing.T) {
	tests := []struct {
		name      string
		errMsg    string
		isAuthErr bool
	}{
		{
			name:      "401 unauthorized",
			errMsg:    "request failed with status 401: unauthorized",
			isAuthErr: true,
		},
		{
			name:      "403 forbidden",
			errMsg:    "request failed with status 403: forbidden",
			isAuthErr: true,
		},
		{
			name:      "authentication failed message",
			errMsg:    "authentication failed",
			isAuthErr: true,
		},
		{
			name:      "permission denied message",
			errMsg:    "permission denied",
			isAuthErr: true,
		},
		{
			name:      "token error message",
			errMsg:    "check your access token",
			isAuthErr: true,
		},
		{
			name:      "404 not auth error",
			errMsg:    "request failed with status 404: not found",
			isAuthErr: false,
		},
		{
			name:      "500 not auth error",
			errMsg:    "request failed with status 500: server error",
			isAuthErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, nil)

			if gitea_errors.IsAuthError(translated) != tt.isAuthErr {
				t.Errorf("IsAuthError() = %v, want %v", gitea_errors.IsAuthError(translated), tt.isAuthErr)
			}
		})
	}
}

func TestErrorScenarios_NetworkErrors(t *testing.T) {
	tests := []struct {
		name          string
		errMsg        string
		isNetworkErr  bool
		isTimeout     bool
	}{
		{
			name:          "connection refused",
			errMsg:        "connection refused",
			isNetworkErr:  true,
			isTimeout:     false,
		},
		{
			name:          "no such host",
			errMsg:        "no such host example.com",
			isNetworkErr:  true,
			isTimeout:     false,
		},
		{
			name:          "network unreachable",
			errMsg:        "network unreachable",
			isNetworkErr:  true,
			isTimeout:     false,
		},
		{
			name:          "timeout error",
			errMsg:        "request timeout",
			isNetworkErr:  true,
			isTimeout:     true,
		},
		{
			name:          "deadline exceeded",
			errMsg:        "context deadline exceeded",
			isNetworkErr:  true,
			isTimeout:     true,
		},
		{
			name:          "dial tcp",
			errMsg:        "dial tcp: connection refused",
			isNetworkErr:  true,
			isTimeout:     false,
		},
		{
			name:          "not network error",
			errMsg:        "file not found",
			isNetworkErr:  false,
			isTimeout:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, nil)

			if gitea_errors.IsNetworkError(translated) != tt.isNetworkErr {
				t.Errorf("IsNetworkError() = %v, want %v", gitea_errors.IsNetworkError(translated), tt.isNetworkErr)
			}
			if gitea_errors.IsTimeout(translated) != tt.isTimeout {
				t.Errorf("IsTimeout() = %v, want %v", gitea_errors.IsTimeout(translated), tt.isTimeout)
			}
		})
	}
}

func TestErrorScenarios_ActionsAPIUnavailable(t *testing.T) {
	tests := []struct {
		name               string
		errMsg             string
		isActionsUnavailable bool
	}{
		{
			name:               "actions 404",
			errMsg:             "actions endpoint returned 404",
			isActionsUnavailable: true,
		},
		{
			name:               "actions 405",
			errMsg:             "actions endpoint returned 405",
			isActionsUnavailable: true,
		},
		{
			name:               "actions not found",
			errMsg:             "actions workflow not found",
			isActionsUnavailable: true,
		},
		{
			name:               "actions method not allowed",
			errMsg:             "actions method not allowed",
			isActionsUnavailable: true,
		},
		{
			name:               "actions enhanced error",
			errMsg:             "not supported on this Gitea version",
			isActionsUnavailable: true,
		},
		{
			name:               "other actions error",
			errMsg:             "actions completed successfully",
			isActionsUnavailable: false,
		},
		{
			name:               "file not found not actions",
			errMsg:             "file not found",
			isActionsUnavailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			if gitea_errors.IsActionsAPIUnavailable(err) != tt.isActionsUnavailable {
				t.Errorf("IsActionsAPIUnavailable() = %v, want %v", gitea_errors.IsActionsAPIUnavailable(err), tt.isActionsUnavailable)
			}
		})
	}
}

func TestErrorScenarios_HTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		isUnauthorized bool
		isForbidden    bool
		isNotFound     bool
		isServerError  bool
	}{
		{
			name:           "HTTP 401",
			statusCode:     401,
			isUnauthorized: true,
			isForbidden:    false,
			isNotFound:     false,
			isServerError:  false,
		},
		{
			name:           "HTTP 403",
			statusCode:     403,
			isUnauthorized: false,
			isForbidden:    true,
			isNotFound:     false,
			isServerError:  false,
		},
		{
			name:           "HTTP 404",
			statusCode:     404,
			isUnauthorized: false,
			isForbidden:    false,
			isNotFound:     true,
			isServerError:  false,
		},
		{
			name:           "HTTP 500",
			statusCode:     500,
			isUnauthorized: false,
			isForbidden:    false,
			isNotFound:     false,
			isServerError:  true,
		},
		{
			name:           "HTTP 502",
			statusCode:     502,
			isUnauthorized: false,
			isForbidden:    false,
			isNotFound:     false,
			isServerError:  true,
		},
		{
			name:           "HTTP 503",
			statusCode:     503,
			isUnauthorized: false,
			isForbidden:    false,
			isNotFound:     false,
			isServerError:  true,
		},
		{
			name:           "HTTP 200",
			statusCode:     200,
			isUnauthorized: false,
			isForbidden:    false,
			isNotFound:     false,
			isServerError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testHTTPError{status: tt.statusCode, message: "test error"}

			if gitea_errors.IsUnauthorized(err) != tt.isUnauthorized {
				t.Errorf("IsUnauthorized() = %v, want %v", gitea_errors.IsUnauthorized(err), tt.isUnauthorized)
			}
			if gitea_errors.IsForbidden(err) != tt.isForbidden {
				t.Errorf("IsForbidden() = %v, want %v", gitea_errors.IsForbidden(err), tt.isForbidden)
			}
			if gitea_errors.IsNotFoundHTTP(err) != tt.isNotFound {
				t.Errorf("IsNotFoundHTTP() = %v, want %v", gitea_errors.IsNotFoundHTTP(err), tt.isNotFound)
			}
			if gitea_errors.IsServerError(err) != tt.isServerError {
				t.Errorf("IsServerError() = %v, want %v", gitea_errors.IsServerError(err), tt.isServerError)
			}
		})
	}
}

func TestErrorScenarios_EnhancedErrorChaining(t *testing.T) {
	original := errors.New("original error")
	
	err := gitea_errors.TranslateError(original, map[string]string{
		"operation": "TestOp",
		"param1":    "value1",
	})

	if !errors.Is(err, original) {
		t.Error("enhanced error should wrap original")
	}

	var enhanced *gitea_errors.EnhancedError
	if !errors.As(err, &enhanced) {
		t.Fatal("should be able to extract EnhancedError")
	}

	if enhanced.Operation != "TestOp" {
		t.Errorf("operation = %q, want %q", enhanced.Operation, "TestOp")
	}

	if enhanced.Context["param1"] != "value1" {
		t.Errorf("context[param1] = %q, want %q", enhanced.Context["param1"], "value1")
	}

	formatted := enhanced.Format()
	if formatted == "" {
		t.Error("Format() should return non-empty string")
	}

	detailed := enhanced.FormatDetailed()
	if detailed == "" {
		t.Error("FormatDetailed() should return non-empty string")
	}
}

func TestErrorScenarios_FluentAPI(t *testing.T) {
	original := errors.New("test error")
	enhanced := gitea_errors.TranslateError(original, nil).(*gitea_errors.EnhancedError)

	result := enhanced.
		WithOperation("GetFile").
		WithParam("owner", "gitea").
		WithParam("repo", "tea").
		WithParam("path", "README.md")

	if result != enhanced {
		t.Error("fluent API should return same error for chaining")
	}

	if enhanced.Operation != "GetFile" {
		t.Errorf("operation = %q, want %q", enhanced.Operation, "GetFile")
	}

	if enhanced.Context["owner"] != "gitea" {
		t.Errorf("context[owner] = %q, want %q", enhanced.Context["owner"], "gitea")
	}
	if enhanced.Context["repo"] != "tea" {
		t.Errorf("context[repo] = %q, want %q", enhanced.Context["repo"], "tea")
	}
	if enhanced.Context["path"] != "README.md" {
		t.Errorf("context[path] = %q, want %q", enhanced.Context["path"], "README.md")
	}
}

type testHTTPError struct {
	status  int
	message string
}

func (e *testHTTPError) Error() string { return e.message }
func (e *testHTTPError) Status() int    { return e.status }

func TestErrorScenarios_CrossToolErrorConsistency(t *testing.T) {
	ctx := map[string]string{
		"owner":     "test-owner",
		"repo":      "test-repo",
		"operation": "CrossToolTest",
	}

	testCases := []struct {
		name     string
		errMsg   string
		category gitea_errors.ErrorCategory
	}{
		{
			name:     "file operation error",
			errMsg:   "GetContents failed",
			category: gitea_errors.CategoryFile,
		},
		{
			name:     "auth operation error",
			errMsg:   "GetUser failed",
			category: gitea_errors.CategoryAuth,
		},
		{
			name:     "repo operation error",
			errMsg:   "GetRepo failed",
			category: gitea_errors.CategoryRepo,
		},
		{
			name:     "issue operation error",
			errMsg:   "GetIssue failed",
			category: gitea_errors.CategoryIssue,
		},
		{
			name:     "pull request operation error",
			errMsg:   "GetPullRequest failed",
			category: gitea_errors.CategoryPull,
		},
		{
			name:     "branch operation error",
			errMsg:   "GetBranch failed",
			category: gitea_errors.CategoryBranch,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, ctx)

			var enhanced *gitea_errors.EnhancedError
			if !errors.As(translated, &enhanced) {
				t.Fatal("expected EnhancedError")
			}

			if enhanced.Category != tt.category {
				t.Errorf("category = %q, want %q", enhanced.Category, tt.category)
			}

			if enhanced.Context["owner"] != "test-owner" {
				t.Errorf("context[owner] not preserved")
			}
			if enhanced.Context["repo"] != "test-repo" {
				t.Errorf("context[repo] not preserved")
			}
		})
	}
}

func TestErrorScenarios_NilHandling(t *testing.T) {
	if gitea_errors.TranslateError(nil, nil) != nil {
		t.Error("TranslateError(nil) should return nil")
	}

	if gitea_errors.IsNotFound(nil) {
		t.Error("IsNotFound(nil) should return false")
	}

	if gitea_errors.IsAuthError(nil) {
		t.Error("IsAuthError(nil) should return false")
	}

	if gitea_errors.IsNetworkError(nil) {
		t.Error("IsNetworkError(nil) should return false")
	}

	if gitea_errors.IsTimeout(nil) {
		t.Error("IsTimeout(nil) should return false")
	}

	if gitea_errors.IsActionsAPIUnavailable(nil) {
		t.Error("IsActionsAPIUnavailable(nil) should return false")
	}

	if gitea_errors.IsServerError(nil) {
		t.Error("IsServerError(nil) should return false")
	}

	if gitea_errors.Wrap(nil, "operation") != nil {
		t.Error("Wrap(nil) should return nil")
	}
}

func TestErrorScenarios_ServerErrors(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		isServerErr bool
	}{
		{
			name:       "500 internal server error",
			errMsg:     "status 500",
			isServerErr: true,
		},
		{
			name:       "502 bad gateway",
			errMsg:     "status 502",
			isServerErr: true,
		},
		{
			name:       "503 service unavailable",
			errMsg:     "status 503",
			isServerErr: true,
		},
		{
			name:       "504 gateway timeout",
			errMsg:     "status 504",
			isServerErr: true,
		},
		{
			name:       "404 not server error",
			errMsg:     "status 404",
			isServerErr: false,
		},
		{
			name:       "200 not server error",
			errMsg:     "status 200",
			isServerErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			if gitea_errors.IsServerError(err) != tt.isServerErr {
				t.Errorf("IsServerError() = %v, want %v", gitea_errors.IsServerError(err), tt.isServerErr)
			}
		})
	}
}

func TestErrorScenarios_ErrorWrapping(t *testing.T) {
	original := errors.New("original error")
	
	wrapped := gitea_errors.Wrap(original, "GetFile")
	
	if !errors.Is(wrapped, original) {
		t.Error("wrapped error should contain original")
	}

	var enhanced *gitea_errors.EnhancedError
	if !errors.As(wrapped, &enhanced) {
		t.Fatal("wrapped should be EnhancedError")
	}

	if enhanced.Operation != "GetFile" {
		t.Errorf("operation = %q, want %q", enhanced.Operation, "GetFile")
	}
}

func TestErrorScenarios_ErrorUnwrap(t *testing.T) {
	original := errors.New("original error")
	enhanced := gitea_errors.NewEnhancedError(original, "translated", gitea_errors.CategoryFile)

	unwrapped := enhanced.Unwrap()
	if unwrapped != original {
		t.Error("Unwrap() should return original error")
	}

	if !errors.Is(enhanced, original) {
		t.Error("errors.Is should find original through unwrapping")
	}
}
