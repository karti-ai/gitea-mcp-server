package actions

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"gitea.com/gitea/gitea-mcp/pkg/errors"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ListActionRunnersToolName = "list_action_runners"
)

// ActionRunner represents a self-hosted action runner
// This is a local type since Gitea SDK v0.23.2 doesn't include it
type ActionRunner struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	UUID       string   `json:"uuid"`
	Status     string   `json:"status"`
	Online     bool     `json:"online"`
	Busy       bool     `json:"busy"`
	Version    string   `json:"version,omitempty"`
	Labels     []string `json:"labels,omitempty"`
	LastOnline string   `json:"last_online,omitempty"`
}

// ActionRunnersResponse represents the API response for listing runners
type ActionRunnersResponse struct {
	TotalCount int             `json:"total_count"`
	Runners    []*ActionRunner `json:"runners"`
}

var (
	ListActionRunnersTool = mcp.NewTool(
		ListActionRunnersToolName,
		mcp.WithDescription("List self-hosted action runners for a repository. Shows runner status, labels, and availability. Filter by status (online/offline). Note: Requires Gitea 1.23+; Gitea 1.22.5 does not support Actions API."),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("status", mcp.Description("optional status filter (online, offline, busy, idle)"), mcp.Enum("online", "offline", "busy", "idle")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1), mcp.Min(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(30), mcp.Min(1)),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListActionRunnersTool,
		Handler: listActionRunnersFn,
	})
}

func listActionRunnersFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listActionRunnersFn")

	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "ListActionRunners",
			"param":     "owner",
		}))
	}

	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "ListActionRunners",
			"param":     "repo",
		}))
	}

	statusFilter, _ := req.GetArguments()["status"].(string)
	page, pageSize := params.GetPagination(req.GetArguments(), 30)

	// Use REST API to get runners
	apiPath := fmt.Sprintf("/repos/%s/%s/actions/runners", owner, repo)
	query := url.Values{
		"page":     []string{fmt.Sprintf("%d", page)},
		"per_page": []string{fmt.Sprintf("%d", pageSize)},
	}

	var runnersResp ActionRunnersResponse
	statusCode, err := gitea.DoJSON(ctx, http.MethodGet, apiPath, query, nil, &runnersResp)
	if err != nil {
		if statusCode == http.StatusNotFound {
			// Gitea 1.22.5 doesn't have Actions API - return empty list with message
			return to.TextResult(map[string]interface{}{
				"total_count": 0,
				"runners":     []interface{}{},
				"note":        "Actions API not available in Gitea 1.22.5. Requires Gitea 1.23+.",
			})
		}
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "ListActionRunners",
			"owner":     owner,
			"repo":      repo,
		}))
	}

	// Filter by status if requested
	filteredRunners := make([]*ActionRunner, 0, len(runnersResp.Runners))
	if statusFilter != "" {
		for _, runner := range runnersResp.Runners {
			switch statusFilter {
			case "online":
				if runner.Online {
					filteredRunners = append(filteredRunners, runner)
				}
			case "offline":
				if !runner.Online {
					filteredRunners = append(filteredRunners, runner)
				}
			case "busy":
				if runner.Busy {
					filteredRunners = append(filteredRunners, runner)
				}
			case "idle":
				if runner.Online && !runner.Busy {
					filteredRunners = append(filteredRunners, runner)
				}
			default:
				if runner.Status == statusFilter {
					filteredRunners = append(filteredRunners, runner)
				}
			}
		}
	} else {
		filteredRunners = runnersResp.Runners
	}

	result := slimActionRunners(filteredRunners)

	return to.TextResult(result)
}

func slimActionRunners(runners []*ActionRunner) map[string]interface{} {
	if len(runners) == 0 {
		return map[string]interface{}{
			"total_count": 0,
			"runners":     []interface{}{},
		}
	}

	slimmed := make([]map[string]interface{}, 0, len(runners))
	for _, runner := range runners {
		slimmed = append(slimmed, slimActionRunner(runner))
	}

	return map[string]interface{}{
		"total_count": len(runners),
		"runners":     slimmed,
	}
}

func slimActionRunner(runner *ActionRunner) map[string]interface{} {
	if runner == nil {
		return nil
	}

	result := map[string]interface{}{
		"id":     runner.ID,
		"name":   runner.Name,
		"uuid":   runner.UUID,
		"status": runner.Status,
		"online": runner.Online,
		"busy":   runner.Busy,
	}

	if runner.Version != "" {
		result["version"] = runner.Version
	}

	if len(runner.Labels) > 0 {
		result["labels"] = runner.Labels
	} else {
		result["labels"] = []string{}
	}

	if runner.LastOnline != "" {
		result["last_online"] = runner.LastOnline
	}

	return result
}
