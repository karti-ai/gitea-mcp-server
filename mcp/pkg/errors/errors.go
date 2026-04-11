// Package errors provides error translation and enhancement for Gitea SDK errors.
// It maps cryptic SDK error messages to human-readable descriptions and adds
// context about the operation being performed.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrorCategory represents the category of an error for easier handling.
type ErrorCategory string

const (
	// CategoryFile represents file/directory related errors.
	CategoryFile ErrorCategory = "file"
	// CategoryAuth represents authentication/authorization errors.
	CategoryAuth ErrorCategory = "auth"
	// CategoryRepo represents repository related errors.
	CategoryRepo ErrorCategory = "repo"
	// CategoryIssue represents issue related errors.
	CategoryIssue ErrorCategory = "issue"
	// CategoryPull represents pull request related errors.
	CategoryPull ErrorCategory = "pull"
	// CategoryBranch represents branch/tag related errors.
	CategoryBranch ErrorCategory = "branch"
	// CategoryActions represents Actions CI/CD related errors.
	CategoryActions ErrorCategory = "actions"
	// CategoryNetwork represents network/timeout related errors.
	CategoryNetwork ErrorCategory = "network"
	// CategoryUnknown represents unknown/uncategorized errors.
	CategoryUnknown ErrorCategory = "unknown"
)

// EnhancedError wraps an error with a human-readable translation and context.
type EnhancedError struct {
	// Original is the underlying error from the SDK or API.
	Original error
	// Translated is the human-readable error message.
	Translated string
	// Category helps identify the type of error for programmatic handling.
	Category ErrorCategory
	// Operation is the name of the operation that failed (e.g., "GetFile").
	Operation string
	// Context contains additional contextual information (e.g., parameters).
	Context map[string]string
	// Timestamp is when the error was created.
	Timestamp time.Time
}

// Error returns the human-readable translated error message.
func (e *EnhancedError) Error() string {
	if e.Translated != "" {
		return e.Translated
	}
	if e.Original != nil {
		return e.Original.Error()
	}
	return "unknown error"
}

// Unwrap returns the original error for error chain inspection.
func (e *EnhancedError) Unwrap() error {
	return e.Original
}

// WithContext adds context information to the error and returns a new EnhancedError.
func (e *EnhancedError) WithContext(key, value string) *EnhancedError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// WithOperation sets the operation name for the error and returns the error for chaining.
// This is a fluent API method for building error context.
//
// Example:
//
//	err := errors.TranslateError(sdkErr, nil).
//	    WithOperation("GetFile").
//	    WithParam("owner", "gitea").
//	    WithParam("repo", "tea")
func (e *EnhancedError) WithOperation(op string) *EnhancedError {
	e.Operation = op
	return e
}

// WithParam adds a single context parameter to the error and returns the error for chaining.
// This is a fluent API method for building error context one parameter at a time.
//
// Example:
//
//	err := errors.TranslateError(sdkErr, nil).
//	    WithOperation("GetFile").
//	    WithParam("owner", "gitea").
//	    WithParam("path", "README.md")
func (e *EnhancedError) WithParam(key, value string) *EnhancedError {
	return e.WithContext(key, value)
}

// FormatDetailed returns a JSON-like structured representation of the error.
// This format is suitable for logging and debugging, providing all error details
// in a machine-readable format.
//
// Example output:
//
//	{
//	  "error": "File or directory not found",
//	  "category": "file",
//	  "operation": "GetFile",
//	  "timestamp": "2024-01-15T10:30:00Z",
//	  "context": {
//	    "owner": "gitea",
//	    "repo": "tea",
//	    "path": "README.md"
//	  },
//	  "original": "GetContents failed with status 404"
//	}
func (e *EnhancedError) FormatDetailed() string {
	details := map[string]any{
		"error":     e.Error(),
		"category":  e.Category,
		"timestamp": e.Timestamp.Format(time.RFC3339),
	}

	if e.Operation != "" {
		details["operation"] = e.Operation
	}

	if len(e.Context) > 0 {
		details["context"] = e.Context
	}

	if e.Original != nil && e.Original.Error() != e.Error() {
		details["original"] = e.Original.Error()
	}

	jsonBytes, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		return e.Format()
	}

	return string(jsonBytes)
}

// Format returns a detailed error message including context.
func (e *EnhancedError) Format() string {
	var parts []string

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("Operation: %s", e.Operation))
	}

	parts = append(parts, fmt.Sprintf("Error: %s", e.Error()))

	if e.Category != "" && e.Category != CategoryUnknown {
		parts = append(parts, fmt.Sprintf("Category: %s", e.Category))
	}

	if len(e.Context) > 0 {
		var ctxParts []string
		for k, v := range e.Context {
			ctxParts = append(ctxParts, fmt.Sprintf("%s=%s", k, v))
		}
		parts = append(parts, fmt.Sprintf("Context: %s", strings.Join(ctxParts, ", ")))
	}

	if e.Original != nil && e.Original.Error() != e.Error() {
		parts = append(parts, fmt.Sprintf("Original: %s", e.Original.Error()))
	}

	return strings.Join(parts, " | ")
}

// TranslateError translates a Gitea SDK error to a human-readable error
// with context enhancement. The context map can contain operation name,
// parameters, or any other relevant information.
//
// Example:
//
//	err := someGiteaOperation()
//	if err != nil {
//	    return TranslateError(err, map[string]string{
//	        "operation": "GetFile",
//	        "owner": "gitea",
//	        "repo": "tea",
//	        "path": "README.md",
//	    })
//	}
func TranslateError(err error, context map[string]string) error {
	if err == nil {
		return nil
	}

	// If already an EnhancedError, just add context
	var existing *EnhancedError
	if errors.As(err, &existing) {
		if context != nil {
			for k, v := range context {
				existing.WithContext(k, v)
			}
		}
		return existing
	}

	// Determine translation based on error content
	translated, category := translateErrorMessage(err)

	operation := ""
	if context != nil {
		operation = context["operation"]
	}

	enhanced := &EnhancedError{
		Original:   err,
		Translated: translated,
		Category:   category,
		Operation:  operation,
		Context:    context,
		Timestamp:  time.Now().UTC(),
	}

	return enhanced
}

// translateErrorMessage maps SDK error strings to human-readable messages.
func translateErrorMessage(err error) (string, ErrorCategory) {
	if err == nil {
		return "", CategoryUnknown
	}

	msg := err.Error()
	lowerMsg := strings.ToLower(msg)

	// HTTP status code based translations (for HTTPError)
	if strings.Contains(msg, "status 404") || strings.Contains(msg, "404") {
		// Check for specific API operations in the error message
		if strings.Contains(lowerMsg, "getcontents") || strings.Contains(lowerMsg, "listcontents") {
			return "File or directory not found", CategoryFile
		}
		if strings.Contains(lowerMsg, "getuser") || strings.Contains(lowerMsg, "getuserbyname") {
			return "User or organization not found", CategoryAuth
		}
		if strings.Contains(lowerMsg, "getrepo") {
			return "Repository not found", CategoryRepo
		}
		if strings.Contains(lowerMsg, "getissue") {
			return "Issue not found", CategoryIssue
		}
		if strings.Contains(lowerMsg, "getpullrequest") || strings.Contains(lowerMsg, "getpull") {
			return "Pull request not found", CategoryPull
		}
		if strings.Contains(lowerMsg, "getbranch") || strings.Contains(lowerMsg, "gettag") {
			return "Branch or tag not found", CategoryBranch
		}
		return "Resource not found", CategoryUnknown
	}

	if strings.Contains(msg, "status 401") || strings.Contains(msg, "401") {
		return "Authentication failed - check your access token", CategoryAuth
	}

	if strings.Contains(msg, "status 403") || strings.Contains(msg, "403") {
		return "Permission denied - you don't have access to this resource", CategoryAuth
	}

	// SDK method name based translations
	translations := []struct {
		pattern  string
		message  string
		category ErrorCategory
	}{
		{"GetContents", "File or directory not found", CategoryFile},
		{"GetContentsOrList", "File or directory not found", CategoryFile},
		{"GetUserByName", "User or organization not found", CategoryAuth},
		{"GetUser", "User or organization not found", CategoryAuth},
		{"GetRepo", "Repository not found", CategoryRepo},
		{"GetIssue", "Issue not found", CategoryIssue},
		{"GetPullRequest", "Pull request not found", CategoryPull},
		{"GetBranch", "Branch not found", CategoryBranch},
		{"GetTag", "Tag not found", CategoryBranch},
		{"ListContents", "Directory not found or empty", CategoryFile},
		{"CreateFile", "Failed to create file - it may already exist", CategoryFile},
		{"UpdateFile", "Failed to update file - it may not exist or SHA mismatch", CategoryFile},
		{"DeleteFile", "Failed to delete file - it may not exist", CategoryFile},
		{"CreateBranch", "Failed to create branch", CategoryBranch},
		{"DeleteBranch", "Failed to delete branch - it may not exist or be protected", CategoryBranch},
		{"CreateIssue", "Failed to create issue", CategoryIssue},
		{"EditIssue", "Failed to update issue - it may not exist", CategoryIssue},
		{"CreatePullRequest", "Failed to create pull request", CategoryPull},
		{"EditPullRequest", "Failed to update pull request", CategoryPull},
		{"CreateRelease", "Failed to create release", CategoryRepo},
		{"EditRelease", "Failed to update release", CategoryRepo},
		{"CreateWikiPage", "Failed to create wiki page", CategoryRepo},
		{"EditWikiPage", "Failed to update wiki page", CategoryRepo},
		{"AddCollaborator", "Failed to add collaborator", CategoryAuth},
		{"RemoveCollaborator", "Failed to remove collaborator", CategoryAuth},
		{"CreateDeployKey", "Failed to create deploy key", CategoryAuth},
		{"DeleteDeployKey", "Failed to delete deploy key", CategoryAuth},
	}

	for _, t := range translations {
		if strings.Contains(msg, t.pattern) {
			return t.message, t.category
		}
	}

	// Timeout and network errors
	if strings.Contains(lowerMsg, "timeout") || strings.Contains(lowerMsg, "deadline exceeded") {
		return "Request timed out - the server took too long to respond", CategoryNetwork
	}

	if strings.Contains(lowerMsg, "connection refused") || strings.Contains(lowerMsg, "no such host") {
		return "Network error - cannot connect to server", CategoryNetwork
	}

	// Default: return original message with unknown category
	return msg, CategoryUnknown
}

// IsNotFound checks if an error is a "not found" type error.
// It works with EnhancedError and HTTPError types.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var enhanced *EnhancedError
	if errors.As(err, &enhanced) {
		switch enhanced.Category {
		case CategoryFile, CategoryRepo, CategoryIssue, CategoryPull, CategoryBranch:
			return true
		}
		return strings.Contains(enhanced.Translated, "not found")
	}

	// Check for HTTP 404
	var httpErr interface{ Error() string }
	if errors.As(err, &httpErr) {
		if strings.Contains(httpErr.Error(), "404") || strings.Contains(httpErr.Error(), "status 404") {
			return true
		}
	}

	// Check error message
	lowerMsg := strings.ToLower(err.Error())
	return strings.Contains(lowerMsg, "not found") ||
		strings.Contains(lowerMsg, "404")
}

// IsAuthError checks if an error is an authentication or authorization error.
// This includes 401 (unauthorized) and 403 (forbidden) HTTP errors.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	var enhanced *EnhancedError
	if errors.As(err, &enhanced) {
		return enhanced.Category == CategoryAuth
	}

	// Check for HTTP 401/403
	msg := err.Error()
	if strings.Contains(msg, "401") || strings.Contains(msg, "status 401") ||
		strings.Contains(msg, "403") || strings.Contains(msg, "status 403") {
		return true
	}

	// Check error message
	lowerMsg := strings.ToLower(msg)
	return strings.Contains(lowerMsg, "authentication") ||
		strings.Contains(lowerMsg, "unauthorized") ||
		strings.Contains(lowerMsg, "permission denied") ||
		strings.Contains(lowerMsg, "forbidden") ||
		strings.Contains(lowerMsg, "access token")
}

// IsActionsAPIUnavailable checks if an error indicates that the Actions API
// is not available on the current Gitea version.
func IsActionsAPIUnavailable(err error) bool {
	if err == nil {
		return false
	}

	var enhanced *EnhancedError
	if errors.As(err, &enhanced) {
		return enhanced.Category == CategoryActions ||
			strings.Contains(enhanced.Translated, "not supported on this Gitea version")
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "actions") &&
		(strings.Contains(msg, "not found") ||
			strings.Contains(msg, "method not allowed") ||
			strings.Contains(msg, "404") ||
			strings.Contains(msg, "405"))
}

// IsTimeout checks if an error is a timeout error.
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	var enhanced *EnhancedError
	if errors.As(err, &enhanced) {
		return enhanced.Category == CategoryNetwork ||
			strings.Contains(enhanced.Translated, "timed out")
	}

	lowerMsg := strings.ToLower(err.Error())
	return strings.Contains(lowerMsg, "timeout") ||
		strings.Contains(lowerMsg, "deadline exceeded") ||
		strings.Contains(lowerMsg, "context deadline")
}

// IsNetworkError checks if an error is a network connectivity error.
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var enhanced *EnhancedError
	if errors.As(err, &enhanced) {
		return enhanced.Category == CategoryNetwork
	}

	lowerMsg := strings.ToLower(err.Error())
	return strings.Contains(lowerMsg, "connection") ||
		strings.Contains(lowerMsg, "network") ||
		strings.Contains(lowerMsg, "no such host") ||
		strings.Contains(lowerMsg, "dial tcp")
}

// NewEnhancedError creates a new EnhancedError with the given parameters.
func NewEnhancedError(original error, translated string, category ErrorCategory) *EnhancedError {
	return &EnhancedError{
		Original:   original,
		Translated: translated,
		Category:   category,
		Context:    make(map[string]string),
		Timestamp:  time.Now().UTC(),
	}
}

// Wrap wraps an error with additional context information.
func Wrap(err error, operation string) error {
	if err == nil {
		return nil
	}
	return TranslateError(err, map[string]string{"operation": operation})
}

// HTTPError represents an HTTP error response.
// This interface is used to check for HTTP status codes.
type HTTPError interface {
	error
	Status() int
}

// statusError is a simple implementation of HTTPError for testing.
type statusError struct {
	status  int
	message string
}

func (e *statusError) Error() string { return e.message }
func (e *statusError) Status() int    { return e.status }

// IsHTTPError checks if an error is an HTTP error with the given status code.
func IsHTTPError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	// Check if it's our HTTPError type
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Status() == statusCode
	}

	// Check error message for status code
	msg := err.Error()
	return strings.Contains(msg, fmt.Sprintf("status %d", statusCode)) ||
		strings.Contains(msg, fmt.Sprintf("%d", statusCode))
}

// Common HTTP status check helpers

// IsUnauthorized checks if the error is an HTTP 401 Unauthorized.
func IsUnauthorized(err error) bool {
	return IsHTTPError(err, http.StatusUnauthorized)
}

// IsForbidden checks if the error is an HTTP 403 Forbidden.
func IsForbidden(err error) bool {
	return IsHTTPError(err, http.StatusForbidden)
}

// IsNotFoundHTTP checks if the error is an HTTP 404 Not Found.
func IsNotFoundHTTP(err error) bool {
	return IsHTTPError(err, http.StatusNotFound)
}

// IsServerError checks if the error is an HTTP 5xx server error.
func IsServerError(err error) bool {
	if err == nil {
		return false
	}

	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Status() >= 500 && httpErr.Status() < 600
	}

	msg := err.Error()
	for i := 500; i < 600; i++ {
		if strings.Contains(msg, fmt.Sprintf("status %d", i)) ||
			strings.Contains(msg, fmt.Sprintf("%d", i)) {
			return true
		}
	}
	return false
}
