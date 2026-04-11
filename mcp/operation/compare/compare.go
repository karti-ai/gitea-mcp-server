package compare

import (
	"context"
	"fmt"

	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"
	"gitea.com/gitea/gitea-mcp/pkg/tool"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	CompareCommitsToolName = "compare_commits"
)

var Tool = tool.New()

var (
	CompareCommitsTool = mcp.NewTool(
		CompareCommitsToolName,
		mcp.WithDescription("Compare commits between two branches/tags/commits"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("base", mcp.Required(), mcp.Description("Base branch/tag/commit")),
		mcp.WithString("head", mcp.Required(), mcp.Description("Head branch/tag/commit")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    CompareCommitsTool,
		Handler: compareCommitsFn,
	})
}

func compareCommitsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Compare] Called compareCommitsFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	base, err := params.GetString(args, "base")
	if err != nil {
		return to.ErrorResult(err)
	}
	head, err := params.GetString(args, "head")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	compare, _, err := client.CompareCommits(owner, repo, base, head)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("compare commits err: %v", err))
	}

	return to.TextResult(slimCompare(compare))
}

func slimCompare(c *gitea_sdk.Compare) map[string]interface{} {
	result := map[string]interface{}{
		"total_commits": c.TotalCommits,
	}

	if c.Commits != nil {
		commits := make([]map[string]interface{}, 0, len(c.Commits))
		for _, cmt := range c.Commits {
			commits = append(commits, map[string]interface{}{
				"sha": cmt.SHA,
				"url": cmt.URL,
			})
		}
		result["commits"] = commits
	}

	return result
}
