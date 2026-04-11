# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Development Commands

**Build**: `make build` - Build the gitea-mcp binary
**Install**: `make install` - Build and install to GOPATH/bin
**Clean**: `make clean` - Remove build artifacts
**Test**: `go test ./...` - Run all tests
**Hot reload**: `make dev` - Start development server with hot reload (requires air)
**Dependencies**: `make vendor` - Tidy and verify module dependencies

## Architecture Overview

This is a **Gitea MCP (Model Context Protocol) Server** written in Go that provides MCP tools for interacting with Gitea repositories, issues, pull requests, users, and more.

**Core Components**:

- `main.go` + `cmd/cmd.go`: CLI entry point and flag parsing
- `operation/operation.go`: Main server setup and tool registration
- `pkg/tool/tool.go`: Tool registry with read/write categorization
- `operation/*/`: Individual tool modules (user, repo, issue, pull, search, wiki, etc.)

**Transport Modes**:

- **stdio** (default): Standard input/output for MCP clients
- **HTTP**: HTTP server mode on configurable port (default 8080)

**Authentication**:

- Global token via `--token` flag or `GITEA_ACCESS_TOKEN` env var
- HTTP mode supports per-request Bearer token override in Authorization header
- Token precedence: HTTP Authorization header > CLI flag > environment variable

**Tool Organization**:

- Tools are categorized as read-only or write operations
- `--read-only` flag exposes only read tools
- Tool modules register via `Tool.RegisterRead()` and `Tool.RegisterWrite()`

**Key Configuration**:

- Default Gitea host: `https://gitea.com` (override with `--host` or `GITEA_HOST`)
- Environment variables can override CLI flags: `MCP_MODE`, `GITEA_READONLY`, `GITEA_DEBUG`, `GITEA_INSECURE`
- Logs are written to `~/.gitea-mcp/gitea-mcp.log` with rotation

## Available Tools

The server provides 40+ MCP tools covering:

- **User**: get_my_user_info, get_user_orgs, search_users
- **Repository**: create_repo, fork_repo, list_my_repos, search_repos
- **Branches/Tags**: create_branch, delete_branch, list_branches, create_tag, list_tags
- **Files**: get_file_content, create_file, update_file, delete_file, get_dir_content
- **Issues**: create_issue, list_repo_issues, create_issue_comment, edit_issue
- **Pull Requests**: create_pull_request, list_repo_pull_requests, get_pull_request_by_index
- **Releases**: create_release, list_releases, get_latest_release
- **Wiki**: create_wiki_page, update_wiki_page, list_wiki_pages
- **Search**: search_repos, search_users, search_org_teams
- **Version**: get_gitea_mcp_server_version

## Error Handling and Logging

The codebase provides comprehensive error handling and structured logging with context support.

### Enhanced Error Handling

The `pkg/errors` package provides enhanced error handling with context and fluent API:

```go
import "gitea.com/gitea/gitea-mcp/pkg/errors"

// Basic error translation
err := someGiteaOperation()
if err != nil {
    return errors.TranslateError(err, map[string]string{
        "operation": "GetFile",
        "owner":     owner,
        "repo":      repo,
        "path":      path,
    })
}

// Using fluent API for building error context
err := someGiteaOperation()
if err != nil {
    return errors.TranslateError(err, nil).
        WithOperation("GetFile").
        WithParam("owner", owner).
        WithParam("repo", repo).
        WithParam("path", path)
}

// Error includes automatic timestamp
enhanced := err.(*errors.EnhancedError)
fmt.Printf("Error occurred at: %v\n", enhanced.Timestamp)

// Format error for logging (human-readable)
fmt.Println(enhanced.Format())
// Output: Operation: GetFile | Error: File or directory not found | Category: file | Context: owner=gitea, path=README.md | Original: GetContents failed

// Format error for structured logging (JSON-like)
fmt.Println(enhanced.FormatDetailed())
// Output:
// {
//   "error": "File or directory not found",
//   "category": "file",
//   "operation": "GetFile",
//   "timestamp": "2024-01-15T10:30:00Z",
//   "context": {
//     "owner": "gitea",
//     "path": "README.md"
//   },
//   "original": "GetContents failed with status 404"
// }
```

### Error Category Checking

```go
// Check error categories
if errors.IsNotFound(err) {
    // Handle not found (404, "not found" messages)
}

if errors.IsAuthError(err) {
    // Handle auth errors (401, 403)
}

if errors.IsTimeout(err) {
    // Handle timeout errors
}

if errors.IsNetworkError(err) {
    // Handle network connectivity issues
}

// Check specific HTTP status codes
if errors.IsUnauthorized(err) {
    // Handle 401
}

if errors.IsForbidden(err) {
    // Handle 403
}

if errors.IsServerError(err) {
    // Handle 5xx errors
}
```

### Structured Logging with Context

The `pkg/log` package provides request-scoped structured logging with correlation IDs:

```go
import (
    "context"
    "gitea.com/gitea/gitea-mcp/pkg/log"
    "go.uber.org/zap"
)

// Create context with correlation ID for request tracing
ctx := log.WithCorrelationID(context.Background(), "req-12345")

// Add operation name to context
ctx = log.WithOperation(ctx, "GetFile")

// Create logger with context
logger := log.WithContext(ctx)

// Log messages - correlation_id and operation are automatically included
logger.Info("processing request")
logger.Error("operation failed", zap.Error(err))

// Log with additional fields
logger.Info("file retrieved",
    zap.String("owner", owner),
    zap.String("repo", repo),
    zap.String("path", path),
)

// Operation logging with timing
op := log.StartOperation(ctx, "CreatePullRequest")
op.Start("beginning pull request creation")
// ... do work ...
op.Success("pull request created successfully")
// Or on failure:
op.Failure("failed to create pull request", err)
```

### REST API Logging

The `pkg/gitea/rest.go` automatically logs all API requests with context:

```go
import (
    "context"
    "gitea.com/gitea/gitea-mcp/pkg/log"
    "gitea.com/gitea/gitea-mcp/pkg/gitea"
)

// Create context with operation name for tracing
ctx := log.WithOperation(context.Background(), "GetRepository")

// All API calls are automatically logged with:
// - operation name
// - HTTP method
// - request path (no sensitive data)
// - response status code
// - duration
// - correlation ID
status, err := gitea.DoJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s", owner, repo), nil, nil, &repo)

// Logs will include:
// - Debug: "sending API request" with operation, method, path, correlation_id
// - Debug: "API request completed" with status_code and duration on success
// - Error: "API request returned error status" with details on failure
```

### Common Development Patterns

**Testing**: Use `go test ./operation -run TestFunctionName` for specific tests

**Token Context**: HTTP requests use `pkg/context.TokenContextKey` for request-scoped token access

**Flag Access**: All packages access configuration via global variables in `pkg/flag/flag.go`

**Graceful Shutdown**: HTTP mode implements graceful shutdown with 10-second timeout on SIGTERM/SIGINT