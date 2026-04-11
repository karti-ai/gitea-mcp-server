package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"gitea.com/gitea/gitea-mcp/pkg/errors"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ListRepoStructureToolName = "list_repo_structure"
)

type TreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size int64  `json:"size,omitempty"`
}

type TreeResponse struct {
	SHA       string      `json:"sha"`
	URL       string      `json:"url,omitempty"`
	Tree      []TreeEntry `json:"tree"`
	Truncated bool        `json:"truncated,omitempty"`
}

var (
	ListRepoStructureTool = mcp.NewTool(
		ListRepoStructureToolName,
		mcp.WithDescription("List the complete directory and file structure of a repository using Git tree API. Supports recursive listing and pattern filtering."),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("ref", mcp.Description("Git reference (branch, tag, or commit SHA). Defaults to default branch.")),
		mcp.WithString("pattern", mcp.Description("Glob pattern to filter files (e.g., '*.yml', '.gitea/*', 'src/**/*.go')")),
		mcp.WithBoolean("recursive", mcp.Description("List contents recursively (default: true)")),
		mcp.WithNumber("page", mcp.Description("Page number for pagination (1-based, default: 1)")),
		mcp.WithNumber("per_page", mcp.Description("Number of items per page (default: 100, max: 1000)")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListRepoStructureTool,
		Handler: ListRepoStructureFn,
	})
}

func ListRepoStructureFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called ListRepoStructureFn")

	args := req.GetArguments()

	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}

	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}

	ref, _ := args["ref"].(string)
	if ref == "" {
		ref = "HEAD"
	}

	pattern, _ := args["pattern"].(string)

	recursive := true
	if recursiveVal, ok := args["recursive"].(bool); ok {
		recursive = recursiveVal
	}

	page := 1
	if pageVal, ok := args["page"].(float64); ok && pageVal > 0 {
		page = int(pageVal)
	}

	perPage := 100
	if perPageVal, ok := args["per_page"].(float64); ok && perPageVal > 0 {
		perPage = int(perPageVal)
		if perPage > 1000 {
			perPage = 1000
		}
	}

	query := url.Values{}
	if recursive {
		query.Set("recursive", "1")
	}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("per_page", fmt.Sprintf("%d", perPage))

	path := fmt.Sprintf("repos/%s/%s/git/trees/%s", owner, repo, ref)

	var treeResp TreeResponse
	statusCode, err := gitea.DoJSON(ctx, "GET", path, query, nil, &treeResp)
	if err != nil {
		translatedErr := errors.TranslateError(err, map[string]string{
			"operation": "ListRepoStructure",
			"owner":     owner,
			"repo":      repo,
			"ref":       ref,
			"status":    fmt.Sprintf("%d", statusCode),
		})
		return to.ErrorResult(translatedErr)
	}

	filteredEntries := filterEntries(treeResp.Tree, pattern)

	result := map[string]any{
		"owner":       owner,
		"repo":        repo,
		"ref":         ref,
		"sha":         treeResp.SHA,
		"truncated":   treeResp.Truncated,
		"total_count": len(filteredEntries),
		"page":        page,
		"per_page":    perPage,
		"tree":        slimTreeEntries(filteredEntries),
	}

	if treeResp.Truncated {
		result["warning"] = "Tree listing was truncated due to size. Consider using pattern filtering or pagination."
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return to.ErrorResult(fmt.Errorf("marshal result: %w", err))
	}

	return to.TextResult(string(resultJSON))
}

func filterEntries(entries []TreeEntry, pattern string) []TreeEntry {
	if pattern == "" {
		return entries
	}

	filtered := make([]TreeEntry, 0, len(entries))
	for _, entry := range entries {
		if matchPattern(entry.Path, pattern) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func matchPattern(path, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true
	}

	negate := false
	if strings.HasPrefix(pattern, "!") {
		negate = true
		pattern = strings.TrimPrefix(pattern, "!")
	}

	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return !negate
	}

	filename := filepath.Base(path)
	matched, err = filepath.Match(pattern, filename)
	if err == nil && matched {
		return !negate
	}

	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		if strings.HasSuffix(path, suffix) {
			return !negate
		}
		parts := strings.Split(path, "/")
		for i := range parts {
			subPath := strings.Join(parts[i:], "/")
			if matched, _ := filepath.Match(suffix, subPath); matched {
				return !negate
			}
		}
	}

	if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
		dirPrefix := strings.TrimSuffix(pattern, "/*")
		dirPrefix = strings.TrimSuffix(dirPrefix, "/**")
		if strings.HasPrefix(path, dirPrefix+"/") {
			return !negate
		}
	}

	if strings.HasPrefix(path, pattern+"/") || path == pattern {
		return !negate
	}

	return negate
}

func slimTreeEntries(entries []TreeEntry) []map[string]any {
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		m := map[string]any{
			"path": e.Path,
			"type": e.Type,
			"sha":  e.SHA,
		}
		if e.Type == "blob" && e.Size > 0 {
			m["size"] = e.Size
		}
		out = append(out, m)
	}
	return out
}
