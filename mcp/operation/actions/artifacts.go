package actions

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"gitea.com/gitea/gitea-mcp/pkg/errors"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	DefaultMaxArtifactSize  = 100 * 1024 * 1024
	ActionsArtifactToolName = "list_action_artifacts"
)

var (
	ActionsArtifactTool = mcp.NewTool(
		ActionsArtifactToolName,
		mcp.WithDescription("List and download artifacts from workflow runs. Use method 'list' to list artifacts, 'get' to get a specific artifact, 'download' to download artifact content."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("list", "get", "download")),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("run_id", mcp.Description("run ID to filter artifacts (optional for list, required for get/download)")),
		mcp.WithString("artifact_name", mcp.Description("artifact name to filter (optional)")),
		mcp.WithNumber("artifact_id", mcp.Description("artifact ID (required for 'get' and 'download' methods)")),
		mcp.WithString("output_path", mcp.Description("output file path (for 'download' method). If not specified, saves to ~/.gitea-mcp/artifacts/")),
		mcp.WithNumber("max_size", mcp.Description("maximum artifact size in bytes to download (default 100MB)"), mcp.DefaultNumber(DefaultMaxArtifactSize), mcp.Min(1024)),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1), mcp.Min(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(30), mcp.Min(1)),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{Tool: ActionsArtifactTool, Handler: artifactHandler})
}

func artifactHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "ListActionArtifacts",
			"param":     "method",
		}))
	}

	switch method {
	case "list":
		return listActionArtifactsFn(ctx, req)
	case "get":
		return getActionArtifactFn(ctx, req)
	case "download":
		return downloadActionArtifactFn(ctx, req)
	default:
		return to.ErrorResult(errors.NewEnhancedError(
			fmt.Errorf("unknown method: %s", method),
			"Invalid method. Use 'list', 'get', or 'download'",
			errors.CategoryActions,
		).WithOperation("ListActionArtifacts"))
	}
}

func listActionArtifactsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listActionArtifactsFn")

	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil || owner == "" {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"owner is required",
			errors.CategoryActions,
		).WithOperation("ListActionArtifacts"))
	}

	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil || repo == "" {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"repo is required",
			errors.CategoryActions,
		).WithOperation("ListActionArtifacts"))
	}

	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	artifactName, _ := req.GetArguments()["artifact_name"].(string)

	var runID int64
	if runIDVal, exists := req.GetArguments()["run_id"]; exists {
		if runIDFloat, ok := runIDVal.(float64); ok {
			runID = int64(runIDFloat)
		}
	}

	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("limit", strconv.Itoa(pageSize))
	if artifactName != "" {
		query.Set("name", artifactName)
	}
	if runID > 0 {
		query.Set("run_id", strconv.FormatInt(runID, 10))
	}

	var result any
	err = doJSONWithFallback(ctx, "GET",
		[]string{
			fmt.Sprintf("repos/%s/%s/actions/artifacts", url.PathEscape(owner), url.PathEscape(repo)),
		},
		query, nil, &result,
	)
	if err != nil {
		if errors.IsActionsAPIUnavailable(err) {
			return to.TextResult(map[string]any{
				"artifacts":   []any{},
				"total_count": 0,
				"message":     "Actions API not available on this Gitea version",
			})
		}
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "ListActionArtifacts",
			"owner":     owner,
			"repo":      repo,
		}))
	}

	return to.TextResult(slimActionArtifacts(result))
}

func getActionArtifactFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called getActionArtifactFn")

	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil || owner == "" {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"owner is required",
			errors.CategoryActions,
		).WithOperation("GetActionArtifact"))
	}

	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil || repo == "" {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"repo is required",
			errors.CategoryActions,
		).WithOperation("GetActionArtifact"))
	}

	artifactID, err := params.GetIndex(req.GetArguments(), "artifact_id")
	if err != nil || artifactID <= 0 {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"artifact_id is required",
			errors.CategoryActions,
		).WithOperation("GetActionArtifact"))
	}

	var result any
	err = doJSONWithFallback(ctx, "GET",
		[]string{
			fmt.Sprintf("repos/%s/%s/actions/artifacts/%d", url.PathEscape(owner), url.PathEscape(repo), artifactID),
		},
		nil, nil, &result,
	)
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation":   "GetActionArtifact",
			"owner":       owner,
			"repo":        repo,
			"artifact_id": strconv.FormatInt(artifactID, 10),
		}))
	}

	return to.TextResult(slimActionArtifact(result))
}

func downloadActionArtifactFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called downloadActionArtifactFn")

	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil || owner == "" {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"owner is required",
			errors.CategoryActions,
		).WithOperation("DownloadActionArtifact"))
	}

	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil || repo == "" {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"repo is required",
			errors.CategoryActions,
		).WithOperation("DownloadActionArtifact"))
	}

	artifactID, err := params.GetIndex(req.GetArguments(), "artifact_id")
	if err != nil || artifactID <= 0 {
		return to.ErrorResult(errors.NewEnhancedError(
			err,
			"artifact_id is required",
			errors.CategoryActions,
		).WithOperation("DownloadActionArtifact"))
	}

	maxSize := int64(params.GetOptionalInt(req.GetArguments(), "max_size", DefaultMaxArtifactSize))
	outputPath, _ := req.GetArguments()["output_path"].(string)

	var artifactInfo any
	err = doJSONWithFallback(ctx, "GET",
		[]string{
			fmt.Sprintf("repos/%s/%s/actions/artifacts/%d", url.PathEscape(owner), url.PathEscape(repo), artifactID),
		},
		nil, nil, &artifactInfo,
	)
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation":   "DownloadActionArtifact",
			"owner":       owner,
			"repo":        repo,
			"artifact_id": strconv.FormatInt(artifactID, 10),
		}))
	}

	var artifactSize int64
	if info, ok := artifactInfo.(map[string]any); ok {
		if size, ok := info["size_in_bytes"].(float64); ok {
			artifactSize = int64(size)
		}
	}

	if artifactSize > maxSize {
		return to.ErrorResult(errors.NewEnhancedError(
			fmt.Errorf("artifact size %d exceeds maximum allowed size %d", artifactSize, maxSize),
			fmt.Sprintf("Artifact size (%s) exceeds maximum allowed size (%s). Use max_size parameter to increase limit.",
				formatBytes(artifactSize), formatBytes(maxSize)),
			errors.CategoryActions,
		).WithOperation("DownloadActionArtifact").
			WithParam("owner", owner).
			WithParam("repo", repo).
			WithParam("artifact_id", strconv.FormatInt(artifactID, 10)))
	}

	artifactBytes, _, err := gitea.DoBytes(ctx, "GET",
		fmt.Sprintf("repos/%s/%s/actions/artifacts/%d/download", url.PathEscape(owner), url.PathEscape(repo), artifactID),
		nil, nil, "application/zip",
	)
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation":   "DownloadActionArtifact",
			"owner":       owner,
			"repo":        repo,
			"artifact_id": strconv.FormatInt(artifactID, 10),
		}))
	}

	if outputPath == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			home = os.TempDir()
		}
		var artifactName string
		if info, ok := artifactInfo.(map[string]any); ok {
			if name, ok := info["name"].(string); ok {
				artifactName = name
			}
		}
		if artifactName == "" {
			artifactName = fmt.Sprintf("artifact-%d", artifactID)
		}
		outputPath = filepath.Join(home, ".gitea-mcp", "artifacts", owner, repo, fmt.Sprintf("%s.zip", artifactName))
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o700); err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "DownloadActionArtifact",
			"action":    "create_output_dir",
		}))
	}

	if err := os.WriteFile(outputPath, artifactBytes, 0o600); err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "DownloadActionArtifact",
			"action":    "write_file",
		}))
	}

	var artifactName string
	if info, ok := artifactInfo.(map[string]any); ok {
		if name, ok := info["name"].(string); ok {
			artifactName = name
		}
	}

	return to.TextResult(map[string]any{
		"artifact_id":   artifactID,
		"name":          artifactName,
		"path":          outputPath,
		"size_in_bytes": len(artifactBytes),
		"message":       "artifact downloaded successfully",
	})
}

func slimActionArtifact(raw any) any {
	if m, ok := raw.(map[string]any); ok {
		return pick(m, "id", "name", "size_in_bytes", "download_url", "run_id", "created_at", "expires_at")
	}
	return raw
}

func slimActionArtifacts(raw any) any {
	return slimPaginated(raw, func(m map[string]any) map[string]any {
		return pick(m, "id", "name", "size_in_bytes", "download_url", "run_id", "created_at", "expires_at")
	})
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
