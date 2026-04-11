package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestEnhancedError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *EnhancedError
		expected string
	}{
		{
			name: "with translated message",
			err: &EnhancedError{
				Translated: "File not found",
				Original:   errors.New("original error"),
			},
			expected: "File not found",
		},
		{
			name: "without translated message uses original",
			err: &EnhancedError{
				Original: errors.New("original error"),
			},
			expected: "original error",
		},
		{
			name:     "empty error",
			err:      &EnhancedError{},
			expected: "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEnhancedError_Unwrap(t *testing.T) {
	original := errors.New("original error")
	enhanced := &EnhancedError{
		Original:   original,
		Translated: "translated error",
	}

	unwrapped := enhanced.Unwrap()
	if unwrapped != original {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, original)
	}

	if !errors.Is(enhanced, original) {
		t.Error("errors.Is should find the original error")
	}
}

func TestEnhancedError_WithContext(t *testing.T) {
	err := &EnhancedError{
		Translated: "File not found",
		Context:    make(map[string]string),
	}

	err.WithContext("owner", "gitea")
	err.WithContext("repo", "tea")

	if err.Context["owner"] != "gitea" {
		t.Errorf("Context['owner'] = %q, want %q", err.Context["owner"], "gitea")
	}
	if err.Context["repo"] != "tea" {
		t.Errorf("Context['repo'] = %q, want %q", err.Context["repo"], "tea")
	}
}

func TestEnhancedError_Format(t *testing.T) {
	err := &EnhancedError{
		Original:   errors.New("original error"),
		Translated: "File not found",
		Category:   CategoryFile,
		Operation:  "GetFile",
		Context: map[string]string{
			"owner": "gitea",
			"path":  "README.md",
		},
	}

	formatted := err.Format()

	if !strings.Contains(formatted, "Operation: GetFile") {
		t.Error("Format() should include operation")
	}
	if !strings.Contains(formatted, "Error: File not found") {
		t.Error("Format() should include error message")
	}
	if !strings.Contains(formatted, "Category: file") {
		t.Error("Format() should include category")
	}
	if !strings.Contains(formatted, "owner=gitea") {
		t.Error("Format() should include context")
	}
	if !strings.Contains(formatted, "Original: original error") {
		t.Error("Format() should include original error")
	}
}

func TestTranslateError_Nil(t *testing.T) {
	result := TranslateError(nil, nil)
	if result != nil {
		t.Error("TranslateError(nil) should return nil")
	}
}

func TestTranslateError_AlreadyEnhanced(t *testing.T) {
	original := errors.New("original")
	enhanced := TranslateError(original, map[string]string{"operation": "GetFile"})

	reEnhanced := TranslateError(enhanced, map[string]string{"owner": "gitea", "path": "README.md"})

	err, ok := reEnhanced.(*EnhancedError)
	if !ok {
		t.Fatal("Expected *EnhancedError")
	}

	if err.Operation != "GetFile" {
		t.Errorf("Operation = %q, want %q", err.Operation, "GetFile")
	}
	if err.Context["owner"] != "gitea" {
		t.Errorf("Context['owner'] = %q, want %q", err.Context["owner"], "gitea")
	}
	if err.Context["path"] != "README.md" {
		t.Errorf("Context['path'] = %q, want %q", err.Context["path"], "README.md")
	}
}

func TestTranslateError_Mappings(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantTranslated  string
		wantCategory    ErrorCategory
	}{
		{
			name:           "HTTP 404 with GetContents",
			input:          "request failed with status 404: GetContents error",
			wantTranslated: "File or directory not found",
			wantCategory:   CategoryFile,
		},
		{
			name:           "HTTP 404 with ListContents",
			input:          "request failed with status 404: ListContents error",
			wantTranslated: "Directory not found or empty",
			wantCategory:   CategoryFile,
		},
		{
			name:           "HTTP 404 with GetUserByName",
			input:          "request failed with status 404: GetUserByName error",
			wantTranslated: "User or organization not found",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "HTTP 404 with GetRepo",
			input:          "request failed with status 404: GetRepo error",
			wantTranslated: "Repository not found",
			wantCategory:   CategoryRepo,
		},
		{
			name:           "HTTP 404 with GetIssue",
			input:          "request failed with status 404: GetIssue error",
			wantTranslated: "Issue not found",
			wantCategory:   CategoryIssue,
		},
		{
			name:           "HTTP 404 with GetPullRequest",
			input:          "request failed with status 404: GetPullRequest error",
			wantTranslated: "Pull request not found",
			wantCategory:   CategoryPull,
		},
		{
			name:           "HTTP 404 with GetBranch",
			input:          "request failed with status 404: GetBranch error",
			wantTranslated: "Branch or tag not found",
			wantCategory:   CategoryBranch,
		},
		{
			name:           "HTTP 404 generic",
			input:          "request failed with status 404",
			wantTranslated: "Resource not found",
			wantCategory:   CategoryUnknown,
		},
		{
			name:           "HTTP 401",
			input:          "request failed with status 401: unauthorized",
			wantTranslated: "Authentication failed - check your access token",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "HTTP 403",
			input:          "request failed with status 403: forbidden",
			wantTranslated: "Permission denied - you don't have access to this resource",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "GetContents",
			input:          "GetContents failed",
			wantTranslated: "File or directory not found",
			wantCategory:   CategoryFile,
		},
		{
			name:           "GetContentsOrList",
			input:          "GetContentsOrList failed",
			wantTranslated: "File or directory not found",
			wantCategory:   CategoryFile,
		},
		{
			name:           "GetUserByName",
			input:          "GetUserByName failed",
			wantTranslated: "User or organization not found",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "GetUser",
			input:          "GetUser failed",
			wantTranslated: "User or organization not found",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "GetRepo",
			input:          "GetRepo failed",
			wantTranslated: "Repository not found",
			wantCategory:   CategoryRepo,
		},
		{
			name:           "GetIssue",
			input:          "GetIssue failed",
			wantTranslated: "Issue not found",
			wantCategory:   CategoryIssue,
		},
		{
			name:           "GetPullRequest",
			input:          "GetPullRequest failed",
			wantTranslated: "Pull request not found",
			wantCategory:   CategoryPull,
		},
		{
			name:           "GetBranch",
			input:          "GetBranch failed",
			wantTranslated: "Branch not found",
			wantCategory:   CategoryBranch,
		},
		{
			name:           "GetTag",
			input:          "GetTag failed",
			wantTranslated: "Tag not found",
			wantCategory:   CategoryBranch,
		},
		{
			name:           "CreateFile",
			input:          "CreateFile failed",
			wantTranslated: "Failed to create file - it may already exist",
			wantCategory:   CategoryFile,
		},
		{
			name:           "UpdateFile",
			input:          "UpdateFile failed",
			wantTranslated: "Failed to update file - it may not exist or SHA mismatch",
			wantCategory:   CategoryFile,
		},
		{
			name:           "DeleteFile",
			input:          "DeleteFile failed",
			wantTranslated: "Failed to delete file - it may not exist",
			wantCategory:   CategoryFile,
		},
		{
			name:           "CreateBranch",
			input:          "CreateBranch failed",
			wantTranslated: "Failed to create branch",
			wantCategory:   CategoryBranch,
		},
		{
			name:           "DeleteBranch",
			input:          "DeleteBranch failed",
			wantTranslated: "Failed to delete branch - it may not exist or be protected",
			wantCategory:   CategoryBranch,
		},
		{
			name:           "CreateIssue",
			input:          "CreateIssue failed",
			wantTranslated: "Failed to create issue",
			wantCategory:   CategoryIssue,
		},
		{
			name:           "EditIssue",
			input:          "EditIssue failed",
			wantTranslated: "Failed to update issue - it may not exist",
			wantCategory:   CategoryIssue,
		},
		{
			name:           "CreatePullRequest",
			input:          "CreatePullRequest failed",
			wantTranslated: "Failed to create pull request",
			wantCategory:   CategoryPull,
		},
		{
			name:           "EditPullRequest",
			input:          "EditPullRequest failed",
			wantTranslated: "Failed to update pull request",
			wantCategory:   CategoryPull,
		},
		{
			name:           "CreateRelease",
			input:          "CreateRelease failed",
			wantTranslated: "Failed to create release",
			wantCategory:   CategoryRepo,
		},
		{
			name:           "EditRelease",
			input:          "EditRelease failed",
			wantTranslated: "Failed to update release",
			wantCategory:   CategoryRepo,
		},
		{
			name:           "CreateWikiPage",
			input:          "CreateWikiPage failed",
			wantTranslated: "Failed to create wiki page",
			wantCategory:   CategoryRepo,
		},
		{
			name:           "EditWikiPage",
			input:          "EditWikiPage failed",
			wantTranslated: "Failed to update wiki page",
			wantCategory:   CategoryRepo,
		},
		{
			name:           "AddCollaborator",
			input:          "AddCollaborator failed",
			wantTranslated: "Failed to add collaborator",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "RemoveCollaborator",
			input:          "RemoveCollaborator failed",
			wantTranslated: "Failed to remove collaborator",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "CreateDeployKey",
			input:          "CreateDeployKey failed",
			wantTranslated: "Failed to create deploy key",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "DeleteDeployKey",
			input:          "DeleteDeployKey failed",
			wantTranslated: "Failed to delete deploy key",
			wantCategory:   CategoryAuth,
		},
		{
			name:           "Timeout",
			input:          "request timeout",
			wantTranslated: "Request timed out - the server took too long to respond",
			wantCategory:   CategoryNetwork,
		},
		{
			name:           "Deadline exceeded",
			input:          "context deadline exceeded",
			wantTranslated: "Request timed out - the server took too long to respond",
			wantCategory:   CategoryNetwork,
		},
		{
			name:           "Connection refused",
			input:          "connection refused",
			wantTranslated: "Network error - cannot connect to server",
			wantCategory:   CategoryNetwork,
		},
		{
			name:           "No such host",
			input:          "no such host example.com",
			wantTranslated: "Network error - cannot connect to server",
			wantCategory:   CategoryNetwork,
		},
		{
			name:           "Unknown error",
			input:          "some random error message",
			wantTranslated: "some random error message",
			wantCategory:   CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.input)
			result := TranslateError(err, nil)

			enhanced, ok := result.(*EnhancedError)
			if !ok {
				t.Fatal("Expected *EnhancedError")
			}

			if enhanced.Translated != tt.wantTranslated {
				t.Errorf("Translated = %q, want %q", enhanced.Translated, tt.wantTranslated)
			}
			if enhanced.Category != tt.wantCategory {
				t.Errorf("Category = %q, want %q", enhanced.Category, tt.wantCategory)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "EnhancedError with CategoryFile",
			err:      &EnhancedError{Category: CategoryFile, Translated: "File not found"},
			expected: true,
		},
		{
			name:     "EnhancedError with CategoryRepo",
			err:      &EnhancedError{Category: CategoryRepo, Translated: "Repo not found"},
			expected: true,
		},
		{
			name:     "EnhancedError with CategoryIssue",
			err:      &EnhancedError{Category: CategoryIssue, Translated: "Issue not found"},
			expected: true,
		},
		{
			name:     "EnhancedError with CategoryPull",
			err:      &EnhancedError{Category: CategoryPull, Translated: "PR not found"},
			expected: true,
		},
		{
			name:     "EnhancedError with CategoryBranch",
			err:      &EnhancedError{Category: CategoryBranch, Translated: "Branch not found"},
			expected: true,
		},
		{
			name:     "EnhancedError with CategoryAuth",
			err:      &EnhancedError{Category: CategoryAuth, Translated: "Auth failed"},
			expected: false,
		},
		{
			name:     "EnhancedError with 'not found' in message",
			err:      &EnhancedError{Category: CategoryUnknown, Translated: "Something not found"},
			expected: true,
		},
		{
			name:     "HTTP 404 error",
			err:      errors.New("request failed with status 404"),
			expected: true,
		},
		{
			name:     "Simple not found message",
			err:      errors.New("file not found"),
			expected: true,
		},
		{
			name:     "Other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "EnhancedError with CategoryAuth",
			err:      &EnhancedError{Category: CategoryAuth, Translated: "Auth failed"},
			expected: true,
		},
		{
			name:     "HTTP 401 error",
			err:      errors.New("request failed with status 401"),
			expected: true,
		},
		{
			name:     "HTTP 403 error",
			err:      errors.New("request failed with status 403"),
			expected: true,
		},
		{
			name:     "Authentication in message",
			err:      errors.New("authentication failed"),
			expected: true,
		},
		{
			name:     "Permission denied message",
			err:      errors.New("permission denied"),
			expected: true,
		},
		{
			name:     "Access token message",
			err:      errors.New("check your access token"),
			expected: true,
		},
		{
			name:     "File not found error",
			err:      errors.New("file not found"),
			expected: false,
		},
		{
			name:     "EnhancedError with other category",
			err:      &EnhancedError{Category: CategoryFile, Translated: "File not found"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAuthError(tt.err)
			if got != tt.expected {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsActionsAPIUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "EnhancedError with CategoryActions",
			err:      &EnhancedError{Category: CategoryActions, Translated: "Actions not available"},
			expected: true,
		},
		{
			name:     "EnhancedError with not supported message",
			err:      &EnhancedError{Translated: "not supported on this Gitea version"},
			expected: true,
		},
		{
			name:     "Actions with 404",
			err:      errors.New("actions endpoint returned 404"),
			expected: true,
		},
		{
			name:     "Actions with not found",
			err:      errors.New("actions workflow not found"),
			expected: true,
		},
		{
			name:     "Actions with method not allowed",
			err:      errors.New("actions method not allowed"),
			expected: true,
		},
		{
			name:     "Actions with 405",
			err:      errors.New("actions endpoint returned 405"),
			expected: true,
		},
		{
			name:     "Other actions error",
			err:      errors.New("actions completed successfully"),
			expected: false,
		},
		{
			name:     "Non-actions error",
			err:      errors.New("file not found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsActionsAPIUnavailable(tt.err)
			if got != tt.expected {
				t.Errorf("IsActionsAPIUnavailable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "EnhancedError with CategoryNetwork",
			err:      &EnhancedError{Category: CategoryNetwork, Translated: "Network error"},
			expected: true,
		},
		{
			name:     "EnhancedError with timed out message",
			err:      &EnhancedError{Translated: "Request timed out"},
			expected: true,
		},
		{
			name:     "Timeout in message",
			err:      errors.New("request timeout"),
			expected: true,
		},
		{
			name:     "Deadline exceeded",
			err:      errors.New("context deadline exceeded"),
			expected: true,
		},
		{
			name:     "Other error",
			err:      errors.New("file not found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTimeout(tt.err)
			if got != tt.expected {
				t.Errorf("IsTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "EnhancedError with CategoryNetwork",
			err:      &EnhancedError{Category: CategoryNetwork},
			expected: true,
		},
		{
			name:     "Connection error",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "Network error",
			err:      errors.New("network unreachable"),
			expected: true,
		},
		{
			name:     "No such host",
			err:      errors.New("no such host"),
			expected: true,
		},
		{
			name:     "Dial TCP",
			err:      errors.New("dial tcp: connection refused"),
			expected: true,
		},
		{
			name:     "Other error",
			err:      errors.New("file not found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNetworkError(tt.err)
			if got != tt.expected {
				t.Errorf("IsNetworkError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewEnhancedError(t *testing.T) {
	original := errors.New("original")
	err := NewEnhancedError(original, "translated", CategoryFile)

	if err.Original != original {
		t.Error("Original should be set correctly")
	}
	if err.Translated != "translated" {
		t.Errorf("Translated = %q, want %q", err.Translated, "translated")
	}
	if err.Category != CategoryFile {
		t.Errorf("Category = %q, want %q", err.Category, CategoryFile)
	}
	if err.Context == nil {
		t.Error("Context should be initialized")
	}
}

func TestWrap(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := Wrap(nil, "GetFile")
		if result != nil {
			t.Error("Wrap(nil) should return nil")
		}
	})

	t.Run("wraps error with operation", func(t *testing.T) {
		original := errors.New("original error")
		result := Wrap(original, "GetFile")

		enhanced, ok := result.(*EnhancedError)
		if !ok {
			t.Fatal("Expected *EnhancedError")
		}

		if enhanced.Operation != "GetFile" {
			t.Errorf("Operation = %q, want %q", enhanced.Operation, "GetFile")
		}
		if !errors.Is(enhanced, original) {
			t.Error("Original error should be preserved")
		}
	})
}

func TestIsHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		expected   bool
	}{
		{
			name:       "nil error",
			err:        nil,
			statusCode: 404,
			expected:   false,
		},
		{
			name:       "HTTPError with matching status",
			err:        &testHTTPError{status: 404, message: "not found"},
			statusCode: 404,
			expected:   true,
		},
		{
			name:       "HTTPError with non-matching status",
			err:        &testHTTPError{status: 500, message: "server error"},
			statusCode: 404,
			expected:   false,
		},
		{
			name:       "Error message with status code",
			err:        errors.New("request failed with status 404"),
			statusCode: 404,
			expected:   true,
		},
		{
			name:       "Error message without status code",
			err:        errors.New("something else"),
			statusCode: 404,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsHTTPError(tt.err, tt.statusCode)
			if got != tt.expected {
				t.Errorf("IsHTTPError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHTTPStatusHelpers(t *testing.T) {
	t.Run("IsUnauthorized", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"HTTP 401", &testHTTPError{status: 401}, true},
			{"HTTP 403", &testHTTPError{status: 403}, false},
			{"Error message with 401", errors.New("status 401"), true},
			{"Other error", errors.New("other"), false},
			{"nil", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := IsUnauthorized(tt.err)
				if got != tt.expected {
					t.Errorf("IsUnauthorized() = %v, want %v", got, tt.expected)
				}
			})
		}
	})

	t.Run("IsForbidden", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"HTTP 403", &testHTTPError{status: 403}, true},
			{"HTTP 401", &testHTTPError{status: 401}, false},
			{"Error message with 403", errors.New("status 403"), true},
			{"Other error", errors.New("other"), false},
			{"nil", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := IsForbidden(tt.err)
				if got != tt.expected {
					t.Errorf("IsForbidden() = %v, want %v", got, tt.expected)
				}
			})
		}
	})

	t.Run("IsNotFoundHTTP", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"HTTP 404", &testHTTPError{status: 404}, true},
			{"HTTP 403", &testHTTPError{status: 403}, false},
			{"Error message with 404", errors.New("status 404"), true},
			{"Other error", errors.New("other"), false},
			{"nil", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := IsNotFoundHTTP(tt.err)
				if got != tt.expected {
					t.Errorf("IsNotFoundHTTP() = %v, want %v", got, tt.expected)
				}
			})
		}
	})

	t.Run("IsServerError", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"HTTP 500", &testHTTPError{status: 500}, true},
			{"HTTP 502", &testHTTPError{status: 502}, true},
			{"HTTP 503", &testHTTPError{status: 503}, true},
			{"HTTP 404", &testHTTPError{status: 404}, false},
			{"Error message with 500", errors.New("status 500"), true},
			{"Error message with 502", errors.New("status 502"), true},
			{"Other error", errors.New("other"), false},
			{"nil", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := IsServerError(tt.err)
				if got != tt.expected {
					t.Errorf("IsServerError() = %v, want %v", got, tt.expected)
				}
			})
		}
	})
}

func TestHTTPErrorInterface(t *testing.T) {
	err := &testHTTPError{status: 404, message: "not found"}

	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		t.Error("testHTTPError should implement HTTPError")
	}

	if httpErr.Status() != 404 {
		t.Errorf("Status() = %d, want 404", httpErr.Status())
	}

	if httpErr.Error() != "not found" {
		t.Errorf("Error() = %q, want %q", httpErr.Error(), "not found")
	}
}

type testHTTPError struct {
	status  int
	message string
}

func (e *testHTTPError) Error() string { return e.message }
func (e *testHTTPError) Status() int    { return e.status }

func BenchmarkTranslateError(b *testing.B) {
	err := errors.New("GetContents failed")
	ctx := map[string]string{
		"operation": "GetFile",
		"owner":     "gitea",
		"repo":      "tea",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TranslateError(err, ctx)
	}
}

func BenchmarkIsNotFound(b *testing.B) {
	err := TranslateError(errors.New("GetContents failed"), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsNotFound(err)
	}
}

// ExampleTranslateError demonstrates how to use TranslateError.
func ExampleTranslateError() {
	sdkErr := errors.New("GetContents failed with status 404")

	err := TranslateError(sdkErr, map[string]string{
		"operation": "GetFile",
		"owner":     "gitea",
		"repo":      "tea",
		"path":      "README.md",
	})

	fmt.Println(err.Error())
}

// ExampleEnhancedError_Format demonstrates the Format method.
func ExampleEnhancedError_Format() {
	err := &EnhancedError{
		Original:   errors.New("GetContents failed"),
		Translated: "File or directory not found",
		Category:   CategoryFile,
		Operation:  "GetFile",
		Context: map[string]string{
			"owner": "gitea",
			"path":  "README.md",
		},
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	formatted := err.Format()
	_ = formatted // Use the formatted string
}

// TestWithOperation tests the fluent API WithOperation method.
func TestWithOperation(t *testing.T) {
	original := errors.New("original error")
	enhanced := TranslateError(original, nil).(*EnhancedError)

	result := enhanced.WithOperation("GetFile")

	// Should return the same error for chaining
	if result != enhanced {
		t.Error("WithOperation should return the same error for chaining")
	}

	if enhanced.Operation != "GetFile" {
		t.Errorf("Operation = %q, want %q", enhanced.Operation, "GetFile")
	}
}

// TestWithParam tests the fluent API WithParam method.
func TestWithParam(t *testing.T) {
	original := errors.New("original error")
	enhanced := TranslateError(original, nil).(*EnhancedError)

	result := enhanced.
		WithOperation("GetFile").
		WithParam("owner", "gitea").
		WithParam("repo", "tea")

	// Should return the same error for chaining
	if result != enhanced {
		t.Error("WithParam should return the same error for chaining")
	}

	if enhanced.Operation != "GetFile" {
		t.Errorf("Operation = %q, want %q", enhanced.Operation, "GetFile")
	}

	if enhanced.Context["owner"] != "gitea" {
		t.Errorf("Context['owner'] = %q, want %q", enhanced.Context["owner"], "gitea")
	}

	if enhanced.Context["repo"] != "tea" {
		t.Errorf("Context['repo'] = %q, want %q", enhanced.Context["repo"], "tea")
	}
}

// TestFormatDetailed tests the JSON-like structured error output.
func TestFormatDetailed(t *testing.T) {
	err := &EnhancedError{
		Original:   errors.New("GetContents failed with status 404"),
		Translated: "File or directory not found",
		Category:   CategoryFile,
		Operation:  "GetFile",
		Context: map[string]string{
			"owner": "gitea",
			"path":  "README.md",
		},
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	detailed := err.FormatDetailed()

	// Check that JSON output contains expected fields
	if !strings.Contains(detailed, `"error"`) {
		t.Error("FormatDetailed should include 'error' field")
	}
	if !strings.Contains(detailed, `"category"`) {
		t.Error("FormatDetailed should include 'category' field")
	}
	if !strings.Contains(detailed, `"operation"`) {
		t.Error("FormatDetailed should include 'operation' field")
	}
	if !strings.Contains(detailed, `"timestamp"`) {
		t.Error("FormatDetailed should include 'timestamp' field")
	}
	if !strings.Contains(detailed, `"context"`) {
		t.Error("FormatDetailed should include 'context' field")
	}
	if !strings.Contains(detailed, `"original"`) {
		t.Error("FormatDetailed should include 'original' field")
	}

	// Check that values are included
	if !strings.Contains(detailed, "File or directory not found") {
		t.Error("FormatDetailed should include the error message")
	}
	if !strings.Contains(detailed, "gitea") {
		t.Error("FormatDetailed should include context values")
	}
	if !strings.Contains(detailed, "2024-01-15T10:30:00Z") {
		t.Error("FormatDetailed should include formatted timestamp")
	}
}

// TestFormatDetailedWithoutOptionalFields tests JSON output with minimal fields.
func TestFormatDetailedWithoutOptionalFields(t *testing.T) {
	err := &EnhancedError{
		Original:   errors.New("some error"),
		Translated: "translated message",
		Category:   CategoryUnknown,
		Timestamp:  time.Now(),
	}

	detailed := err.FormatDetailed()

	// Should not include operation field when empty
	if strings.Contains(detailed, `"operation"`) {
		t.Error("FormatDetailed should not include empty operation field")
	}

	// Should not include context field when empty
	if strings.Contains(detailed, `"context"`) {
		t.Error("FormatDetailed should not include empty context field")
	}

	// Should not include original when same as translated
	if strings.Contains(detailed, `"original"`) {
		t.Error("FormatDetailed should not include original when same as error")
	}
}

// TestTimestampIsSet tests that timestamp is automatically set.
func TestTimestampIsSet(t *testing.T) {
	before := time.Now().UTC()
	err := TranslateError(errors.New("test error"), nil).(*EnhancedError)
	after := time.Now().UTC()

	if err.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if err.Timestamp.Before(before) || err.Timestamp.After(after) {
		t.Errorf("Timestamp %v should be between %v and %v", err.Timestamp, before, after)
	}
}

// TestNewEnhancedErrorSetsTimestamp tests that NewEnhancedError sets timestamp.
func TestNewEnhancedErrorSetsTimestamp(t *testing.T) {
	before := time.Now().UTC()
	err := NewEnhancedError(errors.New("test"), "translated", CategoryFile)
	after := time.Now().UTC()

	if err.Timestamp.IsZero() {
		t.Error("NewEnhancedError should set timestamp")
	}

	if err.Timestamp.Before(before) || err.Timestamp.After(after) {
		t.Errorf("Timestamp %v should be between %v and %v", err.Timestamp, before, after)
	}
}

// TestFluentAPIChaining tests complete fluent API usage.
func TestFluentAPIChaining(t *testing.T) {
	original := errors.New("GetContents failed with status 404")

	err := TranslateError(original, nil).
		(*EnhancedError).
		WithOperation("GetFile").
		WithParam("owner", "gitea").
		WithParam("repo", "tea").
		WithParam("path", "README.md")

	if err.Operation != "GetFile" {
		t.Errorf("Operation = %q, want %q", err.Operation, "GetFile")
	}

	if err.Context["owner"] != "gitea" {
		t.Errorf("Context['owner'] = %q, want %q", err.Context["owner"], "gitea")
	}

	if err.Context["repo"] != "tea" {
		t.Errorf("Context['repo'] = %q, want %q", err.Context["repo"], "tea")
	}

	if err.Context["path"] != "README.md" {
		t.Errorf("Context['path'] = %q, want %q", err.Context["path"], "README.md")
	}

	// Verify FormatDetailed works with fluent API built error
	detailed := err.FormatDetailed()
	if !strings.Contains(detailed, "GetFile") {
		t.Error("FormatDetailed should include operation from fluent API")
	}
}
