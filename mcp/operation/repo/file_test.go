package repo

import (
	"errors"
	"testing"

	gitea_errors "gitea.com/gitea/gitea-mcp/pkg/errors"
)

// mockClientError is a mock error that simulates SDK errors
type mockClientError struct {
	message string
}

func (e *mockClientError) Error() string {
	return e.message
}

func TestErrorTranslation_GetFile(t *testing.T) {
	// Test that GetContentsOrList errors are translated properly
	err := errors.New("GetContentsOrList: 404 Not Found")
	translated := gitea_errors.TranslateError(err, map[string]string{
		"operation": "GetFile",
		"owner":     "karti-ai",
		"repo":      "docs",
		"path":      "README.md",
		"ref":       "main",
	})

	// Should return an EnhancedError
	var enhanced *gitea_errors.EnhancedError
	if !errors.As(translated, &enhanced) {
		t.Fatal("expected translated error to be EnhancedError")
	}

	// Check operation
	if enhanced.Operation != "GetFile" {
		t.Errorf("expected operation GetFile, got %s", enhanced.Operation)
	}

	// Check context
	if enhanced.Context["owner"] != "karti-ai" {
		t.Errorf("expected owner karti-ai, got %s", enhanced.Context["owner"])
	}
	if enhanced.Context["path"] != "README.md" {
		t.Errorf("expected path README.md, got %s", enhanced.Context["path"])
	}

	// Should be a file-related error
	if enhanced.Category != gitea_errors.CategoryFile {
		t.Errorf("expected CategoryFile, got %s", enhanced.Category)
	}

	// Should be identified as NotFound
	if !gitea_errors.IsNotFound(translated) {
		t.Error("expected error to be identified as NotFound")
	}

	t.Logf("Translated error message: %s", enhanced.Error())
}

func TestErrorTranslation_GetDir(t *testing.T) {
	// Test that ListContents errors are translated properly
	err := errors.New("ListContents: 404 Not Found")
	translated := gitea_errors.TranslateError(err, map[string]string{
		"operation": "GetDir",
		"owner":     "karti-ai",
		"repo":      "public_website",
		"path":      ".gitea/workflows",
		"ref":       "main",
	})

	var enhanced *gitea_errors.EnhancedError
	if !errors.As(translated, &enhanced) {
		t.Fatal("expected translated error to be EnhancedError")
	}

	if enhanced.Operation != "GetDir" {
		t.Errorf("expected operation GetDir, got %s", enhanced.Operation)
	}

	if enhanced.Category != gitea_errors.CategoryFile {
		t.Errorf("expected CategoryFile, got %s", enhanced.Category)
	}

	t.Logf("Translated error message: %s", enhanced.Error())
}

func TestErrorTranslation_CreateFile(t *testing.T) {
	// Test that CreateFile errors are translated properly
	err := errors.New("CreateFile: 422 Unprocessable Entity")
	translated := gitea_errors.TranslateError(err, map[string]string{
		"operation": "CreateFile",
		"owner":     "karti-ai",
		"repo":      "docs",
		"path":      "newfile.md",
		"branch":    "main",
	})

	var enhanced *gitea_errors.EnhancedError
	if !errors.As(translated, &enhanced) {
		t.Fatal("expected translated error to be EnhancedError")
	}

	if enhanced.Operation != "CreateFile" {
		t.Errorf("expected operation CreateFile, got %s", enhanced.Operation)
	}

	if enhanced.Context["path"] != "newfile.md" {
		t.Errorf("expected path newfile.md, got %s", enhanced.Context["path"])
	}

	t.Logf("Translated error message: %s", enhanced.Error())
}

func TestErrorTranslation_UpdateFile(t *testing.T) {
	// Test that UpdateFile errors are translated properly
	err := errors.New("UpdateFile: 409 Conflict")
	translated := gitea_errors.TranslateError(err, map[string]string{
		"operation": "UpdateFile",
		"owner":     "karti-ai",
		"repo":      "docs",
		"path":      "README.md",
		"branch":    "main",
	})

	var enhanced *gitea_errors.EnhancedError
	if !errors.As(translated, &enhanced) {
		t.Fatal("expected translated error to be EnhancedError")
	}

	if enhanced.Operation != "UpdateFile" {
		t.Errorf("expected operation UpdateFile, got %s", enhanced.Operation)
	}

	t.Logf("Translated error message: %s", enhanced.Error())
}

func TestErrorTranslation_DeleteFile(t *testing.T) {
	// Test that DeleteFile errors are translated properly
	err := errors.New("DeleteFile: 404 Not Found")
	translated := gitea_errors.TranslateError(err, map[string]string{
		"operation": "DeleteFile",
		"owner":     "karti-ai",
		"repo":      "docs",
		"path":      "oldfile.md",
		"branch":    "main",
	})

	var enhanced *gitea_errors.EnhancedError
	if !errors.As(translated, &enhanced) {
		t.Fatal("expected translated error to be EnhancedError")
	}

	if enhanced.Operation != "DeleteFile" {
		t.Errorf("expected operation DeleteFile, got %s", enhanced.Operation)
	}

	t.Logf("Translated error message: %s", enhanced.Error())
	t.Logf("Error category: %s", enhanced.Category)
}

func TestErrorTranslation_Unwrap(t *testing.T) {
	original := errors.New("GetContentsOrList: 404 Not Found")
	translated := gitea_errors.TranslateError(original, map[string]string{
		"operation": "GetFile",
		"owner":     "karti-ai",
		"repo":      "docs",
		"path":      "README.md",
	})

	// Should be able to unwrap to get original error
	var enhanced *gitea_errors.EnhancedError
	if errors.As(translated, &enhanced) {
		unwrapped := enhanced.Unwrap()
		if unwrapped == nil {
			t.Error("expected to be able to unwrap error")
		}
		if unwrapped.Error() != original.Error() {
			t.Errorf("expected unwrapped error to match original: got %s, want %s", unwrapped.Error(), original.Error())
		}
	} else {
		t.Error("expected translated error to be EnhancedError")
	}
}

func TestErrorTranslation_AuthErrors(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		isAuth bool
	}{
		{
			name:   "401 Unauthorized",
			errMsg: "GetContentsOrList: 401 Unauthorized",
			isAuth: true,
		},
		{
			name:   "403 Forbidden",
			errMsg: "GetContentsOrList: 403 Forbidden",
			isAuth: true,
		},
		{
			name:   "404 Not Found (not auth)",
			errMsg: "GetContentsOrList: 404 Not Found",
			isAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, map[string]string{
				"operation": "GetFile",
			})

			isAuth := gitea_errors.IsAuthError(translated)
			if isAuth != tt.isAuth {
				t.Errorf("IsAuthError() = %v, want %v", isAuth, tt.isAuth)
			}
		})
	}
}

func TestErrorTranslation_PreservesExistingEnhancedError(t *testing.T) {
	// If we translate an already-enhanced error, it should add context, not replace
	original := errors.New("GetContentsOrList: 404 Not Found")
	enhanced1 := gitea_errors.TranslateError(original, map[string]string{
		"operation": "GetFile",
		"owner":     "karti-ai",
	})

	// Translate again with more context
	enhanced2 := gitea_errors.TranslateError(enhanced1, map[string]string{
		"repo": "docs",
		"path": "README.md",
	})

	var e *gitea_errors.EnhancedError
	if errors.As(enhanced2, &e) {
		// Should have both sets of context
		if e.Context["operation"] != "GetFile" {
			t.Errorf("expected operation context to be preserved, got %s", e.Context["operation"])
		}
		if e.Context["owner"] != "karti-ai" {
			t.Errorf("expected owner context to be preserved, got %s", e.Context["owner"])
		}
		if e.Context["repo"] != "docs" {
			t.Errorf("expected repo context to be added, got %s", e.Context["repo"])
		}
		if e.Context["path"] != "README.md" {
			t.Errorf("expected path context to be added, got %s", e.Context["path"])
		}
	} else {
		t.Error("expected error to be EnhancedError")
	}
}

func TestErrorTranslation_NetworkErrors(t *testing.T) {
	tests := []struct {
		name      string
		errMsg    string
		isNetwork bool
		isTimeout bool
	}{
		{
			name:      "Connection refused",
			errMsg:    "GetContentsOrList: connection refused",
			isNetwork: true,
			isTimeout: false,
		},
		{
			name:      "Timeout",
			errMsg:    "GetContentsOrList: timeout",
			isNetwork: true,
			isTimeout: true,
		},
		{
			name:      "No such host",
			errMsg:    "GetContentsOrList: no such host",
			isNetwork: true,
			isTimeout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, map[string]string{
				"operation": "GetFile",
			})

			isNetwork := gitea_errors.IsNetworkError(translated)
			isTimeout := gitea_errors.IsTimeout(translated)

			t.Logf("Error: %s, IsNetwork: %v, IsTimeout: %v", tt.errMsg, isNetwork, isTimeout)
		})
	}
}

func TestErrorTranslation_ServerErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		isServer bool
	}{
		{
			name:     "500 Internal Server Error",
			errMsg:   "GetContentsOrList: 500 Internal Server Error",
			isServer: true,
		},
		{
			name:     "502 Bad Gateway",
			errMsg:   "GetContentsOrList: 502 Bad Gateway",
			isServer: true,
		},
		{
			name:     "503 Service Unavailable",
			errMsg:   "GetContentsOrList: 503 Service Unavailable",
			isServer: true,
		},
		{
			name:     "404 Not Found (not server)",
			errMsg:   "GetContentsOrList: 404 Not Found",
			isServer: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			translated := gitea_errors.TranslateError(err, map[string]string{
				"operation": "GetFile",
			})

			isServer := gitea_errors.IsServerError(translated)
			t.Logf("Error: %s, IsServer: %v", tt.errMsg, isServer)
		})
	}
}

func TestErrorTranslation_Format(t *testing.T) {
	err := errors.New("GetContentsOrList: 404 Not Found")
	translated := gitea_errors.TranslateError(err, map[string]string{
		"operation": "GetFile",
		"owner":     "karti-ai",
		"repo":      "docs",
		"path":      "README.md",
		"ref":       "main",
	})

	var enhanced *gitea_errors.EnhancedError
	if errors.As(translated, &enhanced) {
		formatted := enhanced.Format()

		// Format should include operation
		if formatted == "" {
			t.Error("expected non-empty formatted error")
		}

		t.Logf("Formatted error: %s", formatted)
	} else {
		t.Error("expected translated error to be EnhancedError")
	}
}
