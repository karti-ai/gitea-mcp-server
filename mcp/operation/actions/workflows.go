package actions

import (
	"context"
	"encoding/base64"
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
	"gopkg.in/yaml.v3"
)

const (
	GetWorkflowFileContentToolName = "get_workflow_file_content"
)

var (
	GetWorkflowFileContentTool = mcp.NewTool(
		GetWorkflowFileContentToolName,
		mcp.WithDescription("Get workflow file content from .gitea/workflows/ or .github/workflows/ directories. Auto-discovers workflow files and returns parsed YAML as JSON."),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("ref", mcp.Description("git ref (branch/tag/commit). Defaults to default branch if not specified")),
		mcp.WithString("pattern", mcp.Description("file pattern to match (e.g., '*.yml', 'build-*.yml'). Defaults to all workflow files")),
		mcp.WithString("filename", mcp.Description("specific workflow filename to retrieve. If provided, pattern is ignored")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetWorkflowFileContentTool,
		Handler: getWorkflowFileContentFn,
	})
}

type WorkflowFile struct {
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	SHA        string      `json:"sha"`
	Size       int64       `json:"size"`
	Content    interface{} `json:"content"`
	RawContent string      `json:"raw_content,omitempty"`
	Encoding   string      `json:"encoding,omitempty"`
}

type WorkflowFilesResult struct {
	Directory  string         `json:"directory"`
	Files      []WorkflowFile `json:"files"`
	TotalCount int            `json:"total_count"`
}

func getWorkflowFileContentFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called getWorkflowFileContentFn")

	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "GetWorkflowFileContent",
			"param":     "owner",
		}))
	}

	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "GetWorkflowFileContent",
			"param":     "repo",
		}))
	}

	ref, _ := req.GetArguments()["ref"].(string)
	pattern, _ := req.GetArguments()["pattern"].(string)
	filename, _ := req.GetArguments()["filename"].(string)

	directories := []string{".gitea/workflows", ".github/workflows"}

	var result WorkflowFilesResult
	var lastErr error

	for _, dir := range directories {
		if filename != "" {
			file, err := getWorkflowFile(ctx, owner, repo, ref, filepath.Join(dir, filename))
			if err == nil {
				result.Directory = dir
				result.Files = []WorkflowFile{*file}
				result.TotalCount = 1
				return to.TextResult(result)
			}
			if lastErr == nil {
				lastErr = err
			}
			continue
		}

		files, err := discoverWorkflowFiles(ctx, owner, repo, ref, dir, pattern)
		if err == nil && len(files) > 0 {
			result.Directory = dir
			result.Files = files
			result.TotalCount = len(files)
			return to.TextResult(result)
		}
		if err != nil && lastErr == nil {
			lastErr = err
		}
	}

	return to.TextResult(WorkflowFilesResult{
		Directory:  "",
		Files:      []WorkflowFile{},
		TotalCount: 0,
	})
}

func discoverWorkflowFiles(ctx context.Context, owner, repo, ref, dir, pattern string) ([]WorkflowFile, error) {
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "DiscoverWorkflowFiles",
			"owner":     owner,
			"repo":      repo,
		})
	}

	contents, _, err := client.ListContents(owner, repo, ref, dir)
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "ListWorkflowDirectory",
			"owner":     owner,
			"repo":      repo,
			"path":      dir,
		})
	}

	var files []WorkflowFile

	for _, content := range contents {
		if content.Type != "file" {
			continue
		}

		ext := strings.ToLower(filepath.Ext(content.Name))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		if pattern != "" && !matchPattern(content.Name, pattern) {
			continue
		}

		filePath := filepath.Join(dir, content.Name)
		file, err := getWorkflowFile(ctx, owner, repo, ref, filePath)
		if err != nil {
			log.Debugf("Failed to get workflow file %s: %v", filePath, err)
			continue
		}

		files = append(files, *file)
	}

	return files, nil
}

func getWorkflowFile(ctx context.Context, owner, repo, ref, path string) (*WorkflowFile, error) {
	escapedOwner := url.PathEscape(owner)
	escapedRepo := url.PathEscape(repo)
	escapedPath := url.PathEscape(path)

	var result any
	_, err := gitea.DoJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/contents/%s", escapedOwner, escapedRepo, escapedPath), nil, nil, &result)
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "GetWorkflowFile",
			"owner":     owner,
			"repo":      repo,
			"path":      path,
		})
	}

	contentMap, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for workflow file")
	}

	workflowFile := &WorkflowFile{
		Path: path,
	}

	if name, ok := contentMap["name"].(string); ok {
		workflowFile.Name = name
	}
	if sha, ok := contentMap["sha"].(string); ok {
		workflowFile.SHA = sha
	}
	if size, ok := contentMap["size"].(float64); ok {
		workflowFile.Size = int64(size)
	}
	if encoding, ok := contentMap["encoding"].(string); ok {
		workflowFile.Encoding = encoding
	}

	if contentStr, ok := contentMap["content"].(string); ok && contentStr != "" {
		var rawContent []byte

		if workflowFile.Encoding == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(contentStr)
			if err != nil {
				return nil, errors.TranslateError(
					errors.NewEnhancedError(err, "Failed to decode workflow file content", errors.CategoryActions),
					map[string]string{
						"operation": "DecodeWorkflowFile",
						"path":      path,
						"encoding":  workflowFile.Encoding,
					},
				)
			}
			rawContent = decoded
		} else {
			rawContent = []byte(contentStr)
		}

		workflowFile.RawContent = string(rawContent)

		var yamlContent interface{}
		if err := yaml.Unmarshal(rawContent, &yamlContent); err != nil {
			log.Debugf("Failed to parse YAML for %s: %v", path, err)
			yamlContent = nil
		} else {
			workflowFile.Content = convertYamlToInterface(yamlContent)
		}
	}

	return workflowFile, nil
}

func matchPattern(filename, pattern string) bool {
	pattern = strings.ToLower(pattern)
	filename = strings.ToLower(filename)

	if pattern == "*" || pattern == "*.*" {
		return true
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(filename, suffix)
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(filename, prefix)
	}

	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		mid := pattern[1 : len(pattern)-1]
		return strings.Contains(filename, mid)
	}

	return filename == pattern || strings.HasPrefix(filename, pattern)
}

func convertYamlToInterface(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = convertYamlToInterface(v)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			key := fmt.Sprintf("%v", k)
			result[key] = convertYamlToInterface(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = convertYamlToInterface(v)
		}
		return result
	case []yaml.Node:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = convertYamlToInterface(&v)
		}
		return result
	case *yaml.Node:
		switch val.Kind {
		case yaml.ScalarNode:
			s := val.Value
			var parsed interface{}
			if err := json.Unmarshal([]byte(s), &parsed); err == nil {
				return parsed
			}
			return s
		case yaml.SequenceNode:
			result := make([]interface{}, len(val.Content))
			for i, n := range val.Content {
				result[i] = convertYamlToInterface(&n)
			}
			return result
		case yaml.MappingNode:
			result := make(map[string]interface{})
			for i := 0; i < len(val.Content); i += 2 {
				key := val.Content[i].Value
				result[key] = convertYamlToInterface(&val.Content[i+1])
			}
			return result
		default:
			return val.Value
		}
	default:
		return val
	}
}
